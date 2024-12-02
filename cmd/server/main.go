package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sejo412/ya-metrics/cmd/server/app"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/spf13/pflag"
	"log"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var cfg Config
	pflag.StringVarP(&cfg.Address, "address", "a", DefaultAddress, "Listen address")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()
	store := storage.NewMemoryStorage()
	r.Use(middleware.WithValue("store", store))
	r.Use(middleware.CleanPath)
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
	r.Get("/"+models.MetricPathGetPrefix+"/{kind}/{name}", getValue)
	r.Get("/", getIndex)
	return http.ListenAndServe(cfg.Address, r)
}
