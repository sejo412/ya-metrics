package main

import (
	"time"
)

const (
	ServerScheme          string = "http"
	DefaultServerAddress  string = "localhost:8080"
	DefaultPollInterval   int    = 2
	DefaultReportInterval int    = 10
	DefaultUseOldAPI      bool   = false
	ContextTimeout               = 2 * time.Second
)

type Config struct {
	Address            string `env:"ADDRESS"`
	ReportInterval     int    `env:"REPORT_INTERVAL"`
	PollInterval       int    `env:"POLL_INTERVAL"`
	UseOldAPI          bool   `env:"USE_OLD_API"`
	RealReportInterval time.Duration
	RealPollInterval   time.Duration
}
