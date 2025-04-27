package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/sejo412/ya-metrics/pkg/utils"
)

type Router struct {
	chi.Router
	opts config.Options
}

func NewRouter() *Router {
	return &Router{chi.NewRouter(), config.Options{}}
}

func NewRouterWithConfig(opts *config.Options, logs *logger.Middleware) *Router {
	router := NewRouter()

	// config & storage
	router.opts.Config = opts.Config
	router.opts.Storage = opts.Storage
	router.opts.PrivateKey = opts.PrivateKey
	router.opts.TrustedSubnets = opts.TrustedSubnets

	// middlewares
	router.Use(logs.WithLogging)
	if len(*router.opts.TrustedSubnets) > 0 {
		router.Use(router.checkXRealIPHandler)
	}
	router.Use(middleware.WithValue("key", opts.Config.Key))
	if router.opts.PrivateKey != nil {
		router.Use(router.decryptHandler)
	}
	router.Use(checkHashHandle)
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
		router.postUpdate(w, r)
	})
	router.Post("/"+models.MetricPathPostPrefix+"/", router.postUpdateJSON)
	router.Post("/"+models.MetricPathPostsPrefix+"/", router.postUpdatesJSON)
	router.Get("/"+models.MetricPathGetPrefix+"/{kind}/{name}", router.getValue)
	router.Get("/", router.getIndex)
	router.Post("/"+models.MetricPathGetPrefix+"/", router.getMetricJSON)
	router.Get("/"+models.PingPath, router.pingStorage)

	return router
}

func StartServer(ctx context.Context, opts *config.Options,
	logs *logger.Middleware) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	log := logs.Logger
	if opts.Config.CryptoKey != "" {
		k, err := os.ReadFile(opts.Config.CryptoKey)
		if err != nil {
			return fmt.Errorf("error read crypto key: %w", err)
		}
		opts.PrivateKey, err = utils.LoadRSAPrivateKey(k)
		if err != nil {
			return fmt.Errorf("error load private key: %w", err)
		}
	}

	warnings := make([]string, 0)
	if opts.Config.TrustedSubnet != "" {
		var er error
		opts.TrustedSubnets, er = stringCIDRsToIPNets(opts.Config.TrustedSubnet)
		if er != nil {
			warnings = append(warnings, er.Error())
		}
	}

	// we wan't check error twice (already checked in main)
	dsn, _ := storage.ParseDSN(opts.Config.DatabaseDSN)
	// start flushing metrics on timer
	wg := sync.WaitGroup{}
	if opts.Config.StoreInterval > 0 && dsn.Scheme == "memory" && opts.Config.StoreFile != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			FlushingMetrics(ctx, opts.Storage, opts.Config.StoreFile, opts.Config.StoreInterval)
		}()
	}
	router := NewRouterWithConfig(opts, logs)

	log.Infow("server starting",
		"version", config.GetVersion(),
		"address", opts.Config.Address,
		"storeInterval", opts.Config.StoreInterval,
		"fileStoragePath", opts.Config.StoreFile,
		"restore", opts.Config.Restore)
	server := &http.Server{
		Addr:              opts.Config.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, config.GracefulSignals...)
	go func() {
		<-sigs
		log.Info("shutting down server...")
		cancel()
		ctx, cancel := context.WithTimeout(context.Background(), config.GracefulTimeout)
		defer cancel()
		if er := server.Shutdown(ctx); er != nil {
			log.Errorf("error shutting down server: %v", er)
		}
		close(idleConnsClosed)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting server: %w", err)
	}
	if len(warnings) > 0 {
		log.Warnln("server started with warnings: %v", warnings)
	}
	<-idleConnsClosed
	opts.Storage.Close()
	wg.Wait()
	log.Info("server stopped")
	return nil
}
