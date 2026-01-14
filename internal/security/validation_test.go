package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid namespace",
			namespace: "production",
			wantErr:   false,
		},
		{
			name:      "valid namespace with dash",
			namespace: "prod-app",
			wantErr:   false,
		},
		{
			name:      "valid namespace with numbers",
			namespace: "app-v1",
			wantErr:   false,
		},
		{
			name:        "empty namespace",
			namespace:   "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "uppercase namespace",
			namespace:   "PROD",
			wantErr:     true,
			errContains: "invalid namespace name",
		},
		{
			name:        "namespace with underscore",
			namespace:   "prod_app",
			wantErr:     true,
			errContains: "invalid namespace name",
		},
		{
			name:        "namespace too long",
			namespace:   "this-is-a-very-long-namespace-name-that-exceeds-the-maximum-allowed-length-of-sixty-three-characters",
			wantErr:     true,
			errContains: "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.namespace)
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

func TestValidateLabelSelector(t *testing.T) {
	tests := []struct {
		name        string
		selector    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "empty selector",
			selector: "",
			wantErr:  false,
		},
		{
			name:     "valid selector",
			selector: "app=nginx",
			wantErr:  false,
		},
		{
			name:     "valid selector with comma",
			selector: "app=nginx,env=prod",
			wantErr:  false,
		},
		{
			name:        "selector with single quote",
			selector:    "app='nginx'",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "selector with double quote",
			selector:    `app="nginx"`,
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "selector with semicolon",
			selector:    "app=nginx;env=prod",
			wantErr:     true,
			errContains: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLabelSelector(tt.selector)
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

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid relative path",
			path:    "./out",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			path:    "/tmp/kost",
			wantErr: false,
		},
		{
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "path traversal",
			path:        "../../etc/passwd",
			wantErr:     true,
			errContains: "path traversal detected",
		},
		{
			name:        "sensitive directory /etc",
			path:        "/etc/test",
			wantErr:     true,
			errContains: "sensitive directory",
		},
		{
			name:        "sensitive directory /root",
			path:        "/root/test",
			wantErr:     true,
			errContains: "sensitive directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputPath(tt.path)
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

func TestSanitizePromQLQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid query",
			query:   "container_cpu_usage_seconds_total",
			wantErr: false,
		},
		{
			name:    "valid query with braces",
			query:   `container_cpu_usage_seconds_total{namespace="prod"}`,
			wantErr: false,
		},
		{
			name:        "query with semicolon",
			query:       "container_cpu;drop table",
			wantErr:     true,
			errContains: "suspicious characters",
		},
		{
			name:        "query with comment",
			query:       "container_cpu--comment",
			wantErr:     true,
			errContains: "suspicious characters",
		},
		{
			name:        "query with multiline comment",
			query:       "container_cpu/* comment */",
			wantErr:     true,
			errContains: "suspicious characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizePromQLQuery(tt.query)
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

func TestMaskSensitiveValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "short string",
			value:    "abc",
			expected: "***",
		},
		{
			name:     "8 char string",
			value:    "12345678",
			expected: "***",
		},
		{
			name:     "long string",
			value:    "sk-1234567890abcdef",
			expected: "sk-1...cdef",
		},
		{
			name:     "API key",
			value:    "sk-ant-1234567890",
			expected: "sk-a...7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePrometheusURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid http URL",
			url:     "http://prometheus:9090",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			url:     "https://prometheus.example.com",
			wantErr: false,
		},
		{
			name:        "empty URL",
			url:         "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "invalid scheme",
			url:         "ftp://prometheus:9090",
			wantErr:     true,
			errContains: "must start with http://",
		},
		{
			name:        "URL with space",
			url:         "http://prometheus :9090",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "URL with special chars",
			url:         "http://prometheus<script>",
			wantErr:     true,
			errContains: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrometheusURL(tt.url)
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
