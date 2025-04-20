package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/spf13/pflag"
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
	Address string `env:"ADDRESS" json:"address,omitempty"`
	// CryptoKey - path to public key
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	// Key for crypt data.
	Key string `env:"KEY" json:"key,omitempty"`
	// ReportInterval - how often send reports.
	ReportInterval int `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	// PollInterval - how often poll runtime metrics.
	PollInterval int `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	RateLimit    int `env:"RATE_LIMIT"`
	// RealReportInterval - don't use it from code. It generates from ReportInterval.
	RealReportInterval time.Duration
	// RealPollInterval - don't use it from code. It generates from PollInterval.
	RealPollInterval time.Duration
	PathStyle        bool `env:"PATH_STYLE"`
}

// NewAgentConfig returns new *AgentConfig.
func NewAgentConfig() *AgentConfig {
	return &AgentConfig{}
}

// Load loads options from config file, flags, env.
func (a *AgentConfig) Load() error {
	var cfg AgentConfig
	cfgFile := pflag.StringP("config", "c", "",
		"path to config file in JSON format")
	pflag.StringVarP(&cfg.Address, "address", "a", "",
		fmt.Sprintf("address to connect to (default %q)", DefaultServerAddress))
	pflag.IntVarP(&cfg.ReportInterval, "reportInterval", "r", 0,
		fmt.Sprintf("report interval in seconds (default %d)", DefaultReportInterval))
	pflag.IntVarP(&cfg.PollInterval, "pollInterval", "p", 0,
		fmt.Sprintf("poll interval in seconds (default %d)", DefaultPollInterval))
	pflag.BoolVarP(&cfg.PathStyle, "pathStyle", "o", DefaultPathStyle,
		"use path style for post metrics (deprecated)")
	pflag.StringVarP(&cfg.Key, "key", "k", "",
		fmt.Sprintf("secret key for signing requests (default %q)", DefaultSecretKey))
	pflag.StringVar(&cfg.CryptoKey, "crypto-key", "",
		fmt.Sprintf("path to public key for encrypt requests (default %q)", DefaultCryptoKey))
	pflag.IntVarP(&cfg.RateLimit, "limit", "l", 0,
		fmt.Sprintf("rate limit in seconds (default %d)", DefaultRateLimit))
	pflag.Parse()
	if *cfgFile != "" {
		// rewrite flags from config (needs only for parsing config file)
		f, err := os.ReadFile(*cfgFile)
		if err != nil {
			return fmt.Errorf("error read config file: %w", err)
		}
		if err = json.Unmarshal(f, &cfg); err != nil {
			return fmt.Errorf("error unmarshal config file: %w", err)
		}
		// rewrite config from flags
		pflag.Parse()
	}
	// rewrite flags from envs
	err := env.Parse(&cfg)
	if err != nil {
		return fmt.Errorf("error parsing env: %w", err)
	}
	// moved from flags default values because it overwrites config if not specified
	if cfg.Address == "" {
		cfg.Address = DefaultServerAddress
	}
	if cfg.CryptoKey == "" {
		cfg.CryptoKey = DefaultCryptoKey
	}
	if cfg.Key == "" {
		cfg.Key = DefaultSecretKey
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = DefaultReportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = DefaultPollInterval
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = DefaultRateLimit
	}
	// fill agent params
	a.Address = cfg.Address
	a.CryptoKey = cfg.CryptoKey
	a.Key = cfg.Key
	a.RateLimit = cfg.RateLimit
	a.RealReportInterval = time.Duration(cfg.ReportInterval) * time.Second
	a.RealPollInterval = time.Duration(cfg.PollInterval) * time.Second
	a.PathStyle = cfg.PathStyle
	return nil
}
