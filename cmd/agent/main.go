package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

const (
	GaugePrefix   string = "gauge"
	CounterPrefix string = "counter"
)

const (
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
)

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

var pollCount int64 = 0

func main() {
	m := new(Metrics)
	go pollMetrics(m)
	for _ = range time.Tick(pollInterval) {
		fmt.Println(m.Counter.PollCount)
		fmt.Println(m.Gauge.RandomValue)
		reportMetrics(m)
	}
}

func pollMetrics(m *Metrics) {
	for {
		mem := &runtime.MemStats{}
		runtime.ReadMemStats(mem)
		incPollCount(&pollCount)
		randomValue := rand.Float64() * 1000
		m.Gauge.MemStats = mem
		m.Gauge.RandomValue = randomValue
		m.Counter.PollCount = pollCount
		time.Sleep(pollInterval)
	}
}

func reportMetrics(m *Metrics) {
	memValues := reflect.ValueOf(*m.Gauge.MemStats)
	memTypes := memValues.Type()
	for index := range memValues.NumField() {
		fmt.Printf("%v: %v\n", memTypes.Field(index).Name, memValues.Field(index))
	}
	//fmt.Printf("%#v\n", memTypes.Name())
}

func incPollCount(pollCount *int64) {
	*pollCount++
}
