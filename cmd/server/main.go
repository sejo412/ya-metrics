package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sejo412/ya-metrics/internal/config"
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
	r.Post("/"+config.MetricPathPostPrefix+"/{kind}/{name}/{value}", postUpdate)
	r.Get("/"+config.MetricPathGetPrefix+"/{kind}/{name}", getValue)
	r.Get("/", getIndex)
	return http.ListenAndServe(cfg.Address, r)
}
