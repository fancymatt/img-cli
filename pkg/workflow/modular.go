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
	OverOutfitRef  string // Base layer outfit that the main outfit is worn over
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

// isFilePath checks if a string is a file path or a text description
func isFilePath(input string) bool {
	if input == "" {
		return false
	}

	// Check if it's a path (contains path separators or file extensions)
	if strings.Contains(input, "/") || strings.Contains(input, "\\") || strings.Contains(input, ".") {
		// Try to stat the file to see if it exists
		if _, err := os.Stat(input); err == nil {
			return true
		}
	}

	// If it doesn't look like a path and doesn't exist as a file, it's text
	return false
}

// processComponentInput handles both file paths and text descriptions for a component
func processComponentInput(input string, componentType string) (string, bool) {
	if input == "" {
		return "", false
	}

	// For style, always treat as file path
	if componentType == "style" || componentType == "visual_style" {
		return input, true
	}

	// Check if it's a file path
	if isFilePath(input) {
		return input, true
	}

	// It's a text description
	return input, false
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
		if isFilePath(config.OutfitRef) {
			fmt.Printf("  Analyzing outfit from: %s\n", filepath.Base(config.OutfitRef))

			// Use modular outfit analyzer with exclusions
			modularAnalyzer := analyzer.NewModularOutfitAnalyzer(o.client, excludeOpts)
			data, err := o.analyzeWithCache("outfit", config.OutfitRef, modularAnalyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze outfit: %w", err)
			}

			// If there's an over-outfit, we only want the outer layer from the main outfit
			var desc string
			if config.OverOutfitRef != "" {
				desc = o.extractOuterLayerOnly(data)
				if desc == "" {
					// If no outer layer found, skip this outfit component
					fmt.Printf("    No outer layer (jacket/coat) found in main outfit, will use over-outfit as complete outfit\n")
					// Don't set components.Outfit so we only use the over-outfit
				} else {
					fmt.Printf("    Extracted outer layer only (jacket/coat) from main outfit\n")
					if config.Debug {
						fmt.Printf("  DEBUG: Outer layer only extracted: %s\n", desc)
					}
					components.Outfit = &models.ComponentData{
						Type:        "outfit",
						Description: desc,
						JSONData:    data,
						ImagePath:   config.OutfitRef,
					}
				}
			} else {
				// No over-outfit, use the full outfit description
				desc = o.extractOutfitDescription(data)
				if config.Debug {
					fmt.Printf("  DEBUG: Full outfit description extracted: %s\n", desc)
				}
				components.Outfit = &models.ComponentData{
					Type:        "outfit",
					Description: desc,
					JSONData:    data,
					ImagePath:   config.OutfitRef,
				}
			}
		} else {
			// It's a text description
			fmt.Printf("  Using text description for outfit: %s\n", config.OutfitRef)
			components.Outfit = &models.ComponentData{
				Type:        "outfit",
				Description: config.OutfitRef,
				JSONData:    nil,
				ImagePath:   "",
			}
		}
	}

	// Analyze over-outfit (layered on top)
	if config.OverOutfitRef != "" {
		if isFilePath(config.OverOutfitRef) {
			fmt.Printf("  Analyzing over-outfit from: %s\n", filepath.Base(config.OverOutfitRef))

			// Use modular outfit analyzer with exclusions for the over-outfit too
			modularAnalyzer := analyzer.NewModularOutfitAnalyzer(o.client, excludeOpts)
			data, err := o.analyzeWithCache("outfit", config.OverOutfitRef, modularAnalyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze over-outfit: %w", err)
			}

			desc := o.extractOutfitDescription(data)
			if config.Debug {
				fmt.Printf("  DEBUG: Over-outfit description extracted: %s\n", desc)
			}
			components.OverOutfit = &models.ComponentData{
				Type:        "over_outfit",
				Description: desc,
				JSONData:    data,
				ImagePath:   config.OverOutfitRef,
			}
		} else {
			// It's a text description
			fmt.Printf("  Using text description for over-outfit: %s\n", config.OverOutfitRef)
			components.OverOutfit = &models.ComponentData{
				Type:        "over_outfit",
				Description: config.OverOutfitRef,
				JSONData:    nil,
				ImagePath:   "",
			}
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
		if isFilePath(config.HairStyleRef) {
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
		} else {
			// It's a text description
			fmt.Printf("  Using text description for hair style: %s\n", config.HairStyleRef)
			components.HairStyle = &models.ComponentData{
				Type:        "hair_style",
				Description: config.HairStyleRef,
				JSONData:    nil,
				ImagePath:   "",
			}
		}
	}

	// Analyze hair color
	if config.HairColorRef != "" {
		if isFilePath(config.HairColorRef) {
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
		} else {
			// It's a text description
			fmt.Printf("  Using text description for hair color: %s\n", config.HairColorRef)
			components.HairColor = &models.ComponentData{
				Type:        "hair_color",
				Description: config.HairColorRef,
				JSONData:    nil,
				ImagePath:   "",
			}
		}
	}

	// Analyze makeup
	if config.MakeupRef != "" {
		if isFilePath(config.MakeupRef) {
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
		} else {
			// It's a text description
			fmt.Printf("  Using text description for makeup: %s\n", config.MakeupRef)
			components.Makeup = &models.ComponentData{
				Type:        "makeup",
				Description: config.MakeupRef,
				JSONData:    nil,
				ImagePath:   "",
			}
		}
	}

	// Analyze expression
	if config.ExpressionRef != "" {
		if isFilePath(config.ExpressionRef) {
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
		} else {
			// It's a text description
			fmt.Printf("  Using text description for expression: %s\n", config.ExpressionRef)
			components.Expression = &models.ComponentData{
				Type:        "expression",
				Description: config.ExpressionRef,
				JSONData:    nil,
				ImagePath:   "",
			}
		}
	}

	// Analyze accessories
	if config.AccessoriesRef != "" {
		if isFilePath(config.AccessoriesRef) {
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
		} else {
			// It's a text description
			fmt.Printf("  Using text description for accessories: %s\n", config.AccessoriesRef)
			components.Accessories = &models.ComponentData{
				Type:        "accessories",
				Description: config.AccessoriesRef,
				JSONData:    nil,
				ImagePath:   "",
			}
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

	// Start with critical identity preservation instruction
	parts = append(parts, "🔴 CRITICAL IDENTITY INSTRUCTION:")
	parts = append(parts, "The person in the generated image MUST be the EXACT SAME INDIVIDUAL from the source portrait.")
	parts = append(parts, "This is not about creating someone similar - it must be THEM, recognizable as the same person.")
	parts = append(parts, "Preserve their exact facial features, bone structure, and identity throughout.")
	parts = append(parts, "")

	// Check if this is a POV/first-person style
	isPOV := components.Style != nil && (
		strings.Contains(strings.ToLower(components.Style.Description), "first-person") ||
		strings.Contains(strings.ToLower(components.Style.Description), "first person") ||
		strings.Contains(strings.ToLower(components.Style.Description), "pov") ||
		strings.Contains(strings.ToLower(components.Style.Description), "extreme close-up on the subject's hands"))

	// Only specify portrait format if no style is provided
	// If style is provided, it controls the framing and composition
	if isPOV {
		parts = append(parts, "🚨 THIS IS A FIRST-PERSON POV SHOT - CRITICAL INSTRUCTIONS 🚨")
		parts = append(parts, "")
		parts = append(parts, "🔴 IDENTITY PRESERVATION: This is the SAME PERSON from the provided portrait.")
		parts = append(parts, "Any visible reflections MUST show their EXACT facial features.")
		parts = append(parts, "")
		parts = append(parts, "1. FRAMING: Create a FIRST-PERSON PERSPECTIVE exactly as shown in the style image")
		parts = append(parts, "2. The camera IS the subject's eyes - shoot FROM their viewpoint, not AT them")
		parts = append(parts, "3. COPY THE EXACT FRAMING from the style image")
		parts = append(parts, "")
		parts = append(parts, "IMPORTANT: The person in the reference image IS the subject, but shown from THEIR OWN perspective:")
		parts = append(parts, "- Their hands/arms in frame = the subject's own hands reaching forward")
		parts = append(parts, "- If there's a mirror = show the subject's EXACT face/features reflected in it")
		parts = append(parts, "- Preserve their facial features, hair, skin tone, and identity completely")
		parts = append(parts, "- Apply their outfit to whatever body parts are visible in the POV framing")
		parts = append(parts, "")
	} else if components.Style != nil {
		parts = append(parts, "⚠️ CRITICAL INSTRUCTION: Generate an image of THIS EXACT PERSON with the framing described below.")
		parts = append(parts, "The subject's facial features and identity MUST be preserved exactly.")
		parts = append(parts, "DO NOT create a portrait or full-body shot unless the style explicitly describes one.")
		parts = append(parts, "The provided person is not just for reference - they ARE the subject.")
		parts = append(parts, "If the style shows only legs, show ONLY legs (but they're still this person's legs).")
		parts = append(parts, "If only arms, show ONLY arms (but they're still this person's arms).")
		parts = append(parts, "")
		parts = append(parts, "The style description below controls framing, but this remains the SAME PERSON.")
	} else {
		parts = append(parts, "Generate a professional 9:16 portrait photograph with the following specifications:")
	}
	parts = append(parts, "")

	// Add outfit description
	if components.Outfit != nil && components.OverOutfit != nil {
		// Layered outfit: outer layer from main outfit + complete base outfit from --over-outfit
		parts = append(parts, "LAYERED OUTFIT:")
		parts = append(parts, "")
		parts = append(parts, "COMPLETE BASE OUTFIT (all clothing worn underneath):")
		parts = append(parts, components.OverOutfit.Description)  // --over-outfit provides the full base outfit
		parts = append(parts, "")
		parts = append(parts, "OUTER LAYER ONLY (jacket/coat worn over the base outfit):")
		parts = append(parts, components.Outfit.Description)  // main outfit provides only the outer layer
		parts = append(parts, "")
		parts = append(parts, "IMPORTANT: The base outfit should be complete (shirt, pants/skirt, etc.), with the outer layer (jacket/coat) worn over it. Parts of the base outfit should be visible where the outer layer is open or doesn't cover (e.g., shirt collar, sleeves, pants/skirt).")
		parts = append(parts, "")
	} else if components.Outfit != nil {
		// Single outfit
		parts = append(parts, "OUTFIT:")
		parts = append(parts, components.Outfit.Description)
		parts = append(parts, "")
	} else if components.OverOutfit != nil {
		// Only over-outfit specified (treat as single outfit)
		parts = append(parts, "OUTFIT:")
		parts = append(parts, components.OverOutfit.Description)
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
		// Re-use the isPOV check from above (it's already been calculated)

		parts = append(parts, "")
		parts = append(parts, "==================================================")
		if isPOV {
			parts = append(parts, "🚨 FIRST-PERSON POV STYLE - CRITICAL INSTRUCTIONS 🚨")
		} else {
			parts = append(parts, "🚨 PHOTOGRAPHIC STYLE - THIS IS YOUR PRIMARY INSTRUCTION 🚨")
		}
		parts = append(parts, "==================================================")
		parts = append(parts, "")

		if isPOV {
			parts = append(parts, "⚠️ THIS IS A FIRST-PERSON POV SHOT ⚠️")
			parts = append(parts, "You MUST create the image from the subject's own perspective looking down/forward")
			parts = append(parts, "NOT a third-person view of the subject!")
			parts = append(parts, "")
		}

		parts = append(parts, "RECREATE THIS EXACT COMPOSITION:")
		parts = append(parts, components.Style.Description)
		parts = append(parts, "")
		parts = append(parts, "ABSOLUTE REQUIREMENTS:")

		if isPOV {
			parts = append(parts, "1. This is POV - shoot FROM the subject's eyes, not AT them")
			parts = append(parts, "2. Hands/arms in foreground = the subject's OWN hands (match their skin tone)")
			parts = append(parts, "3. Mirror reflection = the subject's EXACT face (preserve all facial features)")
			parts = append(parts, "4. The subject's identity must be clearly recognizable in any reflections")
			parts = append(parts, "5. Match the subject's: facial structure, eye color, hair color/style, skin tone")
			parts = append(parts, "6. Apply outfit details to visible body parts in the POV framing")
		} else {
			parts = append(parts, "1. Match the framing EXACTLY as described above")
			parts = append(parts, "2. If it says 'only arms visible' - show ONLY arms, NOT the full person")
			parts = append(parts, "3. If it says 'legs only' - show ONLY legs, NOT the full person")
			parts = append(parts, "4. If it says 'person in background' - keep them in background, NOT as main subject")
			parts = append(parts, "5. The person/subject image provided earlier is ONLY for outfit/appearance details")
			parts = append(parts, "6. DO NOT create a portrait unless the style explicitly describes a portrait")
		}

		parts = append(parts, "")
		parts = append(parts, "THINK OF THIS AS: Taking the outfit/appearance from the person image and applying it to")
		parts = append(parts, "the EXACT framing/composition/perspective described in the style above.")
		parts = append(parts, "")
		parts = append(parts, "==================================================")
		parts = append(parts, "")
	}

	// Add standard requirements
	parts = append(parts, "TECHNICAL REQUIREMENTS:")
	if isPOV {
		parts = append(parts, "- 🔴 CRITICAL: This is the SAME PERSON from the source portrait")
		parts = append(parts, "- Mirror reflections must show their EXACT face (same eyes, nose, mouth, bone structure)")
		parts = append(parts, "- This person must be immediately recognizable as the individual from the reference")
		parts = append(parts, "- Visible hands/arms must match the subject's skin tone and body type")
		parts = append(parts, "- Maintain the subject's exact hair color, style, and facial structure")
	} else if components.Style != nil {
		parts = append(parts, "- 🔴 CRITICAL: This must be the EXACT SAME PERSON from the source portrait")
		parts = append(parts, "- If face is visible, it must show their IDENTICAL facial features (not similar, IDENTICAL)")
		parts = append(parts, "- Their identity must be unmistakably preserved - same eyes, nose, mouth, face shape")
		parts = append(parts, "- Apply the clothing to THIS specific person, not a generic model")
	} else {
		parts = append(parts, "- 🔴 CRITICAL: Preserve the EXACT identity of the person from the source portrait")
		parts = append(parts, "- This must be recognizably the SAME individual, not someone who looks similar")
		parts = append(parts, "- Keep their exact facial features: eyes, nose, mouth, face shape, bone structure")
	}
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