package models

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func floatToPointer(value float64) *float64 {
	return &value
}

func intToPointer(value int64) *int64 {
	return &value
}

func TestConvertV1ToV2(t *testing.T) {
	type args struct {
		m *Metric
	}
	tests := []struct {
		args    args
		want    *MetricV2
		name    string
		wantErr bool
	}{
		{
			name: "gauge ok",
			args: args{
				m: &Metric{
					Kind:  MetricKindGauge,
					Name:  "gauge1",
					Value: "99.9",
				},
			},
			want: &MetricV2{
				ID:    "gauge1",
				MType: MetricKindGauge,
				Value: floatToPointer(99.9),
			},
			wantErr: false,
		},
		{
			name: "gauge error",
			args: args{
				m: &Metric{
					Kind:  MetricKindGauge,
					Name:  "gauge2",
					Value: "99.z",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "counter ok",
			args: args{
				m: &Metric{
					Kind:  MetricKindCounter,
					Name:  "counter1",
					Value: "99",
				},
			},
			want: &MetricV2{
				ID:    "counter1",
				MType: MetricKindCounter,
				Delta: intToPointer(99),
			},
			wantErr: false,
		},
		{
			name: "counter error",
			args: args{
				m: &Metric{
					Kind:  MetricKindCounter,
					Name:  "counter2",
					Value: "99z",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertV1ToV2(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertV1ToV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertV1ToV2() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertV2ToV1(t *testing.T) {
	type args struct {
		m *MetricV2
	}
	tests := []struct {
		args    args
		want    *Metric
		name    string
		wantErr bool
	}{
		{
			name: "gauge ok",
			args: args{
				m: &MetricV2{
					ID:    "gauge1",
					MType: MetricKindGauge,
					Value: floatToPointer(99.9),
				},
			},
			want: &Metric{
				Kind:  MetricKindGauge,
				Name:  "gauge1",
				Value: "99.9",
			},
			wantErr: false,
		},
		{
			name: "counter ok",
			args: args{
				m: &MetricV2{
					ID:    "counter1",
					MType: MetricKindCounter,
					Delta: intToPointer(99),
				},
			},
			want: &Metric{
				Kind:  MetricKindCounter,
				Name:  "counter1",
				Value: "99",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertV2ToV1(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertV2ToV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertV2ToV1() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPSMetricsCPU(t *testing.T) {
	type args struct {
		c []float64
	}
	tests := []struct {
		want map[string]float64
		name string
		args args
	}{
		{
			name: "cpu ok",
			args: args{
				c: []float64{
					10.0,
					20.0,
				},
			},
			want: map[string]float64{
				"CPUutilization0": 10.0,
				"CPUutilization1": 20.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PSMetricsCPU(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PSMetricsCPU() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuntimeMetricsMap(t *testing.T) {
	type args struct {
		r *runtime.MemStats
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "ok",
			args: args{
				r: &runtime.MemStats{
					Alloc: 1000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuntimeMetricsMap(tt.args.r)
			assert.Equal(t, got["Alloc"], uint64(1000))
		})
	}
}
