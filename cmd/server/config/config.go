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

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

type Storage interface {
	Open(opts storage.Options) error
	Close()
	Ping(ctx context.Context) error
	AddOrUpdate(models.Metric) error
	Get(kind string, name string) (models.Metric, error)
	GetAll() []models.Metric
	Flush(dst io.Writer) error
	Load(src io.Reader) error
}
type Options struct {
	Config  Config
	Storage Storage
}
