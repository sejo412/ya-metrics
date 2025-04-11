//go:build integration

package storage

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/sejo412/ya-metrics/internal/models"
)

var testDB = NewPostgresStorage()

func TestMain(m *testing.M) {
	fmt.Println("init integration tests")
	/*
	   POSTGRES_USER: metrics
	   POSTGRES_PASSWORD: secret
	   POSTGRES_DB: metrics
	*/
	opts := Options{
		Scheme:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "metrics",
		Password: "secret",
		Database: "metrics",
	}
	_, ok := os.LookupEnv("GITHUB_ACTIONS")
	if ok {
		opts.Host = "postgres"
	} else {
		opts.Port = 15432
	}
	if err := testDB.Open(context.Background(), opts); err != nil {
		panic(err)
	}
	defer testDB.Close()
	exitVal := m.Run()
	fmt.Println("clear integration tests data")
	if err := testDB.Open(context.Background(), opts); err != nil {
		fmt.Println(err)
	}
	if _, err := testDB.Client.Exec("DELETE FROM metric_mapping WHERE name LIKE 'test%'"); err != nil {
		fmt.Println(err)
	}
	os.Exit(exitVal)
}

func TestPostgresStorage_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ping",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testDB.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresStorage_Init(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "init",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testDB.Init(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresStorage_Load(t *testing.T) {
	type args struct {
		ctx context.Context
		src io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "load",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PostgresStorage{}
			if err := p.Load(tt.args.ctx, tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresStorage_Upsert(t *testing.T) {
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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
			if err := testDB.Upsert(tt.args.ctx, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresStorage_Get(t *testing.T) {
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
				ctx:  context.Background(),
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
			name: "get counter OK",
			args: args{
				ctx:  context.Background(),
				kind: models.MetricKindCounter,
				name: "testCounter1",
			},
			want: models.Metric{
				Kind:  models.MetricKindCounter,
				Name:  "testCounter1",
				Value: "3",
			},
		},
		{
			name: "get gauge Error",
			args: args{
				ctx:  context.Background(),
				kind: models.MetricKindGauge,
				name: "testGauge2222",
			},
			want:    models.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testDB.Get(tt.args.ctx, tt.args.kind, tt.args.name)
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

func TestPostgresStorage_MassUpsert(t *testing.T) {
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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
			if err := testDB.MassUpsert(tt.args.ctx, tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("MassUpsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresStorage_GetAll(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "get all OK",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.GetAll(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPostgresStorage_Flush(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PostgresStorage{}
			dst := &bytes.Buffer{}
			err := p.Flush(tt.args.ctx, dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Flush() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPostgresStorage_Close(t *testing.T) {
	opts := Options{
		Scheme:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "metrics",
		Password: "secret",
		Database: "metrics",
	}
	_, ok := os.LookupEnv("GITHUB_ACTIONS")
	if ok {
		opts.Host = "postgres"
	} else {
		opts.Port = 15432
	}
	if err := testDB.Open(context.Background(), opts); err != nil {
		panic(err)
	}
	type fields struct {
		Client *sql.DB
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "close OK",
			fields: fields{
				Client: testDB.Client,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PostgresStorage{
				Client: tt.fields.Client,
			}
			p.Close()
		})
	}
}
