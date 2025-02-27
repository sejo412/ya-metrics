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
	DefaultPathStyle      bool   = false
	ContextTimeout               = 1 * time.Second
	DefaultRateLimit      int    = 2
)

type AgentConfig struct {
	Address            string `env:"ADDRESS"`
	ReportInterval     int    `env:"REPORT_INTERVAL"`
	PollInterval       int    `env:"POLL_INTERVAL"`
	PathStyle          bool   `env:"PATH_STYLE"`
	Key                string `env:"KEY"`
	RateLimit          int    `env:"RATE_LIMIT"`
	RealReportInterval time.Duration
	RealPollInterval   time.Duration
	Logger             *logger.Logger
}
