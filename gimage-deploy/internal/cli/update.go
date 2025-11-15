package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <deployment-id>",
	Short: "Update a deployment",
	Long: `Update an existing deployment's configuration.

You can update:
- Lambda memory, timeout, and concurrency
- Environment variables
- Function code`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]

		// Get flags
		memory, _ := cmd.Flags().GetInt("memory")
		timeout, _ := cmd.Flags().GetInt("timeout")
		envVars, _ := cmd.Flags().GetStringToString("env")
		code, _ := cmd.Flags().GetString("code")

		// Load deployment
		dm := storage.NewDeploymentManager()
		if err := dm.Load(); err != nil {
			return fmt.Errorf("failed to load deployments: %w", err)
		}

		deployment, err := dm.Get(deploymentID)
		if err != nil {
			return err
		}

		// Load AWS config
		ctx := context.Background()
		cfg, err := aws.LoadConfig(ctx, awsProfile, awsRegion)
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		lambdaClient := aws.NewLambdaClient(cfg)

		// Update code if specified
		if code != "" {
			fmt.Printf("Updating function code...\n")
			codeBytes, err := os.ReadFile(code)
			if err != nil {
				return fmt.Errorf("failed to read code file: %w", err)
			}
			if err := lambdaClient.UpdateFunctionCode(ctx, deployment.FunctionName, codeBytes); err != nil {
				return fmt.Errorf("failed to update code: %w", err)
			}
			fmt.Printf("✓ Code updated\n")
		}

		// Update configuration if any flags were set
		updateConfig := memory > 0 || timeout > 0 || len(envVars) > 0

		if updateConfig {
			fmt.Printf("Updating function configuration...\n")

			// Use existing values if not specified
			memoryMB := int32(deployment.Configuration.MemoryMB)
			if memory > 0 {
				memoryMB = int32(memory)
			}

			timeoutSec := int32(deployment.Configuration.TimeoutSeconds)
			if timeout > 0 {
				timeoutSec = int32(timeout)
			}

			// Merge environment variables
			env := deployment.EnvironmentVars
			if env == nil {
				env = make(map[string]string)
			}
			for k, v := range envVars {
				env[k] = v
			}

			if err := lambdaClient.UpdateFunctionConfiguration(ctx, deployment.FunctionName, memoryMB, timeoutSec, env); err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}

			// Update local storage
			if memory > 0 {
				deployment.Configuration.MemoryMB = memory
			}
			if timeout > 0 {
				deployment.Configuration.TimeoutSeconds = timeout
			}
			if len(envVars) > 0 {
				deployment.EnvironmentVars = env
			}

			if err := dm.Update(deployment); err != nil {
				return fmt.Errorf("failed to update local storage: %w", err)
			}

			fmt.Printf("✓ Configuration updated\n")
		}

		if !updateConfig && code == "" {
			fmt.Printf("No updates specified. Use flags like --memory, --timeout, --env, or --code\n")
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().IntP("memory", "m", 0, "Lambda memory in MB")
	updateCmd.Flags().IntP("timeout", "t", 0, "Lambda timeout in seconds")
	updateCmd.Flags().StringToString("env", nil, "Environment variables (key=value)")
	updateCmd.Flags().String("code", "", "Path to new Lambda code (zip)")

	rootCmd.AddCommand(updateCmd)
}
