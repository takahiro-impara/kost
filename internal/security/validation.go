package security

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateNamespace validates Kubernetes namespace name
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	// Kubernetes namespace naming rules: lowercase alphanumeric, dash, max 63 chars
	matched, err := regexp.MatchString(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`, namespace)
	if err != nil {
		return fmt.Errorf("failed to validate namespace: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid namespace name: must be lowercase alphanumeric with dashes, max 63 chars")
	}

	if len(namespace) > 63 {
		return fmt.Errorf("namespace name too long: max 63 characters")
	}

	return nil
}

// ValidateLabelSelector validates Kubernetes label selector
func ValidateLabelSelector(selector string) error {
	if selector == "" {
		return nil // Empty selector is valid
	}

	// Basic validation: check for suspicious characters
	// More comprehensive validation would be done by Kubernetes API
	if strings.ContainsAny(selector, "';\"\\") {
		return fmt.Errorf("invalid characters in label selector")
	}

	return nil
}

// ValidateOutputPath validates output directory path to prevent path traversal
func ValidateOutputPath(outputPath string) error {
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	// Clean the path
	cleanPath := filepath.Clean(outputPath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: output path cannot contain '..'")
	}

	// Check for absolute paths pointing to sensitive directories
	if filepath.IsAbs(cleanPath) {
		sensitiveRoots := []string{"/etc", "/root", "/sys", "/proc", "/dev"}
		for _, root := range sensitiveRoots {
			if strings.HasPrefix(cleanPath, root) {
				return fmt.Errorf("output path points to sensitive directory: %s", root)
			}
		}
	}

	return nil
}

// SanitizePromQLQuery sanitizes PromQL query to prevent injection
func SanitizePromQLQuery(query string) (string, error) {
	// Check for suspicious characters that could be used for injection
	suspicious := []string{";", "--", "/*", "*/", "\\", "\n", "\r"}
	for _, char := range suspicious {
		if strings.Contains(query, char) {
			return "", fmt.Errorf("suspicious characters detected in query")
		}
	}

	return query, nil
}

// MaskSensitiveValue masks sensitive values in logs and outputs
func MaskSensitiveValue(value string) string {
	if value == "" {
		return ""
	}

	if len(value) <= 8 {
		return "***"
	}

	// Show first 4 and last 4 characters
	return value[:4] + "..." + value[len(value)-4:]
}

// ValidatePrometheusURL validates Prometheus URL
func ValidatePrometheusURL(url string) error {
	if url == "" {
		return fmt.Errorf("prometheus URL cannot be empty")
	}

	// Check for http:// or https:// prefix
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("prometheus URL must start with http:// or https://")
	}

	// Check for suspicious characters
	if strings.ContainsAny(url, " \t\n\r\"'<>") {
		return fmt.Errorf("invalid characters in prometheus URL")
	}

	return nil
}
