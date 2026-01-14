package report

import (
	"encoding/json"
	"fmt"
)

// JSONSummary represents the summary data for JSON output
type JSONSummary struct {
	GeneratedAt            string                 `json:"generatedAt"`
	Namespace              string                 `json:"namespace"`
	AnalysisWindow         string                 `json:"analysisWindow"`
	Percentile             string                 `json:"percentile"`
	SafetyFactor           float64                `json:"safetyFactor"`
	TotalContainers        int                    `json:"totalContainers"`
	OverprovisionedCount   int                    `json:"overprovisionedCount"`
	UnderprovisionedCount  int                    `json:"underprovisionedCount"`
	AppropriateCount       int                    `json:"appropriateCount"`
	TotalSavingsPercent    float64                `json:"totalSavingsPercent"`
	ResourceRecommendations []JSONResourceRec     `json:"resourceRecommendations"`
	HPARecommendations     []JSONHPARec          `json:"hpaRecommendations,omitempty"`
}

// JSONResourceRec represents a single resource recommendation in JSON format
type JSONResourceRec struct {
	Deployment             string      `json:"deployment"`
	Container              string      `json:"container"`
	CurrentCPURequestMilli *int        `json:"currentCpuRequestMilli"`
	RecommendedCPUMilli    int         `json:"recommendedCpuMilli"`
	CPUJudgement           string      `json:"cpuJudgement"`
	CPUSavingRatio         float64     `json:"cpuSavingRatio"`
	CurrentMemRequestMi    *int        `json:"currentMemRequestMi"`
	RecommendedMemMi       int         `json:"recommendedMemMi"`
	MemJudgement           string      `json:"memJudgement"`
	MemSavingRatio         float64     `json:"memSavingRatio"`
	Rationale              JSONRationale `json:"rationale"`
}

// JSONRationale represents the rationale for a recommendation
type JSONRationale struct {
	CPUP95       float64 `json:"cpuP95"`
	MemP95       float64 `json:"memP95"`
	SafetyFactor float64 `json:"safetyFactor"`
	MinCPUMilli  int     `json:"minCpuMilli"`
	MinMemMi     int     `json:"minMemMi"`
}

// JSONHPARec represents an HPA recommendation in JSON format
type JSONHPARec struct {
	Deployment             string  `json:"deployment"`
	HPAName                string  `json:"hpaName"`
	CurrentMinReplicas     int     `json:"currentMinReplicas"`
	RecommendedMinReplicas int     `json:"recommendedMinReplicas"`
	MinJudgement           string  `json:"minJudgement"`
	CurrentMaxReplicas     int     `json:"currentMaxReplicas"`
	RecommendedMaxReplicas int     `json:"recommendedMaxReplicas"`
	MaxJudgement           string  `json:"maxJudgement"`
}

// GenerateJSON generates a JSON summary from a report
func GenerateJSON(report *Report) (string, error) {
	// Calculate summary statistics
	overprovisioned := 0
	underprovisioned := 0
	appropriate := 0
	totalSaving := 0.0

	for _, rec := range report.ResourceRecommendations {
		if rec.CPUJudgement == "overprovisioned" || rec.MemJudgement == "overprovisioned" {
			overprovisioned++
		} else if rec.CPUJudgement == "underprovisioned" || rec.MemJudgement == "underprovisioned" {
			underprovisioned++
		} else {
			appropriate++
		}

		// Average saving across CPU and Memory
		avgSaving := (rec.CPUSavingRatio + rec.MemSavingRatio) / 2.0
		if avgSaving > 0 {
			totalSaving += avgSaving
		}
	}

	// Calculate average saving percentage
	totalSavingsPercent := 0.0
	if len(report.ResourceRecommendations) > 0 {
		totalSavingsPercent = (totalSaving / float64(len(report.ResourceRecommendations))) * 100
	}

	// Build JSON resource recommendations
	jsonResourceRecs := []JSONResourceRec{}
	for _, rec := range report.ResourceRecommendations {
		jsonRec := JSONResourceRec{
			Deployment:             rec.Deployment,
			Container:              rec.Container,
			CurrentCPURequestMilli: rec.CurrentCPURequestMilli,
			RecommendedCPUMilli:    rec.RecommendedCPURequestMilli,
			CPUJudgement:           rec.CPUJudgement,
			CPUSavingRatio:         rec.CPUSavingRatio,
			CurrentMemRequestMi:    rec.CurrentMemRequestMi,
			RecommendedMemMi:       rec.RecommendedMemRequestMi,
			MemJudgement:           rec.MemJudgement,
			MemSavingRatio:         rec.MemSavingRatio,
			Rationale: JSONRationale{
				CPUP95:       rec.Rationale.CPUP95,
				MemP95:       rec.Rationale.MemP95,
				SafetyFactor: rec.Rationale.SafetyFactor,
				MinCPUMilli:  rec.Rationale.MinCPUMilli,
				MinMemMi:     rec.Rationale.MinMemMi,
			},
		}
		jsonResourceRecs = append(jsonResourceRecs, jsonRec)
	}

	// Build JSON HPA recommendations
	jsonHPARecs := []JSONHPARec{}
	for _, rec := range report.HPARecommendations {
		jsonRec := JSONHPARec{
			Deployment:             rec.Deployment,
			HPAName:                rec.HPAName,
			CurrentMinReplicas:     rec.CurrentMinReplicas,
			RecommendedMinReplicas: rec.RecommendedMinReplicas,
			MinJudgement:           rec.MinJudgement,
			CurrentMaxReplicas:     rec.CurrentMaxReplicas,
			RecommendedMaxReplicas: rec.RecommendedMaxReplicas,
			MaxJudgement:           rec.MaxJudgement,
		}
		jsonHPARecs = append(jsonHPARecs, jsonRec)
	}

	// Build summary
	summary := JSONSummary{
		GeneratedAt:             report.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
		Namespace:               report.Namespace,
		AnalysisWindow:          report.Window,
		Percentile:              report.Percentile,
		SafetyFactor:            report.SafetyFactor,
		TotalContainers:         len(report.ResourceRecommendations),
		OverprovisionedCount:    overprovisioned,
		UnderprovisionedCount:   underprovisioned,
		AppropriateCount:        appropriate,
		TotalSavingsPercent:     totalSavingsPercent,
		ResourceRecommendations: jsonResourceRecs,
		HPARecommendations:      jsonHPARecs,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}
