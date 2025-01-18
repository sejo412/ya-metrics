package main

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/internal/app/server"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// startup config init
	var cfg config.ServerConfig
	pflag.StringVarP(&cfg.Address, "address", "a", config.DefaultAddress, "Listen address")
	pflag.IntVarP(&cfg.StoreInterval, "storeInterval", "i", config.DefaultStoreInterval, "Store interval")
	pflag.StringVarP(&cfg.FileStoragePath, "fileStoragePath", "f", config.DefaultFileStoragePath, "File storage path")
	pflag.BoolVarP(&cfg.Restore, "restore", "r", config.DefaultRestore, "Restore metrics")
	pflag.StringVarP(&cfg.DatabaseDSN, "database-dsn", "d", config.DefaultDatabaseDSN, "Database DSN")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
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

	ctx := context.Background()
	if err = store.Open(ctx, dsn); err != nil {
		return fmt.Errorf("error open database: %w", err)
	}
	defer store.Close()
	if err = store.Init(ctx); err != nil {
		return fmt.Errorf("error init database: %w", err)
	}

	// try restore metrics
	skipRestore := false
	if cfg.Restore {
		f, err := os.Open(cfg.FileStoragePath)
		if err != nil {
			log.Errorw("error open file",
				"file", cfg.FileStoragePath)
			skipRestore = true
		}
		if !skipRestore {
			if err = store.Load(f); err != nil {
				log.Errorw("error load file",
					"file", cfg.FileStoragePath)
			}
			if err := f.Close(); err != nil {
				log.Errorw("error close file",
					"file", cfg.FileStoragePath)
			}
		}
	}

	// start flushing metrics on timer
	if cfg.StoreInterval > 0 && dsn.Scheme == "memory" {
		go server.FlushingMetrics(store, cfg.FileStoragePath, cfg.StoreInterval)
	}

	return server.StartServer(&config.Options{
		Config:  cfg,
		Storage: store,
	},
		lm)
}
