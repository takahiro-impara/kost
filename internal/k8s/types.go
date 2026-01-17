package k8s

// Deployment represents a Kubernetes Deployment with resource information
type Deployment struct {
	Namespace  string
	Name       string
	Containers []Container
	HasHPA     bool
	HPARef     *HPA
}

// Container represents a container within a Deployment
type Container struct {
	Name            string
	CPURequestMilli *int // CPU requests in millicores (nil if not set)
	MemRequestMi    *int // Memory requests in MiB (nil if not set)
	CPULimitMilli   *int // CPU limits in millicores (nil if not set)
	MemLimitMi      *int // Memory limits in MiB (nil if not set)
}

// HPA represents a Horizontal Pod Autoscaler
type HPA struct {
	Namespace               string
	Name                    string
	TargetDeployment        string
	MinReplicas             int
	MaxReplicas             int
	TargetCPUUtilization    *int // Target CPU utilization percentage (nil if not set)
	TargetMemoryUtilization *int // Target Memory utilization percentage (nil if not set)
}
