package main

import (
	"context"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/internal/app/agent"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/spf13/pflag"
)

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
	pflag.BoolVarP(&cfg.PathStyle, "pathStyle", "o", config.DefaultPathStyle,
		"use path style for post metrics (deprecated)")
	pflag.StringVarP(&cfg.Key, "key", "k", config.DefaultSecretKey, "secret key")
	pflag.StringVar(&cfg.CryptoKey, "crypto-key", config.DefaultCryptoKey, "path to public key")
	pflag.IntVarP(&cfg.RateLimit, "limit", "l", config.DefaultRateLimit, "rate limit")
	pflag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}
	cfg.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	cfg.RealPollInterval = time.Duration(cfg.PollInterval) * time.Second
	cfg.Logger, err = logger.NewLogger()
	if err != nil {
		return err
	}
	defer func() {
		_ = cfg.Logger.Sync()
	}()

	a := agent.NewAgent(&cfg)
	l := a.Config.Logger
	version := config.GetVersion()
	l.Infow("agent starting",
		"version", version,
		"server", cfg.Address,
		"reportInterval", cfg.ReportInterval,
		"pollInterval", cfg.PollInterval,
		"pathStyle", cfg.PathStyle,
		"sign", cfg.Key != "",
		"rateLimit", cfg.RateLimit)
	ctx := context.Background()
	return a.Run(ctx)
}
