// Package workflow provides orchestration for complex image processing workflows.
// It coordinates between analyzers, generators, and caching to execute multi-step
// image transformation pipelines.
package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"img-cli/pkg/analyzer"
	"img-cli/pkg/cache"
	"img-cli/pkg/generator"
	"img-cli/pkg/gemini"
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
	o.generators["art_style"] = generator.NewArtStyleGenerator(client)

	return o
}

func (o *Orchestrator) SetCacheEnabled(enabled bool) {
	o.enableCache = enabled
}

func (o *Orchestrator) GetCache() *cache.Cache {
	// Return the outfit cache by default for backward compatibility
	return o.caches["outfit"]
}

func (o *Orchestrator) GetCacheForType(cacheType string) *cache.Cache {
	if c, exists := o.caches[cacheType]; exists {
		return c
	}
	// Return outfit cache as default
	return o.caches["outfit"]
}

func (o *Orchestrator) GetAnalyzerTypes() []string {
	types := make([]string, 0, len(o.analyzers))
	for typ := range o.analyzers {
		types = append(types, typ)
	}
	return types
}

func (o *Orchestrator) AnalyzeAll(imagePath string) (map[string]json.RawMessage, error) {
	results := make(map[string]json.RawMessage)

	for typ := range o.analyzers {
		result, err := o.AnalyzeImage(typ, imagePath)
		if err != nil {
			fmt.Printf("  Warning: Failed to analyze %s: %v\n", typ, err)
			continue
		}
		results[typ] = result
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("all analyses failed")
	}

	return results, nil
}

func (o *Orchestrator) AnalyzeImage(analyzerType, imagePath string) (json.RawMessage, error) {
	a, exists := o.analyzers[analyzerType]
	if !exists {
		return nil, fmt.Errorf("analyzer type '%s' not found", analyzerType)
	}

	// Check cache first - use the appropriate cache for the type
	if o.enableCache {
		cache := o.GetCacheForType(analyzerType)
		if cachedData, found := cache.Get(analyzerType, imagePath); found {
			fmt.Printf("  [Using cached %s analysis]\n", analyzerType)
			return cachedData, nil
		}
	}

	// Perform analysis
	result, err := a.Analyze(imagePath)
	if err != nil {
		return nil, err
	}

	// Store in cache - use the appropriate cache for the type
	if o.enableCache {
		cache := o.GetCacheForType(analyzerType)
		if err := cache.Set(analyzerType, imagePath, result); err != nil {
			// Log but don't fail on cache errors
			fmt.Printf("  [Warning: Failed to cache result: %v]\n", err)
		}
	}

	return result, nil
}

func (o *Orchestrator) GenerateImage(generatorType string, params generator.GenerateParams) (*generator.GenerateResult, error) {
	g, exists := o.generators[generatorType]
	if !exists {
		return nil, fmt.Errorf("generator type '%s' not found", generatorType)
	}

	return g.Generate(params)
}

func (o *Orchestrator) RunWorkflow(workflow string, imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	switch workflow {
	case "outfit-variations":
		return o.runOutfitVariationsWorkflow(imagePath, options)
	case "style-transfer":
		return o.runStyleTransferWorkflow(imagePath, options)
	case "complete-transformation":
		return o.runCompleteTransformationWorkflow(imagePath, options)
	case "cross-reference":
		return o.runCrossReferenceWorkflow(imagePath, options)
	case "outfit-swap":
		return o.runOutfitSwapWorkflow(imagePath, options)
	case "use-art-style":
		return o.runUseArtStyleWorkflow(imagePath, options)
	case "analyze-style":
		return o.runAnalyzeStyleWorkflow(imagePath, options)
	case "create-style-guide":
		return o.runCreateStyleGuideWorkflow(imagePath, options)
	default:
		return nil, fmt.Errorf("unknown workflow: %s", workflow)
	}
}

func (o *Orchestrator) runOutfitVariationsWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "outfit-variations",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	outfitData, err := o.AnalyzeImage("outfit", imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze outfit: %w", err)
	}

	result.Steps = append(result.Steps, StepResult{
		Type: "analysis",
		Name: "outfit",
		Data: outfitData,
	})

	outfits := options.Outfits
	if len(outfits) == 0 {
		outfits = []string{"business suit", "casual streetwear", "formal evening wear"}
	}

	for _, outfit := range outfits {
		genResult, err := o.GenerateImage("outfit", generator.GenerateParams{
			ImagePath: imagePath,
			Prompt:    outfit,
			OutputDir: options.OutputDir,
		})
		if err != nil {
			fmt.Printf("Warning: failed to generate %s: %v\n", outfit, err)
			continue
		}

		result.Steps = append(result.Steps, StepResult{
			Type:       "generation",
			Name:       "outfit",
			OutputPath: genResult.OutputPath,
			Message:    genResult.Message,
		})

		time.Sleep(2 * time.Second)
	}

	result.EndTime = time.Now()
	return result, nil
}

