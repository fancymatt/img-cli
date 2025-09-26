package cmd

import (
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	workflowTest     string
	workflowOutfitRef string
	workflowStyleRef  string
	workflowPrompt    string
	workflowSendOrig  bool
	workflowVariations int
)

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:   "workflow <type> <input> [options]",
	Short: "Run image processing workflows",
	Long: `Run complex image processing workflows that combine multiple operations.

Available workflows:
  - outfit-variations: Generate multiple outfit variations
  - style-transfer: Apply visual styles to images
  - complete-transformation: Full outfit + style change
  - cross-reference: Combine outfit and style references
  - outfit-swap: Apply outfit to multiple subjects
  - use-art-style: Apply artistic style to image or text
  - analyze-style: Analyze artistic style
  - create-style-guide: Create a style guide from references`,
	Args: cobra.MinimumNArgs(2),
	RunE: runWorkflow,
}

func init() {
	rootCmd.AddCommand(workflowCmd)

	workflowCmd.Flags().StringVar(&workflowTest, "test", "", "Test on single subject from directory")
	workflowCmd.Flags().StringVar(&workflowOutfitRef, "outfit-ref", "", "Path to outfit reference")
	workflowCmd.Flags().StringVar(&workflowStyleRef, "style-ref", "", "Path to style reference")
	workflowCmd.Flags().StringVar(&workflowPrompt, "prompt", "", "Additional prompt text")
	workflowCmd.Flags().BoolVar(&workflowSendOrig, "send-original", false, "Include reference images in requests")
	workflowCmd.Flags().IntVar(&workflowVariations, "variations", 1, "Number of variations to generate per combination")
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	workflowType := args[0]
	inputPath := args[1]

	// Check if input is a file or text prompt (for certain workflows)
	isTextPrompt := false
	if workflowType == "use-art-style" {
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			isTextPrompt = true
		}
	}

	if !isTextPrompt {
		// Validate input path
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return errors.ErrFileNotFound(inputPath)
		}
	}

	// Validate reference paths if provided
	if workflowOutfitRef != "" {
		if _, err := os.Stat(workflowOutfitRef); os.IsNotExist(err) {
			return errors.ErrFileNotFound(workflowOutfitRef)
		}
	}
	if workflowStyleRef != "" {
		if _, err := os.Stat(workflowStyleRef); os.IsNotExist(err) {
			return errors.ErrFileNotFound(workflowStyleRef)
		}
	}

	orchestrator := workflow.NewOrchestrator(apiKey)

	logger.Info("Starting workflow",
		"type", workflowType,
		"input", func() string {
			if isTextPrompt {
				return "text prompt"
			}
			return filepath.Base(inputPath)
		}())

	// Set default output directory with timestamp
	now := time.Now()
	dateFolder := now.Format("2006-01-02")
	timestampFolder := now.Format("150405")
	outputDir := filepath.Join("output", dateFolder, timestampFolder)

	options := workflow.WorkflowOptions{
		OutputDir:       outputDir,
		OutfitReference: workflowOutfitRef,
		StyleReference:  workflowStyleRef,
		Prompt:          workflowPrompt,
		SendOriginal:    workflowSendOrig,
		Variations:      workflowVariations,
	}

	// For outfit-swap workflow with --test flag, set the target image
	if workflowType == "outfit-swap" && workflowTest != "" {
		// Construct the full path to the test subject
		subjectsDir := "subjects"
		testSubjectPath := filepath.Join(subjectsDir, workflowTest)

		// Check if it needs an extension
		if _, err := os.Stat(testSubjectPath); os.IsNotExist(err) {
			// Try with common image extensions
			for _, ext := range []string{".png", ".jpg", ".jpeg"} {
				tryPath := testSubjectPath + ext
				if _, err := os.Stat(tryPath); err == nil {
					testSubjectPath = tryPath
					break
				}
			}
		}

		// Verify the file exists
		if _, err := os.Stat(testSubjectPath); os.IsNotExist(err) {
			return errors.ErrFileNotFound(testSubjectPath)
		}

		options.TargetImage = testSubjectPath
		logger.Info("Using test subject", "path", testSubjectPath)
	}

	result, err := orchestrator.RunWorkflow(workflowType, inputPath, options)
	if err != nil {
		return errors.Wrapf(err, errors.WorkflowError, "workflow %s failed", workflowType)
	}

	// Display results
	fmt.Printf("\n✓ Workflow completed successfully\n")
	fmt.Printf("Workflow: %s\n", result.Workflow)
	fmt.Printf("Duration: %s\n", result.EndTime.Sub(result.StartTime))

	if len(result.Steps) > 0 {
		fmt.Printf("\nCompleted %d steps:\n", len(result.Steps))
		for _, step := range result.Steps {
			if step.OutputPath != "" {
				fmt.Printf("  - %s: %s\n", step.Name, filepath.Base(step.OutputPath))
			} else if step.Message != "" {
				fmt.Printf("  - %s: %s\n", step.Name, step.Message)
			} else {
				fmt.Printf("  - %s\n", step.Name)
			}
		}
	}

	logger.Info("Workflow completed successfully",
		"type", workflowType,
		"steps", len(result.Steps))

	return nil
}