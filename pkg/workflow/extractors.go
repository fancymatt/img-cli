package workflow

import (
	"encoding/json"
	"fmt"
	"strings"
)

// extractOutfitDescription extracts outfit description from analysis
func (o *Orchestrator) extractOutfitDescription(data json.RawMessage) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "Standard outfit"
	}

	var parts []string

	// Check if it's a cached entry with nested structure
	if analysisData, ok := result["analysis"].(map[string]interface{}); ok {
		// It's a cached entry with analysis nested
		if clothing, ok := analysisData["clothing"].([]interface{}); ok {
			for _, item := range clothing {
				if str, ok := item.(string); ok {
					parts = append(parts, str)
				}
			}
		}
		if overall, ok := analysisData["overall"].(string); ok && overall != "" {
			parts = append(parts, overall)
		}
	} else {
		// Direct structure (not cached)
		if clothing, ok := result["clothing"].([]interface{}); ok {
			for _, item := range clothing {
				if str, ok := item.(string); ok {
					parts = append(parts, str)
				}
			}
		}
		if overall, ok := result["overall"].(string); ok && overall != "" {
			parts = append(parts, overall)
		}
	}

	// Also check for description field (in cached data)
	if desc, ok := result["description"].(string); ok && desc != "" && len(parts) == 0 {
		parts = append(parts, desc)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "Standard outfit"
}

// extractStyleDescription extracts visual style description from analysis
func (o *Orchestrator) extractStyleDescription(data json.RawMessage) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "Natural photographic style"
	}

	var parts []string

	if lighting, ok := result["lighting"].(string); ok && lighting != "" {
		parts = append(parts, fmt.Sprintf("Lighting: %s", lighting))
	}

	if background, ok := result["background"].(string); ok && background != "" {
		parts = append(parts, fmt.Sprintf("Background: %s", background))
	}

	if mood, ok := result["mood"].(string); ok && mood != "" {
		parts = append(parts, fmt.Sprintf("Mood: %s", mood))
	}

	if overall, ok := result["overall_style"].(string); ok && overall != "" {
		parts = append(parts, overall)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "Natural photographic style"
}

// extractHairStyleDescription extracts hair style description from analysis
func (o *Orchestrator) extractHairStyleDescription(data json.RawMessage) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "Natural hairstyle"
	}

	var parts []string

	// Check if it's a cached entry with nested structure
	var analysisData map[string]interface{}
	if analysis, ok := result["analysis"].(map[string]interface{}); ok {
		// It's a cached entry with analysis nested
		analysisData = analysis
	} else {
		// Direct structure (not cached)
		analysisData = result
	}

	if style, ok := analysisData["style"].(string); ok && style != "" {
		parts = append(parts, style)
	}

	if length, ok := analysisData["length"].(string); ok && length != "" {
		parts = append(parts, fmt.Sprintf("Length: %s", length))
	}

	if texture, ok := analysisData["texture"].(string); ok && texture != "" {
		parts = append(parts, fmt.Sprintf("Texture: %s", texture))
	}

	if volume, ok := analysisData["volume"].(string); ok && volume != "" {
		parts = append(parts, fmt.Sprintf("Volume: %s", volume))
	}

	if overall, ok := analysisData["overall"].(string); ok && overall != "" {
		parts = append(parts, overall)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "Natural hairstyle"
}

// extractHairColorDescription extracts hair color description from analysis
func (o *Orchestrator) extractHairColorDescription(data json.RawMessage) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "Natural hair color"
	}

	var parts []string

	if baseColor, ok := result["base_color"].(string); ok && baseColor != "" {
		parts = append(parts, fmt.Sprintf("Base color: %s", baseColor))
	}

	if highlights, ok := result["highlights"].(string); ok && highlights != "" {
		parts = append(parts, fmt.Sprintf("Highlights: %s", highlights))
	}

	if technique, ok := result["technique"].(string); ok && technique != "" {
		parts = append(parts, fmt.Sprintf("Coloring technique: %s", technique))
	}

	if overall, ok := result["overall"].(string); ok && overall != "" {
		parts = append(parts, overall)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "Natural hair color"
}

