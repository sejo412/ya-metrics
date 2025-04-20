package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
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
	/*
		pflag.StringVarP(&cfg.Address, "address", "a", config.DefaultAddress, "Listen address")
		pflag.IntVarP(&cfg.StoreInterval, "storeInterval", "i", config.DefaultStoreInterval, "Store interval")
		pflag.StringVarP(&cfg.StoreFile, "fileStoragePath", "f", config.DefaultStoreFile, "File storage path")
		pflag.BoolVarP(&cfg.Restore, "restore", "r", config.DefaultRestore, "Restore metrics")
		pflag.StringVarP(&cfg.DatabaseDSN, "database-dsn", "d", config.DefaultDatabaseDSN, "Database DSN")
		pflag.StringVarP(&cfg.Key, "key", "k", config.DefaultSecretKey, "secret key")
		pflag.StringVar(&cfg.CryptoKey, "crypto-key", config.DefaultCryptoKey, "path to public key")
		pflag.Parse()
		err := env.Parse(&cfg)
		if err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	*/

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

	// start flushing metrics on timer
	if cfg.StoreInterval > 0 && dsn.Scheme == "memory" {
		go server.FlushingMetrics(store, cfg.StoreFile, cfg.StoreInterval)
	}
	return server.StartServer(&config.Options{
		Config:  *cfg,
		Storage: store,
	},
		lm)
}
