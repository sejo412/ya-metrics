package config

import (
	"context"
	"io"

	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
)

const (
	DefaultAddress         string = ":8080"
	DefaultStoreInterval   int    = 300
	DefaultFileStoragePath string = "/tmp/metrics.json"
	DefaultRestore         bool   = true
	DefaultDatabaseDSN     string = ""
)

type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

type Storage interface {
	Open(ctx context.Context, opts storage.Options) error
	Close()
	Ping(ctx context.Context) error
	AddOrUpdate(context.Context, models.Metric) error
	MassAddOrUpdate(context.Context, []models.Metric) error
	Get(ctx context.Context, kind string, name string) (models.Metric, error)
	GetAll(ctx context.Context) ([]models.Metric, error)
	Flush(dst io.Writer) error
	Load(src io.Reader) error
	Init(ctx context.Context) error
}
type Options struct {
	Config  ServerConfig
	Storage Storage
}
