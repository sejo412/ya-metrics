package agent

import (
	"crypto/rsa"
	"runtime"
	"sync"

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
	mutex   sync.Mutex
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
	mutex   sync.Mutex
}

type psStats struct {
	cpuUtilization map[string]float64
	totalMemory    float64
	freeMemory     float64
}

type callOpts struct {
	hash string
}

func newCallOpts() *callOpts {
	return &callOpts{
		hash: "",
	}
}
