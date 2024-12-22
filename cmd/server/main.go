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

	store := storage.NewMemoryStorage()

	// restore metrics
	if cfg.Restore {
		f, err := os.Open(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("error open file %s: %w", cfg.FileStoragePath, err)
		}
		defer func() {
			_ = f.Close()
		}()
		if err = store.Load(f); err != nil {
			log.Error("failed to restore metrics: %v", err)
		}
	}

	// start flushing metrics on timer
	if cfg.StoreInterval > 0 {
		go app.FlushingMetrics(store, cfg.FileStoragePath, cfg.StoreInterval)
	}

	return app.StartServer(&config.Options{
		Config:  &cfg,
		Storage: store,
	},
		lm)
}
