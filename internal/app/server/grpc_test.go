package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/sejo412/ya-metrics/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestMain(m *testing.M) {
	st := storage.NewMemoryStorage()
	_ = st.MassUpsert(context.Background(), []models.Metric{
		{
			Kind:  "gauge",
			Name:  "testMetric1",
			Value: "99.9",
		},
		{
			Kind:  "gauge",
			Name:  "testMetric2",
			Value: "500",
		},
	})
	logs := logger.MustNewLogger(false)
	opts := &config.Options{
		Storage: st,
		Logger:  *logs,
		Config:  cfg,
	}
	go func() {
		err := StartServer(context.Background(), opts)
		if err != nil {
			panic(err)
		}
	}()
	exitVal := m.Run()
	os.Exit(exitVal)

}

func testGRPCClient() proto.MetricsClient {
	client, err := grpc.NewClient("127.0.0.1:3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := proto.NewMetricsClient(client)
	return c
}

func TestGRPCServer_GetMetric(t *testing.T) {
	testKind1 := new(proto.MType)
	*testKind1 = proto.MType_GAUGE
	testName1 := new(string)
	*testName1 = "testMetric1"
	testKind2 := new(proto.MType)
	*testKind2 = proto.MType_GAUGE
	testName2 := new(string)
	*testName2 = "testMetric200"
	type args struct {
		ctx context.Context
		in  *proto.GetMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		want    models.Metric
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				in: &proto.GetMetricRequest{
					Kind: testKind1,
					Id:   testName1,
				},
			},
			want: models.Metric{
				Kind:  "gauge",
				Name:  "testMetric1",
				Value: "99.9",
			},
			wantErr: false,
		},
	}
	client := testGRPCClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetMetric(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			m := resp.GetMetric()
			got := models.ConvertPbToV1(m)
			assert.Equalf(t, tt.want, got, "GetMetric(%v, %v)", tt.args.ctx, tt.args.in)
		})
	}
}

func TestGRPCServer_GetMetrics(t *testing.T) {
	type args struct {
		ctx context.Context
		in1 *emptypb.Empty
	}
	tests := []struct {
		args    args
		wantErr assert.ErrorAssertionFunc
		name    string
		want    []models.Metric
	}{
		{
			name: "success",
			want: []models.Metric{
				{
					Kind:  "gauge",
					Name:  "testMetric1",
					Value: "99.9",
				},
				{
					Kind:  "gauge",
					Name:  "testMetric2",
					Value: "500",
				},
			},
			args: args{
				ctx: context.Background(),
				in1: &emptypb.Empty{},
			},
			wantErr: assert.NoError,
		},
	}
	client := testGRPCClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetMetrics(tt.args.ctx, tt.args.in1)
			if !tt.wantErr(t, err, fmt.Sprintf("GetMetrics(%v, %v)", tt.args.ctx, tt.args.in1)) {
				return
			}
			m := resp.GetMetrics()
			got := models.ConvertPbsToV1s(m)
			assert.Equalf(t, tt.want, got, "GetMetrics(%v, %v)", tt.args.ctx, tt.args.in1)
		})
	}
}
