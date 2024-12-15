package app

import (
	"reflect"
	"testing"

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
				Value: &v,
			},
			wantErr: false,
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
