package report

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kost/internal/engine"
)

func TestGenerateJSON(t *testing.T) {
	currentCPU := 300
	currentMem := 400

	report := &Report{
		Namespace:    "test-namespace",
		Window:       "7d",
		Percentile:   "P95",
		SafetyFactor: 1.2,
		ResourceRecommendations: []engine.ResourceRecommendation{
			{
				Deployment:                 "test-deployment",
				Container:                  "app",
				CurrentCPURequestMilli:     &currentCPU,
				RecommendedCPURequestMilli: 120,
				CurrentMemRequestMi:        &currentMem,
				RecommendedMemRequestMi:    240,
				CPUJudgement:               "overprovisioned",
				MemJudgement:               "appropriate",
				CPUSavingRatio:             0.6,
				MemSavingRatio:             0.4,
				Rationale: engine.Rationale{
					CPUP95:       100,
					MemP95:       200,
					SafetyFactor: 1.2,
					MinCPUMilli:  20,
					MinMemMi:     64,
				},
			},
		},
		HPARecommendations: []engine.HPARecommendation{},
		GeneratedAt:        time.Now(),
	}

	jsonData, err := GenerateJSON(report)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify valid JSON
	var summary JSONSummary
	err = json.Unmarshal([]byte(jsonData), &summary)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, "test-namespace", summary.Namespace)
	assert.Equal(t, "7d", summary.AnalysisWindow)
	assert.Equal(t, "P95", summary.Percentile)
	assert.Equal(t, 1.2, summary.SafetyFactor)
	assert.Equal(t, 1, summary.TotalContainers)
	assert.Equal(t, 1, summary.OverprovisionedCount)
	assert.Equal(t, 0, summary.UnderprovisionedCount)
	assert.Equal(t, 0, summary.AppropriateCount)
	assert.Len(t, summary.ResourceRecommendations, 1)

	// Verify recommendation details
	rec := summary.ResourceRecommendations[0]
	assert.Equal(t, "test-deployment", rec.Deployment)
	assert.Equal(t, "app", rec.Container)
	assert.Equal(t, 300, *rec.CurrentCPURequestMilli)
	assert.Equal(t, 120, rec.RecommendedCPUMilli)
	assert.Equal(t, "overprovisioned", rec.CPUJudgement)
	assert.InDelta(t, 0.6, rec.CPUSavingRatio, 0.01)
}

func TestGenerateJSONWithHPA(t *testing.T) {
	currentCPU := 200
	currentMem := 300

	report := &Report{
		Namespace:    "test-namespace",
		Window:       "7d",
		Percentile:   "P95",
		SafetyFactor: 1.2,
		ResourceRecommendations: []engine.ResourceRecommendation{
			{
				Deployment:                 "test-deployment",
				Container:                  "app",
				CurrentCPURequestMilli:     &currentCPU,
				RecommendedCPURequestMilli: 120,
				CurrentMemRequestMi:        &currentMem,
				RecommendedMemRequestMi:    240,
				CPUJudgement:               "appropriate",
				MemJudgement:               "appropriate",
				CPUSavingRatio:             0.4,
				MemSavingRatio:             0.2,
				Rationale: engine.Rationale{
					CPUP95:       100,
					MemP95:       200,
					SafetyFactor: 1.2,
					MinCPUMilli:  20,
					MinMemMi:     64,
				},
			},
		},
		HPARecommendations: []engine.HPARecommendation{
			{
				Deployment:             "test-deployment",
				HPAName:                "test-hpa",
				CurrentMinReplicas:     2,
				RecommendedMinReplicas: 1,
				MinJudgement:           "too_high",
				CurrentMaxReplicas:     10,
				RecommendedMaxReplicas: 8,
				MaxJudgement:           "appropriate",
			},
		},
		GeneratedAt: time.Now(),
	}

	jsonData, err := GenerateJSON(report)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal([]byte(jsonData), &summary)
	require.NoError(t, err)

	// Verify HPA recommendations
	assert.Len(t, summary.HPARecommendations, 1)
	hpa := summary.HPARecommendations[0]
	assert.Equal(t, "test-deployment", hpa.Deployment)
	assert.Equal(t, "test-hpa", hpa.HPAName)
	assert.Equal(t, 2, hpa.CurrentMinReplicas)
	assert.Equal(t, 1, hpa.RecommendedMinReplicas)
	assert.Equal(t, "too_high", hpa.MinJudgement)
}

func TestGenerateJSONEmptyReport(t *testing.T) {
	report := &Report{
		Namespace:               "empty-namespace",
		Window:                  "7d",
		Percentile:              "P95",
		SafetyFactor:            1.2,
		ResourceRecommendations: []engine.ResourceRecommendation{},
		HPARecommendations:      []engine.HPARecommendation{},
		GeneratedAt:             time.Now(),
	}

	jsonData, err := GenerateJSON(report)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal([]byte(jsonData), &summary)
	require.NoError(t, err)

	assert.Equal(t, "empty-namespace", summary.Namespace)
	assert.Equal(t, 0, summary.TotalContainers)
	assert.Len(t, summary.ResourceRecommendations, 0)
	assert.Len(t, summary.HPARecommendations, 0)
}

func TestGenerateJSONSavingsCalculation(t *testing.T) {
	cpu1 := 300
	mem1 := 400
	cpu2 := 100
	mem2 := 200

	report := &Report{
		Namespace:    "test-namespace",
		Window:       "7d",
		Percentile:   "P95",
		SafetyFactor: 1.2,
		ResourceRecommendations: []engine.ResourceRecommendation{
			{
				Deployment:                 "deployment-1",
				Container:                  "app",
				CurrentCPURequestMilli:     &cpu1,
				RecommendedCPURequestMilli: 120,
				CurrentMemRequestMi:        &mem1,
				RecommendedMemRequestMi:    240,
				CPUJudgement:               "overprovisioned",
				MemJudgement:               "appropriate",
				CPUSavingRatio:             0.6, // (300-120)/300 = 0.6
				MemSavingRatio:             0.4, // (400-240)/400 = 0.4
				Rationale:                  engine.Rationale{},
			},
			{
				Deployment:                 "deployment-2",
				Container:                  "sidecar",
				CurrentCPURequestMilli:     &cpu2,
				RecommendedCPURequestMilli: 80,
				CurrentMemRequestMi:        &mem2,
				RecommendedMemRequestMi:    160,
				CPUJudgement:               "appropriate",
				MemJudgement:               "appropriate",
				CPUSavingRatio:             0.2, // (100-80)/100 = 0.2
				MemSavingRatio:             0.2, // (200-160)/200 = 0.2
				Rationale:                  engine.Rationale{},
			},
		},
		HPARecommendations: []engine.HPARecommendation{},
		GeneratedAt:        time.Now(),
	}

	jsonData, err := GenerateJSON(report)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal([]byte(jsonData), &summary)
	require.NoError(t, err)

	// Average saving: ((0.6+0.4)/2 + (0.2+0.2)/2) / 2 = (0.5 + 0.2) / 2 = 0.35 = 35%
	assert.InDelta(t, 35.0, summary.TotalSavingsPercent, 1.0)
}
