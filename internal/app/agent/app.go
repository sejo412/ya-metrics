package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sejo412/ya-metrics/internal/utils"

	"github.com/sejo412/ya-metrics/internal/models"
)

const (
	maxRand = 10000
	baseInt = 10
)

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
func ReportMetrics(m *Metrics, report *Report, address string, interval, timeout time.Duration, oldAPI bool) {
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

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		// Try post batch and loops if success
		if !oldAPI {
			err := utils.WithRetry(ctx, func(ctx context.Context) error {
				return postBatchMetricV2(ctx, report, address)
			})
			if err != nil {
				log.Printf("failed to post batch: %v", err)
			} else {
				cancel()
				time.Sleep(interval)
				continue
			}
		}

		// Try post without batch
		var allMetrics []string
		for name, value := range report.Gauge {
			mpath := models.MetricPathPostPrefix + "/" + models.MetricKindGauge + "/" + name
			allMetrics = append(allMetrics, fmt.Sprintf("%s/%v", mpath, value))
		}
		for name, value := range report.Counter {
			mpath := models.MetricPathPostPrefix + "/" + models.MetricKindCounter + "/" + name
			allMetrics = append(allMetrics, fmt.Sprintf("%s/%v", mpath, value))
		}

		// old (v2 without batch) and old-old (v1) api
		var wg sync.WaitGroup
		errChan := make(chan error, len(allMetrics))
		for _, metric := range allMetrics {
			wg.Add(1)
			go func(metric string) {
				defer wg.Done()
				var err error
				if oldAPI {
					err = utils.WithRetry(ctx, func(ctx context.Context) error {
						return postMetric(ctx, metric, address)
					})
				} else {
					err = utils.WithRetry(ctx, func(ctx context.Context) error {
						return postMetricV2(ctx, metric, address)
					})
				}
				if err != nil {
					errChan <- err
				}
			}(metric)
		}
		wg.Wait()
		close(errChan)
		cancel()
		for err := range errChan {
			log.Printf("failed to post metrics: %v", err)
		}
		time.Sleep(interval)
	}
}

// postMetric push metrics to server
func postMetric(ctx context.Context, metric, address string) error {
	uri := address + "/" + metric

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, http.NoBody)
	if err != nil {
		log.Printf("failed build request: %v", err)
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed post request: %v", err)
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	log.Printf("Sent %s: %d", metric, resp.StatusCode)
	return nil
}

func postMetricV2(ctx context.Context, metric, address string) error {
	splitedMetric := strings.Split(metric, "/")
	m := models.MetricV2{
		ID:    splitedMetric[2],
		MType: splitedMetric[1],
	}
	switch m.MType {
	case "gauge":
		v, err := strconv.ParseFloat(splitedMetric[3], 64)
		if err != nil {
			log.Printf("failed to parse gauge value: %v", err)
			return nil
		}
		m.Value = &v
	case "counter":
		v, err := strconv.ParseInt(splitedMetric[3], baseInt, 64)
		if err != nil {
			log.Printf("failed to parse counter value: %v", err)
			return nil
		}
		m.Delta = &v
	default:
		log.Printf("unknown metric type: %s", m.MType)
		return nil
	}

	body, err := json.Marshal(m)
	if err != nil {
		log.Printf("failed to marshal metric: %v", err)
		return nil
	}

	gziped, err := utils.Compress(body)
	if err != nil {
		log.Printf("failed to compress metric: %v", err)
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address+"/"+models.MetricPathPostPrefix+"/",
		bytes.NewBuffer(gziped))
	if err != nil {
		log.Printf("failed build request: %v", err)
		return nil
	}
	req.Header.Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationJSON)
	req.Header.Set(models.HTTPHeaderContentEncoding, models.HTTPHeaderEncodingGzip)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed post request: %v", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	fmt.Printf("Sent %s: %d", string(body), resp.StatusCode)
	return nil
}

func postBatchMetricV2(ctx context.Context, report *Report, address string) error {
	metrics := reportToMetricsV2(report)
	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	gziped, err := utils.Compress(body)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}
	uri := address + "/" + models.MetricPathPostsPrefix + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri,
		bytes.NewBuffer(gziped))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationJSON)
	req.Header.Set(models.HTTPHeaderContentEncoding, models.HTTPHeaderEncodingGzip)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Printf("Sent %s: %d", string(body), resp.StatusCode)
	return nil
}

func reportToMetricsV2(report *Report) []models.MetricV2 {
	metrics := make([]models.MetricV2, 0, len(report.Gauge)+len(report.Counter))
	for name, value := range report.Gauge {
		metric := models.MetricV2{
			ID:    name,
			MType: models.MetricKindGauge,
			Value: &value,
		}
		metrics = append(metrics, metric)
	}
	for name, value := range report.Counter {
		metric := models.MetricV2{
			ID:    name,
			MType: models.MetricKindCounter,
			Delta: &value,
		}
		metrics = append(metrics, metric)
	}
	return metrics
}
