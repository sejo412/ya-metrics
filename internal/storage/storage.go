package storage

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	ctxTimeout                 = 1 * time.Second
	defaultPostgresPort int    = 5432
	MemoryScheme        string = "memory"
)

const (
	sslModeDisable    = "disable"
	sslModeAllow      = "allow"
	sslModePrefer     = "prefer"
	sslModeRequire    = "require"
	sslModeVerifyCA   = "verify-ca"
	sslModeVerifyFull = "verify-full"
)

type Options struct {
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

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
	case "postgresql", "postgres":
		port = defaultPostgresPort
		break
	default:
		return opts, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return opts, fmt.Errorf("failed to parse port in dsn: %w", err)
		}
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
		Host:     u.Host,
		Port:     port,
		Username: u.User.Username(),
		Password: password,
		Database: path.Base(u.Path),
		SSLMode:  sslMode,
	}, nil
}
