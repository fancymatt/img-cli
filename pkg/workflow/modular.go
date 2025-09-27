package workflow

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/analyzer"
	"img-cli/pkg/cache"
	"img-cli/pkg/generator"
	"img-cli/pkg/logger"
	"img-cli/pkg/models"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ModularConfig holds configuration for modular generation
type ModularConfig struct {
	SubjectPath    string
	OutfitRef      string
	StyleRef       string
	HairStyleRef   string
	HairColorRef   string
	MakeupRef      string
	ExpressionRef  string
	AccessoriesRef string
	Variations     int
	SendOriginal   bool
	Debug          bool
	OutputDir      string // Optional: if not specified, will generate one
}


// RunModularWorkflow executes the modular generation workflow
func (o *Orchestrator) RunModularWorkflow(config ModularConfig) ([]string, error) {
	start := time.Now()

	// Initialize additional analyzers and caches if needed
	o.initializeModularComponents()

	// Analyze all provided components
	components, err := o.analyzeModularComponents(config)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze components: %w", err)
	}

	// Build the generation prompt
	prompt := o.buildModularPrompt(components)

	if config.Debug {
		fmt.Println("\n=== DEBUG: Generation Prompt ===")
		fmt.Println(prompt)
		fmt.Println("=== END DEBUG ===\n")
	}

	// Generate images
	var results []string
	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = generateOutputDir()
	}

	// Debug: Show the prompt if debug mode is enabled
	if config.Debug {
		fmt.Println("\n=== DEBUG: Final Generation Prompt ===")
		fmt.Println(prompt)
		fmt.Println("=== END PROMPT ===\n")
	}

	for i := 0; i < config.Variations; i++ {
		fmt.Printf("      Generating variation %d/%d...\n", i+1, config.Variations)

		// Use the modular generator
		gen := generator.NewModularGenerator(o.client)

		// Build generation request
		genRequest := generator.ModularRequest{
			SubjectPath:   config.SubjectPath,
			Prompt:        prompt,
			Components:    components,
			SendOriginals: config.SendOriginal,
			OutputDir:     outputDir,
		}

		outputPath, err := gen.Generate(genRequest)
		if err != nil {
			logger.Warn("Failed to generate image", "variation", i+1, "error", err)
			continue
		}

		results = append(results, outputPath)

		// Rate limiting between API calls
		if i < config.Variations-1 {
			time.Sleep(2 * time.Second)
		}
	}

	logger.Info("Modular workflow completed",
		"duration", time.Since(start),
		"images_generated", len(results))

	return results, nil
}

// initializeModularComponents sets up analyzers and caches for new component types
func (o *Orchestrator) initializeModularComponents() {
	// Add new analyzers if not already present
	if _, exists := o.analyzers["hair_style"]; !exists {
		o.analyzers["hair_style"] = analyzer.NewHairStyleAnalyzer(o.client)
		o.caches["hair_style"] = cache.NewCacheForType("hair_style", 0)
	}
	if _, exists := o.analyzers["hair_color"]; !exists {
		o.analyzers["hair_color"] = analyzer.NewHairColorAnalyzer(o.client)
		o.caches["hair_color"] = cache.NewCacheForType("hair_color", 0)
	}
	if _, exists := o.analyzers["makeup"]; !exists {
		o.analyzers["makeup"] = analyzer.NewMakeupAnalyzer(o.client)
		o.caches["makeup"] = cache.NewCacheForType("makeup", 0)
	}
	if _, exists := o.analyzers["expression"]; !exists {
		o.analyzers["expression"] = analyzer.NewExpressionAnalyzer(o.client)
		o.caches["expression"] = cache.NewCacheForType("expression", 0)
	}
	if _, exists := o.analyzers["accessories"]; !exists {
		o.analyzers["accessories"] = analyzer.NewAccessoriesAnalyzer(o.client)
		o.caches["accessories"] = cache.NewCacheForType("accessories", 0)
	}
}

