package agent

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/sejo412/ya-metrics/internal/config"
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
		name   string
		fields fields
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
		name   string
		fields fields
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

func TestAgent_Sign(t *testing.T) {
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
	}
	type args struct {
		body *[]byte
		r    *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			a.Sign(tt.args.body, tt.args.r)
		})
	}
}

func TestAgent_postMetric(t *testing.T) {
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			if err := a.postMetric(tt.args.ctx, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("postMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_postMetricByPath(t *testing.T) {
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			if err := a.postMetricByPath(tt.args.ctx, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("postMetricByPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_postMetricsBatch(t *testing.T) {
	type fields struct {
		Metrics *metrics
		Config  *config.AgentConfig
	}
	type args struct {
		ctx    context.Context
		report *report
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				Metrics: tt.fields.Metrics,
				Config:  tt.fields.Config,
			}
			if err := a.postMetricsBatch(tt.args.ctx, tt.args.report); (err != nil) != tt.wantErr {
				t.Errorf("postMetricsBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
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
