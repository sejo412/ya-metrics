package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/pkg/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	maxRand = 10000
	baseInt = 10
)

// NewAgent returns new Agent object.
func NewAgent(cfg *config.AgentConfig) *Agent {
	m := &metrics{
		gauge: gauge{
			memStats:    nil,
			randomValue: 0,
		},
		counter: counter{
			pollCount: 0,
		},
	}
	return &Agent{
		Metrics: m,
		Config:  cfg,
	}
}

// Run starts agent application.
func (a *Agent) Run(ctx context.Context) error {
	if a.Config.CryptoKey != "" {
		k, err := os.ReadFile(a.Config.CryptoKey)
		if err != nil {
			return fmt.Errorf("error read crypto key: %w", err)
		}
		a.PublicKey, err = utils.LoadRSAPublicKey(k)
		if err != nil {
			return fmt.Errorf("error loading public key: %w", err)
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, config.GracefulSignals...)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	log := a.Config.Logger
	var wg sync.WaitGroup
	go func() {
		<-sigs
		log.Info("shutting down...")
		cancel()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(a.Config.RealPollInterval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				a.Poll()
				timer.Reset(a.Config.RealPollInterval)
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(a.Config.RealPollInterval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				a.PollPS()
				timer.Reset(a.Config.RealPollInterval)
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(a.Config.RealReportInterval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				a.Metrics.mutex.Lock()
				pollCount := a.Metrics.counter.pollCount
				a.Metrics.mutex.Unlock()
				if pollCount > 0 {
					a.Report(ctx)
				}
				timer.Reset(a.Config.RealReportInterval)
			}
		}
	}()
	wg.Wait()
	if a.Metrics.counter.pollCount > 0 {
		// We don't want waiting sends report with retries if server not reachable
		timeoutCtx, cncl := context.WithTimeout(context.Background(), config.GracefulTimeout)
		defer cncl()
		a.Report(timeoutCtx)
	}
	log.Info("shutdown complete")
	return nil
}

// Poll collects runtime metrics.
func (a *Agent) Poll() {
	cryptoRand, _ := rand.Int(rand.Reader, big.NewInt(maxRand))
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	a.Metrics.mutex.Lock()
	a.Metrics.gauge.memStats = m
	a.Metrics.gauge.randomValue = float64(cryptoRand.Uint64())
	a.Metrics.counter.pollCount = 1
	a.Metrics.mutex.Unlock()
}

// PollPS collects ps metrics.
func (a *Agent) PollPS() {
	log := a.Config.Logger
	m, err := mem.VirtualMemory()
	a.Metrics.mutex.Lock()
	defer a.Metrics.mutex.Unlock()
	if err != nil {
		log.Error("failed to get memory info")
	} else {
		a.Metrics.gauge.psStats.totalMemory = float64(m.Total)
		a.Metrics.gauge.psStats.freeMemory = float64(m.Free)
	}
	c, err := cpu.Percent(0, true)
	if err != nil {
		log.Error("failed to get cpu info")
	} else {
		a.Metrics.gauge.psStats.cpuUtilization = models.PSMetricsCPU(c)
	}
}

// Report gets metrics and run postMetricByPath function.
func (a *Agent) Report(ctx context.Context) {
	var err error
	log := a.Config.Logger
	report := new(report)
	report.gauge = make(map[string]float64)
	report.counter = make(map[string]int64)
	a.Metrics.mutex.Lock()
	runtimeMetrics := models.RuntimeMetricsMap(a.Metrics.gauge.memStats)
	a.Metrics.mutex.Unlock()
	for key, value := range runtimeMetrics {
		switch v := value.(type) {
		case uint64:
			report.gauge[key] = float64(v)
		case uint32:
			report.gauge[key] = float64(v)
		case float64:
			report.gauge[key] = v
		}
	}
	a.Metrics.mutex.Lock()
	report.gauge[models.MetricNameRandomValue] = a.Metrics.gauge.randomValue
	report.counter[models.MetricNamePollCount] = a.Metrics.counter.pollCount
	report.gauge[models.MetricNameTotalMemory] = a.Metrics.gauge.psStats.totalMemory
	report.gauge[models.MetricNameFreeMemory] = a.Metrics.gauge.psStats.freeMemory
	for core, value := range a.Metrics.gauge.psStats.cpuUtilization {
		report.gauge[core] = value
	}
	a.Metrics.mutex.Unlock()
	// Try post batch
	if !a.Config.PathStyle {
		err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
			defer cancel()
			return a.postMetricsBatch(ctx, report)
		})
		if err == nil {
			return
		}
		log.Errorf("failed to post batch: %v", err)
	}

	// Try post without batch
	lenMetrics := len(report.gauge) + len(report.counter)
	metricsChan := make(chan string, lenMetrics)
	for name, value := range report.gauge {
		mpath := models.MetricPathPostPrefix + "/" + models.MetricKindGauge + "/" + name
		metricsChan <- fmt.Sprintf("%s/%v", mpath, value)
	}
	for name, value := range report.counter {
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
				if a.Config.PathStyle {
					err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
						return a.postMetricByPath(ctx, metric)
					})
				} else {
					err = utils.WithRetry(ctx, log, func(ctx context.Context) error {
						return a.postMetric(ctx, metric)
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

func (a *Agent) Encrypt(body *[]byte) ([]byte, error) {
	if a.PublicKey == nil {
		return *body, nil
	}
	encrypted, err := utils.Encode(*body, a.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt body: %w", err)
	}
	return encrypted, nil
}

// Sign signs data with key.
func (a *Agent) Sign(body *[]byte, r *http.Request) {
	if a.Config.Key == "" {
		return
	}
	hash := utils.Hash(*body, a.Config.Key)
	r.Header.Set(models.HTTPHeaderSign, hash)
}

// postMetricByPath push metrics to server.
func (a *Agent) postMetricByPath(ctx context.Context, metric string) error {
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

func (a *Agent) postMetric(ctx context.Context, metric string) error {
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
	data, err := a.Encrypt(&gziped)
	if err != nil {
		return fmt.Errorf("failed to encrypt metrics: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address+"/"+models.MetricPathPostPrefix+"/",
		bytes.NewBuffer(data))
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

func (a *Agent) postMetricsBatch(ctx context.Context, report *report) error {
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
	data, err := a.Encrypt(&gziped)
	if err != nil {
		return fmt.Errorf("failed to encrypt metrics: %w", err)
	}
	uri := address + "/" + models.MetricPathPostsPrefix + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri,
		bytes.NewBuffer(data))
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

func reportToMetricsV2(report *report) []models.MetricV2 {
	metrics := make([]models.MetricV2, 0, len(report.gauge)+len(report.counter))
	for name, value := range report.gauge {
		metric := models.MetricV2{
			ID:    name,
			MType: models.MetricKindGauge,
			Value: &value,
		}
		metrics = append(metrics, metric)
	}
	for name, value := range report.counter {
		metric := models.MetricV2{
			ID:    name,
			MType: models.MetricKindCounter,
			Delta: &value,
		}
		metrics = append(metrics, metric)
	}
	return metrics
}
