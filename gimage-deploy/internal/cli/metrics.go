package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics <deployment-id>",
	Short: "Show CloudWatch metrics",
	Long:  `Display CloudWatch metrics for a deployment including invocations, errors, duration, and throttles.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]
		period, _ := cmd.Flags().GetString("period")

		// Parse period
		var duration time.Duration
		switch period {
		case "1h":
			duration = 1 * time.Hour
		case "6h":
			duration = 6 * time.Hour
		case "24h":
			duration = 24 * time.Hour
		case "7d":
			duration = 7 * 24 * time.Hour
		default:
			return fmt.Errorf("invalid period: %s (use 1h, 6h, 24h, or 7d)", period)
		}

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

		cwClient := aws.NewCloudWatchClient(cfg)

		// Get metrics
		endTime := time.Now()
		startTime := endTime.Add(-duration)

		fmt.Printf("Metrics for %s (last %s):\n\n", deployment.FunctionName, period)

		metrics, err := cwClient.GetLambdaMetrics(ctx, deployment.FunctionName, startTime, endTime)
		if err != nil {
			return fmt.Errorf("failed to get metrics: %w", err)
		}

		// Display metrics
		invocations := int64(metrics["invocations"])
		errors := int64(metrics["errors"])
		throttles := int64(metrics["throttles"])
		avgDuration := metrics["avg_duration"]
		concurrentExec := int(metrics["concurrent_executions"])

		fmt.Printf("Invocations:          %d\n", invocations)
		fmt.Printf("Errors:               %d", errors)
		if invocations > 0 {
			errorRate := float64(errors) / float64(invocations) * 100
			fmt.Printf(" (%.2f%%)\n", errorRate)
		} else {
			fmt.Printf("\n")
		}
		fmt.Printf("Throttles:            %d\n", throttles)
		fmt.Printf("Avg Duration:         %.2f ms\n", avgDuration)
		fmt.Printf("Concurrent Exec:      %d / %d\n", concurrentExec, deployment.Configuration.Concurrency)
		fmt.Printf("\n")

		// Calculate error rate
		if invocations > 0 {
			successRate := float64(invocations-errors) / float64(invocations) * 100
			fmt.Printf("Success Rate:         %.2f%%\n", successRate)
		}

		return nil
	},
}

func init() {
	metricsCmd.Flags().StringP("period", "p", "24h", "Time period (1h, 6h, 24h, 7d)")

	rootCmd.AddCommand(metricsCmd)
}
