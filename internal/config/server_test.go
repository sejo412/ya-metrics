package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerConfig_Load(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		config  string
		want    ServerConfig
		wantErr bool
	}{
		{
			name: "Default values",
			args: []string{},
			want: ServerConfig{
				Address: DefaultAddress,
				Restore: boolPtr(DefaultRestore),
			},
		},
		{
			name: "Override via CLI flags",
			args: []string{
				"-a=127.0.0.1:9090",
				"-r=false",
				"-i=2",
				"-f=/tmp/db",
				"-d=memory://",
				"--crypto-key=/tmp/11.pem",
				"--key=123",
			},
			want: ServerConfig{
				Address:       "127.0.0.1:9090",
				StoreInterval: 2,
				StoreFile:     "/tmp/db",
				DatabaseDSN:   "memory://",
				Key:           "123",
				CryptoKey:     "/tmp/11.pem",
				Restore:       boolPtr(false),
			},
		},
		{
			name: "Override via ENV",
			env:  map[string]string{"ADDRESS": "0.0.0.0:80", "RESTORE": "false"},
			want: ServerConfig{
				Address: "0.0.0.0:80",
				Restore: boolPtr(false),
			},
		},
		{
			name:   "Override via config file",
			config: `{"address": "10.0.0.1:443", "restore": true}`,
			want: ServerConfig{
				Address: "10.0.0.1:443",
				Restore: boolPtr(true),
			},
		},
		{
			name:   "CLI flags override config file",
			args:   []string{"-a=localhost:3000", "-c=test.json"},
			config: `{"address": "ignored", "restore": false}`,
			want: ServerConfig{
				Address: "localhost:3000",
				Restore: boolPtr(false),
			},
		},
		{
			name:    "error read config file",
			args:    []string{"-a=localhost:3000", "-c=test.json"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.env {
					_ = os.Unsetenv(k)
				}
			}()

			if tt.config != "" {
				tmpFile, err := os.Create("test.json")
				require.NoError(t, err)
				defer func() {
					_ = os.Remove(tmpFile.Name())
				}()

				_, err = tmpFile.WriteString(tt.config)
				require.NoError(t, err)
				_ = tmpFile.Close()

				hasConfigFlag := false
				for _, arg := range tt.args {
					if strings.HasPrefix(arg, "-c") || strings.HasPrefix(arg, "--config") {
						hasConfigFlag = true
						break
					}
				}
				if !hasConfigFlag {
					tt.args = append(tt.args, "-c="+tmpFile.Name())
				}
			}

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = append([]string{"test"}, tt.args...)

			cfg := NewServerConfig()
			err := cfg.Load()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want.Address, cfg.Address)
			if tt.want.Restore != nil {
				require.NotNil(t, cfg.Restore)
				require.Equal(t, *tt.want.Restore, *cfg.Restore)
			} else {
				require.Nil(t, cfg.Restore)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
