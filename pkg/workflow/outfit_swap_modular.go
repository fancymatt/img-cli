package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// runOutfitSwapModularWorkflow handles outfit-swap with modular components
func (o *Orchestrator) runOutfitSwapModularWorkflow(outfitSourcePath string, options WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Workflow:  "outfit-swap-modular",
		StartTime: time.Now(),
		Steps:     []StepResult{},
	}

	// Collect target images
	var targetImages []string
	if len(options.TargetImages) > 0 {
		targetImages = options.TargetImages
	} else if options.TargetImage != "" {
		targetImages = []string{options.TargetImage}
	} else {
		return nil, fmt.Errorf("target subject must be specified for outfit-swap workflow")
	}

	// Collect files for each modular component that can be directories
	outfitFiles, err := collectFilesForComponent(outfitSourcePath, "outfit")
	if err != nil {
		return nil, err
	}

	styleFiles, err := collectFilesForComponent(options.StyleReference, "style")
	if err != nil {
		return nil, err
	}

	hairStyleFiles, err := collectFilesForComponent(options.HairStyleRef, "hair-style")
	if err != nil {
		return nil, err
	}

	hairColorFiles, err := collectFilesForComponent(options.HairColorRef, "hair-color")
	if err != nil {
		return nil, err
	}

	makeupFiles, err := collectFilesForComponent(options.MakeupRef, "makeup")
	if err != nil {
		return nil, err
	}

	expressionFiles, err := collectFilesForComponent(options.ExpressionRef, "expression")
	if err != nil {
		return nil, err
	}

	accessoriesFiles, err := collectFilesForComponent(options.AccessoriesRef, "accessories")
	if err != nil {
		return nil, err
	}

	overOutfitFiles, err := collectFilesForComponent(options.OverOutfitRef, "over-outfit")
	if err != nil {
		return nil, err
	}

	// Calculate total images
	totalImages := len(targetImages) *
		maxInt(1, len(outfitFiles)) *
		maxInt(1, len(overOutfitFiles)) *
		maxInt(1, len(styleFiles)) *
		maxInt(1, len(hairStyleFiles)) *
		maxInt(1, len(hairColorFiles)) *
		maxInt(1, len(makeupFiles)) *
		maxInt(1, len(expressionFiles)) *
		maxInt(1, len(accessoriesFiles)) *
		options.Variations

	estimatedCost := float64(totalImages) * 0.04

	// Always show cost analysis
	fmt.Printf("\n📊 Workflow Cost Analysis for outfit-swap:\n")
	fmt.Printf("   Images to generate: %d\n", totalImages)
	fmt.Printf("   Cost breakdown: %d images × $0.04 = $%.2f\n", totalImages, estimatedCost)

	// Show component breakdown
	fmt.Println("\n🎨 Component combinations:")
	fmt.Printf("   Subjects: %d\n", len(targetImages))
	if len(outfitFiles) > 0 {
		fmt.Printf("   Outfits: %d\n", len(outfitFiles))
	}
	if len(overOutfitFiles) > 0 {
		fmt.Printf("   Over-outfits: %d\n", len(overOutfitFiles))
	}
	if len(styleFiles) > 0 {
		fmt.Printf("   Styles: %d\n", len(styleFiles))
	}
	if len(hairStyleFiles) > 0 {
		fmt.Printf("   Hair styles: %d\n", len(hairStyleFiles))
	}
	if len(hairColorFiles) > 0 {
		fmt.Printf("   Hair colors: %d\n", len(hairColorFiles))
	}
	if len(makeupFiles) > 0 {
		fmt.Printf("   Makeup: %d\n", len(makeupFiles))
	}
	if len(expressionFiles) > 0 {
		fmt.Printf("   Expressions: %d\n", len(expressionFiles))
	}
	if len(accessoriesFiles) > 0 {
		fmt.Printf("   Accessories: %d\n", len(accessoriesFiles))
	}
	fmt.Printf("   Variations: %d\n", options.Variations)

	// Only ask for confirmation if cost exceeds $5 (unless --no-confirm is used)
	if !options.SkipCostConfirm && estimatedCost > 5.00 {
		fmt.Printf("\n⚠️  This will cost more than $5 ($%.2f)\n", estimatedCost)
		fmt.Print("   Proceed? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("❌ Workflow cancelled by user")
			return result, nil
		}
	}

	// Initialize modular components
	o.initializeModularComponents()

	// Create output directory once for all images
	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir = generateOutputDir()
	}

	// Process each combination
	generatedCount := 0
	for _, subject := range targetImages {
		for _, outfit := range ensureAtLeastOne(outfitFiles) {
			for _, overOutfit := range ensureAtLeastOne(overOutfitFiles) {
				for _, style := range ensureAtLeastOne(styleFiles) {
					for _, hairStyle := range ensureAtLeastOne(hairStyleFiles) {
						for _, hairColor := range ensureAtLeastOne(hairColorFiles) {
							for _, makeup := range ensureAtLeastOne(makeupFiles) {
								for _, expression := range ensureAtLeastOne(expressionFiles) {
									for _, accessories := range ensureAtLeastOne(accessoriesFiles) {
										// Create modular config
										config := ModularConfig{
											SubjectPath:    subject,
											OutfitRef:      outfit,
											OverOutfitRef:  overOutfit,
											StyleRef:       style,
											HairStyleRef:   hairStyle,
											HairColorRef:   hairColor,
											MakeupRef:      makeup,
											ExpressionRef:  expression,
											AccessoriesRef: accessories,
											Variations:     options.Variations,
											SendOriginal:   options.SendOriginal,
											Debug:          options.DebugPrompt,
											OutputDir:      outputDir,
										}

									// Display current combination
									fmt.Printf("\n🎨 Processing combination:\n")
									fmt.Printf("   Subject: %s\n", filepath.Base(subject))
									if outfit != "" {
										fmt.Printf("   Outfit: %s\n", filepath.Base(outfit))
									}
									if overOutfit != "" {
										fmt.Printf("   Over-outfit: %s\n", filepath.Base(overOutfit))
									}
									if style != "" {
										fmt.Printf("   Style: %s\n", filepath.Base(style))
									}
									if hairStyle != "" {
										fmt.Printf("   Hair style: %s\n", filepath.Base(hairStyle))
									}
									if hairColor != "" {
										fmt.Printf("   Hair color: %s\n", filepath.Base(hairColor))
									}
									if makeup != "" {
										fmt.Printf("   Makeup: %s\n", filepath.Base(makeup))
									}
									if expression != "" {
										fmt.Printf("   Expression: %s\n", filepath.Base(expression))
									}
									if accessories != "" {
										fmt.Printf("   Accessories: %s\n", filepath.Base(accessories))
									}

									// Run modular workflow
									results, err := o.RunModularWorkflow(config)
									if err != nil {
										fmt.Printf("   ❌ Error: %v\n", err)
										continue
									}

									// Add results to workflow
									for _, outputPath := range results {
										result.Steps = append(result.Steps, StepResult{
											Type:       "generation",
											Name:       "modular",
											OutputPath: outputPath,
											Message:    fmt.Sprintf("Generated %s", filepath.Base(outputPath)),
										})
										generatedCount++
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Set result counts
	result.SubjectCount = len(targetImages)
	result.OutfitCount = maxInt(1, len(outfitFiles))
	result.StyleCount = maxInt(1, len(styleFiles))
	result.VariationCount = options.Variations
	result.EndTime = time.Now()

	return result, nil
}

// collectFilesForComponent collects files from a path (file or directory) or handles text descriptions
func collectFilesForComponent(path string, componentType string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}

	// For style, always treat as file path
	if componentType == "style" || componentType == "visual_style" {
		// Check if it's a file or directory
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%s path does not exist: %s", componentType, path)
			}
			return nil, err
		}

		if info.IsDir() {
			// Collect all image files from directory
			files, err := collectImageFiles(path)
			if err != nil {
				return nil, err
			}
			if len(files) == 0 {
				return nil, fmt.Errorf("no image files found in %s directory: %s", componentType, path)
			}
			return files, nil
		}

		// Single file
		return []string{path}, nil
	}

	// For other components, check if it's a file path or text description
	if !isFilePath(path) {
		// It's a text description, return it as-is
		return []string{path}, nil
	}

	// Check if it's a file or directory
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If it doesn't exist as a file, treat it as text description
			return []string{path}, nil
		}
		return nil, err
	}

	if info.IsDir() {
		// Collect all image files from directory
		files, err := collectImageFiles(path)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no image files found in %s directory: %s", componentType, path)
		}
		return files, nil
	}

	// Single file
	return []string{path}, nil
}

// ensureAtLeastOne returns the input slice or a slice with one empty string if input is empty
func ensureAtLeastOne(files []string) []string {
	if len(files) == 0 {
		return []string{""}
	}
	return files
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// hasModularComponents checks if any modular components are specified
func hasModularComponents(options WorkflowOptions) bool {
	return options.HairStyleRef != "" ||
		options.HairColorRef != "" ||
		options.MakeupRef != "" ||
		options.ExpressionRef != "" ||
		options.AccessoriesRef != "" ||
		options.OverOutfitRef != ""
}