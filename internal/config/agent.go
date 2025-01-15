package config

import (
	"time"

	"github.com/sejo412/ya-metrics/internal/logger"
)

const (
	ServerScheme          string = "http"
	DefaultServerAddress  string = "localhost:8080"
	DefaultPollInterval   int    = 2
	DefaultReportInterval int    = 10
	DefaultUseOldAPI      bool   = false
	ContextTimeout               = 1 * time.Second
)

type AgentConfig struct {
	Address            string `env:"ADDRESS"`
	ReportInterval     int    `env:"REPORT_INTERVAL"`
	PollInterval       int    `env:"POLL_INTERVAL"`
	UseOldAPI          bool   `env:"USE_OLD_API"`
	RealReportInterval time.Duration
	RealPollInterval   time.Duration
	Logger             *logger.Logger
}
