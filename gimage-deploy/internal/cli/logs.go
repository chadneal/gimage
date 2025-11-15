package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/apresai/gimage-deploy/internal/aws"
	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <deployment-id>",
	Short: "View CloudWatch logs",
	Long:  `View CloudWatch logs for a deployment. Displays the most recent log entries.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentID := args[0]
		limit, _ := cmd.Flags().GetInt32("limit")
		follow, _ := cmd.Flags().GetBool("follow")
		filter, _ := cmd.Flags().GetString("filter")

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
		logGroupName := fmt.Sprintf("/aws/lambda/%s", deployment.FunctionName)

		if follow {
			// Tail logs in real-time
			fmt.Printf("Tailing logs for %s (Ctrl+C to stop)...\n\n", deployment.FunctionName)

			for {
				events, err := cwClient.GetLogEvents(ctx, logGroupName, limit)
				if err != nil {
					return fmt.Errorf("failed to get log events: %w", err)
				}

				for _, event := range events {
					timestamp := event.Timestamp.Format("2006-01-02 15:04:05")
					fmt.Printf("[%s] %s\n", timestamp, event.Message)
				}

				time.Sleep(2 * time.Second)
			}
		} else {
			// Get logs once
			var events []aws.LogEvent
			var err error

			if filter != "" {
				// Filter logs
				startTime := time.Now().Add(-1 * time.Hour)
				endTime := time.Now()
				events, err = cwClient.FilterLogEvents(ctx, logGroupName, filter, startTime, endTime, limit)
			} else {
				// Get recent logs
				events, err = cwClient.GetLogEvents(ctx, logGroupName, limit)
			}

			if err != nil {
				return fmt.Errorf("failed to get log events: %w", err)
			}

			if len(events) == 0 {
				fmt.Printf("No log events found for %s\n", deployment.FunctionName)
				return nil
			}

			fmt.Printf("Recent logs for %s:\n\n", deployment.FunctionName)
			for _, event := range events {
				timestamp := event.Timestamp.Format("2006-01-02 15:04:05")
				fmt.Printf("[%s] %s\n", timestamp, event.Message)
			}
			fmt.Printf("\nShowing %d log entries\n", len(events))
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().Int32P("limit", "n", 50, "Number of log entries to show")
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output (tail)")
	logsCmd.Flags().String("filter", "", "Filter pattern for logs")

	rootCmd.AddCommand(logsCmd)
}
