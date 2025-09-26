package config

import (
	"fmt"
	"os"
	"strconv"
)

// CostConfig holds the configuration for cost tracking and limits
type CostConfig struct {
	// Cost per image generation in dollars
	CostPerImage float64

	// Threshold for requiring user confirmation in dollars
	ConfirmationThreshold float64

	// Maximum allowed cost without override in dollars
	MaximumCost float64
}

// DefaultCostConfig returns the default cost configuration
// These values can be overridden via environment variables:
// - IMG_CLI_COST_PER_IMAGE (default: 0.04)
// - IMG_CLI_CONFIRM_THRESHOLD (default: 5.00)
// - IMG_CLI_MAX_COST (default: 50.00)
func DefaultCostConfig() *CostConfig {
	config := &CostConfig{
		CostPerImage:          0.04,  // $0.04 per image
		ConfirmationThreshold: 5.00,  // Confirm if over $5
		MaximumCost:           50.00, // Hard limit at $50
	}

	// Allow environment variable overrides
	if envCost := getEnvFloat("IMG_CLI_COST_PER_IMAGE", 0); envCost > 0 {
		config.CostPerImage = envCost
	}
	if envThreshold := getEnvFloat("IMG_CLI_CONFIRM_THRESHOLD", 0); envThreshold > 0 {
		config.ConfirmationThreshold = envThreshold
	}
	if envMax := getEnvFloat("IMG_CLI_MAX_COST", 0); envMax > 0 {
		config.MaximumCost = envMax
	}

	return config
}

// CalculateTotalCost calculates the total cost for a given number of images
func (c *CostConfig) CalculateTotalCost(imageCount int) float64 {
	return float64(imageCount) * c.CostPerImage
}

// RequiresConfirmation checks if the cost requires user confirmation
func (c *CostConfig) RequiresConfirmation(imageCount int) bool {
	return c.CalculateTotalCost(imageCount) > c.ConfirmationThreshold
}

// FormatCost formats a cost value as a string
func (c *CostConfig) FormatCost(cost float64) string {
	return fmt.Sprintf("$%.2f", cost)
}

// GetCostBreakdown returns a formatted string explaining the cost calculation
func (c *CostConfig) GetCostBreakdown(imageCount int) string {
	totalCost := c.CalculateTotalCost(imageCount)
	return fmt.Sprintf("%d images × %s = %s",
		imageCount,
		c.FormatCost(c.CostPerImage),
		c.FormatCost(totalCost))
}

// getEnvFloat reads a float value from environment variable
func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultValue
}