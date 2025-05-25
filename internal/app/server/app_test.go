package server

import (
	"context"
	"errors"
	"net"
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

func Test_stringCIDRsToIPNets(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		want    []net.IPNet
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid IPv4 CIDRs",
			args: args{
				s: "192.168.1.0/24,127.0.0.1/8",
			},
			want: []net.IPNet{
				{
					IP:   []byte{192, 168, 1, 0},
					Mask: []byte{255, 255, 255, 0},
				},
				{
					IP:   []byte{127, 0, 0, 0},
					Mask: []byte{255, 0, 0, 0},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid cidr",
			args: args{
				s: "300.300.300.300/12",
			},
			want:    []net.IPNet{},
			wantErr: true,
		},
		{
			name: "first valid, second invalid",
			args: args{
				s: "192.168.1.0/24, 300.300.300.300/12",
			},
			want: []net.IPNet{
				{
					IP:   []byte{192, 168, 1, 0},
					Mask: []byte{255, 255, 255, 0},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringCIDRsToIPNets(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("stringCIDRsToIPNets() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isNetsContainsIP(t *testing.T) {
	nets := []net.IPNet{
		{
			IP:   []byte{192, 168, 1, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		{
			IP:   []byte{127, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
	}
	type args struct {
		nets []net.IPNet
		ip   string
	}
	tests := []struct {
		args args
		name string
		want bool
	}{
		{
			name: "contains ip",
			args: args{
				ip:   "192.168.1.200",
				nets: nets,
			},
			want: true,
		},
		{
			name: "not contains ip",
			args: args{
				ip:   "192.168.2.200",
				nets: nets,
			},
			want: false,
		},
		{
			name: "empty ip",
			args: args{
				ip:   "",
				nets: nets,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isNetsContainsIP(tt.args.ip, tt.args.nets), "isNetsContainsIP(%v, %v)",
				tt.args.ip,
				tt.args.nets)
		})
	}
}
