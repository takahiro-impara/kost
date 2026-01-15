package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/lot-koichi/kost/internal/metrics"
)

func TestCalculateCPURecommendation(t *testing.T) {
	tests := []struct {
		name         string
		cpuP95       float64
		safetyFactor float64
		minCPUMilli  int
		expected     int
	}{
		{
			name:         "normal case",
			cpuP95:       100,
			safetyFactor: 1.2,
			minCPUMilli:  20,
			expected:     120, // 100 * 1.2 = 120
		},
		{
			name:         "below minimum",
			cpuP95:       10,
			safetyFactor: 1.2,
			minCPUMilli:  20,
			expected:     20, // 10 * 1.2 = 12, but min is 20
		},
		{
			name:         "exact minimum",
			cpuP95:       16.66,
			safetyFactor: 1.2,
			minCPUMilli:  20,
			expected:     20, // 16.66 * 1.2 = 19.99, rounded up to 20
		},
		{
			name:         "high value",
			cpuP95:       1000,
			safetyFactor: 1.5,
			minCPUMilli:  20,
			expected:     1500, // 1000 * 1.5 = 1500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metrics.Metrics{
				P95: tt.cpuP95,
			}
			result := CalculateCPURecommendation(m, tt.safetyFactor, tt.minCPUMilli)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateMemoryRecommendation(t *testing.T) {
	tests := []struct {
		name         string
		memP95       float64
		safetyFactor float64
		minMemMi     int
		expected     int
	}{
		{
			name:         "normal case",
			memP95:       200,
			safetyFactor: 1.2,
			minMemMi:     64,
			expected:     240, // 200 * 1.2 = 240
		},
		{
			name:         "below minimum",
			memP95:       40,
			safetyFactor: 1.2,
			minMemMi:     64,
			expected:     64, // 40 * 1.2 = 48, but min is 64
		},
		{
			name:         "exact minimum",
			memP95:       53.33,
			safetyFactor: 1.2,
			minMemMi:     64,
			expected:     64, // 53.33 * 1.2 = 63.99, rounded up to 64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metrics.Metrics{
				P95: tt.memP95,
			}
			result := CalculateMemoryRecommendation(m, tt.safetyFactor, tt.minMemMi)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJudgeCPU(t *testing.T) {
	tests := []struct {
		name        string
		currentCPU  *int
		p95         float64
		expected    string
	}{
		{
			name:       "not set",
			currentCPU: nil,
			p95:        100,
			expected:   "not_set",
		},
		{
			name:       "overprovisioned",
			currentCPU: intPtr(300),
			p95:        100, // 300 > 100 * 2.0
			expected:   "overprovisioned",
		},
		{
			name:       "underprovisioned",
			currentCPU: intPtr(100),
			p95:        100, // 100 < 100 * 1.1
			expected:   "underprovisioned",
		},
		{
			name:       "appropriate",
			currentCPU: intPtr(150),
			p95:        100, // 100 * 1.1 <= 150 <= 100 * 2.0
			expected:   "appropriate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JudgeCPU(tt.currentCPU, tt.p95)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJudgeMemory(t *testing.T) {
	tests := []struct {
		name       string
		currentMem *int
		p95        float64
		expected   string
	}{
		{
			name:       "not set",
			currentMem: nil,
			p95:        200,
			expected:   "not_set",
		},
		{
			name:       "overprovisioned",
			currentMem: intPtr(500),
			p95:        200, // 500 > 200 * 2.0
			expected:   "overprovisioned",
		},
		{
			name:       "underprovisioned",
			currentMem: intPtr(200),
			p95:        200, // 200 < 200 * 1.1
			expected:   "underprovisioned",
		},
		{
			name:       "appropriate",
			currentMem: intPtr(300),
			p95:        200, // 200 * 1.1 <= 300 <= 200 * 2.0
			expected:   "appropriate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JudgeMemory(tt.currentMem, tt.p95)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateSavingRatio(t *testing.T) {
	tests := []struct {
		name        string
		current     *int
		recommended int
		expected    float64
	}{
		{
			name:        "saving case",
			current:     intPtr(1000),
			recommended: 800,
			expected:    0.2, // (1000 - 800) / 1000 = 0.2
		},
		{
			name:        "increase case",
			current:     intPtr(800),
			recommended: 1000,
			expected:    -0.25, // (800 - 1000) / 800 = -0.25
		},
		{
			name:        "no change",
			current:     intPtr(1000),
			recommended: 1000,
			expected:    0.0,
		},
		{
			name:        "current not set",
			current:     nil,
			recommended: 1000,
			expected:    0.0,
		},
		{
			name:        "current is zero",
			current:     intPtr(0),
			recommended: 1000,
			expected:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSavingRatio(tt.current, tt.recommended)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestGenerateResourceRecommendation(t *testing.T) {
	cpuMetrics := &metrics.Metrics{
		P95: 100,
	}
	memMetrics := &metrics.Metrics{
		P95: 200,
	}

	currentCPU := intPtr(300)
	currentMem := intPtr(400)

	rec := GenerateResourceRecommendation(
		"test-deployment",
		"test-container",
		currentCPU,
		currentMem,
		cpuMetrics,
		memMetrics,
		1.2,  // safetyFactor
		20,   // minCPUMilli
		64,   // minMemMi
	)

	assert.NotNil(t, rec)
	assert.Equal(t, "test-deployment", rec.Deployment)
	assert.Equal(t, "test-container", rec.Container)
	assert.Equal(t, currentCPU, rec.CurrentCPURequestMilli)
	assert.Equal(t, currentMem, rec.CurrentMemRequestMi)
	assert.Equal(t, 120, rec.RecommendedCPURequestMilli) // 100 * 1.2 = 120
	assert.Equal(t, 240, rec.RecommendedMemRequestMi)    // 200 * 1.2 = 240
	assert.Equal(t, "overprovisioned", rec.CPUJudgement) // 300 > 100 * 2.0
	assert.Equal(t, "appropriate", rec.MemJudgement)      // 200 * 1.1 <= 400 <= 200 * 2.0
	assert.InDelta(t, 0.6, rec.CPUSavingRatio, 0.01)      // (300 - 120) / 300
	assert.InDelta(t, 0.4, rec.MemSavingRatio, 0.01)      // (400 - 240) / 400
	assert.Equal(t, 100.0, rec.Rationale.CPUP95)
	assert.Equal(t, 200.0, rec.Rationale.MemP95)
	assert.Equal(t, 1.2, rec.Rationale.SafetyFactor)
	assert.Equal(t, 20, rec.Rationale.MinCPUMilli)
	assert.Equal(t, 64, rec.Rationale.MinMemMi)
}

// Helper function
func intPtr(i int) *int {
	return &i
}
