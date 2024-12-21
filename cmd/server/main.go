package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sejo412/ya-metrics/cmd/server/app"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var cfg Config
	pflag.StringVarP(&cfg.Address, "address", "a", DefaultAddress, "Listen address")
	pflag.IntVarP(&cfg.StoreInterval, "storeInterval", "i", DefaultStoreInterval, "Store interval")
	pflag.StringVarP(&cfg.FileStoragePath, "fileStoragePath", "f", DefaultFileStoragePath, "File storage path")
	pflag.BoolVarP(&cfg.Restore, "restore", "r", DefaultRestore, "Restore metrics")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = logger.Sync()
	}()
	sugar = *logger.Sugar()
	r := chi.NewRouter()
	store := storage.NewMemoryStorage()
	r.Use(WithLogging)
	r.Use(middleware.WithValue("store", store))
	r.Use(middleware.WithValue("config", cfg))
	r.Use(gzipHandle)
	r.Post("/"+models.MetricPathPostPrefix+"/{kind}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		metric := models.Metric{
			Kind:  chi.URLParam(r, "kind"),
			Value: chi.URLParam(r, "value"),
		}
		if err = app.CheckMetricKind(metric); err != nil {
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			return
		}
		postUpdate(w, r)
	})
	r.Post("/"+models.MetricPathPostPrefix+"/", postUpdateJSON)
	r.Get("/"+models.MetricPathGetPrefix+"/{kind}/{name}", getValue)
	r.Get("/", getIndex)
	r.Post("/"+models.MetricPathGetPrefix+"/", getMetricJSON)
	sugar.Infow("server starting",
		"address", cfg.Address,
		"storeInterval", cfg.StoreInterval,
		"fileStoragePath", cfg.FileStoragePath,
		"restore", cfg.Restore)
	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if cfg.Restore {
		if err = store.Load(cfg.FileStoragePath); err != nil {
			log.Printf("failed to restore metrics: %v", err)
		}
	}
	if cfg.StoreInterval > 0 {
		go app.FlushingMetrics(store, cfg.FileStoragePath, cfg.StoreInterval)
	}
	return server.ListenAndServe()
}
