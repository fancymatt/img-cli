// Package workflow provides orchestration for the outfit-swap workflow.
// It coordinates between analyzers, generators, and caching to execute
// the image transformation pipeline.
package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"img-cli/pkg/analyzer"
	"img-cli/pkg/cache"
	"img-cli/pkg/generator"
	"img-cli/pkg/gemini"
	"img-cli/pkg/logger"
	"path/filepath"
	"strings"
	"time"
)

type Orchestrator struct {
	client      *gemini.Client
	analyzers   map[string]analyzer.Analyzer
	generators  map[string]generator.Generator
	caches      map[string]*cache.Cache // Separate cache for each type
	enableCache bool
}

func NewOrchestrator(apiKey string) *Orchestrator {
	client := gemini.NewClient(apiKey)

	o := &Orchestrator{
		client:      client,
		analyzers:   make(map[string]analyzer.Analyzer),
		generators:  make(map[string]generator.Generator),
		caches:      make(map[string]*cache.Cache),
		enableCache: true,
	}

	// Initialize separate caches for different types
	o.caches["outfit"] = cache.NewCacheForType("outfit", 0)
	o.caches["visual_style"] = cache.NewCacheForType("visual_style", 0)
	o.caches["art_style"] = cache.NewCacheForType("art_style", 0)

	o.analyzers["outfit"] = analyzer.NewOutfitAnalyzer(client)
	o.analyzers["visual_style"] = analyzer.NewVisualStyleAnalyzer(client)
	o.analyzers["art_style"] = analyzer.NewArtStyleAnalyzer(client)

	o.generators["outfit"] = generator.NewOutfitGenerator(client)
	o.generators["style_transfer"] = generator.NewStyleTransferGenerator(client)
	o.generators["combined"] = generator.NewCombinedGenerator(client)
	o.generators["style_guide"] = generator.NewStyleGuideGenerator(client)

	return o
}

// SetCacheEnabled enables or disables caching
func (o *Orchestrator) SetCacheEnabled(enabled bool) {
	o.enableCache = enabled
}

// GetCacheForType returns the cache for a specific analyzer type
func (o *Orchestrator) GetCacheForType(analyzerType string) *cache.Cache {
	return o.caches[analyzerType]
}

// AnalyzeAll analyzes an image with all available analyzers
func (o *Orchestrator) AnalyzeAll(imagePath string) (map[string]json.RawMessage, error) {
	results := make(map[string]json.RawMessage)

	for analyzerType := range o.analyzers {
		result, err := o.AnalyzeImage(analyzerType, imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s: %w", analyzerType, err)
		}
		results[analyzerType] = result
	}

	return results, nil
}

// AnalyzeImage analyzes an image using the specified analyzer
func (o *Orchestrator) AnalyzeImage(analyzerType string, imagePath string) (json.RawMessage, error) {
	analyzer, ok := o.analyzers[analyzerType]
	if !ok {
		return nil, fmt.Errorf("analyzer not found: %s", analyzerType)
	}

	// Get the appropriate cache for this analyzer type
	c := o.caches[analyzerType]
	if c == nil || !o.enableCache {
		// No cache configured or caching disabled
		return analyzer.Analyze(imagePath)
	}

	// Try to get from cache
	cached, found := c.Get(analyzerType, imagePath)
	if found {
		logger.Info("Using cached analysis",
			"type", analyzerType,
			"file", filepath.Base(imagePath))
		// Also print to console for visibility
		fmt.Printf("✓ Using cached %s analysis for %s\n", analyzerType, filepath.Base(imagePath))

		// Check if cached data is the raw analysis or wrapped in a cache entry
		// First try to parse as cache entry structure
		var cacheEntry struct {
			Timestamp   time.Time       `json:"timestamp"`
			Description string          `json:"description"`
			Analysis    json.RawMessage `json:"analysis"`
		}
		if err := json.Unmarshal(cached, &cacheEntry); err == nil && cacheEntry.Analysis != nil {
			return cacheEntry.Analysis, nil
		}
		// If that fails, try using the cached data directly as analysis
		// This handles manually edited cache files that might only contain the analysis
		return cached, nil
	}

	// Not in cache, perform analysis
	logger.Debug("Performing new analysis",
		"type", analyzerType,
		"file", filepath.Base(imagePath))
	result, err := analyzer.Analyze(imagePath)
	if err != nil {
		return nil, err
	}

	// Store in cache with the appropriate structure
	cacheEntry := struct {
		Timestamp   time.Time       `json:"timestamp"`
		Description string          `json:"description"`
		Analysis    json.RawMessage `json:"analysis"`
	}{
		Analysis:    result,
		Timestamp:   time.Now(),
		Description: extractDescriptionFromAnalysis(analyzerType, result),
	}

	cacheData, err := json.Marshal(cacheEntry)
	if err == nil {
		c.Set(analyzerType, imagePath, cacheData)
	}

	return result, nil
}