// extractMakeupDescription extracts makeup description from analysis
func (o *Orchestrator) extractMakeupDescription(data json.RawMessage) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "Natural makeup"
	}

	var parts []string

	// Extract complexion details
	if complexion, ok := result["complexion"].(map[string]interface{}); ok {
		var complexionParts []string
		if foundation, ok := complexion["foundation"].(string); ok && foundation != "" {
			complexionParts = append(complexionParts, fmt.Sprintf("Foundation: %s", foundation))
		}
		if blush, ok := complexion["blush"].(string); ok && blush != "" {
			complexionParts = append(complexionParts, fmt.Sprintf("Blush: %s", blush))
		}
		if highlighter, ok := complexion["highlighter"].(string); ok && highlighter != "" {
			complexionParts = append(complexionParts, fmt.Sprintf("Highlighter: %s", highlighter))
		}
		if len(complexionParts) > 0 {
			parts = append(parts, "Complexion: "+strings.Join(complexionParts, ", "))
		}
	}

	// Extract eye makeup
	if eyes, ok := result["eyes"].(map[string]interface{}); ok {
		var eyeParts []string
		if eyeshadow, ok := eyes["eyeshadow"].(string); ok && eyeshadow != "" {
			eyeParts = append(eyeParts, fmt.Sprintf("Eyeshadow: %s", eyeshadow))
		}
		if eyeliner, ok := eyes["eyeliner"].(string); ok && eyeliner != "" {
			eyeParts = append(eyeParts, fmt.Sprintf("Eyeliner: %s", eyeliner))
		}
		if mascara, ok := eyes["mascara"].(string); ok && mascara != "" {
			eyeParts = append(eyeParts, fmt.Sprintf("Mascara: %s", mascara))
		}
		if len(eyeParts) > 0 {
			parts = append(parts, "Eyes: "+strings.Join(eyeParts, ", "))
		}
	}

	// Extract lip makeup
	if lips, ok := result["lips"].(map[string]interface{}); ok {
		if color, ok := lips["color"].(string); ok && color != "" {
			parts = append(parts, fmt.Sprintf("Lips: %s", color))
		}
	}

	if style, ok := result["style"].(string); ok && style != "" {
		parts = append(parts, fmt.Sprintf("Overall style: %s", style))
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "Natural makeup"
}

