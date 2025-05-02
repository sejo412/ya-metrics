package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sejo412/ya-metrics/internal/app/server"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/storage"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var err error
	cfg := config.NewServerConfig()
	if err = cfg.Load(); err != nil {
		return fmt.Errorf("error load config: %w", err)
	}
	// logger init
	logs, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logs.Sync()
	}()
	sugar := logs.Sugar()
	lm := logger.NewMiddleware(sugar)
	log := lm.Logger

	var store config.Storage
	dsn, err := storage.ParseDSN(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("parse database DSN: %w", err)
	}

	switch dsn.Scheme {
	case "memory":
		store = storage.NewMemoryStorage()
	case "postgres", "postgresql":
		store = storage.NewPostgresStorage()
	default:
		return fmt.Errorf("database \"%s\" not supported", cfg.DatabaseDSN)
	}

	ctxStore := context.Background()
	if err = store.Open(ctxStore, dsn); err != nil {
		return fmt.Errorf("error open database: %w", err)
	}
	defer store.Close()
	if err = store.Init(ctxStore); err != nil {
		return fmt.Errorf("error init database: %w", err)
	}

	// try restore metrics
	skipRestore := false
	if *cfg.Restore {
		f, err := os.Open(cfg.StoreFile)
		if err != nil {
			log.Errorw("error open file",
				"file", cfg.StoreFile)
			skipRestore = true
		}
		if !skipRestore {
			if err = store.Load(context.TODO(), f); err != nil {
				log.Errorw("error load file",
					"file", cfg.StoreFile)
			}
			if err := f.Close(); err != nil {
				log.Errorw("error close file",
					"file", cfg.StoreFile)
			}
		}
	}

	ctx := context.Background()
	return server.StartServer(ctx, &config.Options{
		Config:  *cfg,
		Storage: store,
	},
		lm)
}
