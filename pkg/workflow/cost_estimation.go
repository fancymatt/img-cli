package workflow

import (
	"fmt"
	"img-cli/pkg/config"
	"img-cli/pkg/prompt"
)

// calculateOutfitSwapImageCount calculates how many images will be generated
func calculateOutfitSwapImageCount(numSubjects, numOutfits, numStyles, numVariations int) int {
	// Default values if not specified
	if numVariations < 1 {
		numVariations = 1
	}
	if numStyles < 1 {
		numStyles = 1
	}

	return numSubjects * numOutfits * numStyles * numVariations
}

// checkWorkflowCost checks if a workflow will exceed cost thresholds and prompts for confirmation
func checkWorkflowCost(workflowName string, imageCount int, skipConfirm bool) error {
	costConfig := config.DefaultCostConfig()
	totalCost := costConfig.CalculateTotalCost(imageCount)

	// Show cost breakdown
	fmt.Printf("\n📊 Workflow Cost Analysis for %s:\n", workflowName)
	fmt.Printf("   Images to generate: %d\n", imageCount)
	fmt.Printf("   Cost breakdown: %s\n", costConfig.GetCostBreakdown(imageCount))

	// Check if confirmation is needed (unless skipped)
	if !skipConfirm && costConfig.RequiresConfirmation(imageCount) {
		message := fmt.Sprintf("This workflow will generate %d images", imageCount)
		confirmed, err := prompt.ConfirmExpensiveOperation(
			message,
			costConfig.FormatCost(totalCost),
		)
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("workflow cancelled by user")
		}
		fmt.Println("✅ Proceeding with workflow...")
	} else if imageCount > 10 {
		// For moderately large batches, still show the cost
		prompt.ShowCostEstimate(
			fmt.Sprintf("This workflow will generate %d images", imageCount),
			costConfig.FormatCost(totalCost),
		)
	}

	// Check hard limit
	if totalCost > costConfig.MaximumCost {
		return fmt.Errorf("workflow cost ($%.2f) exceeds maximum allowed ($%.2f)",
			totalCost, costConfig.MaximumCost)
	}

	return nil
}