// Helper function to extract description from analysis result
func extractDescriptionFromAnalysis(analyzerType string, analysis json.RawMessage) string {
	switch analyzerType {
	case "outfit":
		var outfit gemini.OutfitDescription
		if err := json.Unmarshal(analysis, &outfit); err == nil {
			// Build a comprehensive description
			var parts []string
			// Handle clothing items
			for _, item := range outfit.Clothing {
				if str, ok := item.(string); ok {
					parts = append(parts, str)
				} else if clothingItem, ok := item.(map[string]interface{}); ok {
					if itemName, exists := clothingItem["item"].(string); exists {
						parts = append(parts, itemName)
					}
				}
			}
			if outfit.Style != "" {
				parts = append(parts, fmt.Sprintf("Style: %s", outfit.Style))
			}
			return strings.Join(parts, ", ")
		}
	case "visual_style":
		var style gemini.VisualStyle
		if err := json.Unmarshal(analysis, &style); err == nil {
			return style.Composition
		}
	}
	return ""
}

// GenerateImage generates an image using the specified generator
func (o *Orchestrator) GenerateImage(generatorType string, params generator.GenerateParams) (*generator.GenerateResult, error) {
	gen, ok := o.generators[generatorType]
	if !ok {
		return nil, fmt.Errorf("generator not found: %s", generatorType)
	}

	return gen.Generate(params)
}

// RunWorkflow runs the outfit-swap workflow
func (o *Orchestrator) RunWorkflow(workflow string, imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	if workflow != "outfit-swap" {
		return nil, fmt.Errorf("unsupported workflow: %s (only 'outfit-swap' is supported)", workflow)
	}

	// Check if modular components are specified
	if hasModularComponents(options) {
		logger.Info("Using modular workflow due to modular components")
		return o.runOutfitSwapModularWorkflow(imagePath, options)
	}
	logger.Info("Using standard outfit-swap workflow")
	return o.runOutfitSwapWorkflow(imagePath, options)
}

