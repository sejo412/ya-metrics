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
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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
			a.PollPS()
			time.Sleep(a.Config.RealPollInterval)
		}
	}()
	time.Sleep(1 * time.Second)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// skip if function start before polling
			if a.Metrics.Counter.PollCount == 0 {
				continue
			}
			a.Report(ctx)
			time.Sleep(a.Config.RealReportInterval)
		}
	}()
	wg.Wait()
	return err
}

// Poll collects runtime metrics
func (a *Agent) Poll() {
	cryptoRand, _ := rand.Int(rand.Reader, big.NewInt(maxRand))
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	a.Metrics.Gauge.MemStats = m
	a.Metrics.Gauge.RandomValue = float64(cryptoRand.Uint64())
	a.Metrics.Counter.PollCount = 1
}

// PollPS collects ps metrics
func (a *Agent) PollPS() {
	log := a.Config.Logger
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Error("failed to get memory info")
	} else {
		a.Metrics.Gauge.PSStats.TotalMemory = float64(m.Total)
		a.Metrics.Gauge.PSStats.FreeMemory = float64(m.Free)
	}
	c, err := cpu.Percent(0, true)
	if err != nil {
		log.Error("failed to get cpu info")
	} else {
		a.Metrics.Gauge.PSStats.CPUUtilization = models.PSMetricsCPU(c)
	}
}

// Report gets metrics and run postMetric function
func (a *Agent) Report(ctx context.Context) {
	var err error
	log := a.Config.Logger
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
	report.Gauge[models.MetricNameTotalMemory] = a.Metrics.Gauge.PSStats.TotalMemory
	report.Gauge[models.MetricNameFreeMemory] = a.Metrics.Gauge.PSStats.FreeMemory
	for core, value := range a.Metrics.Gauge.PSStats.CPUUtilization {
		report.Gauge[core] = value
	}

	// Try post batch
	if !a.Config.UseOldAPI {
		err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
			defer cancel()
			return a.postBatchMetricV2(ctx, report)
		})
		if err == nil {
			return
		}
		log.Errorf("failed to post batch: %v", err)
	}

	// Try post without batch
	// old (v2 without batch) and old-old (v1) api
	lenMetrics := len(report.Gauge) + len(report.Counter)
	metricsChan := make(chan string, lenMetrics)
	for name, value := range report.Gauge {
		mpath := models.MetricPathPostPrefix + "/" + models.MetricKindGauge + "/" + name
		metricsChan <- fmt.Sprintf("%s/%v", mpath, value)
	}
	for name, value := range report.Counter {
		mpath := models.MetricPathPostPrefix + "/" + models.MetricKindCounter + "/" + name
		metricsChan <- fmt.Sprintf("%s/%v", mpath, value)
	}
	close(metricsChan)

	errChan := make(chan error, lenMetrics)
	var wg sync.WaitGroup
	for w := 0; w < a.Config.RateLimit; w++ {
		wg.Add(1)
		go func(ch <-chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for metric := range ch {
				if a.Config.UseOldAPI {
					err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
						return a.postMetric(ctx, metric)
					})
				} else {
					err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
						return a.postMetricV2(ctx, metric)
					})
				}
			}
			if err != nil {
				errChan <- err
			}
		}(metricsChan, &wg)
	}
	wg.Wait()
	close(errChan)
	for e := range errChan {
		log.Errorw("failed to post metrics",
			"error", e)
	}
}

func (a *Agent) Sign(body *[]byte, r *http.Request) {
	if a.Config.Key == "" {
		return
	}
	hash := utils.Hash(*body, a.Config.Key)
	r.Header.Set(models.HTTPHeaderSign, hash)
}

// postMetric push metrics to server
func (a *Agent) postMetric(ctx context.Context, metric string) error {
	address := fmt.Sprintf("%s://%s", config.ServerScheme, a.Config.Address)
	log := a.Config.Logger
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

func (a *Agent) postMetricV2(ctx context.Context, metric string) error {
	address := fmt.Sprintf("%s://%s", config.ServerScheme, a.Config.Address)
	log := a.Config.Logger
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
	a.Sign(&gziped, req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed post request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Infof("Sent %s: %d", string(body), resp.StatusCode)
	return nil
}

func (a *Agent) postBatchMetricV2(ctx context.Context, report *Report) error {
	address := fmt.Sprintf("%s://%s", config.ServerScheme, a.Config.Address)
	log := a.Config.Logger
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
	a.Sign(&gziped, req)
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
