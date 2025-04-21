package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestAgentConfig_Load(t *testing.T) {
	oldArgs := os.Args
	oldEnv := os.Environ()
	defer func() {
		os.Args = oldArgs
		pflag.CommandLine = pflag.NewFlagSet("", pflag.ContinueOnError)
		for _, env := range oldEnv {
			_ = os.Setenv(env, "")
		}
	}()

	tests := []struct {
		env         map[string]string
		name        string
		configFile  string
		errContains string
		args        []string
		want        AgentConfig
		wantErr     bool
	}{
		{
			name: "Default values",
			want: AgentConfig{
				Address:            DefaultServerAddress,
				ReportInterval:     DefaultReportInterval,
				PollInterval:       DefaultPollInterval,
				Key:                DefaultSecretKey,
				CryptoKey:          DefaultCryptoKey,
				RateLimit:          DefaultRateLimit,
				PathStyle:          DefaultPathStyle,
				RealReportInterval: time.Duration(DefaultReportInterval) * time.Second,
				RealPollInterval:   time.Duration(DefaultPollInterval) * time.Second,
			},
		},
		{
			name: "Flags override defaults",
			args: []string{
				"-a=localhost:1234",
				"-r=5",
				"-p=2",
				"-k=flagKey",
				"--crypto-key=/flag/key.pem",
				"-l=10",
			},
			want: AgentConfig{
				Address:            "localhost:1234",
				ReportInterval:     5,
				PollInterval:       2,
				Key:                "flagKey",
				CryptoKey:          "/flag/key.pem",
				RateLimit:          10,
				PathStyle:          DefaultPathStyle,
				RealReportInterval: 5 * time.Second,
				RealPollInterval:   2 * time.Second,
			},
		},
		{
			name: "Config file overrides defaults",
			configFile: `{
				"address": "file:8080",
				"report_interval": 7,
				"poll_interval": 3,
				"key": "fileKey",
				"crypto_key": "/file/key.pem",
				"rate_limit": 15
			}`,
			want: AgentConfig{
				Address:            "file:8080",
				ReportInterval:     7,
				PollInterval:       3,
				Key:                "fileKey",
				CryptoKey:          "/file/key.pem",
				RateLimit:          15,
				PathStyle:          DefaultPathStyle,
				RealReportInterval: 7 * time.Second,
				RealPollInterval:   3 * time.Second,
			},
		},
		{
			name: "Env overrides defaults",
			env: map[string]string{
				"ADDRESS":         "env:8080",
				"REPORT_INTERVAL": "8",
				"POLL_INTERVAL":   "4",
				"KEY":             "envKey",
				"CRYPTO_KEY":      "/env/key.pem",
				"RATE_LIMIT":      "20",
			},
			want: AgentConfig{
				Address:            "env:8080",
				ReportInterval:     8,
				PollInterval:       4,
				Key:                "envKey",
				CryptoKey:          "/env/key.pem",
				RateLimit:          20,
				PathStyle:          DefaultPathStyle,
				RealReportInterval: 8 * time.Second,
				RealPollInterval:   4 * time.Second,
			},
		},
		{
			name: "env overrides flags which overrides config",
			args: []string{
				"-a=flag:8080",
				"-r=10",
				"-c=test.json",
			},
			env: map[string]string{
				"ADDRESS":         "env:8080",
				"REPORT_INTERVAL": "5",
				"POLL_INTERVAL":   "2",
			},
			configFile: `{
				"address": "file:8080",
				"report_interval": 7,
				"poll_interval": 3
			}`,
			want: AgentConfig{
				Address:            "env:8080",
				ReportInterval:     5,
				PollInterval:       2,
				Key:                DefaultSecretKey,
				CryptoKey:          DefaultCryptoKey,
				RateLimit:          DefaultRateLimit,
				PathStyle:          DefaultPathStyle,
				RealReportInterval: 5 * time.Second,
				RealPollInterval:   2 * time.Second,
			},
		},
		{
			name: "Invalid config file",
			args: []string{"-c=test.json"},
			configFile: `{
				"address": "invalid",
				"report_interval": "should be number"
			}`,
			wantErr:     true,
			errContains: "unmarshal config file",
		},
		{
			name:        "Non-existent config file",
			args:        []string{"-c=missing.json"},
			wantErr:     true,
			errContains: "read config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pflag.CommandLine = pflag.NewFlagSet("", pflag.ContinueOnError) // Сброс флагов
			os.Clearenv()
			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}
			var configPath string
			if tt.configFile != "" {
				tmpFile, err := os.Create("test.json")
				require.NoError(t, err)
				defer func() {
					_ = os.Remove(tmpFile.Name())
				}()

				_, err = tmpFile.WriteString(tt.configFile)
				require.NoError(t, err)
				_ = tmpFile.Close()

				configPath = tmpFile.Name()

				hasConfig := false
				for _, arg := range tt.args {
					if arg == "-c" || arg == "--config" || strings.HasPrefix(arg, "-c=") || strings.HasPrefix(arg,
						"--config=") {
						hasConfig = true
						break
					}
				}
				if !hasConfig {
					tt.args = append(tt.args, "-c="+configPath)
				}
			}

			os.Args = append([]string{"test"}, tt.args...)
			cfg := NewAgentConfig()
			err := cfg.Load()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want.Address, cfg.Address)
			require.Equal(t, tt.want.ReportInterval, cfg.ReportInterval)
			require.Equal(t, tt.want.PollInterval, cfg.PollInterval)
			require.Equal(t, tt.want.Key, cfg.Key)
			require.Equal(t, tt.want.CryptoKey, cfg.CryptoKey)
			require.Equal(t, tt.want.RateLimit, cfg.RateLimit)
			require.Equal(t, tt.want.PathStyle, cfg.PathStyle)
			require.Equal(t, tt.want.RealReportInterval, cfg.RealReportInterval)
			require.Equal(t, tt.want.RealPollInterval, cfg.RealPollInterval)
		})
	}
}
