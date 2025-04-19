package config

import (
	"context"
	"crypto/rsa"
	"io"

	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
)

// Default settings for server.
const (
	DefaultAddress         string = ":8080"             // listen address
	DefaultStoreInterval   int    = 300                 // how often flush metrics from memory to disk
	DefaultFileStoragePath string = "/tmp/metrics.json" // file for saved metrics
	DefaultRestore         bool   = true                // restore metrics from file at startup
	DefaultDatabaseDSN     string = ""                  // default dsn string
)

// ServerConfig contains configuration for server application.
type ServerConfig struct {
	// Address - listen address.
	Address string `env:"ADDRESS"`
	// CryptoKey - path to private key
	CryptoKey string `env:"CRYPTO_KEY"`
	// FileStoragePath - file for saved metrics.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// DatabaseDSN - dsn string.
	DatabaseDSN string `env:"DATABASE_DSN"`
	// Key - string for crypt data.
	Key string `env:"KEY"`
	// StoreInterval - how often flush metrics from memory to disk.
	StoreInterval int `env:"STORE_INTERVAL"`
	// Restore -  restore metrics from file at startup.
	Restore bool `env:"RESTORE"`
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
	// PrivateKey - used for decrypt messages
	PrivateKey *rsa.PrivateKey
	// Config - used configuration.
	Config ServerConfig
}
