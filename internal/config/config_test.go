package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		envVars     map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:       "default configuration",
			configFile: "",
			wantErr:    false,
		},
		{
			name:       "valid configuration file",
			configFile: "testdata/valid-config.yaml",
			wantErr:    false,
		},
		{
			name:        "invalid prometheus URL",
			configFile:  "testdata/invalid-prometheus-url.yaml",
			wantErr:     true,
			errContains: "invalid prometheus URL",
		},
		{
			name:        "invalid output path",
			configFile:  "testdata/invalid-output-path.yaml",
			wantErr:     true,
			errContains: "invalid output directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
				defer func(k string) {
					_ = os.Unsetenv(k)
				}(key)
			}

			cfg, err := Load(tt.configFile)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Prometheus: PrometheusConfig{
					URL:            "http://prometheus:9090",
					TimeoutSeconds: 10,
				},
				Analysis: AnalysisConfig{
					Window:        "7d",
					CPUPercentile: 0.95,
					MemPercentile: 0.95,
					SafetyFactor:  1.2,
					MinCPUMilli:   20,
					MinMemMi:      64,
				},
				HPA: HPAConfig{
					MinReplicasFactor: 0.8,
					MaxReplicasFactor: 1.3,
				},
				Filters: FiltersConfig{
					NamespacesExclude: []string{"kube-system"},
					LabelSelector:     "",
				},
				Output: OutputConfig{
					Dir:    "./out",
					Format: []string{"md"},
				},
				LLM: LLMConfig{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "empty prometheus URL",
			config: &Config{
				Prometheus: PrometheusConfig{
					URL: "",
				},
			},
			wantErr:     true,
			errContains: "prometheus.url is required",
		},
		{
			name: "invalid prometheus URL scheme",
			config: &Config{
				Prometheus: PrometheusConfig{
					URL: "ftp://prometheus:9090",
				},
			},
			wantErr:     true,
			errContains: "must start with http://",
		},
		{
			name: "invalid CPU percentile (too high)",
			config: &Config{
				Prometheus: PrometheusConfig{
					URL: "http://prometheus:9090",
				},
				Analysis: AnalysisConfig{
					CPUPercentile: 1.5,
					MemPercentile: 0.95,
				},
				Output: OutputConfig{
					Dir: "./out",
				},
			},
			wantErr:     true,
			errContains: "cpuPercentile must be between 0.0 and 1.0",
		},
		{
			name: "invalid safety factor",
			config: &Config{
				Prometheus: PrometheusConfig{
					URL: "http://prometheus:9090",
				},
				Analysis: AnalysisConfig{
					CPUPercentile: 0.95,
					MemPercentile: 0.95,
					SafetyFactor:  0.5,
				},
				Output: OutputConfig{
					Dir: "./out",
				},
			},
			wantErr:     true,
			errContains: "safetyFactor must be >= 1.0",
		},
		// Note: LLM API key validation test is environment-dependent and skipped here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "http://prometheus-operated.monitoring:9090", cfg.Prometheus.URL)
	assert.Equal(t, 10, cfg.Prometheus.TimeoutSeconds)
	assert.Equal(t, "7d", cfg.Analysis.Window)
	assert.Equal(t, 0.95, cfg.Analysis.CPUPercentile)
	assert.Equal(t, 0.95, cfg.Analysis.MemPercentile)
	assert.Equal(t, 1.2, cfg.Analysis.SafetyFactor)
	assert.Equal(t, 20, cfg.Analysis.MinCPUMilli)
	assert.Equal(t, 64, cfg.Analysis.MinMemMi)
	assert.Equal(t, 0.8, cfg.HPA.MinReplicasFactor)
	assert.Equal(t, 1.3, cfg.HPA.MaxReplicasFactor)
	assert.Contains(t, cfg.Filters.NamespacesExclude, "kube-system")
	assert.Contains(t, cfg.Filters.NamespacesExclude, "monitoring")
	assert.Equal(t, "./out", cfg.Output.Dir)
	assert.Contains(t, cfg.Output.Format, "md")
	assert.Contains(t, cfg.Output.Format, "json")
	assert.Contains(t, cfg.Output.Format, "patch")
	assert.False(t, cfg.LLM.Enabled)
	assert.Equal(t, "openai", cfg.LLM.Provider)
	assert.Equal(t, "gpt-4", cfg.LLM.Model)
}
