package agent

import (
	"runtime"

	"github.com/sejo412/ya-metrics/internal/config"
)

type Agent struct {
	Metrics *Metrics
	Config  *config.AgentConfig
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
