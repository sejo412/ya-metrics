package agent

import (
	"crypto/rsa"
	"runtime"

	"github.com/sejo412/ya-metrics/internal/config"
)

type Agent struct {
	Metrics   *metrics
	Config    *config.AgentConfig
	PublicKey *rsa.PublicKey
}

type metrics struct {
	gauge   gauge
	counter counter
}

type gauge struct {
	memStats    *runtime.MemStats
	psStats     psStats
	randomValue float64
}

type counter struct {
	pollCount int64
}

type report struct {
	gauge   map[string]float64
	counter map[string]int64
}

type psStats struct {
	cpuUtilization map[string]float64
	totalMemory    float64
	freeMemory     float64
}
