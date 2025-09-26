// Package cmd provides the CLI command structure for the image generation application.
// It uses the cobra library for command parsing and execution.
package cmd

import (
	"fmt"
	"img-cli/pkg/logger"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	logLevel   string
	jsonLog    bool
	configFile string
	apiKey     string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "img-cli",
	Short: "Apply outfits and styles to portrait images",
	Long: `Apply outfit and style transformations to portrait images using Gemini API.

Primary Command:
  outfit-swap - Apply an outfit to one or more test subjects with optional style

Examples:
  # Use all defaults (shearling-black outfit, plain-white style, jaimee subject)
  img-cli outfit-swap

  # Specify outfit with custom style and subjects
  img-cli outfit-swap ./outfits/suit.png -s ./styles/night.png -t "jaimee kat"

Additional Commands:
  analyze - Analyze images for outfit, visual style, or art style
  generate - Generate images with specific transformations
  cache - Manage analysis cache`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set up logging
		level := logger.ParseLevel(logLevel)
		log := logger.NewLogger(level, jsonLog)
		logger.SetDefault(log)

		// Load environment variables
		if configFile != "" {
			if err := godotenv.Load(configFile); err != nil {
				logger.Warnf("Failed to load config file %s: %v", configFile, err)
			}
		} else {
			godotenv.Load() // Try to load .env file
		}

		// Get API key from flag or environment
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}

		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required. Set via --api-key flag or GEMINI_API_KEY environment variable")
		}

		return nil
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "Log level (DEBUG, INFO, WARN, ERROR)")
	rootCmd.PersistentFlags().BoolVar(&jsonLog, "json-log", false, "Output logs in JSON format")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file path (default: .env)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Gemini API key")
}