func (o *Orchestrator) runStyleTransferWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "style-transfer",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	var styleData json.RawMessage
	if options.StyleReference != "" {
		var err error
		styleData, err = o.AnalyzeImage("visual_style", options.StyleReference)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze style reference: %w", err)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "visual_style",
			Data: styleData,
		})
	}

	genResult, err := o.GenerateImage("style_transfer", generator.GenerateParams{
		ImagePath: imagePath,
		Prompt:    options.StylePrompt,
		StyleData: styleData,
		OutputDir: options.OutputDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate styled image: %w", err)
	}

	result.Steps = append(result.Steps, StepResult{
		Type:       "generation",
		Name:       "style_transfer",
		OutputPath: genResult.OutputPath,
		Message:    genResult.Message,
	})

	result.EndTime = time.Now()
	return result, nil
}

func (o *Orchestrator) runCompleteTransformationWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "complete-transformation",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	outfitData, err := o.AnalyzeImage("outfit", imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze outfit: %w", err)
	}
	result.Steps = append(result.Steps, StepResult{
		Type: "analysis",
		Name: "outfit",
		Data: outfitData,
	})

	styleData, err := o.AnalyzeImage("visual_style", imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze visual style: %w", err)
	}
	result.Steps = append(result.Steps, StepResult{
		Type: "analysis",
		Name: "visual_style",
		Data: styleData,
	})

	if options.NewOutfit != "" {
		genResult, err := o.GenerateImage("outfit", generator.GenerateParams{
			ImagePath: imagePath,
			Prompt:    options.NewOutfit,
			OutputDir: options.OutputDir,
		})
		if err == nil {
			result.Steps = append(result.Steps, StepResult{
				Type:       "generation",
				Name:       "outfit",
				OutputPath: genResult.OutputPath,
				Message:    genResult.Message,
			})

			time.Sleep(2 * time.Second)

			styledResult, err := o.GenerateImage("style_transfer", generator.GenerateParams{
				ImagePath: genResult.OutputPath,
				StyleData: styleData,
				OutputDir: options.OutputDir,
			})
			if err == nil {
				result.Steps = append(result.Steps, StepResult{
					Type:       "generation",
					Name:       "style_transfer",
					OutputPath: styledResult.OutputPath,
					Message:    "Applied original style to new outfit",
				})
			}
		}
	}

	result.EndTime = time.Now()
	return result, nil
}

