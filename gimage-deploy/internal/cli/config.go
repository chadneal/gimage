package cli

import (
	"fmt"

	"github.com/apresai/gimage-deploy/internal/storage"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and manage gimage-deploy configuration settings.`,
}

// configGetCmd shows current configuration
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Long:  `Get a configuration value. If no key is provided, shows all configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cm := storage.NewConfigManager()
		if err := cm.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		config := cm.Get()

		if len(args) == 0 {
			// Show all configuration
			fmt.Println("Configuration:")
			fmt.Printf("  AWS Profile:         %s\n", config.AWSProfile)
			fmt.Printf("  Default Region:      %s\n", config.DefaultRegion)
			fmt.Printf("  Default Stage:       %s\n", config.DefaultStage)
			fmt.Printf("  Default Memory:      %d MB\n", config.DefaultMemoryMB)
			fmt.Printf("  Default Timeout:     %d seconds\n", config.DefaultTimeoutSec)
			fmt.Printf("  Default Concurrency: %d\n", config.DefaultConcurrency)
			fmt.Printf("  Auto Refresh:        %v\n", config.AutoRefreshMetrics)
			fmt.Printf("  Refresh Interval:    %d seconds\n", config.RefreshIntervalSec)
			fmt.Printf("  Log Retention:       %d days\n", config.LogRetentionDays)
			fmt.Printf("  Encryption:          %v\n", config.EncryptionEnabled)
			return nil
		}

		// Show specific key
		key := args[0]
		var value interface{}

		switch key {
		case "aws_profile":
			value = config.AWSProfile
		case "default_region":
			value = config.DefaultRegion
		case "default_stage":
			value = config.DefaultStage
		case "default_memory_mb":
			value = config.DefaultMemoryMB
		case "default_timeout_sec":
			value = config.DefaultTimeoutSec
		case "default_concurrency":
			value = config.DefaultConcurrency
		case "auto_refresh_metrics":
			value = config.AutoRefreshMetrics
		case "refresh_interval_sec":
			value = config.RefreshIntervalSec
		case "log_retention_days":
			value = config.LogRetentionDays
		case "encryption_enabled":
			value = config.EncryptionEnabled
		default:
			return fmt.Errorf("unknown configuration key: %s", key)
		}

		fmt.Printf("%v\n", value)
		return nil
	},
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long:  `Set a configuration value.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cm := storage.NewConfigManager()
		if err := cm.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		key := args[0]
		value := args[1]

		// Type conversion based on key
		var typedValue interface{}
		switch key {
		case "default_memory_mb", "default_timeout_sec", "default_concurrency", "refresh_interval_sec", "log_retention_days":
			var intVal int
			if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
				return fmt.Errorf("value must be an integer")
			}
			typedValue = intVal
		case "auto_refresh_metrics", "encryption_enabled":
			var boolVal bool
			if _, err := fmt.Sscanf(value, "%t", &boolVal); err != nil {
				return fmt.Errorf("value must be a boolean (true/false)")
			}
			typedValue = boolVal
		default:
			typedValue = value
		}

		if err := cm.Set(key, typedValue); err != nil {
			return err
		}

		fmt.Printf("Set %s = %v\n", key, typedValue)
		return nil
	},
}

// configResetCmd resets configuration to defaults
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		cm := storage.NewConfigManager()
		if err := cm.Reset(); err != nil {
			return fmt.Errorf("failed to reset config: %w", err)
		}

		fmt.Println("Configuration reset to defaults")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
}
