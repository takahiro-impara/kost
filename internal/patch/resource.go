package patch

import (
	"fmt"

	"github.com/lot-koichi/kost/internal/engine"
	"github.com/lot-koichi/kost/internal/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

// ResourcePatch represents a Strategic Merge Patch for Deployment resources
type ResourcePatch struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Template struct {
			Spec struct {
				Containers []ContainerPatch `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}

// ContainerPatch represents a container resources patch
type ContainerPatch struct {
	Name      string `json:"name"`
	Resources struct {
		Requests map[string]string `json:"requests,omitempty"`
		Limits   map[string]string `json:"limits,omitempty"`
	} `json:"resources"`
}

// GenerateResourcePatch generates a Strategic Merge Patch for Deployment resources
func GenerateResourcePatch(deploymentName string, recommendations []engine.ResourceRecommendation) (string, error) {
	patch := ResourcePatch{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}
	patch.Metadata.Name = deploymentName

	// Build container patches
	containers := []ContainerPatch{}
	for _, rec := range recommendations {
		if rec.Deployment != deploymentName {
			continue
		}

		container := ContainerPatch{
			Name: rec.Container,
		}

		// Set recommended requests
		container.Resources.Requests = make(map[string]string)
		cpuQuantity := k8s.ConvertMillicoresToQuantity(rec.RecommendedCPURequestMilli)
		memQuantity := k8s.ConvertMiBToQuantity(rec.RecommendedMemRequestMi)
		container.Resources.Requests["cpu"] = cpuQuantity.String()
		container.Resources.Requests["memory"] = memQuantity.String()

		containers = append(containers, container)
	}

	patch.Spec.Template.Spec.Containers = containers

	// Marshal to YAML
	data, err := yaml.Marshal(patch)
	if err != nil {
		return "", fmt.Errorf("failed to marshal patch to YAML: %w", err)
	}

	return string(data), nil
}

// GenerateResourcePatchWithLimits generates a patch with both requests and limits
func GenerateResourcePatchWithLimits(deploymentName string, recommendations []engine.ResourceRecommendation, limitsFactor float64) (string, error) {
	patch := ResourcePatch{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}
	patch.Metadata.Name = deploymentName

	// Build container patches
	containers := []ContainerPatch{}
	for _, rec := range recommendations {
		if rec.Deployment != deploymentName {
			continue
		}

		container := ContainerPatch{
			Name: rec.Container,
		}

		// Set recommended requests
		container.Resources.Requests = make(map[string]string)
		cpuQuantity := k8s.ConvertMillicoresToQuantity(rec.RecommendedCPURequestMilli)
		memQuantity := k8s.ConvertMiBToQuantity(rec.RecommendedMemRequestMi)
		container.Resources.Requests["cpu"] = cpuQuantity.String()
		container.Resources.Requests["memory"] = memQuantity.String()

		// Set limits (requests * limitsFactor)
		if limitsFactor > 1.0 {
			container.Resources.Limits = make(map[string]string)
			cpuLimitMilli := int(float64(rec.RecommendedCPURequestMilli) * limitsFactor)
			memLimitMi := int(float64(rec.RecommendedMemRequestMi) * limitsFactor)
			cpuLimitQuantity := k8s.ConvertMillicoresToQuantity(cpuLimitMilli)
			memLimitQuantity := k8s.ConvertMiBToQuantity(memLimitMi)
			container.Resources.Limits["cpu"] = cpuLimitQuantity.String()
			container.Resources.Limits["memory"] = memLimitQuantity.String()
		}

		containers = append(containers, container)
	}

	patch.Spec.Template.Spec.Containers = containers

	// Marshal to YAML
	data, err := yaml.Marshal(patch)
	if err != nil {
		return "", fmt.Errorf("failed to marshal patch to YAML: %w", err)
	}

	return string(data), nil
}

// ValidatePatch validates that a patch is valid Strategic Merge Patch format
func ValidatePatch(patchYAML string) error {
	var patch ResourcePatch
	if err := yaml.Unmarshal([]byte(patchYAML), &patch); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	if patch.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}

	if patch.Kind == "" {
		return fmt.Errorf("kind is required")
	}

	if patch.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if len(patch.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	// Validate resource quantities
	for _, container := range patch.Spec.Template.Spec.Containers {
		if container.Name == "" {
			return fmt.Errorf("container name is required")
		}

		for resourceName, value := range container.Resources.Requests {
			if _, err := resource.ParseQuantity(value); err != nil {
				return fmt.Errorf("invalid request quantity for %s: %w", resourceName, err)
			}
		}

		for resourceName, value := range container.Resources.Limits {
			if _, err := resource.ParseQuantity(value); err != nil {
				return fmt.Errorf("invalid limit quantity for %s: %w", resourceName, err)
			}
		}
	}

	return nil
}
