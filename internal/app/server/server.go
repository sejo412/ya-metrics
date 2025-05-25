package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/sejo412/ya-metrics/pkg/utils"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	HTTPServer *Router
	GRPCServer *GRPCServer
}

func NewServerWithOptions(opts *config.Options) *Server {
	return &Server{
		HTTPServer: NewRouterWithOptions(opts),
		GRPCServer: NewGRPCServerWithOptions(opts),
	}
}

func StartServer(ctx context.Context, opts *config.Options,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	log := opts.Logger.Logger
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

	cfg := opts.Config
	warnings := make([]string, 0)
	if cfg.TrustedSubnet != "" {
		var er error
		opts.TrustedSubnets, er = stringCIDRsToIPNets(cfg.TrustedSubnet)
		if er != nil {
			warnings = append(warnings, er.Error())
		}
	} else {
		opts.TrustedSubnets = []net.IPNet{}
	}

	// we don't want check error twice (already checked in main)
	dsn, _ := storage.ParseDSN(cfg.DatabaseDSN)
	// start flushing metrics on timer
	wg := sync.WaitGroup{}
	if cfg.StoreInterval > 0 && dsn.Scheme == "memory" && cfg.StoreFile != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			FlushingMetrics(ctx, opts.Storage, cfg.StoreFile, cfg.StoreInterval)
		}()
	}

	// convert trusted subnets to human readable format
	hrTrustedSubnets := make([]string, 0, len(opts.TrustedSubnets))
	for _, subnet := range opts.TrustedSubnets {
		hrTrustedSubnets = append(hrTrustedSubnets, subnet.String())
	}

	setKey := false
	if cfg.Key != "" {
		setKey = true
	}
	log.Infow("server starting",
		"version", config.GetVersion(),
		"address", cfg.Address,
		"address_grpc", cfg.AddressGRPC,
		"storeInterval", cfg.StoreInterval,
		"fileStoragePath", cfg.StoreFile,
		"restore", cfg.Restore,
		"setKey", setKey,
		"trustedSubnets", hrTrustedSubnets)
	if len(warnings) > 0 {
		log.Warnln("warnings: ", warnings)
	}
	server := NewServerWithOptions(opts)
	httpServer := &http.Server{
		Addr:              cfg.Address,
		Handler:           server.HTTPServer,
		ReadHeaderTimeout: 5 * time.Second,
	}

	grpcServer := grpc.NewServer(gRPCServerOptions(server.GRPCServer, cfg.Key)...)
	RegisterGRPCServer(grpcServer, server.GRPCServer)

	// for debug with grpcurl
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", cfg.AddressGRPC)
	if err != nil {
		log.Errorw("failed to listen", "address", cfg.AddressGRPC, "error", err)
		return err
	}

	idleConnsClosed := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, config.GracefulSignals...)
	go func() {
		<-sigs
		log.Info("shutting down server...")
		cancel()
		ct, cnl := context.WithTimeout(context.Background(), config.GracefulTimeout)
		defer cnl()
		if er := httpServer.Shutdown(ct); er != nil {
			log.Errorw("shutting down server", "error", er)
		}
		grpcServer.GracefulStop()
		close(idleConnsClosed)
	}()
	var errGroup errgroup.Group
	errGroup.Go(func() error {
		return grpcServer.Serve(grpcListener)
	})
	errGroup.Go(func() error {
		return httpServer.ListenAndServe()
	})
	if err = errGroup.Wait(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting server: %w", err)
	}
	<-idleConnsClosed
	opts.Storage.Close()
	wg.Wait()
	log.Info("server stopped")
	return nil
}
