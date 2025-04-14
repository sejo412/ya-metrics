package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

var testCfg = config.AgentConfig{
	Address:            "",
	ReportInterval:     0,
	PollInterval:       0,
	PathStyle:          false,
	Key:                "",
	RateLimit:          0,
	RealReportInterval: 0,
	RealPollInterval:   0,
	Logger:             nil,
}

var testAgent = NewAgent(&testCfg)

func TestAgent_Poll(t *testing.T) {
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
	}
	tests := []struct {
		fields fields
		name   string
	}{
		{
			name: "Poll",
			fields: fields{
				Metrics: &metrics{
					gauge: gauge{
						memStats:    nil,
						randomValue: 0,
						psStats:     psStats{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			a.Poll()
		})
	}
}

func TestAgent_PollPS(t *testing.T) {
	m, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
	}
	tests := []struct {
		fields fields
		name   string
	}{
		{
			name: "PollPS",
			fields: fields{
				Metrics: &metrics{
					gauge: gauge{
						psStats: psStats{
							totalMemory:    float64(m.Total),
							freeMemory:     float64(m.Free),
							cpuUtilization: models.PSMetricsCPU(c),
						},
					},
				},
				Config: &testCfg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			a.PollPS()
		})
	}
}

func Test_reportToMetricsV2(t *testing.T) {
	type args struct {
		report *report
	}
	tests := []struct {
		name string
		args args
		want []models.MetricV2
	}{
		{
			name: "report to metrics v2",
			args: args{
				report: &report{
					gauge: map[string]float64{
						"testGauge1": 1.0,
					},
					counter: map[string]int64{
						"testCounter1": 1,
					},
				},
			},
			want: []models.MetricV2{
				{
					ID:    "testGauge1",
					MType: models.MetricKindGauge,
					Value: float64Ptr(1.0),
				},
				{
					ID:    "testCounter1",
					MType: models.MetricKindCounter,
					Delta: int64Ptr(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reportToMetricsV2(tt.args.report); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportToMetricsV2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_Report(t *testing.T) {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	m, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	address, _ := strings.CutPrefix(server.URL, "http://")
	logs, _ := logger.NewLogger()
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		fields fields
		args   args
		name   string
	}{
		{
			name: "Report batch",
			fields: fields{
				Metrics: &metrics{
					gauge: gauge{
						memStats: &ms,
						psStats: psStats{
							totalMemory:    float64(m.Total),
							freeMemory:     float64(m.Free),
							cpuUtilization: models.PSMetricsCPU(c),
						},
						randomValue: 99.99,
					},
					counter: counter{
						pollCount: 1,
					},
				},
				Config: &config.AgentConfig{
					Logger:    logs,
					Address:   address,
					PathStyle: false,
				},
			},
			args: args{
				ctx: context.Background(),
			},
		},
		{
			name: "Report batch by path style",
			fields: fields{
				Metrics: &metrics{
					gauge: gauge{
						memStats: &ms,
						psStats: psStats{
							totalMemory:    float64(m.Total),
							freeMemory:     float64(m.Free),
							cpuUtilization: models.PSMetricsCPU(c),
						},
						randomValue: 99.99,
					},
					counter: counter{
						pollCount: 1,
					},
				},
				Config: &config.AgentConfig{
					Logger:    logs,
					Address:   address,
					PathStyle: true,
					RateLimit: 2,
				},
			},
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			a.Report(tt.args.ctx)
		})
	}
}

func TestAgent_postMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	address, _ := strings.CutPrefix(server.URL, "http://")
	logs, _ := logger.NewLogger()
	agentConfig := &config.AgentConfig{
		Logger:    logs,
		Address:   address,
		PathStyle: false,
	}
	type fields struct {
		Config *config.AgentConfig
	}
	type args struct {
		ctx    context.Context
		metric string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "PostMetric counter ok",
			fields: fields{
				Config: agentConfig,
			},
			args: args{
				ctx:    context.Background(),
				metric: "updates/counter/poll/4",
			},
			wantErr: false,
		},
		{
			name: "PostMetric counter fail",
			fields: fields{
				Config: agentConfig,
			},
			args: args{
				ctx:    context.Background(),
				metric: "updates/counter/poll2/4xc",
			},
			wantErr: true,
		},
		{
			name: "PostMetric gauge ok",
			fields: fields{
				Config: agentConfig,
			},
			args: args{
				ctx:    context.Background(),
				metric: "updates/gauge/poll/42.011",
			},
			wantErr: false,
		},
		{
			name: "PostMetric gauge fail",
			fields: fields{
				Config: agentConfig,
			},
			args: args{
				ctx:    context.Background(),
				metric: "updates/gauge/poll/42.011x",
			},
			wantErr: true,
		},
		{
			name: "PostMetric unknown",
			fields: fields{
				Config: agentConfig,
			},
			args: args{
				ctx:    context.Background(),
				metric: "updates/zzzz/poll/42.011",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Config: tt.fields.Config,
			}
			if err := a.postMetric(tt.args.ctx, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("postMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
