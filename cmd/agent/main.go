package main

import (
	"context"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/internal/app/agent"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/spf13/pflag"
)

type Service interface {
	Poll()
	Report(ctx context.Context, opts *config.AgentConfig)
	Run(ctx context.Context, opts *config.AgentConfig) error
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var cfg config.AgentConfig
	pflag.StringVarP(&cfg.Address, "address", "a", config.DefaultServerAddress, "addressFlag to connect to")
	pflag.IntVarP(&cfg.ReportInterval, "reportInterval", "r", config.DefaultReportInterval,
		"report interval (in seconds)")
	pflag.IntVarP(&cfg.PollInterval, "pollInterval", "p", config.DefaultPollInterval, "poll interval (in seconds)")
	pflag.BoolVarP(&cfg.UseOldAPI, "oldApi", "o", config.DefaultUseOldAPI, "use old api (deprecated)")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second

	metrics, err := agent.NewMetrics()
	if err != nil {
		return err
	}
	defer func() {
		_ = metrics.Logger.Sync()
	}()
	metrics.Logger.Infow("starting agent", "server", cfg.Address,
		"reportInterval", cfg.ReportInterval,
		"pollInterval", cfg.PollInterval,
		"oldApi", cfg.UseOldAPI)
	ctx, cancel := context.WithTimeout(context.Background(), config.ContextTimeout)
	defer cancel()
	return metrics.Run(ctx, &cfg)
}