// extractExpressionDescription extracts expression description from analysis
// If excludeGaze is true, gaze direction information will be filtered out
func (o *Orchestrator) extractExpressionDescription(data json.RawMessage, excludeGaze ...bool) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "Natural expression"
	}

	// Check if we should exclude gaze (when style is also specified)
	shouldExcludeGaze := len(excludeGaze) > 0 && excludeGaze[0]

	// Check if it's a cached entry with nested structure
	var analysisData map[string]interface{}
	if dataField, ok := result["data"].(map[string]interface{}); ok {
		if analysis, ok := dataField["analysis"].(map[string]interface{}); ok {
			// It's a cached entry with analysis nested under data.analysis
			analysisData = analysis
		}
	} else if analysis, ok := result["analysis"].(map[string]interface{}); ok {
		// It's a cached entry with analysis directly nested
		analysisData = analysis
	} else {
		// Direct structure (not cached)
		analysisData = result
	}

	var parts []string

	if emotion, ok := analysisData["primary_emotion"].(string); ok && emotion != "" {
		parts = append(parts, fmt.Sprintf("Primary emotion: %s", emotion))
	}

	if intensity, ok := analysisData["intensity"].(string); ok && intensity != "" {
		parts = append(parts, fmt.Sprintf("Intensity: %s", intensity))
	}

	// Extract facial features
	if features, ok := analysisData["facial_features"].(map[string]interface{}); ok {
		if eyes, ok := features["eyes"].(string); ok && eyes != "" {
			parts = append(parts, fmt.Sprintf("Eyes: %s", eyes))
		}
		if mouth, ok := features["mouth"].(string); ok && mouth != "" {
			parts = append(parts, fmt.Sprintf("Mouth: %s", mouth))
		}
	}

	// Only extract gaze if not excluded (style controls this when present)
	if !shouldExcludeGaze {
		if gaze, ok := analysisData["gaze"].(map[string]interface{}); ok {
			if direction, ok := gaze["direction"].(string); ok && direction != "" {
				parts = append(parts, fmt.Sprintf("Gaze: %s", direction))
			}
		}
	}

	if mood, ok := analysisData["mood"].(string); ok && mood != "" {
		parts = append(parts, fmt.Sprintf("Mood: %s", mood))
	}

	// Handle overall description - filter out gaze-related phrases when needed
	if overall, ok := analysisData["overall"].(string); ok && overall != "" {
		if shouldExcludeGaze {
			// Remove common gaze-related phrases
			overall = strings.ReplaceAll(overall, ", with the gaze directly engaging the viewer in this moment of astonishment", "")
			overall = strings.ReplaceAll(overall, ", with the gaze directly engaging the viewer", "")
			overall = strings.ReplaceAll(overall, " with the gaze directly engaging the viewer", "")
			overall = strings.ReplaceAll(overall, ", gazing directly at the camera", "")
			overall = strings.ReplaceAll(overall, " gazing directly at the camera", "")
			overall = strings.ReplaceAll(overall, ", looking directly at the viewer", "")
			overall = strings.ReplaceAll(overall, " looking directly at the viewer", "")
			overall = strings.ReplaceAll(overall, ", eyes locked on the camera", "")
			overall = strings.ReplaceAll(overall, " eyes locked on the camera", "")
		}
		parts = append(parts, overall)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "Natural expression"
}

// extractAccessoriesDescription extracts accessories description from analysis
func (o *Orchestrator) extractAccessoriesDescription(data json.RawMessage) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "No accessories"
	}

	var parts []string

	// Extract jewelry
	if jewelry, ok := result["jewelry"].(map[string]interface{}); ok {
		var jewelryParts []string
		if earrings, ok := jewelry["earrings"].(string); ok && earrings != "" {
			jewelryParts = append(jewelryParts, fmt.Sprintf("Earrings: %s", earrings))
		}
		if necklaces, ok := jewelry["necklaces"].(string); ok && necklaces != "" {
			jewelryParts = append(jewelryParts, fmt.Sprintf("Necklaces: %s", necklaces))
		}
		if bracelets, ok := jewelry["bracelets"].(string); ok && bracelets != "" {
			jewelryParts = append(jewelryParts, fmt.Sprintf("Bracelets: %s", bracelets))
		}
		if rings, ok := jewelry["rings"].(string); ok && rings != "" {
			jewelryParts = append(jewelryParts, fmt.Sprintf("Rings: %s", rings))
		}
		if len(jewelryParts) > 0 {
			parts = append(parts, "Jewelry: "+strings.Join(jewelryParts, ", "))
		}
	}

	// Extract other accessories
	if bags, ok := result["bags"].(string); ok && bags != "" {
		parts = append(parts, fmt.Sprintf("Bags: %s", bags))
	}

	if belts, ok := result["belts"].(string); ok && belts != "" {
		parts = append(parts, fmt.Sprintf("Belts: %s", belts))
	}

	if scarves, ok := result["scarves"].(string); ok && scarves != "" {
		parts = append(parts, fmt.Sprintf("Scarves: %s", scarves))
	}

	if hats, ok := result["hats"].(string); ok && hats != "" {
		parts = append(parts, fmt.Sprintf("Hats: %s", hats))
	}

	if watches, ok := result["watches"].(string); ok && watches != "" {
		parts = append(parts, fmt.Sprintf("Watches: %s", watches))
	}

	if overall, ok := result["overall"].(string); ok && overall != "" {
		parts = append(parts, overall)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}

	return "No accessories"
}