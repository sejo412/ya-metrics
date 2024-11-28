package server

import (
	"github.com/sejo412/ya-metrics/internal/storage"
	"reflect"
	"testing"
)

func TestCheckMetricType(t *testing.T) {
	type args struct {
		metricType  string
		metricValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckMetricType(tt.args.metricType, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("CheckMetricType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMetricSum(t *testing.T) {
	type args struct {
		store  *storage.MemoryStorage
		metric storage.Metric
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetricSum(tt.args.store, tt.args.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetricSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetricSum() got = %v, want %v", got, tt.want)
			}
		})
	}
}
