package cmd

import (
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	outfitStyleRef    string
	outfitTestSubjects string
	outfitVariations  int
	outfitSendOriginal bool
	outfitNoConfirm   bool
	outfitDebugPrompt bool
)

// Default values for common parameters
const (
	defaultOutfit = "./outfits/shearling-black.png"
	defaultStyle  = "./styles/plain-white.png"
	defaultSubject = "jaimee"
)

// outfitSwapCmd represents the simplified outfit-swap command
var outfitSwapCmd = &cobra.Command{
	Use:   "outfit-swap [outfit]",
	Short: "Apply outfit and style to test subjects",
	Long: `Apply an outfit to one or more test subjects with optional style.

Examples:
  # Use all defaults (shearling-black outfit, plain-white style, jaimee subject)
  img-cli outfit-swap

  # Specify outfit, use default style and subject
  img-cli outfit-swap ./outfits/business-suit.png

  # Full specification with shortcuts
  img-cli outfit-swap ./outfits/suit.png -s ./styles/night.png -t "jaimee kat"

  # Directory of outfits with multiple subjects
  img-cli outfit-swap ./outfits/batch/ -t "jaimee kat izzy" -v 3

Default values:
  Outfit: ./outfits/shearling-black.png
  Style:  ./styles/plain-white.png
  Subject: jaimee (when -t is used without value)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runOutfitSwap,
}

func init() {
	rootCmd.AddCommand(outfitSwapCmd)

	// Shortcuts and full flags
	outfitSwapCmd.Flags().StringVarP(&outfitStyleRef, "style", "s", "", "Style reference image (default: ./styles/plain-white.png)")
	outfitSwapCmd.Flags().StringVarP(&outfitTestSubjects, "test", "t", "", "Test subjects from subjects/ directory (default: jaimee)")
	outfitSwapCmd.Flags().IntVarP(&outfitVariations, "variations", "v", 1, "Number of variations per combination")

	// Additional options
	outfitSwapCmd.Flags().BoolVar(&outfitSendOriginal, "send-original", false, "Include reference images in API requests")
	outfitSwapCmd.Flags().BoolVar(&outfitNoConfirm, "no-confirm", false, "Skip cost confirmation prompts")
	outfitSwapCmd.Flags().BoolVar(&outfitDebugPrompt, "debug", false, "Show debug information including prompts")

	// Mark test flag as not requiring a value
	outfitSwapCmd.Flags().Lookup("test").NoOptDefVal = defaultSubject
}

func runOutfitSwap(cmd *cobra.Command, args []string) error {
	// Determine outfit source
	var outfitPath string
	if len(args) > 0 {
		outfitPath = args[0]
	} else {
		outfitPath = defaultOutfit
		logger.Info("Using default outfit", "path", outfitPath)
	}

	// Validate outfit path exists
	if _, err := os.Stat(outfitPath); os.IsNotExist(err) {
		// Try without extension if it's not a directory
		if !strings.Contains(outfitPath, ".") {
			for _, ext := range []string{".png", ".jpg", ".jpeg"} {
				tryPath := outfitPath + ext
				if _, err := os.Stat(tryPath); err == nil {
					outfitPath = tryPath
					break
				}
			}
		}
		// Check again after trying extensions
		if _, err := os.Stat(outfitPath); os.IsNotExist(err) {
			return errors.ErrFileNotFound(outfitPath)
		}
	}

	// Set default style if not specified
	if outfitStyleRef == "" {
		outfitStyleRef = defaultStyle
		logger.Info("Using default style", "path", outfitStyleRef)
	}

	// Handle test subjects
	var targetImages []string
	if cmd.Flags().Changed("test") {
		// Flag was used
		if outfitTestSubjects == "" || outfitTestSubjects == defaultSubject {
			// -t or --test with no value, or explicitly set to default
			outfitTestSubjects = defaultSubject
		}

		// Parse subjects and build paths
		subjects := strings.Fields(outfitTestSubjects)
		subjectsDir := "subjects"

		for _, subject := range subjects {
			subjectPath := filepath.Join(subjectsDir, subject)

			// Try to find the file with common extensions
			if _, err := os.Stat(subjectPath); os.IsNotExist(err) {
				found := false
				for _, ext := range []string{".png", ".jpg", ".jpeg"} {
					tryPath := subjectPath + ext
					if _, err := os.Stat(tryPath); err == nil {
						subjectPath = tryPath
						found = true
						break
					}
				}
				if !found {
					return errors.ErrFileNotFound(subjectPath)
				}
			}

			targetImages = append(targetImages, subjectPath)
		}
	} else {
		// No test flag, use default subject
		outfitTestSubjects = defaultSubject
		subjectPath := filepath.Join("subjects", defaultSubject)

		// Try with extensions
		if _, err := os.Stat(subjectPath); os.IsNotExist(err) {
			for _, ext := range []string{".png", ".jpg", ".jpeg"} {
				tryPath := subjectPath + ext
				if _, err := os.Stat(tryPath); err == nil {
					subjectPath = tryPath
					break
				}
			}
		}

		if _, err := os.Stat(subjectPath); os.IsNotExist(err) {
			return errors.ErrFileNotFound(subjectPath)
		}

		targetImages = []string{subjectPath}
		logger.Info("Using default subject", "name", defaultSubject)
	}

	// Set up output directory with timestamp
	now := time.Now()
	dateFolder := now.Format("2006-01-02")
	timestampFolder := now.Format("150405")
	outputDir := filepath.Join("output", dateFolder, timestampFolder)

	// Create workflow options
	options := workflow.WorkflowOptions{
		OutputDir:       outputDir,
		StyleReference:  outfitStyleRef,
		TargetImages:    targetImages,
		Variations:      outfitVariations,
		SendOriginal:    outfitSendOriginal,
		SkipCostConfirm: outfitNoConfirm,
		DebugPrompt:     outfitDebugPrompt,
	}

	// Initialize orchestrator
	orchestrator := workflow.NewOrchestrator(apiKey)

	// Log the operation
	logger.Info("Starting outfit-swap",
		"outfit", filepath.Base(outfitPath),
		"style", filepath.Base(outfitStyleRef),
		"subjects", len(targetImages),
		"variations", outfitVariations)

	// Run the workflow
	result, err := orchestrator.RunWorkflow("outfit-swap", outfitPath, options)
	if err != nil {
		return errors.Wrapf(err, errors.WorkflowError, "outfit-swap failed")
	}

	// Display results
	fmt.Printf("\n✓ Outfit swap completed successfully\n")
	fmt.Printf("Duration: %s\n", result.EndTime.Sub(result.StartTime))
	fmt.Printf("Images generated: %d\n", len(result.Steps))

	logger.Info("Outfit swap completed",
		"duration", result.EndTime.Sub(result.StartTime),
		"images", len(result.Steps))

	return nil
}