func (o *Orchestrator) runOutfitSwapWorkflow(outfitSourcePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "outfit-swap",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	// Collect target images - use TargetImages if available, otherwise fall back to TargetImage
	var targetImages []string
	if len(options.TargetImages) > 0 {
		targetImages = options.TargetImages
	} else if options.TargetImage != "" {
		targetImages = []string{options.TargetImage}
	} else {
		return nil, fmt.Errorf("target subject must be specified for outfit-swap workflow")
	}

	if len(targetImages) == 1 {
		fmt.Printf("Applying to subject: %s\n", filepath.Base(targetImages[0]))
	} else {
		fmt.Printf("Applying to %d subjects\n", len(targetImages))
	}

	// Determine number of variations to generate
	variations := options.Variations
	if variations < 1 {
		variations = 1
	}

	// Collect outfit files
	var outfitFiles []string
	if outfitSourcePath == "" && options.OutfitText != "" {
		outfitFiles = []string{""} // Empty string signals text mode
		fmt.Printf("Using text outfit description\n")
	} else if outfitSourcePath != "" {
		var err error
		outfitFiles, err = collectImageFiles(outfitSourcePath)
		if err != nil {
			return nil, err
		}
		if len(outfitFiles) > 1 {
			fmt.Printf("Found %d outfit images in directory\n", len(outfitFiles))
		}
	} else {
		return nil, fmt.Errorf("no outfit source provided: either specify an outfit image path or use --outfit-text")
	}

	// Pre-count style files for accurate cost estimation
	// We need to determine the style source to count properly
	var numStyles int
	if options.StyleReference != "" {
		styleFiles, err := collectImageFiles(options.StyleReference)
		if err != nil {
			// If we can't count styles, assume 1
			numStyles = 1
		} else {
			numStyles = len(styleFiles)
		}
	} else {
		// When no style specified or using outfit as style, count as 1
		numStyles = 1
	}

	// Calculate and check total cost before processing
	estimatedImages := calculateOutfitSwapImageCount(
		len(targetImages),
		len(outfitFiles),
		numStyles,
		variations,
	)

	// Check cost and get user confirmation if needed
	if err := checkWorkflowCost("outfit-swap", estimatedImages, options.SkipCostConfirm); err != nil {
		return nil, err
	}

	// Process each subject
	for subjectIndex, targetImage := range targetImages {
		if len(targetImages) > 1 {
			fmt.Printf("\n=== Subject %d/%d: %s ===\n", subjectIndex+1, len(targetImages), filepath.Base(targetImage))
		}

		// Process each outfit for this subject
		for outfitIndex, outfitPath := range outfitFiles {
		var outfitPrompt string
		var hairDataFromOutfit json.RawMessage
		var outfitSourceName string

		// Handle text outfit vs image outfit
		if outfitPath == "" && options.OutfitText != "" {
			// Text outfit mode
			outfitPrompt = options.OutfitText
			outfitSourceName = "text_outfit"
			if len(outfitFiles) > 1 {
				fmt.Printf("\n[Outfit %d/%d] Using text description\n", outfitIndex+1, len(outfitFiles))
			}

			result.Steps = append(result.Steps, StepResult{
				Type:    "text_outfit",
				Name:    "outfit_description",
				Message: outfitPrompt,
			})
		} else {
			// Image outfit mode
			outfitSourceName = strings.TrimSuffix(filepath.Base(outfitPath), filepath.Ext(outfitPath))
			if len(outfitFiles) > 1 {
				fmt.Printf("\n[Outfit %d/%d] Processing: %s\n", outfitIndex+1, len(outfitFiles), filepath.Base(outfitPath))
			} else {
				fmt.Printf("Analyzing outfit from: %s\n", filepath.Base(outfitPath))
			}

			// Analyze outfit from the source image
			outfitData, err := o.AnalyzeImage("outfit", outfitPath)
			if err != nil {
				fmt.Printf("  Warning: Failed to analyze outfit %s: %v\n", filepath.Base(outfitPath), err)
				continue
			}

			result.Steps = append(result.Steps, StepResult{
				Type: "analysis",
				Name: "outfit_source",
				Data: outfitData,
			})

			// Extract outfit description and hair data
			outfitPrompt, hairDataFromOutfit = extractOutfitPromptAndHair(outfitData)

			// Debug output
			if options.DebugPrompt {
				fmt.Printf("\n[DEBUG] Outfit prompt built from analysis:\n%s\n\n", outfitPrompt)
			}
		}

		// Determine style source - use style-ref if provided, otherwise use the outfit source
		styleSourcePath := options.StyleReference
		if styleSourcePath == "" && outfitPath != "" {
			// Only use outfit source for style if we have an outfit image
			styleSourcePath = outfitPath
			fmt.Printf("  Using same image for style: %s\n", filepath.Base(outfitPath))
		} else if styleSourcePath != "" {
			fmt.Printf("  Using style from: %s\n", filepath.Base(styleSourcePath))
		}

		// Determine hair source and data
		var hairData json.RawMessage
		var hairSourceName string
		if options.HairReference == "USE_OUTFIT_REF" {
			// Use hair from outfit reference
			hairData = hairDataFromOutfit
			if outfitPath != "" {
				hairSourceName = strings.TrimSuffix(filepath.Base(outfitPath), filepath.Ext(outfitPath))
			}
			if hairData != nil {
				fmt.Printf("  Using hair from outfit reference\n")
			}
		} else if options.HairReference != "" {
		// Analyze hair from specified reference image
		fmt.Printf("  Analyzing hair from: %s\n", filepath.Base(options.HairReference))
		hairAnalysisResult, err := o.AnalyzeImage("outfit", options.HairReference)
		if err != nil {
			fmt.Printf("    Warning: Failed to analyze hair from %s: %v\n", filepath.Base(options.HairReference), err)
		} else {
			// Extract hair from analysis
			var outfit gemini.OutfitDescription
			if err := json.Unmarshal(hairAnalysisResult, &outfit); err == nil && outfit.Hair != nil {
				hairData, _ = json.Marshal(outfit.Hair)
			}
			if hairData != nil {
				hairSourceName = strings.TrimSuffix(filepath.Base(options.HairReference), filepath.Ext(options.HairReference))
				fmt.Printf("    Successfully extracted hair data\n")
			} else {
				fmt.Printf("    Warning: No hair data found in analysis\n")
			}

			result.Steps = append(result.Steps, StepResult{
				Type: "analysis",
				Name: "hair_source",
				Data: hairAnalysisResult,
			})
		}
	}
	// If no hair reference specified, hairData remains nil and original hair will be preserved

	// Collect style sources
	styleFiles, err := collectImageFiles(styleSourcePath)
	if err != nil {
		fmt.Printf("  Warning: Failed to collect style files: %v\n", err)
		styleFiles = []string{""} // Use default style
	} else if len(styleFiles) > 1 {
		fmt.Printf("  Found %d style images in directory\n", len(styleFiles))
	}

	// Loop through all style files
	for styleIndex, stylePath := range styleFiles {
		var styleData json.RawMessage
		styleSourceName := "default_style"

		// Analyze style if we have a style file
		if stylePath != "" {
			if len(styleFiles) > 1 {
				fmt.Printf("    [Style %d/%d] Processing: %s\n", styleIndex+1, len(styleFiles), filepath.Base(stylePath))
			}

			var err error
			styleData, err = o.AnalyzeImage("visual_style", stylePath)
			if err != nil {
				fmt.Printf("    Warning: Failed to analyze style %s: %v\n", filepath.Base(stylePath), err)
				continue
			}

			styleSourceName = strings.TrimSuffix(filepath.Base(stylePath), filepath.Ext(stylePath))

			result.Steps = append(result.Steps, StepResult{
				Type: "analysis",
				Name: "style_source",
				Data: styleData,
			})
		}

		// Generate the specified number of variations for this combination
		for v := 1; v <= variations; v++ {
			if variations > 1 {
				fmt.Printf("      Generating variation %d of %d...\n", v, variations)
			} else {
				fmt.Printf("      Generating image...\n")
			}

			// Pass outfit reference image if SendOriginal is true and we have an image
			outfitRef := ""
			promptToUse := outfitPrompt
			if options.SendOriginal && outfitPath != "" {
				outfitRef = outfitPath
				// When using --send-original, use minimal prompt to let the image speak for itself
				promptToUse = ""
			}

			combinedResult, err := o.GenerateImage("combined", generator.GenerateParams{
				ImagePath:       targetImage,
				Prompt:          promptToUse,
				StyleData:       styleData,
				HairData:        hairData,
				OutputDir:       options.OutputDir,
				DebugPrompt:     options.DebugPrompt,
				OutfitSource:    outfitSourceName,
				StyleSource:     styleSourceName,
				HairSource:      hairSourceName,
				VariationIndex:  v,
				TotalVariations: variations,
				OutfitReference: outfitRef,
				SendOriginal:    options.SendOriginal,
			})
			if err != nil {
				fmt.Printf("    Warning: Failed to generate image with style %s: %v\n", styleSourceName, err)
				continue
			}

			message := fmt.Sprintf("Generated with %s outfit and %s style", outfitSourceName, styleSourceName)
			if len(targetImages) > 1 {
				message = fmt.Sprintf("Generated %s with %s outfit and %s style", filepath.Base(targetImage), outfitSourceName, styleSourceName)
			}
			result.Steps = append(result.Steps, StepResult{
				Type:       "generation",
				Name:       "combined",
				OutputPath: combinedResult.OutputPath,
				Message:    message,
			})

			// Brief pause between generations
			if v < variations || styleIndex < len(styleFiles)-1 || outfitIndex < len(outfitFiles)-1 || subjectIndex < len(targetImages)-1 {
				time.Sleep(1 * time.Second)
			}
		}
	}
	} // End of outfit loop
	} // End of subject loop

	result.EndTime = time.Now()
	result.SubjectCount = len(targetImages)
	result.OutfitCount = len(outfitFiles)
	result.StyleCount = numStyles
	result.VariationCount = variations
	return result, nil
}


// formatDescription formats a description with a label
func formatDescription(label, description string) string {
	if description == "" {
		return ""
	}

	// For descriptions that already form complete sentences,
	// just add the label as a prefix
	if strings.Contains(description, ".") {
		return fmt.Sprintf("%s: %s", label, description)
	}

	// For shorter descriptions, create a complete sentence
	return fmt.Sprintf("%s: %s.", label, description)
}

// Buffer wraps bytes.Buffer to make it implement io.Writer
type Buffer struct {
	*bytes.Buffer
}

func (b *Buffer) Close() error {
	return nil
}