package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/models"
)

type Router struct {
	chi.Router
	opts config.Options
}

func NewRouter() *Router {
	return &Router{chi.NewRouter(), config.Options{}}
}

func NewRouterWithOptions(opts *config.Options) *Router {
	router := NewRouter()
	router.opts.Config = opts.Config
	router.opts.Storage = opts.Storage
	router.opts.PrivateKey = opts.PrivateKey
	if opts.TrustedSubnets != nil {
		router.opts.TrustedSubnets = opts.TrustedSubnets
	} else {
		router.opts.TrustedSubnets = []net.IPNet{}
	}
	router.opts.Logger = opts.Logger
	router.SetMiddlewares()
	router.SetHandlers()
	return router
}

func (r *Router) SetMiddlewares() {
	r.Use(r.opts.Logger.WithLogging)
	if len(r.opts.TrustedSubnets) > 0 {
		r.Use(r.checkXRealIPHandler)
	}
	r.Use(middleware.WithValue("key", r.opts.Config.Key))
	if r.opts.PrivateKey != nil {
		r.Use(r.decryptHandler)
	}
	r.Use(checkHashHandle)
	r.Use(gzipHandle)
}

func (r *Router) SetHandlers() {
	r.Post("/"+models.MetricPathPostPrefix+"/{kind}/{name}/{value}", func(w http.ResponseWriter, req *http.Request) {
		metric := models.Metric{
			Kind:  chi.URLParam(req, "kind"),
			Value: chi.URLParam(req, "value"),
		}
		if err := CheckMetricKind(metric); err != nil {
			http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
			return
		}
		r.postUpdate(w, req)
	})
	r.Post("/"+models.MetricPathPostPrefix+"/", r.postUpdateJSON)
	r.Post("/"+models.MetricPathPostsPrefix+"/", r.postUpdatesJSON)
	r.Get("/"+models.MetricPathGetPrefix+"/{kind}/{name}", r.getValue)
	r.Get("/", r.getIndex)
	r.Post("/"+models.MetricPathGetPrefix+"/", r.getMetricJSON)
	r.Get("/"+models.PingPath, r.pingStorage)
}