func (o *Orchestrator) runCrossReferenceWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "cross-reference",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	// Get outfit from outfit reference image (or use provided description)
	var outfitPrompt string
	var hairDataFromOutfit json.RawMessage
	if options.OutfitReference != "" {
		outfitData, err := o.AnalyzeImage("outfit", options.OutfitReference)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze outfit reference: %w", err)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "outfit_reference",
			Data: outfitData,
		})

		// Extract outfit description from analysis
		var outfit gemini.OutfitDescription
		if err := json.Unmarshal(outfitData, &outfit); err == nil {
			// Build comprehensive prompt with ALL details
			var promptBuilder strings.Builder
			promptBuilder.WriteString("wearing exactly: ")

			// Include ALL clothing items with full details
			if len(outfit.Clothing) > 0 {
				for i, item := range outfit.Clothing {
					if i > 0 {
						promptBuilder.WriteString("; ")
					}

					// Handle both string and ClothingItem object formats
					switch v := item.(type) {
					case string:
						promptBuilder.WriteString(v)
					case map[string]interface{}:
						// This is a ClothingItem object
						if desc, ok := v["description"].(string); ok {
							promptBuilder.WriteString(desc)
						} else if itemName, ok := v["item"].(string); ok {
							promptBuilder.WriteString(itemName)
						}
						// Add specific color requirements if present
						if mainColor, ok := v["main_body_color"].(string); ok && mainColor != "" && mainColor != "none" {
							promptBuilder.WriteString(fmt.Sprintf(" with %s main body", mainColor))
						}
						if collarColor, ok := v["collar_color"].(string); ok && collarColor != "" && collarColor != "none" {
							promptBuilder.WriteString(fmt.Sprintf(", %s collar", collarColor))
						}
					}
				}
			}

			// Add color specifications
			if len(outfit.Colors) > 0 {
				promptBuilder.WriteString(". CRITICAL COLOR REQUIREMENTS: ")
				promptBuilder.WriteString(strings.Join(outfit.Colors, ", "))
			}

			// Add accessories if present
			if len(outfit.Accessories) > 0 {
				promptBuilder.WriteString(". Accessories: ")
				for i, acc := range outfit.Accessories {
					if i > 0 {
						promptBuilder.WriteString(", ")
					}
					// Handle both string and object formats
					switch v := acc.(type) {
					case string:
						promptBuilder.WriteString(v)
					case map[string]interface{}:
						if desc, ok := v["description"].(string); ok {
							promptBuilder.WriteString(desc)
						} else if itemName, ok := v["item"].(string); ok {
							promptBuilder.WriteString(itemName)
						}
					}
				}
			}

			// Add the overall description for context
			if outfit.Overall != "" {
				promptBuilder.WriteString(". Overall styling: ")
				promptBuilder.WriteString(outfit.Overall)
			}

			outfitPrompt = promptBuilder.String()

			// Store hair data from outfit if available
			if outfit.Hair != nil {
				hairDataFromOutfit, _ = json.Marshal(outfit.Hair)
			}
		}
	} else if options.NewOutfit != "" {
		outfitPrompt = options.NewOutfit
	} else {
		return nil, fmt.Errorf("either --outfit-ref or --outfit must be provided")
	}

	// Get style from style reference image
	var styleData json.RawMessage
	if options.StyleReference != "" {
		var err error
		styleData, err = o.AnalyzeImage("visual_style", options.StyleReference)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze style reference: %w", err)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "style_reference",
			Data: styleData,
		})
	}

	// Handle hair reference
	var hairData json.RawMessage
	if options.HairReference == "USE_OUTFIT_REF" {
		// Use hair from outfit reference
		hairData = hairDataFromOutfit
		if hairData != nil {
			fmt.Printf("Using hair from outfit reference\n")
		}
	} else if options.HairReference != "" {
		// Analyze hair from specified reference image
		fmt.Printf("Analyzing hair from: %s\n", filepath.Base(options.HairReference))
		hairAnalysisResult, err := o.AnalyzeImage("outfit", options.HairReference)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze hair from %s: %w", filepath.Base(options.HairReference), err)
		}

		// Extract hair data from the analysis
		var hairOutfit gemini.OutfitDescription
		if err := json.Unmarshal(hairAnalysisResult, &hairOutfit); err == nil && hairOutfit.Hair != nil {
			hairData, _ = json.Marshal(hairOutfit.Hair)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "hair_reference",
			Data: hairAnalysisResult,
		})
	}

	// Use combined generator to apply outfit, style, and hair all at once
	combinedResult, err := o.GenerateImage("combined", generator.GenerateParams{
		ImagePath: imagePath,
		Prompt:    outfitPrompt,
		StyleData: styleData,
		HairData:  hairData,
		OutputDir: options.OutputDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate cross-referenced image: %w", err)
	}

	result.Steps = append(result.Steps, StepResult{
		Type:       "generation",
		Name:       "combined",
		OutputPath: combinedResult.OutputPath,
		Message:    fmt.Sprintf("Generated with outfit, style, and hair references applied"),
	})

	result.EndTime = time.Now()
	return result, nil
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
			hairData = extractHairFromAnalysis(hairAnalysisResult)
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
	return result, nil
}

func (o *Orchestrator) runUseArtStyleWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "use-art-style",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	// Check if style reference is provided
	if options.StyleReference == "" {
		return nil, fmt.Errorf("style reference image required (--style-ref)")
	}

	// Analyze the style reference if it's not a JSON file
	var styleAnalysis json.RawMessage
	if strings.HasSuffix(options.StyleReference, ".json") {
		// Load existing style analysis
		data, err := gemini.LoadFile(options.StyleReference)
		if err != nil {
			return nil, fmt.Errorf("failed to load style analysis: %w", err)
		}
		styleAnalysis = json.RawMessage(data)
	} else {
		// Analyze the style reference image
		var err error
		styleAnalysis, err = o.AnalyzeImage("art_style", options.StyleReference)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze style reference: %w", err)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "art_style",
		})
	}

	// Generate styled image
	params := generator.GenerateParams{
		ImagePath:      imagePath,
		Prompt:         options.Prompt,
		StyleReference: options.StyleReference,
		StyleAnalysis:  styleAnalysis,
		OutputDir:      options.OutputDir,
	}

	genResult, err := o.GenerateImage("art_style", params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate styled image: %w", err)
	}

	result.Steps = append(result.Steps, StepResult{
		Type:       "generation",
		Name:       "art_style",
		OutputPath: genResult.OutputPath,
	})

	result.EndTime = time.Now()
	return result, nil
}

