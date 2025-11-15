package cli

import (
	"context"
	"fmt"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/deploy"
	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/apresai/gimage-deploy/pkg/utils"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Create a new deployment",
	Long: `Create a new Lambda deployment with API Gateway and S3 bucket.

This command will:
- Create an S3 bucket for image storage
- Create an IAM role for Lambda execution
- Upload and deploy the Lambda function
- Create an API Gateway REST API
- Configure environment variables`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		id, _ := cmd.Flags().GetString("id")
		stage, _ := cmd.Flags().GetString("stage")
		region, _ := cmd.Flags().GetString("region")
		memory, _ := cmd.Flags().GetInt("memory")
		timeout, _ := cmd.Flags().GetInt("timeout")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		architecture, _ := cmd.Flags().GetString("architecture")
		envVars, _ := cmd.Flags().GetStringToString("env")
		lambdaCode, _ := cmd.Flags().GetString("lambda-code")

		// Validate inputs
		if err := utils.ValidateDeploymentID(id); err != nil {
			return err
		}

		// Use defaults from config if not specified
		if stage == "" {
			cfg := storage.NewConfigManager()
			cfg.Load()
			stage = cfg.Get().DefaultStage
		}

		if err := utils.ValidateStage(stage); err != nil {
			return err
		}
		if err := utils.ValidateMemory(memory); err != nil {
			return err
		}
		if err := utils.ValidateTimeout(timeout); err != nil {
			return err
		}
		if err := utils.ValidateConcurrency(concurrency); err != nil {
			return err
		}

		// Load AWS config
		ctx := context.Background()
		cfg, err := aws.LoadConfig(ctx, awsProfile, region)
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Create deployment manager
		mgr := deploy.NewManager(cfg)

		// Deploy
		deployment, err := mgr.Deploy(ctx, deploy.DeployInput{
			ID:             id,
			Stage:          stage,
			Region:         aws.GetRegion(cfg),
			MemoryMB:       memory,
			TimeoutSec:     timeout,
			Concurrency:    concurrency,
			Architecture:   architecture,
			Environment:    envVars,
			Description:    fmt.Sprintf("gimage deployment for %s", stage),
			LambdaCodePath: lambdaCode,
		})
		if err != nil {
			return err
		}

		fmt.Printf("\nDeployment Details:\n")
		fmt.Printf("  ID:       %s\n", deployment.ID)
		fmt.Printf("  Region:   %s\n", deployment.Region)
		fmt.Printf("  Stage:    %s\n", deployment.Stage)
		fmt.Printf("  Endpoint: %s\n", deployment.APIGatewayURL)
		fmt.Printf("\nNext Steps:\n")
		fmt.Printf("  1. Create an API key: gimage-deploy keys create --name <key-name> --deployment %s\n", deployment.ID)
		fmt.Printf("  2. Test the deployment: curl %s/health\n", deployment.APIGatewayURL)

		return nil
	},
}

func init() {
	deployCmd.Flags().StringP("id", "i", "", "Deployment ID (required)")
	deployCmd.Flags().StringP("stage", "s", "", "Stage (prod, staging, dev, test)")
	deployCmd.Flags().StringP("region", "r", "", "AWS region")
	deployCmd.Flags().IntP("memory", "m", 512, "Lambda memory in MB")
	deployCmd.Flags().IntP("timeout", "t", 30, "Lambda timeout in seconds")
	deployCmd.Flags().IntP("concurrency", "c", 10, "Reserved concurrent executions")
	deployCmd.Flags().String("architecture", "arm64", "Lambda architecture (arm64 or x86_64)")
	deployCmd.Flags().StringToString("env", nil, "Environment variables (key=value)")
	deployCmd.Flags().String("lambda-code", "", "Path to Lambda deployment package (zip)")

	deployCmd.MarkFlagRequired("id")
}
