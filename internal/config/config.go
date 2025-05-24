package config

import (
	"os"
	"strings"
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

// Mode grpc or http mode.
type Mode int

const (
	UnknownMode Mode = iota
	HTTPMode
	GRPCMode
)

const (
	UnknownModeName string = "unknown"
	HTTPModeName    string = "http"
	GRPCModeName    string = "grpc"
)

// String returns string of mode.
func (m Mode) String() string {
	switch m {
	case HTTPMode:
		return HTTPModeName
	case GRPCMode:
		return GRPCModeName
	default:
		return UnknownModeName
	}
}

// IsValid returns true if mode is valid.
func (m Mode) IsValid() bool {
	return m != UnknownMode
}

// ModeFromString returns Mode from string or unknown.
func ModeFromString(s string) Mode {
	switch strings.ToLower(s) {
	case HTTPModeName:
		return HTTPMode
	case GRPCModeName:
		return GRPCMode
	default:
		return UnknownMode
	}
}
