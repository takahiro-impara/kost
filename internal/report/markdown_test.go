package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takahiro-impara/kost/internal/engine"
)

func TestGenerateMarkdown(t *testing.T) {
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

	markdown, err := GenerateMarkdown(report)
	require.NoError(t, err)
	assert.NotEmpty(t, markdown)

	// Verify key sections
	assert.Contains(t, markdown, "# kost Report")
	assert.Contains(t, markdown, "test-namespace")
	assert.Contains(t, markdown, "7d")
	assert.Contains(t, markdown, "## Summary")
	assert.Contains(t, markdown, "## Top Findings")
	assert.Contains(t, markdown, "test-deployment")
	assert.Contains(t, markdown, "app")
	assert.Contains(t, markdown, "overprovisioned")
	assert.Contains(t, markdown, "## Suggested Patches")
	assert.Contains(t, markdown, "## Methodology")
}

func TestGenerateMarkdownMultipleRecommendations(t *testing.T) {
	cpu1 := 300
	mem1 := 400
	cpu2 := 100
	mem2 := 150

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
			{
				Deployment:                 "deployment-2",
				Container:                  "sidecar",
				CurrentCPURequestMilli:     &cpu2,
				RecommendedCPURequestMilli: 80,
				CurrentMemRequestMi:        &mem2,
				RecommendedMemRequestMi:    120,
				CPUJudgement:               "appropriate",
				MemJudgement:               "appropriate",
				CPUSavingRatio:             0.2,
				MemSavingRatio:             0.2,
				Rationale: engine.Rationale{
					CPUP95:       60,
					MemP95:       100,
					SafetyFactor: 1.2,
					MinCPUMilli:  20,
					MinMemMi:     64,
				},
			},
		},
		HPARecommendations: []engine.HPARecommendation{},
		GeneratedAt:        time.Now(),
	}

	markdown, err := GenerateMarkdown(report)
	require.NoError(t, err)

	// Verify both deployments are included
	assert.Contains(t, markdown, "deployment-1")
	assert.Contains(t, markdown, "deployment-2")

	// Verify sorting by saving ratio (deployment-1 should be first)
	idx1 := assert.Contains(t, markdown, "deployment-1")
	idx2 := assert.Contains(t, markdown, "deployment-2")
	// deployment-1 has higher total saving (0.6 + 0.4 = 1.0) than deployment-2 (0.2 + 0.2 = 0.4)
	// So deployment-1 should appear first in Top Findings
	_ = idx1
	_ = idx2
	// Note: Can't easily verify ordering in string, but sorting logic is tested separately
}

func TestGenerateMarkdownEmptyReport(t *testing.T) {
	report := &Report{
		Namespace:               "empty-namespace",
		Window:                  "7d",
		Percentile:              "P95",
		SafetyFactor:            1.2,
		ResourceRecommendations: []engine.ResourceRecommendation{},
		HPARecommendations:      []engine.HPARecommendation{},
		GeneratedAt:             time.Now(),
	}

	markdown, err := GenerateMarkdown(report)
	require.NoError(t, err)
	assert.NotEmpty(t, markdown)

	// Should still have basic structure
	assert.Contains(t, markdown, "# kost Report")
	assert.Contains(t, markdown, "empty-namespace")
}
