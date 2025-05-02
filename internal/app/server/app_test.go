package server

import (
	"context"
	"errors"
	"testing"

	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/stretchr/testify/assert"

	"github.com/sejo412/ya-metrics/internal/models"
)

func TestParsePostUpdateRequestJSON(t *testing.T) {
	v := 100.99
	type args struct {
		request []byte
	}
	tests := []struct {
		want    models.MetricV2
		wantErr error
		name    string
		args    args
	}{
		{
			name: "Valid JSON",
			args: args{
				request: []byte(`{"id": "testMetric", "type": "gauge", "value": 100.99}`),
			},
			want: models.MetricV2{
				ID:    "testMetric",
				MType: "gauge",
				Value: &v,
			},
			wantErr: nil,
		},
		{
			name: "Not JSON",
			args: args{
				request: []byte(`preved`),
			},
			wantErr: errors.New(models.MessageBadRequest),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePostRequestJSON(tt.args.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAllMetricValues(t *testing.T) {
	tests := []struct {
		want map[string]string
		name string
	}{
		{
			name: "Get all metric values",
			want: map[string]string{
				"test1": "12",
				"test2": "15",
			},
		},
	}
	store := storage.NewMemoryStorage()
	_ = store.Upsert(context.Background(), models.Metric{
		Kind:  "gauge",
		Name:  "test1",
		Value: "12",
	})
	_ = store.Upsert(context.Background(), models.Metric{
		Kind:  "gauge",
		Name:  "test2",
		Value: "15",
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetAllMetricValues(store), "GetAllMetricValues(%v)", store)
		})
	}
}
