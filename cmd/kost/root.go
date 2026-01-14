package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/your-org/kost/internal/config"
)

var (
	cfgFile   string
	namespace string
	cfg       *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "kost",
	Short: "Kubernetes Optimization & Sizing Tool",
	Long: `kost (Kubernetes Optimization & Sizing Tool) analyzes Kubernetes Deployments
and provides optimization recommendations for resource requests/limits and HPA settings.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return nil
	},
}

func init() {
	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "target namespace")
}
