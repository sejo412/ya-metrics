package main

const (
	DefaultAddress         string = ":8080"
	DefaultStoreInterval   int    = 300
	DefaultFileStoragePath string = "/tmp/metrics.json"
	DefaultRestore         bool   = true
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}
