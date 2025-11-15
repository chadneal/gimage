package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/apresai/gimage-deploy/internal/apikeys"
	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/models"
	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/apresai/gimage-deploy/pkg/utils"
	"github.com/spf13/cobra"
)

// keysCmd represents the keys command
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long:  `Create, list, update, and delete API Gateway API keys.`,
}

// keysListCmd lists all API keys
var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Long:  `List all API keys with their status and usage information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		am := storage.NewAPIKeyManager()
		if err := am.Load(); err != nil {
			return fmt.Errorf("failed to load API keys: %w", err)
		}

		deploymentFilter, _ := cmd.Flags().GetString("deployment")

		var keys []*models.APIKey
		if deploymentFilter != "" {
			keys = am.ListByDeployment(deploymentFilter)
		} else {
			keys = am.List()
		}

		if len(keys) == 0 {
			fmt.Println("No API keys found.")
			fmt.Println("\nCreate your first API key with:")
			fmt.Println("  gimage-deploy keys create --name <key-name> --deployment <deployment-id>")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDEPLOYMENT\tSTATUS\tKEY VALUE")
		fmt.Fprintln(w, "────\t──────────\t──────\t─────────")

		for _, key := range keys {
			maskedKey := utils.MaskAPIKey(key.KeyValue)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				key.Name,
				key.DeploymentID,
				key.Status,
				maskedKey,
			)
		}

		w.Flush()
		fmt.Printf("\nTotal: %d API key(s)\n", len(keys))
		return nil
	},
}

// keysCreateCmd creates a new API key
var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		deploymentID, _ := cmd.Flags().GetString("deployment")
		description, _ := cmd.Flags().GetString("description")
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")
		burstLimit, _ := cmd.Flags().GetInt("burst-limit")
		quotaLimit, _ := cmd.Flags().GetInt("quota-limit")

		// Validate
		if err := utils.ValidateAPIKeyName(name); err != nil {
			return err
		}

		// Load AWS config
		ctx := context.Background()
		cfg, err := aws.LoadConfig(ctx, awsProfile, awsRegion)
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Create API key manager
		mgr := apikeys.NewManager(cfg)

		// Create API key
		_, err = mgr.Create(ctx, apikeys.CreateInput{
			Name:         name,
			DeploymentID: deploymentID,
			Description:  description,
			RateLimit:    int32(rateLimit),
			BurstLimit:   int32(burstLimit),
			QuotaLimit:   int32(quotaLimit),
		})

		return err
	},
}

// keysDeleteCmd deletes an API key
var keysDeleteCmd = &cobra.Command{
	Use:   "delete <key-id>",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]

		am := storage.NewAPIKeyManager()
		if err := am.Load(); err != nil {
			return fmt.Errorf("failed to load API keys: %w", err)
		}

		if !am.Exists(keyID) {
			return fmt.Errorf("API key %s not found", keyID)
		}

		// Confirmation
		fmt.Printf("Are you sure you want to delete API key '%s'? (y/N): ", keyID)
		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}

		if err := am.Delete(keyID); err != nil {
			return fmt.Errorf("failed to delete API key: %w", err)
		}

		fmt.Printf("API key '%s' deleted successfully\n", keyID)
		return nil
	},
}

func init() {
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysCreateCmd)
	keysCmd.AddCommand(keysDeleteCmd)

	// Flags for list
	keysListCmd.Flags().String("deployment", "", "Filter by deployment ID")

	// Flags for create
	keysCreateCmd.Flags().StringP("name", "n", "", "API key name (required)")
	keysCreateCmd.Flags().StringP("deployment", "d", "", "Deployment ID (required)")
	keysCreateCmd.Flags().String("description", "", "API key description")
	keysCreateCmd.Flags().Int("rate-limit", 100, "Rate limit (requests per second)")
	keysCreateCmd.Flags().Int("burst-limit", 200, "Burst limit")
	keysCreateCmd.Flags().Int("quota-limit", 5000, "Daily quota limit")

	keysCreateCmd.MarkFlagRequired("name")
	keysCreateCmd.MarkFlagRequired("deployment")
}
