package main

const (
	DefaultAddress string = ":8080"
)

type Config struct {
	Address string `env:"ADDRESS"`
}
