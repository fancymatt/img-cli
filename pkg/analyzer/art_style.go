package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type ArtStyleAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewArtStyleAnalyzer(client *gemini.Client) *ArtStyleAnalyzer {
	return &ArtStyleAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "art_style"},
		client:       client,
	}
}

func (a *ArtStyleAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
	imageData, mimeType, err := gemini.LoadImageAsBase64(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}

	request := gemini.Request{
		Contents: []gemini.Content{
			{
				Parts: []interface{}{
					gemini.BlobPart{
						InlineData: gemini.InlineData{
							MimeType: mimeType,
							Data:     imageData,
						},
					},
					gemini.TextPart{
						Text: `Analyze the artistic style and illustration techniques of this image in extreme detail. Return a JSON object with the following structure:
{
  "style_name": "concise name for this style (e.g., 'Retro Comic Book', 'Watercolor Fantasy', 'Digital Anime')",
  "medium": "apparent medium (e.g., digital art, watercolor, oil painting, pencil sketch, vector illustration)",
  "technique": {
    "line_work": "detailed description of line quality, weight, style (e.g., bold outlines, delicate sketching, no lines)",
    "shading": "shading techniques used (e.g., cell shading, soft gradients, cross-hatching, flat colors)",
    "textures": "texture treatment and details (e.g., paper texture, digital brushes, smooth vectors)",
    "brushwork": "brush or tool characteristics if apparent"
  },
  "color_approach": {
    "palette_type": "color palette approach (e.g., limited palette, vibrant, monochromatic, pastel)",
    "dominant_colors": ["list of main colors used"],
    "color_harmony": "color harmony type (e.g., complementary, analogous, triadic)",
    "saturation": "saturation level (e.g., highly saturated, muted, desaturated)",
    "contrast": "contrast approach (e.g., high contrast, subtle, dramatic)"
  },
  "artistic_movement": "art movement or genre it resembles (e.g., Art Nouveau, Pop Art, Impressionism, Manga, Contemporary Digital)",
  "visual_characteristics": {
    "level_of_detail": "detail level (e.g., highly detailed, minimalist, moderate)",
    "stylization": "degree of stylization vs realism",
    "perspective": "perspective approach (e.g., flat/2D, isometric, realistic perspective)",
    "composition_style": "compositional approach specific to the art style"
  },
  "influences": ["list of apparent artistic influences or similar styles"],
  "mood_aesthetic": "overall aesthetic mood created by the style",
  "distinctive_features": ["list of unique or standout stylistic elements"],
  "reproduction_notes": "specific technical notes for reproducing this style"
}

CRITICAL: Focus ONLY on the artistic style, technique, and visual language. Do NOT describe the subject matter or content, only HOW it's rendered artistically.

Be extremely specific about:
- The exact illustration/art techniques used
- Color treatment and application methods
- Line quality and characteristics
- Any unique stylistic signatures
- Technical aspects that define this style

Return ONLY the JSON object, no additional text.`,
					},
				},
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.3,
			TopK:        20,
			TopP:        0.8,
		},
	}

	resp, err := a.client.SendRequest(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	textResp := gemini.ExtractTextFromResponse(resp)
	if textResp == "" {
		return nil, fmt.Errorf("no text response from API")
	}

	// Clean the response - remove markdown code blocks if present
	cleaned := strings.TrimSpace(textResp)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}

	// Validate it's proper JSON
	var styleData map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &styleData); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return json.RawMessage(cleaned), nil
}

// AnalyzeMultiple analyzes multiple images and combines their style characteristics
func (a *ArtStyleAnalyzer) AnalyzeMultiple(imagePaths []string) (json.RawMessage, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	// Analyze each image
	var styles []json.RawMessage
	for _, path := range imagePaths {
		style, err := a.Analyze(path)
		if err != nil {
			fmt.Printf("Warning: Failed to analyze %s: %v\n", path, err)
			continue
		}
		styles = append(styles, style)
	}

	if len(styles) == 0 {
		return nil, fmt.Errorf("all analyses failed")
	}

	// If only one style, return it
	if len(styles) == 1 {
		return styles[0], nil
	}

	// Combine multiple style analyses into a comprehensive style guide
	combinedRequest := a.createCombinedAnalysisPrompt(styles)

	resp, err := a.client.SendRequest(combinedRequest)
	if err != nil {
		return nil, fmt.Errorf("error combining styles: %w", err)
	}

	textResp := gemini.ExtractTextFromResponse(resp)
	if textResp == "" {
		return nil, fmt.Errorf("no text response from API")
	}

	// Clean the response
	cleaned := strings.TrimSpace(textResp)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}

	return json.RawMessage(cleaned), nil
}

func (a *ArtStyleAnalyzer) createCombinedAnalysisPrompt(styles []json.RawMessage) gemini.Request {
	// Convert styles to string for the prompt
	var styleStrings []string
	for i, style := range styles {
		styleStrings = append(styleStrings, fmt.Sprintf("Style %d: %s", i+1, string(style)))
	}

	promptText := fmt.Sprintf(`Given these %d style analyses from related images, create a unified style guide that captures the common artistic DNA while noting variations.

%s

Create a comprehensive style guide JSON that:
1. Identifies the core consistent style elements across all samples
2. Notes any variations or range within the style
3. Provides a cohesive style name and description

Return a JSON object with this structure:
{
  "style_name": "unified name for this art style",
  "style_summary": "comprehensive description of the overall style",
  "consistent_elements": {
    "medium": "consistent medium across samples",
    "technique": "core techniques used throughout",
    "color_approach": "consistent color approach",
    "visual_characteristics": "shared visual traits"
  },
  "variations": {
    "noted_differences": ["list of variations observed between samples"],
    "flexibility": "areas where the style allows for variation"
  },
  "comprehensive_guide": {
    "essential_elements": ["must-have elements to reproduce this style"],
    "color_palette": ["comprehensive color list from all samples"],
    "technical_approach": "detailed technical guide for reproduction",
    "distinctive_signatures": ["unique elements that define this style"]
  },
  "implementation_notes": "specific instructions for applying this style to new images"
}

Return ONLY the JSON object.`, len(styles), strings.Join(styleStrings, "\n\n"))

	return gemini.Request{
		Contents: []gemini.Content{
			{
				Parts: []interface{}{
					gemini.TextPart{Text: promptText},
				},
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.4,
			TopK:        30,
			TopP:        0.85,
		},
	}
}