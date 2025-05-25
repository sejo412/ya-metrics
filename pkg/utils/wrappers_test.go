package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
)

func TestWithRetry(t *testing.T) {
	type args struct {
		ctx context.Context
		f   func(ctx context.Context) error
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				f: func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "not retryable error",
			args: args{
				ctx: context.Background(),
				f: func(ctx context.Context) error {
					return errors.New("error")
				},
			},
			wantErr: true,
		},
		{
			name: "retryable error",
			args: args{
				ctx: context.Background(),
				f: func(ctx context.Context) error {
					return models.ErrRetryable[0]
				},
			},
			wantErr: true,
		},
	}
	logs, _ := logger.MustNewLogger(false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WithRetry(tt.args.ctx, logs, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("WithRetry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
