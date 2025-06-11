package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"hypr-input-switcher/internal/app"
	"hypr-input-switcher/pkg/logger"
)

var rootCmd = &cobra.Command{
	Use:   "hypr-input-switcher",
	Short: "Hyprland input method switcher",
	Long:  "Automatically switches input methods based on active window in Hyprland",
	Run:   runApp,
}

func init() {
	// Get default config path
	defaultConfigPath := getDefaultConfigPath()

	// Add flags
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warning, error)")
	rootCmd.PersistentFlags().StringP("config", "c", defaultConfigPath, fmt.Sprintf("config file (default: %s)", defaultConfigPath))
	rootCmd.PersistentFlags().Bool("log-stdout", false, "Force log output to stdout")
	rootCmd.PersistentFlags().BoolP("watch", "w", false, "Watch config file for changes and hot reload")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log.stdout", rootCmd.PersistentFlags().Lookup("log-stdout"))
	viper.BindPFlag("watch", rootCmd.PersistentFlags().Lookup("watch"))

	// Bind environment variables
	viper.SetEnvPrefix("HYPR_INPUT_SWITCHER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("config", defaultConfigPath)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.stdout", true)
	viper.SetDefault("watch", false)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runApp(cmd *cobra.Command, args []string) {
	// Setup logging first
	setupLogging()

	// Get configuration values
	configPath := viper.GetString("config")
	watchConfig := viper.GetBool("watch")

	// Expand environment variables in config path
	configPath = os.ExpandEnv(configPath)

	logger.Debugf("Using config path: %s", configPath)

	// Initialize and run the application with config path
	application := app.NewApplication()
	if err := application.Run(configPath, watchConfig); err != nil {
		logger.Errorf("Application failed to run: %v", err)
		os.Exit(1)
	}
}

func setupLogging() {
	// Get log level from viper
	logLevel := viper.GetString("log.level")
	logStdout := viper.GetBool("log.stdout")

	// Set log level
	logger.SetLevel(logLevel)

	// Set output to stdout if requested
	if logStdout {
		logger.SetOutput(os.Stdout)
	}

	logger.Infof("Log level set to: %s", logLevel)
	logger.Infof("Logging to stdout: %v", logStdout)
}

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if can't get home directory
		return "./config.yaml"
	}
	return filepath.Join(homeDir, ".config", "hypr-input-switcher", "config.yaml")
}
