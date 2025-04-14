package storage

import (
	"fmt"
	"net/url"
	"path"
	"time"
)

const (
	ctxTimeout                 = 10 * time.Second
	defaultPostgresPort int    = 5432
	MemoryScheme        string = "memory"     // memory storage scheme
	PostgresSchemeLong  string = "postgresql" // long string for postgres scheme
	PostgresSchemeShort string = "postgres"   // short string for postgres scheme
)

const (
	sslModeDisable    = "disable"
	sslModeAllow      = "allow"
	sslModePrefer     = "prefer"
	sslModeRequire    = "require"
	sslModeVerifyCA   = "verify-ca"
	sslModeVerifyFull = "verify-full"
)

// Options defines settings to communicate with Storage.
type Options struct {
	// Scheme - memory or postgres.
	Scheme string
	// Host - host for connect to backend.
	Host string
	// Username - login for connect to backend.
	Username string
	// Password - password for connect to backend.
	Password string
	// Database - database to use.
	Database string
	// SSLMode - settings for SSL.
	SSLMode string
	// Port - TCP port for connect to backend.
	Port int
}

// ParseDSN parses DSN string to Options type.
func ParseDSN(dsn string) (opts Options, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return opts, fmt.Errorf("failed to parse dsn: %w", err)
	}
	var port int
	switch u.Scheme {
	case MemoryScheme, "":
		return Options{
			Scheme: MemoryScheme,
		}, nil
	case PostgresSchemeLong, PostgresSchemeShort:
		port = defaultPostgresPort
	default:
		return opts, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}
	password, _ := u.User.Password()
	paramSslMode := u.Query().Get("sslmode")
	var sslMode string
	switch paramSslMode {
	case "":
		sslMode = "disable"
	case sslModeAllow, sslModePrefer, sslModeRequire, sslModeVerifyCA, sslModeVerifyFull, sslModeDisable:
		sslMode = paramSslMode
	default:
		return opts, fmt.Errorf("unsupported sslmode: %s", sslMode)
	}
	return Options{
		Scheme:   u.Scheme,
		Host:     u.Hostname(),
		Port:     port,
		Username: u.User.Username(),
		Password: password,
		Database: path.Base(u.Path),
		SSLMode:  sslMode,
	}, nil
}