// analyzeModularComponents analyzes all provided component images
func (o *Orchestrator) analyzeModularComponents(config ModularConfig) (*models.ModularComponents, error) {
	components := &models.ModularComponents{}

	// Determine which components are excluded (have separate inputs)
	excludeOpts := analyzer.ExcludeOptions{
		Hair:        config.HairStyleRef != "" || config.HairColorRef != "",
		Makeup:      config.MakeupRef != "",
		Accessories: config.AccessoriesRef != "",
	}

	// Analyze outfit with exclusions
	if config.OutfitRef != "" {
		fmt.Printf("  Analyzing outfit from: %s\n", filepath.Base(config.OutfitRef))

		// Use modular outfit analyzer with exclusions
		modularAnalyzer := analyzer.NewModularOutfitAnalyzer(o.client, excludeOpts)
		data, err := o.analyzeWithCache("outfit", config.OutfitRef, modularAnalyzer)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze outfit: %w", err)
		}

		desc := o.extractOutfitDescription(data)
		if config.Debug {
			fmt.Printf("  DEBUG: Outfit description extracted: %s\n", desc)
		}
		components.Outfit = &models.ComponentData{
			Type:        "outfit",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.OutfitRef,
		}
	}

	// Analyze style
	if config.StyleRef != "" {
		fmt.Printf("  Analyzing style from: %s\n", filepath.Base(config.StyleRef))
		data, err := o.AnalyzeImage("visual_style", config.StyleRef)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze style: %w", err)
		}

		desc := o.extractStyleDescription(data)
		components.Style = &models.ComponentData{
			Type:        "visual_style",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.StyleRef,
		}
	}

	// Analyze hair style
	if config.HairStyleRef != "" {
		fmt.Printf("  Analyzing hair style from: %s\n", filepath.Base(config.HairStyleRef))

		// Check if it's cached
		if cache, exists := o.caches["hair_style"]; exists && o.enableCache {
			if cachedData, found := cache.Get("hair_style", config.HairStyleRef); found {
				fmt.Printf("    Using cached hair style analysis\n")
				if config.Debug {
					fmt.Printf("    DEBUG: Cached hair style data: %s\n", string(cachedData))
				}
			}
		}

		data, err := o.AnalyzeImage("hair_style", config.HairStyleRef)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze hair style: %w", err)
		}

		desc := o.extractHairStyleDescription(data)
		if config.Debug {
			fmt.Printf("  DEBUG: Raw hair style JSON: %s\n", string(data))
			fmt.Printf("  DEBUG: Hair style description extracted: %s\n", desc)
		}
		components.HairStyle = &models.ComponentData{
			Type:        "hair_style",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.HairStyleRef,
		}
	}

	// Analyze hair color
	if config.HairColorRef != "" {
		fmt.Printf("  Analyzing hair color from: %s\n", filepath.Base(config.HairColorRef))
		data, err := o.AnalyzeImage("hair_color", config.HairColorRef)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze hair color: %w", err)
		}

		desc := o.extractHairColorDescription(data)
		components.HairColor = &models.ComponentData{
			Type:        "hair_color",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.HairColorRef,
		}
	}

	// Analyze makeup
	if config.MakeupRef != "" {
		fmt.Printf("  Analyzing makeup from: %s\n", filepath.Base(config.MakeupRef))
		data, err := o.AnalyzeImage("makeup", config.MakeupRef)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze makeup: %w", err)
		}

		desc := o.extractMakeupDescription(data)
		components.Makeup = &models.ComponentData{
			Type:        "makeup",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.MakeupRef,
		}
	}

	// Analyze expression
	if config.ExpressionRef != "" {
		fmt.Printf("  Analyzing expression from: %s\n", filepath.Base(config.ExpressionRef))
		data, err := o.AnalyzeImage("expression", config.ExpressionRef)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze expression: %w", err)
		}

		// Extract expression, excluding gaze if style is also specified
		desc := o.extractExpressionDescription(data, config.StyleRef != "")
		components.Expression = &models.ComponentData{
			Type:        "expression",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.ExpressionRef,
		}
	}

	// Analyze accessories
	if config.AccessoriesRef != "" {
		fmt.Printf("  Analyzing accessories from: %s\n", filepath.Base(config.AccessoriesRef))
		data, err := o.AnalyzeImage("accessories", config.AccessoriesRef)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze accessories: %w", err)
		}

		desc := o.extractAccessoriesDescription(data)
		components.Accessories = &models.ComponentData{
			Type:        "accessories",
			Description: desc,
			JSONData:    data,
			ImagePath:   config.AccessoriesRef,
		}
	}

	return components, nil
}

