package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/takahiro-impara/kost/internal/engine"
	"github.com/takahiro-impara/kost/internal/k8s"
	"github.com/takahiro-impara/kost/internal/metrics"
	"github.com/takahiro-impara/kost/internal/security"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Generate resource optimization suggestions",
	Long: `Analyze Prometheus metrics and generate CPU/Memory optimization suggestions.
This command queries Prometheus for historical metrics and calculates recommended values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		if namespace == "" {
			return fmt.Errorf("namespace is required (use --namespace or -n)")
		}

		// Validate namespace to prevent injection
		if err := security.ValidateNamespace(namespace); err != nil {
			return fmt.Errorf("invalid namespace: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		fmt.Printf("Analyzing namespace: %s\n", namespace)
		fmt.Printf("Analysis window: %s\n", cfg.Analysis.Window)
		fmt.Printf("CPU percentile: P%.0f, Memory percentile: P%.0f\n",
			cfg.Analysis.CPUPercentile*100, cfg.Analysis.MemPercentile*100)
		fmt.Printf("Safety factor: %.1f\n\n", cfg.Analysis.SafetyFactor)

		// Create Kubernetes client
		k8sClient, err := k8s.NewClient(cfg.Kube.Context)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		// Create Prometheus client
		promClient, err := metrics.NewClient(cfg.Prometheus.URL, cfg.Prometheus.TimeoutSeconds)
		if err != nil {
			return fmt.Errorf("failed to create prometheus client: %w\nPlease verify:\n  - Prometheus URL is accessible: %s\n  - Prometheus is running and healthy\n  - Network connectivity to Prometheus", err, cfg.Prometheus.URL)
		}

		// Parse analysis window
		window, err := time.ParseDuration(cfg.Analysis.Window)
		if err != nil {
			return fmt.Errorf("invalid analysis window: %s (use format like '7d', '168h', etc.): %w", cfg.Analysis.Window, err)
		}

		// List Deployments
		fmt.Println("Fetching Deployments...")
		deployments, err := k8sClient.ListDeployments(ctx, namespace, cfg.Filters.LabelSelector)
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w\nPlease verify:\n  - Namespace exists: kubectl get namespace %s\n  - RBAC permissions: kubectl auth can-i list deployments -n %s\n  - Kubeconfig is valid: kubectl config view", err, namespace, namespace)
		}

		// Filter excluded namespaces
		deployments = k8s.FilterDeployments(deployments, cfg.Filters.NamespacesExclude)

		if len(deployments) == 0 {
			fmt.Printf("No Deployments found in namespace: %s\n", namespace)
			fmt.Println("Please verify:")
			fmt.Printf("  - Deployments exist: kubectl get deployments -n %s\n", namespace)
			if cfg.Filters.LabelSelector != "" {
				fmt.Printf("  - Label selector matches: %s\n", cfg.Filters.LabelSelector)
			}
			return nil
		}

		fmt.Printf("Found %d Deployments\n", len(deployments))

		// Analyze each Deployment
		var recommendations []engine.ResourceRecommendation
		overprovisioned := 0
		underprovisioned := 0

		for _, deployment := range deployments {
			fmt.Printf("  Analyzing %s/%s (%d containers)...\n",
				deployment.Namespace, deployment.Name, len(deployment.Containers))

			for _, container := range deployment.Containers {
				// Query CPU metrics
				cpuMetrics, err := promClient.QueryCPUMetrics(ctx, namespace, deployment.Name, container.Name, window)
				if err != nil {
					fmt.Printf("    Warning: Failed to query CPU metrics for %s: %v\n", container.Name, err)
					continue
				}

				// Query Memory metrics
				memMetrics, err := promClient.QueryMemoryMetrics(ctx, namespace, deployment.Name, container.Name, window)
				if err != nil {
					fmt.Printf("    Warning: Failed to query Memory metrics for %s: %v\n", container.Name, err)
					continue
				}

				// Calculate statistics
				if len(cpuMetrics.Values) == 0 || len(memMetrics.Values) == 0 {
					fmt.Printf("    Warning: No metrics found for %s (window: %s)\n", container.Name, cfg.Analysis.Window)
					fmt.Println("    Please verify:")
					fmt.Printf("      - Pod is running: kubectl get pods -n %s -l app=%s\n", namespace, deployment.Name)
					fmt.Println("      - Metrics are being scraped by Prometheus")
					fmt.Printf("      - Metrics retention period (need at least %s)\n", cfg.Analysis.Window)
					continue
				}

				if err := cpuMetrics.CalculateStatistics(); err != nil {
					fmt.Printf("    Warning: Failed to calculate CPU statistics for %s: %v\n", container.Name, err)
					continue
				}

				if err := memMetrics.CalculateStatistics(); err != nil {
					fmt.Printf("    Warning: Failed to calculate Memory statistics for %s: %v\n", container.Name, err)
					continue
				}

				// Generate recommendation
				rec := engine.GenerateResourceRecommendation(
					deployment.Name,
					container.Name,
					container.CPURequestMilli,
					container.MemRequestMi,
					cpuMetrics,
					memMetrics,
					cfg.Analysis.SafetyFactor,
					cfg.Analysis.MinCPUMilli,
					cfg.Analysis.MinMemMi,
				)

				recommendations = append(recommendations, *rec)

				// Count judgements
				if rec.CPUJudgement == "overprovisioned" || rec.MemJudgement == "overprovisioned" {
					overprovisioned++
				}
				if rec.CPUJudgement == "underprovisioned" || rec.MemJudgement == "underprovisioned" {
					underprovisioned++
				}
			}
		}

		// Print summary
		fmt.Printf("\nAnalysis complete!\n")
		fmt.Printf("  Total containers analyzed: %d\n", len(recommendations))
		fmt.Printf("  Overprovisioned containers: %d\n", overprovisioned)
		fmt.Printf("  Underprovisioned containers: %d\n", underprovisioned)
		fmt.Printf("  Appropriate containers: %d\n", len(recommendations)-overprovisioned-underprovisioned)
		fmt.Printf("\nCompleted in %.1fs\n", time.Since(startTime).Seconds())
		fmt.Printf("\nNext step: Run 'kost report -n %s' to generate detailed report\n", namespace)

		// Store recommendations in a temporary location for report command
		// In a real implementation, this could be saved to a file or passed via context
		// For now, we just print the next step

		return nil
	},
}

func init() {
	rootCmd.AddCommand(suggestCmd)
}
