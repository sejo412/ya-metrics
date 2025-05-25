package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/pkg/utils"
	pb "github.com/sejo412/ya-metrics/proto"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
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
	log := a.Config.Logger.Logger
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
	log := a.Config.Logger.Logger
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

// Report gets metrics and send their via http or grpc.
func (a *Agent) Report(ctx context.Context) {
	logger := a.Config.Logger
	log := logger.Logger
	report := new(report)
	report.mutex.Lock()
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
	report.mutex.Unlock()

	// Try to send report via grpc and TODO fallback to http if error
	if config.ModeFromString(a.Config.Mode) == config.GRPCMode {
		err := utils.WithRetry(ctx, logger, func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
			defer cancel()
			return a.sendViaGRPC(ctx, reportToPbMetrics(report))
		})
		if err != nil {
			log.Error("failed to send via gRPC metrics: ", err)
		}
		return
	}

	// Try post batch
	if !a.Config.PathStyle {
		err := utils.WithRetry(ctx, logger, func(ctx context.Context) error {
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
	report.mutex.Lock()
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
	report.mutex.Unlock()
	close(metricsChan)

	errChan := make(chan error, lenMetrics)
	var wg sync.WaitGroup
	for w := 0; w < a.Config.RateLimit; w++ {
		wg.Add(1)
		go func(ch <-chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			var err error
			for metric := range ch {
				if a.Config.PathStyle {
					err = utils.WithRetry(ctx, logger, func(ctx context.Context) error {
						return a.postMetricByPath(ctx, metric)
					})
				} else {
					err = utils.WithRetry(ctx, logger, func(ctx context.Context) error {
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

// Encrypt body with public key
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
	log := a.Config.Logger.Logger
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
	log := a.Config.Logger.Logger
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
	a.setXRealIP(req)
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
	log := a.Config.Logger.Logger
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
	a.setXRealIP(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Infof("Sent %s: %d", string(body), resp.StatusCode)
	return nil
}

func reportToMetricsV2(report *report) []models.MetricV2 {
	report.mutex.Lock()
	defer report.mutex.Unlock()
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

func reportToPbMetrics(report *report) []*pb.Metric {
	report.mutex.Lock()
	defer report.mutex.Unlock()
	m := make([]*pb.Metric, 0, len(report.gauge)+len(report.counter))
	for name, value := range report.gauge {
		kind := pb.MType_GAUGE
		m = append(m, &pb.Metric{
			Type:  &kind,
			Id:    &name,
			Value: &value,
		})
	}
	for name, value := range report.counter {
		kind := pb.MType_COUNTER
		m = append(m, &pb.Metric{
			Type:  &kind,
			Id:    &name,
			Delta: &value,
		})
	}
	return m
}

// getOutboundIP determine outgoing IP from fake UDP request to server.
func (a *Agent) getOutboundIP() net.IP {
	log := a.Config.Logger.Logger
	conn, err := net.Dial("udp4", a.Config.Address)
	if err != nil {
		log.Warn("failed to dial server. Skipping")
		return nil
	}
	defer func() {
		_ = conn.Close()
	}()
	return conn.LocalAddr().(*net.UDPAddr).IP
}

func (a *Agent) setXRealIP(req *http.Request) {
	addr := a.getOutboundIP()
	if addr != nil {
		req.Header.Set("X-Real-IP", addr.String())
	}
}

func (a *Agent) grpcCallOptions(opts callOpts) []grpc.CallOption {
	res := make([]grpc.CallOption, 0)
	res = append(res, grpc.UseCompressor(gzip.Name))
	addr := a.getOutboundIP()
	if addr != nil {
		md := metadata.Pairs(realip.XRealIp, addr.String())
		res = append(res, grpc.Header(&md))
	}
	if opts.hash != "" {
		md := metadata.Pairs(models.HTTPHeaderSign, opts.hash)
		res = append(res, grpc.Header(&md))
	}
	return res
}

func (a *Agent) sendViaGRPC(ctx context.Context, metrics []*pb.Metric) error {
	log := a.Config.Logger.Logger
	client, err := grpc.NewClient(a.Config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()
	c := pb.NewMetricsClient(client)

	opts := newCallOpts()
	if a.Config.Key != "" {
		metricsBytes, er := json.Marshal(metrics)
		if er != nil {
			log.Errorw("marshal metrics", "error", er)
			return fmt.Errorf("error marshal metrics: %w", er)
		}
		opts.hash = utils.Hash(metricsBytes, a.Config.Key)
	}
	resp, err := c.SendMetrics(ctx, &pb.SendMetricsRequest{
		Metrics: metrics,
	}, a.grpcCallOptions(*opts)...)
	if err != nil || (resp != nil && resp.Error != nil) {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	log.Info("Sent via gRPC: ", metrics)
	return nil
}
