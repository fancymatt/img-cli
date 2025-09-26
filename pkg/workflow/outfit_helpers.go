package workflow

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

// collectImageFiles collects image files from a path (single file or directory)
func collectImageFiles(path string) ([]string, error) {
	if path == "" {
		return []string{""}, nil // Empty string for default/no file
	}

	fileInfo, err := gemini.GetFileInfo(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", path, err)
	}

	if !fileInfo.IsDir() {
		return []string{path}, nil
	}

	// Directory - get all images
	images, err := gemini.GetImagesFromDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no image files found in directory %s", path)
	}
	return images, nil
}

// buildOutfitPrompt builds a detailed outfit prompt from analysis data
func buildOutfitPrompt(outfit *gemini.OutfitDescription) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("wearing exactly: ")

	// Include clothing items
	if len(outfit.Clothing) > 0 {
		for i, item := range outfit.Clothing {
			if i > 0 {
				promptBuilder.WriteString("; ")
			}
			appendClothingItem(&promptBuilder, item)
		}
	}

	// Add color specifications
	if len(outfit.Colors) > 0 {
		promptBuilder.WriteString(". CRITICAL COLOR REQUIREMENTS: ")
		promptBuilder.WriteString(strings.Join(outfit.Colors, ", "))
	}

	// Add accessories
	if len(outfit.Accessories) > 0 {
		promptBuilder.WriteString(". Accessories: ")
		for i, acc := range outfit.Accessories {
			if i > 0 {
				promptBuilder.WriteString(", ")
			}
			appendAccessoryItem(&promptBuilder, acc)
		}
	}

	// Add overall description
	if outfit.Overall != "" {
		promptBuilder.WriteString(". Overall styling: ")
		promptBuilder.WriteString(outfit.Overall)
	}

	// Add style notes
	if outfit.Style != "" {
		promptBuilder.WriteString(". Style notes: ")
		promptBuilder.WriteString(outfit.Style)
	}

	result := promptBuilder.String()
	if result == "wearing exactly: " {
		return "wearing the same outfit as shown in the reference image"
	}
	return result
}

// appendClothingItem appends clothing item details to the builder
func appendClothingItem(builder *strings.Builder, item interface{}) {
	switch v := item.(type) {
	case string:
		builder.WriteString(v)
	case map[string]interface{}:
		if desc, ok := v["description"].(string); ok {
			builder.WriteString(desc)
		} else if itemName, ok := v["item"].(string); ok {
			builder.WriteString(itemName)
		}

		// Add color details
		if mainColor, ok := v["main_body_color"].(string); ok && mainColor != "" && mainColor != "none" {
			builder.WriteString(fmt.Sprintf(" with %s main body", mainColor))
		}
		if collarColor, ok := v["collar_color"].(string); ok && collarColor != "" && collarColor != "none" {
			builder.WriteString(fmt.Sprintf(", %s collar", collarColor))
		}
		if trimColor, ok := v["trim_color"].(string); ok && trimColor != "" && trimColor != "none" {
			builder.WriteString(fmt.Sprintf(", %s trim", trimColor))
		}
	}
}

// appendAccessoryItem appends accessory item to the builder
func appendAccessoryItem(builder *strings.Builder, item interface{}) {
	switch v := item.(type) {
	case string:
		builder.WriteString(v)
	case map[string]interface{}:
		if desc, ok := v["description"].(string); ok {
			builder.WriteString(desc)
		} else if itemName, ok := v["item"].(string); ok {
			builder.WriteString(itemName)
		}
	}
}

// extractOutfitPromptAndHair extracts the outfit prompt and hair data from outfit analysis
func extractOutfitPromptAndHair(outfitData json.RawMessage) (string, json.RawMessage) {
	var outfit gemini.OutfitDescription
	if err := json.Unmarshal(outfitData, &outfit); err != nil {
		// Try to use the raw JSON as a string
		var rawText string
		if json.Unmarshal(outfitData, &rawText) == nil && rawText != "" {
			return rawText, nil
		}
		return "wearing the same outfit as shown in the reference image", nil
	}

	outfitPrompt := buildOutfitPrompt(&outfit)

	var hairData json.RawMessage
	if outfit.Hair != nil {
		hairData, _ = json.Marshal(outfit.Hair)
	}

	return outfitPrompt, hairData
}

// extractHairFromAnalysis extracts hair data from an outfit analysis
func extractHairFromAnalysis(analysisData json.RawMessage) json.RawMessage {
	var outfit gemini.OutfitDescription
	if err := json.Unmarshal(analysisData, &outfit); err == nil && outfit.Hair != nil {
		hairData, _ := json.Marshal(outfit.Hair)
		return hairData
	}
	return nil
}