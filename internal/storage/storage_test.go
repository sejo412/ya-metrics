package storage

import (
	"reflect"
	"testing"
)

func TestParseDSN(t *testing.T) {
	type args struct {
		dsn string
	}
	tests := []struct {
		name     string
		args     args
		wantOpts Options
		wantErr  bool
	}{
		{
			name: "valid DSN postgres",
			args: args{
				dsn: "postgres://user:password@host.com:5432/dbname",
			},
			wantOpts: Options{
				Scheme:   "postgres",
				Username: "user",
				Password: "password",
				Host:     "host.com",
				Port:     5432,
				Database: "dbname",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "invalid DSN postgres (sslmode)",
			args: args{
				dsn: "postgres://user:password@host.com:5432/dbname?sslmode=preved",
			},
			wantOpts: Options{},
			wantErr:  true,
		},
		{
			name: "invalid DSN postgres (port)",
			args: args{
				dsn: "postgres://user:password@host.com:5432z/dbname?sslmode=preved",
			},
			wantOpts: Options{},
			wantErr:  true,
		},
		{
			name: "valid DSN memory",
			args: args{
				dsn: "memory:",
			},
			wantOpts: Options{
				Scheme: "memory",
			},
			wantErr: false,
		},
		{
			name: "unknown scheme",
			args: args{
				dsn: "mysql://user:password@host.com:5432/dbname",
			},
			wantOpts: Options{},
			wantErr:  true,
		},
		{
			name: "invalid DSN",
			args: args{
				dsn: "qweqwe:asdasd",
			},
			wantOpts: Options{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpts, err := ParseDSN(tt.args.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDSN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOpts, tt.wantOpts) {
				t.Errorf("ParseDSN() gotOpts = %v, want %v", gotOpts, tt.wantOpts)
			}
		})
	}
}
