package app

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"time"

	"github.com/sejo412/ya-metrics/internal/models"
)

const maxRand = 10000

// PollMetrics collects runtime metrics in infinite loop
func PollMetrics(m *Metrics, interval time.Duration) {
	for {
		cryptoRand, _ := rand.Int(rand.Reader, big.NewInt(maxRand))
		mem := &runtime.MemStats{}
		runtime.ReadMemStats(mem)
		m.Gauge.MemStats = mem
		m.Gauge.RandomValue = float64(cryptoRand.Uint64())
		m.Counter.PollCount = 1
		time.Sleep(interval)
	}
}

// ReportMetrics gets metrics and run postMetric function
func ReportMetrics(m *Metrics, report *Report, address string, interval, timeout time.Duration) {
	for {
		// skip if function start before polling
		if m.Counter.PollCount == 0 {
			continue
		}

		runtimeMetrics := models.RuntimeMetricsMap(m.Gauge.MemStats)
		for key, value := range runtimeMetrics {
			switch v := value.(type) {
			case uint64:
				report.Gauge[key] = float64(v)
			case uint32:
				report.Gauge[key] = float64(v)
			case float64:
				report.Gauge[key] = v
			}
		}
		report.Gauge[models.MetricNameRandomValue] = m.Gauge.RandomValue
		report.Counter[models.MetricNamePollCount] = m.Counter.PollCount

		var allMetrics []string
		for name, value := range report.Gauge {
			mpath := models.MetricPathPostPrefix + "/" + models.MetricKindGauge + "/" + name
			allMetrics = append(allMetrics, fmt.Sprintf("%s/%v", mpath, value))
		}
		for name, value := range report.Counter {
			mpath := models.MetricPathPostPrefix + "/" + models.MetricKindCounter + "/" + name
			allMetrics = append(allMetrics, fmt.Sprintf("%s/%v", mpath, value))
		}

		ch := make(chan string, len(allMetrics))
		chErr := make(chan error, len(allMetrics))

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		for _, metric := range allMetrics {
			go postMetric(ctx, metric, address, ch, chErr)
			select {
			case <-ctx.Done():
				log.Printf("Context canceled: %v", ctx.Err())
			case res := <-ch:
				log.Println(res)
			case err := <-chErr:
				log.Printf("Error: %v", err)
			}
		}
		cancel()
		time.Sleep(interval)
	}
}

// postMetric push metrics to server
func postMetric(ctx context.Context, metric, address string, ch chan string, chErr chan error) {
	uri := address + "/" + metric

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, http.NoBody)
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
