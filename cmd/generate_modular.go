package cmd

import (
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Modular component references
	modOutfitRef      string
	modStyleRef       string
	modHairStyleRef   string
	modHairColorRef   string
	modMakeupRef      string
	modExpressionRef  string
	modAccessoriesRef string

	// Target options
	modSubjects      string
	modVariations    int
	modSendOriginal  bool
	modNoConfirm     bool
	modDebug         bool
)

// generateModularCmd represents the new modular generation command
var generateModularCmd = &cobra.Command{
	Use:   "generate-modular [subject]",
	Short: "Generate images with modular control over each aspect",
	Long: `Generate images with fine-grained control over each visual component.
Each component can be specified independently or left to defaults.

Examples:
  # Full Japanese theme example
  img-cli generate-modular subjects/person.png \
    --outfit outfits/kimono.png \
    --style styles/japan.png \
    --hair-style hair-style/ornate.png \
    --hair-color hair-color/black.png \
    --makeup makeup/geisha.png \
    --accessories accessories/umbrella.png \
    --expression expressions/serene.png

  # Mix and match components
  img-cli generate-modular subjects/person.png \
    --outfit outfits/business-suit.png \
    --hair-style hair-style/professional-bun.png \
    --expression expressions/confident.png

  # Change only hair color, keep natural style
  img-cli generate-modular subjects/person.png \
    --hair-color hair-color/platinum-blonde.png

Component Independence:
  - Each component is analyzed and applied independently
  - Unspecified components use the subject's natural appearance
  - Components don't influence each other (e.g., outfit won't affect hair)`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerateModular,
}

func init() {
	rootCmd.AddCommand(generateModularCmd)

	// Component flags
	generateModularCmd.Flags().StringVar(&modOutfitRef, "outfit", "", "Outfit reference image")
	generateModularCmd.Flags().StringVar(&modStyleRef, "style", "", "Photo style reference image")
	generateModularCmd.Flags().StringVar(&modHairStyleRef, "hair-style", "", "Hair style reference image")
	generateModularCmd.Flags().StringVar(&modHairColorRef, "hair-color", "", "Hair color reference image")
	generateModularCmd.Flags().StringVar(&modMakeupRef, "makeup", "", "Makeup reference image")
	generateModularCmd.Flags().StringVar(&modExpressionRef, "expression", "", "Expression reference image")
	generateModularCmd.Flags().StringVar(&modAccessoriesRef, "accessories", "", "Accessories reference image")

	// Generation options
	generateModularCmd.Flags().IntVarP(&modVariations, "variations", "v", 1, "Number of variations to generate")
	generateModularCmd.Flags().BoolVar(&modSendOriginal, "send-original", false, "Include reference images in API requests")
	generateModularCmd.Flags().BoolVar(&modNoConfirm, "no-confirm", false, "Skip cost confirmation")
	generateModularCmd.Flags().BoolVar(&modDebug, "debug", false, "Show debug information including prompts")
}

func runGenerateModular(cmd *cobra.Command, args []string) error {
	subjectPath := args[0]

	// Validate subject exists
	if !fileExists(subjectPath) {
		return errors.ErrInvalidInput("subject", fmt.Sprintf("file not found: %s", subjectPath))
	}

	// Log what components are being used
	logger.Info("Starting modular generation",
		"subject", filepath.Base(subjectPath),
		"variations", modVariations)

	// Create workflow configuration
	config := workflow.ModularConfig{
		SubjectPath:    subjectPath,
		OutfitRef:      modOutfitRef,
		StyleRef:       modStyleRef,
		HairStyleRef:   modHairStyleRef,
		HairColorRef:   modHairColorRef,
		MakeupRef:      modMakeupRef,
		ExpressionRef:  modExpressionRef,
		AccessoriesRef: modAccessoriesRef,
		Variations:     modVariations,
		SendOriginal:   modSendOriginal,
		Debug:          modDebug,
	}

	// Calculate cost
	totalImages := modVariations
	estimatedCost := float64(totalImages) * 0.04

	if !modNoConfirm {
		fmt.Printf("\n📊 Generation Cost Analysis:\n")
		fmt.Printf("   Images to generate: %d\n", totalImages)
		fmt.Printf("   Cost breakdown: %d images × $0.04 = $%.2f\n", totalImages, estimatedCost)

		// Show which components will be applied
		fmt.Println("\n🎨 Components to apply:")
		if modOutfitRef != "" {
			fmt.Printf("   ✓ Outfit: %s\n", filepath.Base(modOutfitRef))
		}
		if modStyleRef != "" {
			fmt.Printf("   ✓ Style: %s\n", filepath.Base(modStyleRef))
		}
		if modHairStyleRef != "" {
			fmt.Printf("   ✓ Hair Style: %s\n", filepath.Base(modHairStyleRef))
		}
		if modHairColorRef != "" {
			fmt.Printf("   ✓ Hair Color: %s\n", filepath.Base(modHairColorRef))
		}
		if modMakeupRef != "" {
			fmt.Printf("   ✓ Makeup: %s\n", filepath.Base(modMakeupRef))
		}
		if modExpressionRef != "" {
			fmt.Printf("   ✓ Expression: %s\n", filepath.Base(modExpressionRef))
		}
		if modAccessoriesRef != "" {
			fmt.Printf("   ✓ Accessories: %s\n", filepath.Base(modAccessoriesRef))
		}

		fmt.Print("\n   Proceed? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("❌ Generation cancelled by user")
			return nil
		}
	}

	// Create orchestrator and run workflow
	orchestrator := workflow.NewOrchestrator(apiKey)

	// Run the modular workflow
	results, err := orchestrator.RunModularWorkflow(config)
	if err != nil {
		return errors.Wrap(err, errors.WorkflowError, "modular generation failed")
	}

	// Display results
	fmt.Printf("\n✅ Generation completed successfully!\n")
	fmt.Printf("   Generated %d images\n", len(results))

	if len(results) > 0 {
		fmt.Printf("   Output directory: %s\n", filepath.Dir(results[0]))
	}

	return nil
}

func fileExists(path string) bool {
	_, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}