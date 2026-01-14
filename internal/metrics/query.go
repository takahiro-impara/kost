package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

// QueryCPUMetrics queries CPU usage metrics for a deployment/container
func (c *Client) QueryCPUMetrics(ctx context.Context, namespace, deployment, container string, window time.Duration) (*Metrics, error) {
	// PromQL query for CPU usage rate over 5m
	// Note: container label may not be present in all Prometheus setups (e.g., minikube)
	// Try with container label first, fall back to pod-only if needed
	query := fmt.Sprintf(
		`rate(container_cpu_usage_seconds_total{namespace="%s",pod=~"%s-.*",id=~".*/.*"}[5m])`,
		namespace, deployment,
	)

	end := time.Now()
	start := end.Add(-window)
	step := 5 * time.Minute

	result, err := c.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("failed to query CPU metrics: %w", err)
	}

	metrics := &Metrics{
		DeploymentName: deployment,
		ContainerName:  container,
		MetricType:     "cpu",
	}

	// Extract values from result
	if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
		for _, sample := range matrix[0].Values {
			// Convert cores to millicores
			value := float64(sample.Value) * 1000
			metrics.Values = append(metrics.Values, value)
		}
	}

	if len(metrics.Values) == 0 {
		return nil, fmt.Errorf("no CPU metrics found for %s/%s", deployment, container)
	}

	return metrics, nil
}

// QueryMemoryMetrics queries memory usage metrics for a deployment/container
func (c *Client) QueryMemoryMetrics(ctx context.Context, namespace, deployment, container string, window time.Duration) (*Metrics, error) {
	// PromQL query for memory working set
	// Note: container label may not be present in all Prometheus setups (e.g., minikube)
	query := fmt.Sprintf(
		`container_memory_working_set_bytes{namespace="%s",pod=~"%s-.*",id=~".*/.*"}`,
		namespace, deployment,
	)

	end := time.Now()
	start := end.Add(-window)
	step := 5 * time.Minute

	result, err := c.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory metrics: %w", err)
	}

	metrics := &Metrics{
		DeploymentName: deployment,
		ContainerName:  container,
		MetricType:     "memory",
	}

	// Extract values from result
	if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
		for _, sample := range matrix[0].Values {
			// Convert bytes to MiB
			value := float64(sample.Value) / (1024 * 1024)
			metrics.Values = append(metrics.Values, value)
		}
	}

	if len(metrics.Values) == 0 {
		return nil, fmt.Errorf("no memory metrics found for %s/%s", deployment, container)
	}

	return metrics, nil
}

// QueryReplicasMetrics queries replica count metrics for a deployment
func (c *Client) QueryReplicasMetrics(ctx context.Context, namespace, deployment string, window time.Duration) (*Metrics, error) {
	// PromQL query for deployment replicas
	query := fmt.Sprintf(
		`kube_deployment_status_replicas{namespace="%s",deployment="%s"}`,
		namespace, deployment,
	)

	end := time.Now()
	start := end.Add(-window)
	step := 5 * time.Minute

	result, err := c.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("failed to query replicas metrics: %w", err)
	}

	metrics := &Metrics{
		DeploymentName: deployment,
		MetricType:     "replicas",
	}

	// Extract values from result
	if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
		for _, sample := range matrix[0].Values {
			value := float64(sample.Value)
			metrics.Values = append(metrics.Values, value)
		}
	}

	if len(metrics.Values) == 0 {
		return nil, fmt.Errorf("no replicas metrics found for %s", deployment)
	}

	return metrics, nil
}
