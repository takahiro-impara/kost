package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/takahiro-impara/kost/internal/engine"
	"github.com/takahiro-impara/kost/internal/k8s"
	"github.com/takahiro-impara/kost/internal/metrics"
	"github.com/takahiro-impara/kost/internal/patch"
	"github.com/takahiro-impara/kost/internal/report"
	"github.com/takahiro-impara/kost/internal/security"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate optimization report",
	Long: `Generate a comprehensive optimization report including:
- Markdown report with top findings
- JSON summary for programmatic access
- YAML patches ready to apply`,
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

		fmt.Printf("Generating report for namespace: %s\n", namespace)

		// Create Kubernetes client
		k8sClient, err := k8s.NewClient(cfg.Kube.Context)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		// Create Prometheus client
		promClient, err := metrics.NewClient(cfg.Prometheus.URL, cfg.Prometheus.TimeoutSeconds)
		if err != nil {
			return fmt.Errorf("failed to create prometheus client: %w", err)
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
			return fmt.Errorf("failed to list deployments: %w", err)
		}

		// Filter excluded namespaces
		deployments = k8s.FilterDeployments(deployments, cfg.Filters.NamespacesExclude)

		if len(deployments) == 0 {
			fmt.Printf("No Deployments found in namespace: %s\n", namespace)
			return nil
		}

		fmt.Printf("Found %d Deployments\n", len(deployments))

		// Analyze and generate recommendations
		fmt.Println("Analyzing metrics...")
		var recommendations []engine.ResourceRecommendation

		for _, deployment := range deployments {
			for _, container := range deployment.Containers {
				// Query CPU metrics
				cpuMetrics, err := promClient.QueryCPUMetrics(ctx, namespace, deployment.Name, container.Name, window)
				if err != nil {
					fmt.Printf("  Warning: Failed to query CPU metrics for %s/%s: %v\n", deployment.Name, container.Name, err)
					continue
				}

				// Query Memory metrics
				memMetrics, err := promClient.QueryMemoryMetrics(ctx, namespace, deployment.Name, container.Name, window)
				if err != nil {
					fmt.Printf("  Warning: Failed to query Memory metrics for %s/%s: %v\n", deployment.Name, container.Name, err)
					continue
				}

				// Calculate statistics
				if len(cpuMetrics.Values) == 0 || len(memMetrics.Values) == 0 {
					fmt.Printf("  Warning: No metrics found for %s/%s\n", deployment.Name, container.Name)
					continue
				}

				if err := cpuMetrics.CalculateStatistics(); err != nil {
					fmt.Printf("  Warning: Failed to calculate CPU statistics for %s/%s: %v\n", deployment.Name, container.Name, err)
					continue
				}

				if err := memMetrics.CalculateStatistics(); err != nil {
					fmt.Printf("  Warning: Failed to calculate Memory statistics for %s/%s: %v\n", deployment.Name, container.Name, err)
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
			}
		}

		if len(recommendations) == 0 {
			fmt.Println("No recommendations generated (no metrics available)")
			return nil
		}

		fmt.Printf("Generated %d recommendations\n", len(recommendations))

		// Create Report
		rep := &report.Report{
			Namespace:               namespace,
			Window:                  cfg.Analysis.Window,
			Percentile:              fmt.Sprintf("P%.0f", cfg.Analysis.CPUPercentile*100),
			SafetyFactor:            cfg.Analysis.SafetyFactor,
			ResourceRecommendations: recommendations,
			HPARecommendations:      []engine.HPARecommendation{}, // TODO: Implement HPA recommendations
			Narrative:               "",                           // TODO: Implement LLM narrative
			GeneratedAt:             time.Now(),
		}

		// Create writer
		writer, err := report.NewWriter(cfg.Output.Dir)
		if err != nil {
			return fmt.Errorf("failed to create report writer: %w", err)
		}

		// Generate outputs based on configured formats
		for _, format := range cfg.Output.Format {
			switch format {
			case "md", "markdown":
				fmt.Println("Generating Markdown report...")
				markdown, err := report.GenerateMarkdown(rep)
				if err != nil {
					return fmt.Errorf("failed to generate markdown: %w", err)
				}
				if err := writer.WriteMarkdown("report.md", markdown); err != nil {
					return fmt.Errorf("failed to write markdown: %w", err)
				}
				fmt.Printf("  ✓ Written to: %s\n", writer.GetFilePath("report.md"))

			case "json":
				fmt.Println("Generating JSON summary...")
				jsonData, err := report.GenerateJSON(rep)
				if err != nil {
					return fmt.Errorf("failed to generate JSON: %w", err)
				}
				if err := writer.WriteJSON("summary.json", jsonData); err != nil {
					return fmt.Errorf("failed to write JSON: %w", err)
				}
				fmt.Printf("  ✓ Written to: %s\n", writer.GetFilePath("summary.json"))

			case "patch":
				fmt.Println("Generating YAML patches...")
				// Group recommendations by deployment
				recsByDeployment := make(map[string][]engine.ResourceRecommendation)
				for _, rec := range recommendations {
					recsByDeployment[rec.Deployment] = append(recsByDeployment[rec.Deployment], rec)
				}

				// Generate patch for each deployment
				patchCount := 0
				for deploymentName, recs := range recsByDeployment {
					patchYAML, err := patch.GenerateResourcePatch(deploymentName, recs)
					if err != nil {
						fmt.Printf("  Warning: Failed to generate patch for %s: %v\n", deploymentName, err)
						continue
					}

					// Validate patch
					if err := patch.ValidatePatch(patchYAML); err != nil {
						fmt.Printf("  Warning: Invalid patch for %s: %v\n", deploymentName, err)
						continue
					}

					// Write patch
					filename := fmt.Sprintf("%s.yaml", deploymentName)
					if err := writer.WritePatch(fmt.Sprintf("patches/%s", namespace), filename, patchYAML); err != nil {
						fmt.Printf("  Warning: Failed to write patch for %s: %v\n", deploymentName, err)
						continue
					}

					fmt.Printf("  ✓ Patch written: %s\n", writer.GetPatchFilePath(fmt.Sprintf("patches/%s", namespace), filename))
					patchCount++
				}

				if patchCount > 0 {
					fmt.Printf("  Total: %d patches generated\n", patchCount)
				}
			}
		}

		fmt.Printf("\n✓ Report generation complete (%.1fs)\n", time.Since(startTime).Seconds())
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. Review the report: cat %s\n", writer.GetFilePath("report.md"))
		fmt.Printf("  2. Verify patches: kubectl apply --dry-run=client -f %s\n", writer.GetPatchFilePath(fmt.Sprintf("patches/%s", namespace), ""))
		fmt.Printf("  3. Apply when ready: kubectl apply -f %s\n", writer.GetPatchFilePath(fmt.Sprintf("patches/%s", namespace), ""))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
