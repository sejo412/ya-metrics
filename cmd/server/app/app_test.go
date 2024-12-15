package app

import (
	"github.com/sejo412/ya-metrics/internal/models"
	"reflect"
	"testing"
)

func TestParsePostUpdateRequestJSON(t *testing.T) {
	type args struct {
		request []byte
	}
	tests := []struct {
		name    string
		args    args
		want    models.MetricV2
		wantErr bool
	}{
		{
			name: "Valid JSON",
			args: args{
				request: []byte(`{"id": "testMetric", "type": "gauge", "value": 100.99}`),
			},
			want: models.MetricV2{
				ID:    "testMetric",
				MType: "gauge",
				Value: 100.99,
			},
			wantErr: false,
		},
		{
			name: "Invalid JSON",
			args: args{
				request: []byte(`{"value": "invalid"}`),
			},
			wantErr: true,
		},
		{
			name: "Not JSON",
			args: args{
				request: []byte(`preved`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePostRequestJSON(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePostRequestJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePostRequestJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}
