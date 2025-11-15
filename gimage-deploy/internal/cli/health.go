package cli

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health <deployment-id>",
	Short: "Check deployment health",
	Long:  `Perform a health check on a deployment by calling the /health endpoint.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		// Load deployment
		dm := storage.NewDeploymentManager()
		if err := dm.Load(); err != nil {
			return fmt.Errorf("failed to load deployments: %w", err)
		}

		deployment, err := dm.Get(deploymentID)
		if err != nil {
			return err
		}

		fmt.Printf("Checking health for %s...\n", deployment.ID)
		fmt.Printf("Endpoint: %s/health\n\n", deployment.APIGatewayURL)

		// Make HTTP request to /health
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		healthURL := fmt.Sprintf("%s/health", deployment.APIGatewayURL)
		resp, err := client.Get(healthURL)
		if err != nil {
			fmt.Printf("❌ Health check failed: %v\n", err)
			fmt.Printf("\nDeployment may not be healthy or endpoint is not accessible.\n")
			return nil
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == 200 {
			fmt.Printf("✓ Health check passed!\n")
			fmt.Printf("  Status: %d %s\n", resp.StatusCode, resp.Status)
			fmt.Printf("  Response:\n%s\n", string(body))

			// Update health in local storage
			deployment.Health.IsHealthy = true
			deployment.Health.Score = 100
			deployment.Health.LastChecked = time.Now()
			dm.Update(deployment)
		} else {
			fmt.Printf("⚠ Health check returned non-200 status\n")
			fmt.Printf("  Status: %d %s\n", resp.StatusCode, resp.Status)
			fmt.Printf("  Response:\n%s\n", string(body))

			// Update health
			deployment.Health.IsHealthy = false
			deployment.Health.Score = 0
			deployment.Health.LastChecked = time.Now()
			deployment.Health.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
			dm.Update(deployment)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
