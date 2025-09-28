package cmd

import (
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"
	"io"
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
	// Modular component flags
	outfitHairStyle   string
	outfitHairColor   string
	outfitMakeup      string
	outfitExpression  string
	outfitAccessories string
	outfitOverOutfit  string
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
	Long: `Apply an outfit to one or more test subjects with optional style and modular components.

Examples:
  # Use all defaults (shearling-black outfit, plain-white style, jaimee subject)
  img-cli outfit-swap

  # Specify outfit, use default style and subject
  img-cli outfit-swap ./outfits/business-suit.png

  # Full specification with shortcuts
  img-cli outfit-swap ./outfits/suit.png -s ./styles/night.png -t "jaimee kat"

  # Directory of outfits with multiple subjects
  img-cli outfit-swap ./outfits/batch/ -t "jaimee kat izzy" -v 3

  # Japanese theme with modular components
  img-cli outfit-swap ./outfits/kimono.png \
    --style ./styles/japan.png \
    --hair-style ./hair-style/geisha.png \
    --makeup ./makeup/geisha.png \
    --accessories ./accessories/parasol.png

  # Mix and match with directories (creates all combinations)
  img-cli outfit-swap ./outfits/ \
    --hair-style ./hair-style/ \
    --makeup ./makeup/natural.png \
    -t "jaimee kat"

  # Layered outfits (jacket from first outfit worn over complete second outfit)
  img-cli outfit-swap ./outfits/punk-jacket.png \
    --over-outfit ./outfits/dress.png \
    --style ./styles/winter.png \
    -t sarah
  # Result: dress + only the jacket from punk-jacket outfit

Default values:
  Outfit:  ./outfits/shearling-black.png
  Style:   ./styles/plain-white.png
  Subject: all subjects (when -t is not specified)
           jaimee (when -t is specified without a value)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runOutfitSwap,
}

func init() {
	rootCmd.AddCommand(outfitSwapCmd)

	// Shortcuts and full flags
	outfitSwapCmd.Flags().StringVarP(&outfitStyleRef, "style", "s", "", "Style reference image (default: ./styles/plain-white.png)")
	outfitSwapCmd.Flags().StringVarP(&outfitTestSubjects, "test", "t", "", "Test subjects from subjects/ directory (omit flag for all subjects, use -t alone for jaimee)")
	outfitSwapCmd.Flags().IntVarP(&outfitVariations, "variations", "v", 1, "Number of variations per combination")

	// Modular component flags
	outfitSwapCmd.Flags().StringVar(&outfitHairStyle, "hair-style", "", "Hair style reference image or directory")
	outfitSwapCmd.Flags().StringVar(&outfitHairColor, "hair-color", "", "Hair color reference image or directory")
	outfitSwapCmd.Flags().StringVar(&outfitMakeup, "makeup", "", "Makeup reference image or directory")
	outfitSwapCmd.Flags().StringVar(&outfitExpression, "expression", "", "Expression reference image or directory")
	outfitSwapCmd.Flags().StringVarP(&outfitAccessories, "accessories", "a", "", "Accessories reference image or directory")
	outfitSwapCmd.Flags().StringVar(&outfitAccessories, "accessory", "", "Accessories reference image or directory (alias for --accessories)")
	outfitSwapCmd.Flags().MarkHidden("accessory") // Hide from help to avoid clutter, but still works
	outfitSwapCmd.Flags().StringVar(&outfitOverOutfit, "over-outfit", "", "Complete base outfit; main outfit's outer layer (jacket/coat) will be worn over this")

	// Additional options
	outfitSwapCmd.Flags().BoolVar(&outfitSendOriginal, "send-original", false, "Include reference images in API requests")
	outfitSwapCmd.Flags().BoolVar(&outfitNoConfirm, "no-confirm", false, "Skip cost confirmation prompts")
	outfitSwapCmd.Flags().BoolVar(&outfitDebugPrompt, "debug", false, "Show debug information including prompts")
}

func runOutfitSwap(cmd *cobra.Command, args []string) error {
	// Debug: log all arguments received
	if len(args) > 1 {
		logger.Debug("Received multiple arguments", "count", len(args), "args", args)
	}

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

	// Move external images to outfits folder if needed
	outfitPath, err := moveToOutfitsIfExternal(outfitPath)
	if err != nil {
		return errors.Wrapf(err, errors.FileError, "failed to move outfit to outfits folder")
	}

	// Set default style if not specified
	if outfitStyleRef == "" {
		outfitStyleRef = defaultStyle
		logger.Info("Using default style", "path", outfitStyleRef)
	}

	// Handle test subjects
	var targetImages []string
	subjectsDir := "subjects"

	// Check if test flag was provided
	if !cmd.Flags().Changed("test") {
		// No -t flag provided at all: use ALL subjects
		logger.Info("No test subjects specified, using all subjects")
		files, err := os.ReadDir(subjectsDir)
		if err != nil {
			return errors.Wrapf(err, errors.FileError, "failed to read subjects directory")
		}

		for _, file := range files {
			if !file.IsDir() {
				ext := filepath.Ext(file.Name())
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					targetImages = append(targetImages, filepath.Join(subjectsDir, file.Name()))
				}
			}
		}

		if len(targetImages) == 0 {
			return errors.New(errors.FileError, "no image files found in subjects directory")
		}
	} else {
		// -t flag was provided
		if outfitTestSubjects == "" {
			// -t provided with no value: use default "jaimee"
			outfitTestSubjects = defaultSubject
			logger.Info("Using default subject", "name", defaultSubject)
		}

		// Parse subjects and build paths
		subjects := strings.Fields(outfitTestSubjects)
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
		// Modular components
		HairStyleRef:   outfitHairStyle,
		HairColorRef:   outfitHairColor,
		MakeupRef:      outfitMakeup,
		ExpressionRef:  outfitExpression,
		AccessoriesRef: outfitAccessories,
		OverOutfitRef:  outfitOverOutfit,
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

	// Count actual generated images (only "combined" type steps)
	generatedCount := 0
	for _, step := range result.Steps {
		if step.Type == "generation" && step.Name == "combined" {
			generatedCount++
		}
	}

	// Build the summary based on what was actually done
	var summary string
	if result.SubjectCount > 0 && result.OutfitCount > 0 {
		parts := []string{}
		if result.SubjectCount > 1 {
			parts = append(parts, fmt.Sprintf("%d subjects", result.SubjectCount))
		} else {
			parts = append(parts, "1 subject")
		}
		if result.OutfitCount > 1 {
			parts = append(parts, fmt.Sprintf("%d outfits", result.OutfitCount))
		} else {
			parts = append(parts, "1 outfit")
		}
		if result.StyleCount > 1 {
			parts = append(parts, fmt.Sprintf("%d styles", result.StyleCount))
		} else {
			parts = append(parts, "1 style")
		}
		if result.VariationCount > 1 {
			parts = append(parts, fmt.Sprintf("%d variations", result.VariationCount))
		}
		summary = fmt.Sprintf("Created %d images (%s)", generatedCount, strings.Join(parts, " × "))
	} else {
		summary = fmt.Sprintf("Created %d images", generatedCount)
	}

	fmt.Println(summary)

	logger.Info("Outfit swap completed",
		"duration", result.EndTime.Sub(result.StartTime),
		"images", len(result.Steps))

	return nil
}

// moveToOutfitsIfExternal moves an image to the outfits folder if it's from an external location
func moveToOutfitsIfExternal(imagePath string) (string, error) {
	// Clean and convert to absolute path for comparison
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return imagePath, err
	}

	// Get the absolute path of the outfits directory
	outfitsDir, err := filepath.Abs("outfits")
	if err != nil {
		return imagePath, err
	}

	// Create outfits directory if it doesn't exist
	if err := os.MkdirAll(outfitsDir, 0755); err != nil {
		return imagePath, err
	}

	// Check if the image is already in the outfits folder or a subfolder
	relPath, err := filepath.Rel(outfitsDir, absPath)
	if err == nil && !strings.HasPrefix(relPath, "..") {
		// Image is already in outfits folder or subfolder
		logger.Debug("Image already in outfits folder", "path", imagePath)
		return imagePath, nil
	}

	// Check if file is a directory (batch processing case)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return imagePath, err
	}

	if fileInfo.IsDir() {
		// Don't move directories, just return the original path
		return imagePath, nil
	}

	// Image is external, move it to outfits folder
	filename := filepath.Base(absPath)
	destPath := filepath.Join(outfitsDir, filename)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		// File with same name exists, add timestamp to make it unique
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
		destPath = filepath.Join(outfitsDir, filename)
	}

	// Open source file
	sourceFile, err := os.Open(absPath)
	if err != nil {
		return imagePath, err
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return imagePath, err
	}
	defer destFile.Close()

	// Copy the file
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return imagePath, err
	}

	logger.Info("Moved external image to outfits folder",
		"from", absPath,
		"to", destPath)

	// Return the new path relative to current directory
	relPath, err = filepath.Rel(".", destPath)
	if err != nil {
		// If relative path fails, just use the destination path
		return destPath, nil
	}
	return relPath, nil
}