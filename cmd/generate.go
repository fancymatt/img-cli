package cmd

import (
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/generator"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	generateType     string
	sendOriginal     bool
	outfitRef        string
	styleRef         string
	outputDir        string
	temperature      float64
	debugPrompt      bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate <image-path> [prompt]",
	Short: "Generate transformed images",
	Long: `Generate transformed images using outfit, style, or artistic transformations.

Examples:
  img-cli generate portrait.jpg "business suit" --type outfit
  img-cli generate portrait.jpg --type outfit --outfit-ref outfits/suit.png
  img-cli generate image.jpg "dramatic lighting" --type style_transfer`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&generateType, "type", "t", "outfit", "Type of generation: outfit, style_transfer, art_style")
	generateCmd.Flags().BoolVar(&sendOriginal, "send-original", false, "Include reference image in the request")
	generateCmd.Flags().StringVar(&outfitRef, "outfit-ref", "", "Path to outfit reference image")
	generateCmd.Flags().StringVar(&styleRef, "style-ref", "", "Path to style reference image")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: output/YYYY-MM-DD/HHMMSS)")
	generateCmd.Flags().Float64Var(&temperature, "temperature", 0.7, "Generation temperature (0.0-1.0)")
	generateCmd.Flags().BoolVar(&debugPrompt, "debug-prompt", false, "Show the generation prompt")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	prompt := ""
	if len(args) > 1 {
		prompt = strings.Join(args[1:], " ")
	}

	// Validate input
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return errors.ErrFileNotFound(imagePath)
	}

	// Validate references if provided
	if outfitRef != "" {
		if _, err := os.Stat(outfitRef); os.IsNotExist(err) {
			return errors.ErrFileNotFound(outfitRef)
		}
	}
	if styleRef != "" {
		if _, err := os.Stat(styleRef); os.IsNotExist(err) {
			return errors.ErrFileNotFound(styleRef)
		}
	}

	// Set default output directory if not specified
	if outputDir == "" {
		now := time.Now()
		outputDir = filepath.Join("output",
			now.Format("2006-01-02"),
			now.Format("150405"))
	}

	orchestrator := workflow.NewOrchestrator(apiKey)

	logger.Info("Starting generation",
		"type", generateType,
		"image", filepath.Base(imagePath),
		"output", outputDir)

	params := generator.GenerateParams{
		ImagePath:       imagePath,
		Prompt:          prompt,
		OutputDir:       outputDir,
		SendOriginal:    sendOriginal,
		OutfitReference: outfitRef,
		StyleReference:  styleRef,
		Temperature:     temperature,
		DebugPrompt:     debugPrompt,
	}

	result, err := orchestrator.GenerateImage(generateType, params)
	if err != nil {
		return errors.Wrap(err, errors.GenerationError, "failed to generate image")
	}

	fmt.Printf("✓ %s\n", result.Message)
	fmt.Printf("Saved to: %s\n", result.OutputPath)

	logger.Info("Generation completed successfully",
		"output", result.OutputPath)

	return nil
}