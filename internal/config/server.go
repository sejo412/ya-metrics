package config

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/spf13/pflag"
)

// Default settings for server.
const (
	DefaultAddress       string = ":8080"             // listen address
	DefaultStoreInterval int    = 300                 // how often flush metrics from memory to disk
	DefaultStoreFile     string = "/tmp/metrics.json" // file for saved metrics
	DefaultRestore       bool   = true                // restore metrics from file at startup
	DefaultDatabaseDSN   string = ""                  // default dsn string
)

// ServerConfig contains configuration for server application.
type ServerConfig struct {
	// Restore -  restore metrics from file at startup.
	Restore *bool `env:"RESTORE" json:"restore,omitempty"`
	// Address - listen address.
	Address string `env:"ADDRESS" json:"address,omitempty"`
	// CryptoKey - path to private key
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	// StoreFile - file for saved metrics.
	StoreFile string `env:"STORE_FILE" json:"store_file,omitempty"`
	// DatabaseDSN - dsn string.
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	// Key - string for sign data.
	Key string `env:"KEY" json:"key,omitempty"`
	// StoreInterval - how often flush metrics from memory to disk.
	StoreInterval int `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
}

// Storage interface for used backend.
type Storage interface {
	// Open opens connection with backend.
	Open(ctx context.Context, opts storage.Options) error
	// Close closes connection.
	Close()
	// Ping checks backend for receive requests.
	Ping(ctx context.Context) error
	// Upsert inserts or updates metric.
	Upsert(context.Context, models.Metric) error
	// MassUpsert inserts or updates slice of metrics.
	MassUpsert(context.Context, []models.Metric) error
	// Get returns metric by kind and name.
	Get(ctx context.Context, kind string, name string) (models.Metric, error)
	// GetAll returns all metrics.
	GetAll(ctx context.Context) ([]models.Metric, error)
	// Flush saves metrics to file.
	Flush(ctx context.Context, dst io.Writer) error
	// Load resores metrics from file.
	Load(ctx context.Context, src io.Reader) error
	// Init initialized backend database.
	Init(ctx context.Context) error
}

// Options contains server's options for startup.
type Options struct {
	// Storage - used storage backend.
	Storage Storage
	// PrivateKey - used for decrypt data.
	PrivateKey *rsa.PrivateKey
	// Config - used configuration.
	Config ServerConfig
}

// NewServerConfig returns new *ServerConfig
func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Restore: new(bool),
	}
	*cfg.Restore = DefaultRestore
	return cfg
}

func (s *ServerConfig) Load() error {
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	cfgFile := flagSet.StringP("config", "c", "",
		"path to config file in JSON format")
	flagAddress := flagSet.StringP("address", "a", "",
		fmt.Sprintf("Listen address (default: \"%s\")", DefaultAddress))
	flagStoreInterval := flagSet.IntP("storeInterval", "i", 0,
		fmt.Sprintf("Store interval (default: %d)", DefaultStoreInterval))
	flagStoreFile := flagSet.StringP("storeFile", "f", "",
		fmt.Sprintf("File storage path (default: \"%s\")", DefaultStoreFile))
	flagRestore := flagSet.BoolP("restore", "r", false,
		fmt.Sprintf("Restore metrics (default: %t)", DefaultRestore))
	flagDatabaseDSN := flagSet.StringP("database-dsn", "d", "",
		fmt.Sprintf("Database DSN (default: \"%s\")", DefaultDatabaseDSN))
	flagKey := flagSet.StringP("key", "k", "",
		fmt.Sprintf("secret key (default: \"%s\")", DefaultSecretKey))
	flagCryptoKey := flagSet.String("crypto-key", "",
		fmt.Sprintf("path to public key (default: \"%s\")", DefaultCryptoKey))

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("error parse flags: %w", err)
	}
	if *cfgFile != "" {
		f, err := os.ReadFile(*cfgFile)
		if err != nil {
			return fmt.Errorf("error read config file: %w", err)
		}
		if err = json.Unmarshal(f, s); err != nil {
			return fmt.Errorf("error unmarshal config file: %w", err)
		}
	}

	// Workaround: double parse flags overwrites bool with default even it's not present in command line
	if flagSet.Changed("address") {
		s.Address = *flagAddress
	}
	if flagSet.Changed("store_interval") {
		s.StoreInterval = *flagStoreInterval
	}
	if flagSet.Changed("store_file") {
		s.StoreFile = *flagStoreFile
	}
	if flagSet.Changed("restore") {
		s.Restore = flagRestore
	}
	if flagSet.Changed("database_dsn") {
		s.DatabaseDSN = *flagDatabaseDSN
	}
	if flagSet.Changed("key") {
		s.Key = *flagKey
	}
	if flagSet.Changed("crypto_key") {
		s.CryptoKey = *flagCryptoKey
	}

	// rewrite flags from envs
	err := env.Parse(s)
	if err != nil {
		return fmt.Errorf("error parsing env: %w", err)
	}
	// moved from flags default values because it overwrites config if not specified
	if s.Address == "" {
		s.Address = DefaultAddress
	}
	if s.StoreInterval == 0 {
		s.StoreInterval = DefaultStoreInterval
	}
	if s.StoreFile == "" {
		s.StoreFile = DefaultStoreFile
	}
	if s.Restore == nil {
		s.Restore = new(bool)
		*s.Restore = DefaultRestore
	}
	if s.DatabaseDSN == "" {
		s.DatabaseDSN = DefaultDatabaseDSN
	}
	if s.Key == "" {
		s.Key = DefaultSecretKey
	}
	if s.CryptoKey == "" {
		s.CryptoKey = DefaultCryptoKey
	}
	return nil
}
