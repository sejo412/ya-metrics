package main

import (
	"fmt"
	"math/rand"
	"path"
	"reflect"
	"runtime"
	"time"
)

const (
	UploadsPrefix string = "update"
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

var (
	pollCount   int64  = 0
	root        string = ""
	metricName  string = ""
	float64Type        = reflect.TypeOf(float64(0))
	int64Type          = reflect.TypeOf(int64(0))
)

func main() {
	m := new(Metrics)
	pollMetrics(m)

	//	go pollMetrics(m)
	//	for _ = range time.Tick(pollInterval) {
	//		fmt.Println(m.Counter.PollCount)
	//		fmt.Println(m.Gauge.RandomValue)
	//		reportMetrics(m)
	//	}
	//parseMetric(*m)
	z := reflect.ValueOf(*m)
	r := new(Report)
	r.Gauge = make(map[string]float64)
	r.Counter = make(map[string]int64)
	parseMetric(&root, &metricName, z, r)
	//fmt.Printf("%#v", r)
}

func pollMetrics(m *Metrics) {
	//for {
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	incPollCount(&pollCount)
	randomValue := rand.Float64() * 1000
	m.Gauge.MemStats = mem
	m.Gauge.RandomValue = randomValue
	m.Counter.PollCount = pollCount
	//	time.Sleep(pollInterval)
	//}
}

func parseMetric(root, metricName *string, data reflect.Value, report *Report) {
	types := data.Type()
	switch types.Name() {
	case "Gauge":
		*root = GaugePrefix
	case "Counter":
		*root = CounterPrefix
	}

	switch data.Kind() {
	case reflect.Struct:
		for i := 0; i < data.NumField(); i++ {
			value := data.Field(i)
			*metricName = types.Field(i).Name
			parseMetric(root, metricName, reflect.ValueOf(value.Interface()), report)
		}
	case reflect.Ptr:
		f := data.Elem()
		parseMetric(root, metricName, f, report)
	default:
		prefix := path.Join(UploadsPrefix, *root, *metricName)
		switch *root {
		case GaugePrefix:
			if data.Type().ConvertibleTo(float64Type) {
				v := data.Convert(float64Type)
				report.Gauge[prefix] = v.Float()
			}
		case CounterPrefix:
			if data.Type().ConvertibleTo(int64Type) {
				v := data.Convert(int64Type)
				report.Counter[prefix] = v.Int()
			}
		}
	}
}

func reportMetrics(report *Report) {
	for path, value := range report.Gauge {
		fmt.Printf("%s: %f\n", path, value/1000)
	}
}

func incPollCount(pollCount *int64) {
	*pollCount++
}
