package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	analyzeNoCache bool
	analyzeType    string
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze <image-path>",
	Short: "Analyze an image for outfit, visual style, or art style",
	Long: `Analyze an image to extract detailed information about outfits,
visual/photographic styles, or artistic styles.

The analysis results are cached by default to improve performance.`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().BoolVar(&analyzeNoCache, "no-cache", false, "Disable cache for this analysis")
	analyzeCmd.Flags().StringVarP(&analyzeType, "type", "t", "", "Type of analysis: outfit, visual_style, art_style (default: all)")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Validate input
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return errors.ErrFileNotFound(imagePath)
	}

	orchestrator := workflow.NewOrchestrator(apiKey)

	if analyzeNoCache {
		orchestrator.SetCacheEnabled(false)
		defer orchestrator.SetCacheEnabled(true)
	}

	logger.Info("Starting analysis",
		"image", filepath.Base(imagePath),
		"type", analyzeType)

	// Perform analysis
	if analyzeType == "" {
		// Analyze all types
		results, err := orchestrator.AnalyzeAll(imagePath)
		if err != nil {
			return errors.Wrap(err, errors.AnalysisError, "failed to analyze image")
		}

		// Print results
		for typ, result := range results {
			fmt.Printf("\n=== %s Analysis ===\n", typ)
			printJSON(result)
		}
	} else {
		// Analyze specific type
		result, err := orchestrator.AnalyzeImage(analyzeType, imagePath)
		if err != nil {
			return errors.Wrapf(err, errors.AnalysisError, "failed to analyze %s", analyzeType)
		}

		fmt.Printf("\n=== %s Analysis ===\n", analyzeType)
		printJSON(result)
	}

	logger.Info("Analysis completed successfully")
	return nil
}

func printJSON(data json.RawMessage) {
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, data, "", "  "); err != nil {
		fmt.Println(string(data))
	} else {
		fmt.Println(formatted.String())
	}
}