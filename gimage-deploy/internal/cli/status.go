package cli

import (
	"fmt"

	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status <deployment-id>",
	Short: "Show deployment status",
	Long:  `Display detailed information about a deployment including configuration, health, and endpoints.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		dm := storage.NewDeploymentManager()
		if err := dm.Load(); err != nil {
			return fmt.Errorf("failed to load deployments: %w", err)
		}

		deployment, err := dm.Get(deploymentID)
		if err != nil {
			return err
		}

		// Display deployment info
		fmt.Printf("Deployment: %s\n", deployment.ID)
		fmt.Printf("─────────────────────────────────────────\n\n")

		fmt.Printf("Status:    %s\n", deployment.Status)
		fmt.Printf("Region:    %s\n", deployment.Region)
		fmt.Printf("Stage:     %s\n", deployment.Stage)
		fmt.Printf("Created:   %s\n", deployment.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:   %s\n", deployment.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("\n")

		fmt.Printf("Resources:\n")
		fmt.Printf("  Function:  %s\n", deployment.FunctionName)
		fmt.Printf("  API ID:    %s\n", deployment.APIGatewayID)
		fmt.Printf("  S3 Bucket: %s\n", deployment.S3Bucket)
		fmt.Printf("\n")

		fmt.Printf("Endpoint:\n")
		fmt.Printf("  %s\n", deployment.APIGatewayURL)
		fmt.Printf("\n")

		fmt.Printf("Configuration:\n")
		fmt.Printf("  Memory:       %d MB\n", deployment.Configuration.MemoryMB)
		fmt.Printf("  Timeout:      %d seconds\n", deployment.Configuration.TimeoutSeconds)
		fmt.Printf("  Concurrency:  %d\n", deployment.Configuration.Concurrency)
		fmt.Printf("  Architecture: %s\n", deployment.Configuration.Architecture)
		fmt.Printf("  Runtime:      %s\n", deployment.Configuration.Runtime)
		fmt.Printf("\n")

		if len(deployment.EnvironmentVars) > 0 {
			fmt.Printf("Environment Variables: (%d)\n", len(deployment.EnvironmentVars))
			for key := range deployment.EnvironmentVars {
				fmt.Printf("  %s: ***\n", key)
			}
			fmt.Printf("\n")
		}

		fmt.Printf("Health:\n")
		fmt.Printf("  Healthy: %v\n", deployment.Health.IsHealthy)
		fmt.Printf("  Score:   %d/100\n", deployment.Health.Score)
		if deployment.Health.LastChecked.IsZero() {
			fmt.Printf("  Last Check: Never\n")
		} else {
			fmt.Printf("  Last Check: %s\n", deployment.Health.LastChecked.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
