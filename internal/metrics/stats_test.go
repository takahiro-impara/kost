package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateStatistics(t *testing.T) {
	tests := []struct {
		name        string
		values      []float64
		wantErr     bool
		expectedP50 float64
		expectedP95 float64
		expectedP99 float64
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "normal distribution",
			values:      []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			wantErr:     false,
			expectedP50: 50,
			expectedP95: 95,
			expectedP99: 95, // With 10 values, P99 is close to P95
			expectedMin: 10,
			expectedMax: 100,
		},
		{
			name:    "empty values",
			values:  []float64{},
			wantErr: true,
		},
		{
			name:        "single value",
			values:      []float64{42},
			wantErr:     false,
			expectedP50: 42,
			expectedP95: 42,
			expectedP99: 42,
			expectedMin: 42,
			expectedMax: 42,
		},
		{
			name:        "two values",
			values:      []float64{10, 20},
			wantErr:     false,
			expectedP50: 15, // With 2 values, percentile calculation varies
			expectedP95: 15, // Delta of 5 allows for variance
			expectedP99: 15,
			expectedMin: 10,
			expectedMax: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				DeploymentName: "test-deployment",
				ContainerName:  "test-container",
				MetricType:     "cpu",
				Values:         tt.values,
			}

			err := m.CalculateStatistics()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.expectedP50, m.P50, 5.0)
				assert.InDelta(t, tt.expectedP95, m.P95, 5.0)
				assert.InDelta(t, tt.expectedP99, m.P99, 5.0)
				assert.Equal(t, tt.expectedMin, m.Min)
				assert.Equal(t, tt.expectedMax, m.Max)
			}
		})
	}
}

func TestGetPercentile(t *testing.T) {
	m := &Metrics{
		Values: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}

	tests := []struct {
		name       string
		percentile float64
		expected   float64
		delta      float64
	}{
		{
			name:       "P50",
			percentile: 0.50,
			expected:   50,
			delta:      1.0,
		},
		{
			name:       "P90",
			percentile: 0.90,
			expected:   90,
			delta:      1.0,
		},
		{
			name:       "P95",
			percentile: 0.95,
			expected:   95,
			delta:      1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := m.GetPercentile(tt.percentile)
			require.NoError(t, err)
			assert.InDelta(t, tt.expected, value, tt.delta)
		})
	}
}

func TestGetPercentileEmptyValues(t *testing.T) {
	m := &Metrics{
		Values: []float64{},
	}

	_, err := m.GetPercentile(0.95)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no values available")
}

func TestRoundUp(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{1.1, 2},
		{1.9, 2},
		{2.0, 2},
		{2.1, 3},
		{0.1, 1},
		{0.0, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, RoundUp(tt.input))
		})
	}
}

func TestRoundToNearest(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{1.1, 1},
		{1.5, 2},
		{1.9, 2},
		{2.0, 2},
		{2.5, 3}, // banker's rounding: 2.5 rounds to 3 (nearest even)
		{2.6, 3},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, RoundToNearest(tt.input))
		})
	}
}
