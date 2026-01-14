package metrics

// Metrics represents time series metrics data with statistics
type Metrics struct {
	DeploymentName string
	ContainerName  string
	MetricType     string // "cpu", "memory", "replicas"
	Values         []float64
	P50            float64
	P95            float64
	P99            float64
	Min            float64
	Max            float64
	Avg            float64
}
