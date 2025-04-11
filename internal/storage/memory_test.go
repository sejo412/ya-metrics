package storage

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"reflect"
	"strconv"
	"testing"

	"github.com/sejo412/ya-metrics/internal/models"
)

var memoryStorageTest = NewMemoryStorage()
var bufferTest = bytes.NewBuffer(nil)

func TestMemoryStorage_Upsert(t *testing.T) {
	type args struct {
		ctx    context.Context
		metric models.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "new gauge OK",
			args: args{
				metric: models.Metric{
					Kind:  models.MetricKindGauge,
					Name:  "testGauge1",
					Value: "9999.11",
				},
			},
			wantErr: false,
		},
		{
			name: "new counter OK",
			args: args{
				metric: models.Metric{
					Kind:  models.MetricKindCounter,
					Name:  "testCounter1",
					Value: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "update counter OK",
			args: args{
				metric: models.Metric{
					Kind:  models.MetricKindCounter,
					Name:  "testCounter1",
					Value: "2",
				},
			},
			wantErr: false,
		},
		{
			name: "update counter Error",
			args: args{
				metric: models.Metric{
					Kind:  models.MetricKindCounter,
					Name:  "testCounter1",
					Value: "preved",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := memoryStorageTest.Upsert(tt.args.ctx, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_Get(t *testing.T) {
	type args struct {
		ctx  context.Context
		kind string
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    models.Metric
		wantErr bool
	}{
		{
			name: "get gauge OK",
			args: args{
				kind: models.MetricKindGauge,
				name: "testGauge1",
			},
			want: models.Metric{
				Kind:  models.MetricKindGauge,
				Name:  "testGauge1",
				Value: "9999.11",
			},
			wantErr: false,
		},
		{
			name: "get gauge Error",
			args: args{
				kind: models.MetricKindGauge,
				name: "testGauge2",
			},
			want:    models.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := memoryStorageTest.Get(tt.args.ctx, tt.args.kind, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStorage_MassUpsert(t *testing.T) {
	type args struct {
		ctx     context.Context
		metrics []models.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "mass upsert OK",
			args: args{
				metrics: []models.Metric{
					{
						Kind:  models.MetricKindGauge,
						Name:  "testGauge2",
						Value: "9999.22",
					},
					{
						Kind:  models.MetricKindCounter,
						Name:  "testCounter2",
						Value: "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mass upsert Error",
			args: args{
				metrics: []models.Metric{
					{
						Kind:  models.MetricKindGauge,
						Name:  "testGauge3",
						Value: "9999.33",
					},
					{
						Kind:  models.MetricKindCounter,
						Name:  "testCounter2",
						Value: "preved",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := memoryStorageTest.MassUpsert(tt.args.ctx, tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("MassUpsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_GetAll(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []models.Metric
		wantErr bool
	}{
		{
			name: "get all OK",
			want: []models.Metric{
				{
					Kind:  models.MetricKindCounter,
					Name:  "testCounter1",
					Value: "3",
				},
				{
					Kind:  models.MetricKindCounter,
					Name:  "testCounter2",
					Value: "1",
				},
				{
					Kind:  models.MetricKindGauge,
					Name:  "testGauge1",
					Value: "9999.11",
				},
				{
					Kind:  models.MetricKindGauge,
					Name:  "testGauge2",
					Value: "9999.22",
				},
				{
					Kind:  models.MetricKindGauge,
					Name:  "testGauge3",
					Value: "9999.33",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := memoryStorageTest.GetAll(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStorage_Flush(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantDst string
		wantErr bool
	}{
		{
			name: "flush OK",
			wantDst: `{"delta":3,"id":"testCounter1","type":"counter"}
{"delta":1,"id":"testCounter2","type":"counter"}
{"value":9999.11,"id":"testGauge1","type":"gauge"}
{"value":9999.22,"id":"testGauge2","type":"gauge"}
{"value":9999.33,"id":"testGauge3","type":"gauge"}
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := memoryStorageTest.Flush(tt.args.ctx, bufferTest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Flush() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDst := bufferTest.String(); gotDst != tt.wantDst {
				t.Errorf("Flush() gotDst = %v, want %v", gotDst, tt.wantDst)
			}
		})
	}
}

func TestMemoryStorage_Load(t *testing.T) {
	var testError = []byte("zzz")
	type args struct {
		ctx context.Context
		src io.Reader
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "load OK",
			args: args{
				src: bufferTest,
			},
			wantErr: false,
		},
		{
			name: "load Error",
			args: args{
				src: bytes.NewReader(testError),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := memoryStorageTest.Load(tt.args.ctx, tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var benchStorage *MemoryStorage

func BenchmarkMemoryStorage_MassUpsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tmpSlice := genMetrics(models.TotalCountMetrics)
		b.StartTimer()
		benchStorage = NewMemoryStorage()
		_ = benchStorage.MassUpsert(context.Background(), tmpSlice)
	}
}

func BenchmarkMemoryStorage_GetAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = benchStorage.GetAll(context.Background())
	}
}

func genMetrics(count int) []models.Metric {
	metrics := make([]models.Metric, count)
	for i := 0; i < count; i++ {
		metrics[i] = models.Metric{
			Kind:  models.MetricKindGauge,
			Name:  "testGauge" + strconv.Itoa(i),
			Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
		}
	}
	return metrics
}

func TestMemoryStorage_Init(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "init OK",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{}
			if err := s.Init(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_Open(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "open OK",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{}
			if err := s.Open(tt.args.ctx, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ping OK",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{}
			if err := s.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "close OK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{}
			s.metrics = make(map[string]models.Metric)
			s.metrics["testGauge1"] = models.Metric{
				Kind:  models.MetricKindGauge,
				Name:  "testGauge1",
				Value: "9999.11",
			}
			s.Close()
			if s.metrics != nil {
				t.Errorf("Close() metrics = %v", s.metrics)
			}
		})
	}
}