// analyzeWithCache analyzes an image using a custom analyzer with caching
func (o *Orchestrator) analyzeWithCache(cacheType string, imagePath string, analyzer analyzer.Analyzer) (json.RawMessage, error) {
	// Try cache first
	if cache, exists := o.caches[cacheType]; exists && o.enableCache {
		if cached, found := cache.Get(cacheType, imagePath); found {
			logger.Info("Using cached analysis",
				"type", cacheType,
				"file", filepath.Base(imagePath))
			fmt.Printf("✓ Using cached %s analysis for %s\n", cacheType, filepath.Base(imagePath))
			return cached, nil
		}
	}

	// Analyze
	result, err := analyzer.Analyze(imagePath)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cache, exists := o.caches[cacheType]; exists && o.enableCache {
		cache.Set(cacheType, imagePath, result)
	}

	return result, nil
}

// buildModularPrompt builds the generation prompt from components
func (o *Orchestrator) buildModularPrompt(components *models.ModularComponents) string {
	var parts []string

	parts = append(parts, "Generate a professional 9:16 portrait photograph with the following specifications:")
	parts = append(parts, "")

	// Add outfit description
	if components.Outfit != nil {
		parts = append(parts, "OUTFIT:")
		parts = append(parts, components.Outfit.Description)
		parts = append(parts, "")
	}

	// Add hair style description
	if components.HairStyle != nil {
		// If no hair color is specified, make preservation VERY clear upfront
		if components.HairColor == nil {
			parts = append(parts, "⚠️ CRITICAL HAIR COLOR PRESERVATION ⚠️")
			parts = append(parts, "DO NOT CHANGE THE SUBJECT'S HAIR COLOR! The subject's original hair color from the source portrait MUST be preserved EXACTLY.")
			parts = append(parts, "If the subject has blonde hair, they MUST still have blonde hair in the result.")
			parts = append(parts, "If the subject has red hair, they MUST still have red hair in the result.")
			parts = append(parts, "If the subject has black hair, they MUST still have black hair in the result.")
			parts = append(parts, "")
		}

		parts = append(parts, "HAIR STYLE (STRUCTURE/CUT/SHAPE ONLY - NOT COLOR):")
		parts = append(parts, components.HairStyle.Description)

		// Add another reminder if no color specified
		if components.HairColor == nil {
			parts = append(parts, "")
			parts = append(parts, "REMINDER: Apply ONLY the hairstyle structure, cut, shape, and styling from the description above.")
			parts = append(parts, "DO NOT change the hair color - keep the subject's ORIGINAL hair color from the source image.")
			parts = append(parts, "The hair style description is about the CUT and STYLE only, not the color.")
		}
		parts = append(parts, "")
	}

	// Add hair color description
	if components.HairColor != nil {
		parts = append(parts, "HAIR COLOR:")
		parts = append(parts, components.HairColor.Description)
		parts = append(parts, "")
	}

	// Add makeup description
	if components.Makeup != nil {
		parts = append(parts, "MAKEUP (COSMETIC APPLICATION ONLY):")
		parts = append(parts, components.Makeup.Description)
		parts = append(parts, "CRITICAL: Apply makeup as a SURFACE LAYER ONLY. Do NOT alter facial bone structure, face shape, eye shape, nose shape, lip shape, or any anatomical features. Makeup should only add color, shading, and highlights to the existing facial features without changing their underlying structure or proportions.")
		parts = append(parts, "")
	}

	// Add expression description
	if components.Expression != nil {
		parts = append(parts, "FACIAL EXPRESSION (EMOTION ONLY - NOT GAZE DIRECTION):")
		parts = append(parts, components.Expression.Description)
		if components.Style != nil {
			parts = append(parts, "IMPORTANT: The PHOTOGRAPHIC STYLE section below controls where the subject looks and camera angle. Apply only the emotional expression from above, not any gaze direction.")
		}
		parts = append(parts, "")
	}

	// Add accessories description
	if components.Accessories != nil {
		parts = append(parts, "ACCESSORIES:")
		parts = append(parts, components.Accessories.Description)
		parts = append(parts, "")
	}

	// Add style description last (photographic style)
	if components.Style != nil {
		parts = append(parts, "PHOTOGRAPHIC STYLE (CONTROLS CAMERA ANGLE AND COMPOSITION):")
		parts = append(parts, components.Style.Description)
		parts = append(parts, "CRITICAL: This photographic style section OVERRIDES all other components for:")
		parts = append(parts, "- Camera angle and framing")
		parts = append(parts, "- Where the subject is looking (e.g., at camera, in mirror, to the side)")
		parts = append(parts, "- Composition and perspective")
		parts = append(parts, "- Lighting and atmosphere")
		parts = append(parts, "If the style shows someone looking in a mirror, the subject MUST be looking in a mirror, not at the camera.")
		parts = append(parts, "")
	}

	// Add standard requirements
	parts = append(parts, "TECHNICAL REQUIREMENTS:")
	parts = append(parts, "- Maintain exact facial features and identity from the original photo")
	// Add makeup preservation note
	if components.Makeup != nil {
		parts = append(parts, "- PRESERVE facial bone structure, face shape, and all anatomical features - makeup is cosmetic only")
	}
	// Add hair color preservation if only style is specified
	if components.HairStyle != nil && components.HairColor == nil {
		parts = append(parts, "- ⚠️ CRITICAL: PRESERVE the subject's ORIGINAL HAIR COLOR exactly as shown in the source portrait")
		parts = append(parts, "- The subject's hair color MUST NOT change - if they have blonde hair, keep it blonde")
		parts = append(parts, "- Apply ONLY the hair CUT/STYLE/SHAPE, NOT the color")
	}
	parts = append(parts, "- Professional 9:16 vertical portrait format")
	parts = append(parts, "- Waist-up framing showing outfit details")
	parts = append(parts, "- Natural, professional pose")
	parts = append(parts, "- High quality, detailed rendering")
	parts = append(parts, "")
	parts = append(parts, "IMPORTANT: Each component specified above should be applied independently without influencing other components.")

	// Add extra emphasis on facial preservation when makeup is involved
	if components.Makeup != nil {
		parts = append(parts, "")
		parts = append(parts, "FACIAL STRUCTURE PRESERVATION:")
		parts = append(parts, "The subject's facial anatomy, bone structure, and features must remain EXACTLY as in the original portrait.")
		parts = append(parts, "Makeup is ONLY a cosmetic surface application - like painting on skin.")
		parts = append(parts, "Do NOT reshape eyes, nose, lips, jawline, or any facial features.")
	}

	return strings.Join(parts, "\n")
}

// generateOutputDir creates a timestamped output directory
func generateOutputDir() string {
	baseDir := "output"
	dateDir := time.Now().Format("2006-01-02")
	timeDir := time.Now().Format("150405")

	outputDir := filepath.Join(baseDir, dateDir, timeDir)
	os.MkdirAll(outputDir, 0755)

	return outputDir
}