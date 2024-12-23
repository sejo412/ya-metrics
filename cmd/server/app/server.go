package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sejo412/ya-metrics/cmd/server/config"
	"github.com/sejo412/ya-metrics/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	chi.Router
}

func NewRouter() *Router {
	return &Router{chi.NewRouter()}
}

func NewRouterWithConfig(opts *config.Options, logger *LoggerMiddleware) *Router {
	router := NewRouter()

	// middlewares
	router.Use(logger.WithLogging)
	router.Use(middleware.WithValue("store", opts.Storage))
	router.Use(middleware.WithValue("config", opts.Config))
	router.Use(gzipHandle)

	// requests
	router.Post("/"+models.MetricPathPostPrefix+"/{kind}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		metric := models.Metric{
			Kind:  chi.URLParam(r, "kind"),
			Value: chi.URLParam(r, "value"),
		}
		if err := CheckMetricKind(metric); err != nil {
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			return
		}
		postUpdate(w, r)
	})
	router.Post("/"+models.MetricPathPostPrefix+"/", postUpdateJSON)
	router.Get("/"+models.MetricPathGetPrefix+"/{kind}/{name}", getValue)
	router.Get("/", getIndex)
	router.Post("/"+models.MetricPathGetPrefix+"/", getMetricJSON)

	return &Router{router}
}

func StartServer(opts *config.Options, logger *LoggerMiddleware) error {
	log := logger.Logger
	router := NewRouterWithConfig(opts, logger)

	log.Infow("server starting",
		"address", opts.Config.Address,
		"storeInterval", opts.Config.StoreInterval,
		"fileStoragePath", opts.Config.FileStoragePath,
		"restore", opts.Config.Restore)
	server := &http.Server{
		Addr:              opts.Config.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server.ListenAndServe()
}
