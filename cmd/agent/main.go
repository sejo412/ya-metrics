package main

import (
	"context"
	"fmt"
	"github.com/sejo412/ya-metrics/internal/config"
	"log"
	"math/rand"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"sync"
	"time"
)

const (
	UpdatePrefix  string = "update"
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
	r := new(Report)
	r.Gauge = make(map[string]float64)
	r.Counter = make(map[string]int64)
	var wg sync.WaitGroup
	wg.Add(1)
	go pollMetrics(m)
	wg.Add(1)
	go reportMetrics(m, r)
	wg.Wait()
	defer wg.Done()
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
		prefix := path.Join(UpdatePrefix, *root, *metricName)
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

func reportMetrics(m *Metrics, report *Report) {
	for {
		// skip if function start before polling
		if m.Counter.PollCount == 0 {
			continue
		}
		metrics := reflect.ValueOf(*m)
		parseMetric(&root, &metricName, metrics, report)
		var allMetrics []string
		for mpath, value := range report.Gauge {
			allMetrics = append(allMetrics, fmt.Sprintf("%s/%v", mpath, value))
		}
		for mpath, value := range report.Counter {
			allMetrics = append(allMetrics, fmt.Sprintf("%s/%v", mpath, value))
		}

		ch := make(chan string, len(allMetrics))

		ctx, cancel := context.WithTimeout(context.Background(), reportInterval/2)
		for _, metric := range allMetrics {
			go postMetric(ctx, metric, ch)
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled: %v", ctx.Err())
			case res := <-ch:
				log.Println(res)
			}
		}
		cancel()
		time.Sleep(reportInterval)
	}
}

func incPollCount(pollCount *int64) {
	*pollCount++
}

func postMetric(ctx context.Context, metric string, ch chan string) {
	uri := fmt.Sprintf("%s://%s:%s/%s",
		config.ServerScheme,
		config.ServerAddress,
		config.ListenPort,
		metric)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	ch <- fmt.Sprintf("Sent %s: %d", metric, resp.StatusCode)
}