func (o *Orchestrator) runAnalyzeStyleWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "analyze-style",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	// Check if it's a directory for batch analysis
	fileInfo, err := gemini.GetFileInfo(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	var analysisResult json.RawMessage

	if fileInfo.IsDir() {
		// Batch analyze all images in directory
		images, err := gemini.GetImagesFromDirectory(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		if len(images) == 0 {
			return nil, fmt.Errorf("no images found in directory")
		}

		// Use the multi-image analyzer
		if artAnalyzer, ok := o.analyzers["art_style"].(*analyzer.ArtStyleAnalyzer); ok {
			analysisResult, err = artAnalyzer.AnalyzeMultiple(images)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze images: %w", err)
			}
		} else {
			// Fallback to single analysis of first image
			analysisResult, err = o.AnalyzeImage("art_style", images[0])
			if err != nil {
				return nil, fmt.Errorf("failed to analyze style: %w", err)
			}
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: fmt.Sprintf("art_style (batch: %d images)", len(images)),
		})
	} else {
		// Single image analysis
		analysisResult, err = o.AnalyzeImage("art_style", imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze style: %w", err)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "art_style",
		})
	}

	// Print the analysis
	fmt.Println("\n=== Art Style Analysis ===")
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, analysisResult, "", "  "); err != nil {
		fmt.Println(string(analysisResult))
	} else {
		fmt.Println(prettyJSON.String())
	}

	// Save analysis to file if output dir specified
	if options.OutputDir != "" {
		outputPath := filepath.Join(options.OutputDir, "style_analysis.json")
		if err := gemini.SaveFile(outputPath, analysisResult); err != nil {
			fmt.Printf("Warning: Could not save analysis: %v\n", err)
		} else {
			fmt.Printf("\nAnalysis saved to: %s\n", outputPath)
		}
	}

	result.EndTime = time.Now()
	return result, nil
}

func (o *Orchestrator) runCreateStyleGuideWorkflow(imagePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "create-style-guide",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	// First analyze the style(s)
	var styleAnalysis json.RawMessage
	var err error

	// Check if it's a directory for batch analysis
	fileInfo, err := gemini.GetFileInfo(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	if fileInfo.IsDir() {
		// Batch analyze all images in directory
		images, err := gemini.GetImagesFromDirectory(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		if len(images) == 0 {
			return nil, fmt.Errorf("no images found in directory")
		}

		// Use the multi-image analyzer
		if artAnalyzer, ok := o.analyzers["art_style"].(*analyzer.ArtStyleAnalyzer); ok {
			styleAnalysis, err = artAnalyzer.AnalyzeMultiple(images)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze images: %w", err)
			}
		} else {
			return nil, fmt.Errorf("art style analyzer not available")
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: fmt.Sprintf("art_style (batch: %d images)", len(images)),
		})
	} else {
		// Single image analysis
		styleAnalysis, err = o.AnalyzeImage("art_style", imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze style: %w", err)
		}

		result.Steps = append(result.Steps, StepResult{
			Type: "analysis",
			Name: "art_style",
		})
	}

	// Generate style guide
	params := generator.GenerateParams{
		StyleAnalysis: styleAnalysis,
		OutputDir:     "styles", // Always save to styles directory
	}

	// Override with custom name if provided in prompt
	if options.Prompt != "" {
		params.OutputDir = filepath.Join("styles", options.Prompt)
	}

	genResult, err := o.GenerateImage("style_guide", params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate style guide: %w", err)
	}

	result.Steps = append(result.Steps, StepResult{
		Type:       "generation",
		Name:       "style_guide",
		OutputPath: genResult.OutputPath,
	})

	fmt.Printf("\n=== Style Guide Created ===\n")
	fmt.Printf("Style guide saved to: %s\n", genResult.OutputPath)
	fmt.Printf("You can now use this style with: --style-ref %s\n", genResult.OutputPath)

	result.EndTime = time.Now()
	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}