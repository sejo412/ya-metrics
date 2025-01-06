package agent

import (
	"runtime"

	"github.com/sejo412/ya-metrics/internal/logger"
)

type Metrics struct {
	Gauge   Gauge
	Counter Counter
	Logger  *logger.Logger
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
