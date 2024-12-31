package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/cmd/server/app"
	"github.com/sejo412/ya-metrics/cmd/server/config"
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
	var cfg config.Config
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
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()
	lm := app.NewLoggerMiddleware(sugar)
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

	if err = store.Open(dsn); err != nil {
		return fmt.Errorf("error open database: %w", err)
	}
	defer store.Close()

	// restore metrics
	if cfg.Restore {
		f, err := os.Open(cfg.FileStoragePath)
		if err != nil {
			log.Errorw("error open file",
				"file", cfg.FileStoragePath)
		}
		defer func() {
			_ = f.Close()
		}()
		if err = store.Load(f); err != nil {
			log.Errorw("error load file",
				"file", cfg.FileStoragePath)
		}
	}

	// start flushing metrics on timer
	if cfg.StoreInterval > 0 {
		go app.FlushingMetrics(store, cfg.FileStoragePath, cfg.StoreInterval)
	}

	return app.StartServer(&config.Options{
		Config:  cfg,
		Storage: store,
	},
		lm)
}
