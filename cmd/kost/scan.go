package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/lot-koichi/kost/internal/k8s"
	"github.com/lot-koichi/kost/internal/security"
)

// validateNamespace validates namespace input
func validateNamespace(ns string) error {
	return security.ValidateNamespace(ns)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Kubernetes Deployments and HPAs",
	Long: `Scan Kubernetes Deployments and Horizontal Pod Autoscalers in the specified namespace.
This command collects current resource configurations to prepare for analysis.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if namespace == "" {
			return fmt.Errorf("namespace is required (use --namespace or -n)")
		}

		// Validate namespace to prevent injection
		if err := validateNamespace(namespace); err != nil {
			return fmt.Errorf("invalid namespace: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fmt.Printf("Scanning namespace: %s\n", namespace)

		// Create Kubernetes client
		k8sClient, err := k8s.NewClient(cfg.Kube.Context)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		// List Deployments
		deployments, err := k8sClient.ListDeployments(ctx, namespace, cfg.Filters.LabelSelector)
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w", err)
		}

		// Filter excluded namespaces
		deployments = k8s.FilterDeployments(deployments, cfg.Filters.NamespacesExclude)

		// Print summary
		fmt.Printf("Found %d Deployments\n", len(deployments))
		for _, d := range deployments {
			fmt.Printf("  - %s/%s (%d containers)\n", d.Namespace, d.Name, len(d.Containers))
		}

		fmt.Printf("Scan completed in %.1fs\n", time.Since(time.Now()).Seconds())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
