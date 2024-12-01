package main

import (
	"context"
	"fmt"
	"github.com/sejo412/ya-metrics/internal/domain"
	"log"
	"math/rand"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"time"
)

// pollMetrics collects runtime metrics in infinite loop
func pollMetrics(m *Metrics, c *Config) {
	for {
		mem := &runtime.MemStats{}
		runtime.ReadMemStats(mem)
		randomValue := rand.Float64() * 1000
		m.Gauge.MemStats = mem
		m.Gauge.RandomValue = randomValue
		m.Counter.PollCount = 1
		time.Sleep(c.RealPollInterval)
	}
}

// parseMetrics parses polled metrics
func parseMetric(root, metricName *string, data reflect.Value, report *Report) {
	types := data.Type()
	switch types.Name() {
	case "Gauge":
		*root = domain.MetricKindGauge
	case "Counter":
		*root = domain.MetricKindCounter
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
		prefix := path.Join(domain.MetricPathPostPrefix, *root, *metricName)
		switch *root {
		case domain.MetricKindGauge:
			if data.Type().ConvertibleTo(float64Type) {
				v := data.Convert(float64Type)
				report.Gauge[prefix] = v.Float()
			}
		case domain.MetricKindCounter:
			if data.Type().ConvertibleTo(int64Type) {
				v := data.Convert(int64Type)
				report.Counter[prefix] = v.Int()
			}
		}
	}
}

// reportMetrics gets metrics and run postMetric function
func reportMetrics(m *Metrics, report *Report, c *Config) {
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
		chErr := make(chan error, len(allMetrics))

		ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
		for _, metric := range allMetrics {
			go postMetric(ctx, metric, c, ch, chErr)
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled: %v", ctx.Err())
			case res := <-ch:
				log.Println(res)
			case err := <-chErr:
				log.Printf("Error: %v", err)
			}
		}
		cancel()
		time.Sleep(c.RealReportInterval)
	}
}

// postMetric push metrics to server
func postMetric(ctx context.Context, metric string, c *Config, ch chan string, chErr chan error) {
	uri := fmt.Sprintf("%s://%s/%s",
		ServerScheme,
		c.Address,
		metric)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		chErr <- err
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		chErr <- err
		return
	}
	defer resp.Body.Close()
	ch <- fmt.Sprintf("Sent %s: %d", metric, resp.StatusCode)
}
