package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sejo412/ya-metrics/internal/app/agent"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var err error
	cfg := config.NewAgentConfig()
	if err = cfg.Load(); err != nil {
		return fmt.Errorf("error load config: %w", err)
	}
	cfg.Logger = logger.MustNewLogger(false)
	defer func() {
		_ = cfg.Logger.Logger.Sync()
	}()

	a := agent.NewAgent(cfg)
	l := a.Config.Logger.Logger
	version := config.GetVersion()
	l.Infow("agent starting",
		"version", version,
		"server", cfg.Address,
		"mode", cfg.Mode,
		"reportInterval", cfg.RealReportInterval,
		"pollInterval", cfg.RealPollInterval,
		"pathStyle", cfg.PathStyle,
		"sign", cfg.Key != "",
		"rateLimit", cfg.RateLimit)
	return a.Run(context.Background())
}
