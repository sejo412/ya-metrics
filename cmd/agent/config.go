package main

import (
	"reflect"
	"runtime"
	"time"
)

const (
	ServerScheme          string = "http"
	DefaultServerAddress  string = "localhost:8080"
	DefaultPollInterval   int    = 2
	DefaultReportInterval int    = 10
	ContextTimeout               = 2 * time.Second
)

type Config struct {
	Address            string `env:"ADDRESS"`
	ReportInterval     int    `env:"REPORT_INTERVAL"`
	PollInterval       int    `env:"POLL_INTERVAL"`
	RealReportInterval time.Duration
	RealPollInterval   time.Duration
}

type Metrics struct {
	Gauge   Gauge
	Counter Counter
}

type Gauge struct {
	MemStats    *runtime.MemStats
	RandomValue float64
}

type Counter struct {
	PollCount int64
}

type Report struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

var (
	root        string = ""
	metricName  string = ""
	float64Type        = reflect.TypeOf(float64(0))
	int64Type          = reflect.TypeOf(int64(0))
)
