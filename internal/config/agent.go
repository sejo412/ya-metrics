package config

import (
	"time"

	"github.com/sejo412/ya-metrics/internal/logger"
)

// Constants for agent settings.
const (
	ServerScheme          string = "http"           // scheme to communicate with server
	DefaultServerAddress  string = "localhost:8080" // default server endpoint
	DefaultPollInterval   int    = 2                // default interval for poll runtime metrics
	DefaultReportInterval int    = 10               // default interval for report
	DefaultPathStyle      bool   = false            // default uses path-style
	ContextTimeout               = 1 * time.Second  // timeout for network communications
	DefaultRateLimit      int    = 2
)

// AgentConfig contains configuration for agent application.
type AgentConfig struct {
	// Logger - used logger.
	Logger *logger.Logger
	// Address - server endpoint.
	Address string `env:"ADDRESS"`
	// CryptoKey - path to public key
	CryptoKey string `env:"CRYPTO_KEY"`
	// Key for crypt data.
	Key string `env:"KEY"`
	// ReportInterval - how often send reports.
	ReportInterval int `env:"REPORT_INTERVAL"`
	// PollInterval - how often poll runtime metrics.
	PollInterval int `env:"POLL_INTERVAL"`
	RateLimit    int `env:"RATE_LIMIT"`
	// RealReportInterval - don't use it from code. It generates from ReportInterval.
	RealReportInterval time.Duration
	// RealPollInterval - don't use it from code. It generates from PollInterval.
	RealPollInterval time.Duration
	PathStyle        bool `env:"PATH_STYLE"`
}
