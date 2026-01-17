package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/takahiro-impara/kost/internal/security"
)

// Config represents the application configuration
type Config struct {
	Kube       KubeConfig       `mapstructure:"kube"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Analysis   AnalysisConfig   `mapstructure:"analysis"`
	HPA        HPAConfig        `mapstructure:"hpa"`
	Filters    FiltersConfig    `mapstructure:"filters"`
	Output     OutputConfig     `mapstructure:"output"`
	LLM        LLMConfig        `mapstructure:"llm"`
}

// KubeConfig represents Kubernetes client configuration
type KubeConfig struct {
	Context string `mapstructure:"context"`
}

// PrometheusConfig represents Prometheus client configuration
type PrometheusConfig struct {
	URL            string `mapstructure:"url"`
	TimeoutSeconds int    `mapstructure:"timeoutSeconds"`
}

// AnalysisConfig represents analysis parameters
type AnalysisConfig struct {
	Window        string  `mapstructure:"window"`
	CPUPercentile float64 `mapstructure:"cpuPercentile"`
	MemPercentile float64 `mapstructure:"memPercentile"`
	SafetyFactor  float64 `mapstructure:"safetyFactor"`
	MinCPUMilli   int     `mapstructure:"minCpuMilli"`
	MinMemMi      int     `mapstructure:"minMemMi"`
}

// HPAConfig represents HPA-specific analysis parameters
type HPAConfig struct {
	MinReplicasFactor float64 `mapstructure:"minReplicasFactor"`
	MaxReplicasFactor float64 `mapstructure:"maxReplicasFactor"`
}

// FiltersConfig represents resource filtering configuration
type FiltersConfig struct {
	NamespacesExclude []string `mapstructure:"namespacesExclude"`
	LabelSelector     string   `mapstructure:"labelSelector"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Dir    string   `mapstructure:"dir"`
	Format []string `mapstructure:"format"`
}

// LLMConfig represents LLM provider configuration
type LLMConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	Provider    string  `mapstructure:"provider"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"maxTokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// Load loads configuration from file and environment variables
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Load from config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variables override config file
	v.AutomaticEnv()
	v.SetEnvPrefix("KOST")

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("kube.context", "")
	v.SetDefault("prometheus.url", "http://prometheus-operated.monitoring:9090")
	v.SetDefault("prometheus.timeoutSeconds", 10)
	v.SetDefault("analysis.window", "7d")
	v.SetDefault("analysis.cpuPercentile", 0.95)
	v.SetDefault("analysis.memPercentile", 0.95)
	v.SetDefault("analysis.safetyFactor", 1.2)
	v.SetDefault("analysis.minCpuMilli", 20)
	v.SetDefault("analysis.minMemMi", 64)
	v.SetDefault("hpa.minReplicasFactor", 0.8)
	v.SetDefault("hpa.maxReplicasFactor", 1.3)
	v.SetDefault("filters.namespacesExclude", []string{"kube-system", "monitoring"})
	v.SetDefault("filters.labelSelector", "")
	v.SetDefault("output.dir", "./out")
	v.SetDefault("output.format", []string{"md", "json", "patch"})
	v.SetDefault("llm.enabled", false)
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4")
	v.SetDefault("llm.maxTokens", 1200)
	v.SetDefault("llm.temperature", 0.2)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Prometheus URL
	if c.Prometheus.URL == "" {
		return fmt.Errorf("prometheus.url is required")
	}
	if err := security.ValidatePrometheusURL(c.Prometheus.URL); err != nil {
		return fmt.Errorf("invalid prometheus URL: %w", err)
	}

	// Validate output directory path
	if err := security.ValidateOutputPath(c.Output.Dir); err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}

	// Validate label selector
	if err := security.ValidateLabelSelector(c.Filters.LabelSelector); err != nil {
		return fmt.Errorf("invalid label selector: %w", err)
	}

	// Validate percentiles
	if c.Analysis.CPUPercentile <= 0.0 || c.Analysis.CPUPercentile > 1.0 {
		return fmt.Errorf("analysis.cpuPercentile must be between 0.0 and 1.0")
	}
	if c.Analysis.MemPercentile <= 0.0 || c.Analysis.MemPercentile > 1.0 {
		return fmt.Errorf("analysis.memPercentile must be between 0.0 and 1.0")
	}

	// Validate safety factors
	if c.Analysis.SafetyFactor < 1.0 {
		return fmt.Errorf("analysis.safetyFactor must be >= 1.0")
	}
	if c.HPA.MinReplicasFactor <= 0.0 || c.HPA.MinReplicasFactor > 1.0 {
		return fmt.Errorf("hpa.minReplicasFactor must be between 0.0 and 1.0")
	}
	if c.HPA.MaxReplicasFactor < 1.0 {
		return fmt.Errorf("hpa.maxReplicasFactor must be >= 1.0")
	}

	// Validate LLM provider
	if c.LLM.Enabled {
		if c.LLM.Provider != "openai" && c.LLM.Provider != "claude" {
			return fmt.Errorf("llm.provider must be 'openai' or 'claude'")
		}

		// Check for API keys in environment
		if c.LLM.Provider == "openai" && os.Getenv("OPENAI_API_KEY") == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is required when llm.provider is 'openai'")
		}
		if c.LLM.Provider == "claude" && os.Getenv("ANTHROPIC_API_KEY") == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY environment variable is required when llm.provider is 'claude'")
		}
	}

	return nil
}
