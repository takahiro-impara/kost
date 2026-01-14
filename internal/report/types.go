package report

import (
	"time"

	"github.com/your-org/kost/internal/engine"
)

// Report represents the final output report
type Report struct {
	Namespace               string
	Window                  string
	Percentile              string
	SafetyFactor            float64
	ResourceRecommendations []engine.ResourceRecommendation
	HPARecommendations      []engine.HPARecommendation
	Narrative               string
	GeneratedAt             time.Time
}
