package models

import (
	"errors"
	"syscall"
	"testing"
)

func TestErrIsRetryable(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "error is retryable",
			args: args{
				err: syscall.ECONNRESET,
			},
			want: true,
		},
		{
			name: "error is not retryable",
			args: args{
				err: errors.New("some error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrIsRetryable(tt.args.err); got != tt.want {
				t.Errorf("ErrIsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}
