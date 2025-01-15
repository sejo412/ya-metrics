package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/utils"
)

const (
	maxRand = 10000
	baseInt = 10
)

func NewAgent(cfg *config.AgentConfig) *Agent {
	m := &Metrics{
		Gauge: Gauge{
			MemStats:    nil,
			RandomValue: 0,
		},
		Counter: Counter{
			PollCount: 0,
		},
	}
	return &Agent{
		Metrics: m,
		Config:  cfg,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			a.Poll()
			time.Sleep(a.Config.RealPollInterval)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// skip if function start before polling
			if a.Metrics.Counter.PollCount == 0 {
				continue
			}
			ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
			a.Report(ctx)
			cancel()
			time.Sleep(a.Config.RealReportInterval)
		}
	}()
	wg.Wait()
	return err
}

// Poll collects runtime metrics
func (a *Agent) Poll() {
	cryptoRand, _ := rand.Int(rand.Reader, big.NewInt(maxRand))
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	a.Metrics.Gauge.MemStats = mem
	a.Metrics.Gauge.RandomValue = float64(cryptoRand.Uint64())
	a.Metrics.Counter.PollCount = 1
}

// Report gets metrics and run postMetric function
func (a *Agent) Report(ctx context.Context) {
	var err error
	log := a.Config.Logger
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()
	report := new(Report)
	report.Gauge = make(map[string]float64)
	report.Counter = make(map[string]int64)
	runtimeMetrics := models.RuntimeMetricsMap(a.Metrics.Gauge.MemStats)
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
	report.Gauge[models.MetricNameRandomValue] = a.Metrics.Gauge.RandomValue
	report.Counter[models.MetricNamePollCount] = a.Metrics.Counter.PollCount

	address := fmt.Sprintf("%s://%s", config.ServerScheme, a.Config.Address)

	// Try post batch
	if !a.Config.UseOldAPI {
		err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
			return postBatchMetricV2(ctx, report, address, log)
		})
		if err == nil {
			return
		}
	}
	log.Errorf("failed to post batch: %v", err)

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
			if a.Config.UseOldAPI {
				err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
					return postMetric(ctx, metric, address, log)
				})
			} else {
				err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
					return postMetricV2(ctx, metric, address, log)
				})
			}
			if err != nil {
				errChan <- err
			}
		}(metric)
	}
	wg.Wait()
	close(errChan)
	for e := range errChan {
		log.Errorw("failed to post metrics",
			"error", e)
	}
}

// postMetric push metrics to server
func postMetric(ctx context.Context, metric, address string, log *logger.Logger) error {
	uri := address + "/" + metric

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed post request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	log.Infof("Sent %s: %d", metric, resp.StatusCode)
	return nil
}

func postMetricV2(ctx context.Context, metric, address string, log *logger.Logger) error {
	splitedMetric := strings.Split(metric, "/")
	m := models.MetricV2{
		ID:    splitedMetric[2],
		MType: splitedMetric[1],
	}
	switch m.MType {
	case "gauge":
		v, err := strconv.ParseFloat(splitedMetric[3], 64)
		if err != nil {
			return fmt.Errorf("failed parse gauge value: %w", err)
		}
		m.Value = &v
	case "counter":
		v, err := strconv.ParseInt(splitedMetric[3], baseInt, 64)
		if err != nil {
			return fmt.Errorf("failed parse counter value: %w", err)
		}
		m.Delta = &v
	default:
		return fmt.Errorf("unknown metric type: %s", m.MType)
	}

	body, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal metric: %w", err)
	}

	gziped, err := utils.Compress(body)
	if err != nil {
		return fmt.Errorf("failed compress metric: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address+"/"+models.MetricPathPostPrefix+"/",
		bytes.NewBuffer(gziped))
	if err != nil {
		return fmt.Errorf("failed build request: %w", err)
	}
	req.Header.Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationJSON)
	req.Header.Set(models.HTTPHeaderContentEncoding, models.HTTPHeaderEncodingGzip)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed post request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Infof("Sent %s: %d", string(body), resp.StatusCode)
	return nil
}

func postBatchMetricV2(ctx context.Context, report *Report, address string, log *logger.Logger) error {
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
	log.Infof("Sent %s: %d", string(body), resp.StatusCode)
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
