package k8s

import (
	"context"
	"fmt"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployments lists all Deployments in the given namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string, labelSelector string) ([]Deployment, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	deploymentList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var deployments []Deployment
	for _, d := range deploymentList.Items {
		deployment := convertDeployment(&d)
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// convertDeployment converts a Kubernetes Deployment to our internal Deployment type
func convertDeployment(d *v1.Deployment) Deployment {
	deployment := Deployment{
		Namespace:  d.Namespace,
		Name:       d.Name,
		Containers: []Container{},
		HasHPA:     false,
	}

	// Extract container resources
	for _, container := range d.Spec.Template.Spec.Containers {
		c := extractContainerResources(container)
		deployment.Containers = append(deployment.Containers, c)
	}

	return deployment
}

// extractContainerResources extracts resource requests and limits from a container
func extractContainerResources(container corev1.Container) Container {
	c := Container{
		Name: container.Name,
	}

	// Extract CPU requests
	if cpuReq, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
		cpuMilli := int(cpuReq.MilliValue())
		c.CPURequestMilli = &cpuMilli
	}

	// Extract Memory requests
	if memReq, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
		memMi := int(memReq.Value() / (1024 * 1024)) // Convert bytes to MiB
		c.MemRequestMi = &memMi
	}

	// Extract CPU limits
	if cpuLimit, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
		cpuMilli := int(cpuLimit.MilliValue())
		c.CPULimitMilli = &cpuMilli
	}

	// Extract Memory limits
	if memLimit, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
		memMi := int(memLimit.Value() / (1024 * 1024)) // Convert bytes to MiB
		c.MemLimitMi = &memMi
	}

	return c
}

// FilterDeployments filters deployments based on namespace exclusion rules
func FilterDeployments(deployments []Deployment, excludeNamespaces []string) []Deployment {
	if len(excludeNamespaces) == 0 {
		return deployments
	}

	excludeMap := make(map[string]bool)
	for _, ns := range excludeNamespaces {
		excludeMap[ns] = true
	}

	var filtered []Deployment
	for _, d := range deployments {
		if !excludeMap[d.Namespace] {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// ConvertMillicoresToQuantity converts millicores to resource.Quantity
func ConvertMillicoresToQuantity(milli int) resource.Quantity {
	return *resource.NewMilliQuantity(int64(milli), resource.DecimalSI)
}

// ConvertMiBToQuantity converts MiB to resource.Quantity
func ConvertMiBToQuantity(mi int) resource.Quantity {
	return *resource.NewQuantity(int64(mi)*1024*1024, resource.BinarySI)
}
