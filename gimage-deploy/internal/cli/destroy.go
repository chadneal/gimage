package cli

import (
	"context"
	"fmt"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/deploy"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy <deployment-id>",
	Short: "Destroy a deployment",
	Long: `Destroy a deployment and all associated AWS resources.

This will delete:
- API Gateway REST API
- Lambda function
- S3 bucket (after emptying)
- IAM role and policies

WARNING: This action cannot be undone!`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Confirmation prompt
		if !skipConfirm {
			fmt.Printf("⚠️  WARNING: This will permanently delete deployment '%s' and all associated resources.\n", deploymentID)
			fmt.Printf("This action cannot be undone!\n\n")
			fmt.Printf("Are you sure you want to continue? (yes/no): ")

			var response string
			fmt.Scanln(&response)

			if response != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		// Load AWS config
		ctx := context.Background()
		cfg, err := aws.LoadConfig(ctx, awsProfile, awsRegion)
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Create deployment manager
		mgr := deploy.NewManager(cfg)

		// Destroy
		return mgr.Destroy(ctx, deploymentID)
	},
}

func init() {
	destroyCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(destroyCmd)
}
