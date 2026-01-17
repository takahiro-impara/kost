package engine

import (
	"math"

	"github.com/takahiro-impara/kost/internal/metrics"
)

// CalculateCPURecommendation calculates recommended CPU requests
func CalculateCPURecommendation(cpuMetrics *metrics.Metrics, safetyFactor float64, minCPUMilli int) int {
	// Use P95 as the baseline
	recommended := cpuMetrics.P95 * safetyFactor

	// Apply minimum threshold
	if recommended < float64(minCPUMilli) {
		recommended = float64(minCPUMilli)
	}

	// Round up to nearest integer
	return int(math.Ceil(recommended))
}

// CalculateMemoryRecommendation calculates recommended Memory requests
func CalculateMemoryRecommendation(memMetrics *metrics.Metrics, safetyFactor float64, minMemMi int) int {
	// Use P95 as the baseline
	recommended := memMetrics.P95 * safetyFactor

	// Apply minimum threshold
	if recommended < float64(minMemMi) {
		recommended = float64(minMemMi)
	}

	// Round up to nearest integer
	return int(math.Ceil(recommended))
}

// JudgeCPU determines if current CPU request is appropriate
func JudgeCPU(currentMilli *int, p95 float64) string {
	if currentMilli == nil {
		return "not_set"
	}

	current := float64(*currentMilli)

	// Overprovisioned: current > P95 * 2.0
	if current > p95*2.0 {
		return "overprovisioned"
	}

	// Underprovisioned: current < P95 * 1.1
	if current < p95*1.1 {
		return "underprovisioned"
	}

	return "appropriate"
}

// JudgeMemory determines if current Memory request is appropriate
func JudgeMemory(currentMi *int, p95 float64) string {
	if currentMi == nil {
		return "not_set"
	}

	current := float64(*currentMi)

	// Overprovisioned: current > P95 * 2.0
	if current > p95*2.0 {
		return "overprovisioned"
	}

	// Underprovisioned: current < P95 * 1.1
	if current < p95*1.1 {
		return "underprovisioned"
	}

	return "appropriate"
}

// CalculateSavingRatio calculates the saving ratio (positive means saving, negative means increase needed)
func CalculateSavingRatio(current *int, recommended int) float64 {
	if current == nil || *current == 0 {
		return 0.0
	}

	saving := float64(*current - recommended)
	ratio := saving / float64(*current)

	return ratio
}

// GenerateResourceRecommendation generates a complete resource recommendation
func GenerateResourceRecommendation(
	deployment, container string,
	currentCPU, currentMem *int,
	cpuMetrics, memMetrics *metrics.Metrics,
	safetyFactor float64,
	minCPUMilli, minMemMi int,
) *ResourceRecommendation {
	// Calculate recommendations
	recCPU := CalculateCPURecommendation(cpuMetrics, safetyFactor, minCPUMilli)
	recMem := CalculateMemoryRecommendation(memMetrics, safetyFactor, minMemMi)

	// Judge current settings
	cpuJudge := JudgeCPU(currentCPU, cpuMetrics.P95)
	memJudge := JudgeMemory(currentMem, memMetrics.P95)

	// Calculate saving ratios
	cpuSaving := CalculateSavingRatio(currentCPU, recCPU)
	memSaving := CalculateSavingRatio(currentMem, recMem)

	return &ResourceRecommendation{
		Deployment:                 deployment,
		Container:                  container,
		CurrentCPURequestMilli:     currentCPU,
		RecommendedCPURequestMilli: recCPU,
		CurrentMemRequestMi:        currentMem,
		RecommendedMemRequestMi:    recMem,
		CPUJudgement:               cpuJudge,
		MemJudgement:               memJudge,
		CPUSavingRatio:             cpuSaving,
		MemSavingRatio:             memSaving,
		Rationale: Rationale{
			CPUP95:       cpuMetrics.P95,
			MemP95:       memMetrics.P95,
			SafetyFactor: safetyFactor,
			MinCPUMilli:  minCPUMilli,
			MinMemMi:     minMemMi,
		},
	}
}
