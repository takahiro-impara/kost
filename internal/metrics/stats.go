package metrics

import (
	"fmt"
	"math"
	"sort"

	"github.com/montanaflynn/stats"
)

// CalculateStatistics calculates percentiles and basic statistics for metrics
func (m *Metrics) CalculateStatistics() error {
	if len(m.Values) == 0 {
		return fmt.Errorf("no values to calculate statistics")
	}

	// Calculate percentiles
	p50, err := stats.Percentile(m.Values, 50)
	if err != nil {
		return fmt.Errorf("failed to calculate P50: %w", err)
	}
	m.P50 = p50

	p95, err := stats.Percentile(m.Values, 95)
	if err != nil {
		return fmt.Errorf("failed to calculate P95: %w", err)
	}
	m.P95 = p95

	p99, err := stats.Percentile(m.Values, 99)
	if err != nil {
		return fmt.Errorf("failed to calculate P99: %w", err)
	}
	m.P99 = p99

	// Calculate min, max, avg
	sortedValues := make([]float64, len(m.Values))
	copy(sortedValues, m.Values)
	sort.Float64s(sortedValues)

	m.Min = sortedValues[0]
	m.Max = sortedValues[len(sortedValues)-1]

	sum := 0.0
	for _, v := range m.Values {
		sum += v
	}
	m.Avg = sum / float64(len(m.Values))

	return nil
}

// GetPercentile returns the specified percentile value
func (m *Metrics) GetPercentile(percentile float64) (float64, error) {
	if len(m.Values) == 0 {
		return 0, fmt.Errorf("no values available")
	}

	value, err := stats.Percentile(m.Values, percentile*100)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate percentile: %w", err)
	}

	return value, nil
}

// RoundUp rounds up a float64 to the nearest integer
func RoundUp(value float64) int {
	return int(math.Ceil(value))
}

// RoundToNearest rounds to the nearest integer
func RoundToNearest(value float64) int {
	return int(math.Round(value))
}
