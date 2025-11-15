package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	noColor bool
	jsonOutput bool
	awsProfile string
	awsRegion string
	appVersion string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gimage-deploy",
	Short: "Manage gimage Lambda deployments and API keys",
	Long: `gimage-deploy is a CLI tool for managing gimage Lambda deployments and API Gateway API keys.

It provides both interactive TUI (via 'gimage-deploy tui') and headless CLI modes
for deployment lifecycle management, monitoring, and API key administration.`,
	SilenceUsage: true,
}

// Execute executes the root command
func Execute(version string) error {
	appVersion = version
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gimage-deploy/config.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().StringVar(&awsProfile, "profile", "", "AWS profile to use")
	rootCmd.PersistentFlags().StringVar(&awsRegion, "region", "", "AWS region")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(keysCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Use default config location
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}

		viper.AddConfigPath(filepath.Join(home, ".gimage-deploy"))
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gimage-deploy version %s\n", appVersion)
	},
}
