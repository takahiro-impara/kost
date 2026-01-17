package engine

// ResourceRecommendation represents resource optimization recommendations
type ResourceRecommendation struct {
	Deployment                 string
	Container                  string
	CurrentCPURequestMilli     *int
	RecommendedCPURequestMilli int
	CurrentMemRequestMi        *int
	RecommendedMemRequestMi    int
	CPUJudgement               string  // "overprovisioned", "underprovisioned", "appropriate"
	MemJudgement               string  // "overprovisioned", "underprovisioned", "appropriate"
	CPUSavingRatio             float64 // 0.0-1.0, negative means increase
	MemSavingRatio             float64 // 0.0-1.0, negative means increase
	Rationale                  Rationale
}

// Rationale contains the reasoning behind recommendations
type Rationale struct {
	CPUP95       float64
	MemP95       float64
	SafetyFactor float64
	MinCPUMilli  int
	MinMemMi     int
}

// HPARecommendation represents HPA optimization recommendations
type HPARecommendation struct {
	Deployment             string
	HPAName                string
	CurrentMinReplicas     int
	RecommendedMinReplicas int
	CurrentMaxReplicas     int
	RecommendedMaxReplicas int
	MinJudgement           string // "too_high", "too_low", "appropriate"
	MaxJudgement           string // "too_high", "too_low", "appropriate"
	Rationale              HPARationale
}

// HPARationale contains the reasoning behind HPA recommendations
type HPARationale struct {
	ReplicasP5        float64
	ReplicasP99       float64
	MinReplicasFactor float64
	MaxReplicasFactor float64
}
