package server

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sejo412/ya-metrics/internal/models"
)

func TestParsePostUpdateRequestJSON(t *testing.T) {
	v := 100.99
	type args struct {
		request []byte
	}
	tests := []struct {
		name    string
		args    args
		want    models.MetricV2
		wantErr error
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
