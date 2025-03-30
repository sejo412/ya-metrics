package agent

import (
	"runtime"

	"github.com/sejo412/ya-metrics/internal/config"
)

type Agent struct {
	Metrics *metrics
	Config  *config.AgentConfig
}

type metrics struct {
	gauge   gauge
	counter counter
}

type gauge struct {
	memStats    *runtime.MemStats
	randomValue float64
	psStats     psStats
}

type counter struct {
	pollCount int64
}

type report struct {
	gauge   map[string]float64
	counter map[string]int64
}

type psStats struct {
	totalMemory    float64
	freeMemory     float64
	cpuUtilization map[string]float64
}
