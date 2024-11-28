package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	. "github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/storage"
	"net/http"
)

func main() {
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
	return http.ListenAndServe(fmt.Sprintf("%s:%s", ListenAddress, ListenPort), r)
}
