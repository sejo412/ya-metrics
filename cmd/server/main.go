package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/spf13/pflag"
	"net/http"
)

var address string

func main() {
	pflag.StringVarP(&address, "address", "a", "localhost:8080", "Listen address")
	pflag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	r := chi.NewRouter()
	store := storage.NewMemoryStorage()
	r.Use(middleware.WithValue("store", store))
	r.Use(middleware.CleanPath)
	r.Post("/update/{kind}/{name}/{value}", postUpdate)
	r.Get("/value/{kind}/{name}", getValue)
	r.Get("/", getIndex)
	return http.ListenAndServe(address, r)
}
