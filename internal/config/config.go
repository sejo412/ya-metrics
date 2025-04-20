package config

import (
	"os"
	"syscall"
	"time"
)

const (
	DefaultSecretKey string = "" // default key for crypt
	DefaultCryptoKey string = "" // default path to private/public key
)

const GracefulTimeout = time.Second * 1 // timeout for graceful shutdown functions

// GracefulSignals - signals for graceful shutdown
var GracefulSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGQUIT,
	syscall.SIGINT,
}
