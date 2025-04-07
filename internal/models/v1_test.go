package models

import "testing"

func TestGetMetricValueString(t *testing.T) {
	type args struct {
		metric Metric
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "gauge metric to string OK",
			args: args{
				metric: Metric{
					Kind:  MetricKindGauge,
					Name:  "testGauge1",
					Value: "99.9",
				},
			},
			want:    "99.9",
			wantErr: false,
		},
		{
			name: "gauge metric to string ERROR",
			args: args{
				metric: Metric{
					Kind:  MetricKindGauge,
					Name:  "testGauge2",
					Value: "99.z",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "counter metric to string OK",
			args: args{
				metric: Metric{
					Kind:  MetricKindCounter,
					Name:  "testCounter1",
					Value: "99",
				},
			},
			want:    "99",
			wantErr: false,
		},
		{
			name: "counter metric to string ERROR",
			args: args{
				metric: Metric{
					Kind:  MetricKindCounter,
					Name:  "testCounter2",
					Value: "99.33",
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetricValueString(tt.args.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetricValueString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMetricValueString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundFloatToString(t *testing.T) {
	type args struct {
		val float64
	}
	tests := []struct {
		name string
		want string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundFloatToString(tt.args.val); got != tt.want {
				t.Errorf("RoundFloatToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
