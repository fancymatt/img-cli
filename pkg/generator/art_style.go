package generator

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ArtStyleGenerator struct {
	BaseGenerator
	client *gemini.Client
}

func NewArtStyleGenerator(client *gemini.Client) *ArtStyleGenerator {
	return &ArtStyleGenerator{
		BaseGenerator: BaseGenerator{Type: "art_style"},
		client:        client,
	}
}

func (a *ArtStyleGenerator) Generate(params GenerateParams) (*GenerateResult, error) {
	// This generator can work in two modes:
	// 1. Text-to-image with style reference
	// 2. Image-to-image style transfer

	var request gemini.Request

	if params.ImagePath != "" && !strings.HasSuffix(params.ImagePath, ".json") {
		// Image-to-image style transfer mode
		request = a.createImageStyleTransferRequest(params)
	} else {
		// Text-to-image with style mode
		request = a.createTextToImageWithStyleRequest(params)
	}

	resp, err := a.client.SendRequest(request)
	if err != nil {
		return nil, fmt.Errorf("error generating styled image: %w", err)
	}

	// Extract generated image
	imageData := gemini.ExtractImageFromResponse(resp)
	if imageData == nil {
		textResp := gemini.ExtractTextFromResponse(resp)
		if textResp != "" {
			return &GenerateResult{
				Message:    fmt.Sprintf("Response: %s", textResp),
				OutputPath: "",
			}, fmt.Errorf("no image generated")
		}
		return nil, fmt.Errorf("no image generated in response")
	}

	// Create output directory
	if params.OutputDir == "" {
		now := time.Now()
		dateFolder := now.Format("2006-01-02")
		timestampFolder := now.Format("150405")
		params.OutputDir = filepath.Join("output", dateFolder, timestampFolder)
	}

	if err := os.MkdirAll(params.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	// Generate output filename
	baseName := "styled_image"
	if params.ImagePath != "" {
		baseName = strings.TrimSuffix(filepath.Base(params.ImagePath), filepath.Ext(params.ImagePath))
		baseName = fmt.Sprintf("%s_styled", baseName)
	} else if params.Prompt != "" {
		// Use first few words of prompt for filename
		words := strings.Fields(params.Prompt)
		if len(words) > 3 {
			words = words[:3]
		}
		baseName = strings.Join(words, "_")
		baseName = strings.ToLower(baseName)
		// Sanitize
		baseName = strings.ReplaceAll(baseName, "/", "_")
		baseName = strings.ReplaceAll(baseName, "\\", "_")
		baseName = strings.ReplaceAll(baseName, ".", "_")
	}

	outputPath := filepath.Join(params.OutputDir, fmt.Sprintf("%s.png", baseName))

	// Handle file conflicts
	if _, err := os.Stat(outputPath); err == nil {
		timestamp := time.Now().Format("150405")
		outputPath = filepath.Join(params.OutputDir, fmt.Sprintf("%s_%s.png", baseName, timestamp))
	}

	if err := os.WriteFile(outputPath, imageData.Data, 0644); err != nil {
		return nil, fmt.Errorf("error saving image: %w", err)
	}

	return &GenerateResult{
		Message:    "Styled image generated successfully",
		OutputPath: outputPath,
	}, nil
}

func (a *ArtStyleGenerator) createTextToImageWithStyleRequest(params GenerateParams) gemini.Request {
	parts := []interface{}{}

	// Add style reference image if provided
	if params.StyleReference != "" {
		styleData, mimeType, err := gemini.LoadImageAsBase64(params.StyleReference)
		if err == nil {
			parts = append(parts, gemini.BlobPart{
				InlineData: gemini.InlineData{
					MimeType: mimeType,
					Data:     styleData,
				},
			})
		}
	}

	// Build the prompt
	promptText := a.buildTextToImagePrompt(params)
	parts = append(parts, gemini.TextPart{Text: promptText})

	return gemini.Request{
		Contents: []gemini.Content{
			{Parts: parts},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.8,
			TopK:        40,
			TopP:        0.95,
		},
	}
}

func (a *ArtStyleGenerator) createImageStyleTransferRequest(params GenerateParams) gemini.Request {
	parts := []interface{}{}

	// Load the input image
	imageData, mimeType, err := gemini.LoadImageAsBase64(params.ImagePath)
	if err != nil {
		// Return request with error in prompt
		return gemini.Request{
			Contents: []gemini.Content{
				{Parts: []interface{}{
					gemini.TextPart{Text: fmt.Sprintf("Error loading image: %v", err)},
				}},
			},
		}
	}

	parts = append(parts, gemini.BlobPart{
		InlineData: gemini.InlineData{
			MimeType: mimeType,
			Data:     imageData,
		},
	})

	// Add style reference if provided
	if params.StyleReference != "" {
		styleData, styleMimeType, err := gemini.LoadImageAsBase64(params.StyleReference)
		if err == nil {
			parts = append(parts, gemini.BlobPart{
				InlineData: gemini.InlineData{
					MimeType: styleMimeType,
					Data:     styleData,
				},
			})
		}
	}

	// Build the prompt
	promptText := a.buildImageStyleTransferPrompt(params)
	parts = append(parts, gemini.TextPart{Text: promptText})

	return gemini.Request{
		Contents: []gemini.Content{
			{Parts: parts},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.7,
			TopK:        35,
			TopP:        0.9,
		},
	}
}

func (a *ArtStyleGenerator) buildTextToImagePrompt(params GenerateParams) string {
	prompt := params.Prompt
	if prompt == "" {
		prompt = "a beautiful landscape"
	}

	// Parse style analysis if provided
	styleDescription := a.parseStyleDescription(params)

	if params.StyleReference != "" && styleDescription != "" {
		return fmt.Sprintf(`Using the artistic style shown in the reference image, create a new illustration of: %s

Style Analysis:
%s

CRITICAL INSTRUCTIONS:
- Match the EXACT artistic style, medium, and technique from the reference
- Use the same color palette approach and visual characteristics
- Apply the same level of stylization and detail
- Maintain consistency with the line work, shading, and texture techniques
- The subject should be: %s
- But the STYLE must perfectly match the reference

Generate a high-quality image that looks like it was created by the same artist using the same techniques.`, prompt, styleDescription, prompt)
	}

	if styleDescription != "" {
		return fmt.Sprintf(`Create an illustration in this specific artistic style:

%s

Subject: %s

Apply all the stylistic elements described above to create a cohesive artwork that perfectly embodies this artistic approach.`, styleDescription, prompt)
	}

	// Fallback if no style information
	return fmt.Sprintf(`Create a beautiful artistic illustration of: %s

Use a distinctive and appealing art style with careful attention to composition, color, and technique.`, prompt)
}

func (a *ArtStyleGenerator) buildImageStyleTransferPrompt(params GenerateParams) string {
	styleDescription := a.parseStyleDescription(params)

	if params.StyleReference != "" && styleDescription != "" {
		return fmt.Sprintf(`Transform the first image into the artistic style of the second reference image.

Style to Apply:
%s

CRITICAL INSTRUCTIONS:
- Keep the EXACT same subject, composition, and content from the first image
- Transform ONLY the artistic style to match the reference
- Apply the reference's medium, technique, color approach, and visual characteristics
- Match the line work, shading, textures, and overall aesthetic
- The result should look like the original subject was illustrated by the artist of the style reference

Generate a high-quality transformation that perfectly applies the reference style while preserving the original content.`)
	}

	if styleDescription != "" {
		return fmt.Sprintf(`Transform this image into the following artistic style:

%s

Keep the subject and composition identical, but completely reimagine it in the described artistic style.
Apply all the stylistic elements to create a cohesive transformation.`, styleDescription)
	}

	// Fallback
	return `Transform this image into a beautiful artistic illustration.
Keep the subject and composition, but apply an appealing artistic style with attention to color, technique, and visual appeal.`
}

func (a *ArtStyleGenerator) parseStyleDescription(params GenerateParams) string {
	if params.StyleAnalysis == nil {
		return ""
	}

	var styleData map[string]interface{}
	if err := json.Unmarshal(params.StyleAnalysis, &styleData); err != nil {
		return ""
	}

	var desc []string

	// Extract key style elements for the prompt
	if styleName, ok := styleData["style_name"].(string); ok {
		desc = append(desc, fmt.Sprintf("Style: %s", styleName))
	}

	if medium, ok := styleData["medium"].(string); ok {
		desc = append(desc, fmt.Sprintf("Medium: %s", medium))
	}

	if technique, ok := styleData["technique"].(map[string]interface{}); ok {
		if lineWork, ok := technique["line_work"].(string); ok {
			desc = append(desc, fmt.Sprintf("Line work: %s", lineWork))
		}
		if shading, ok := technique["shading"].(string); ok {
			desc = append(desc, fmt.Sprintf("Shading: %s", shading))
		}
		if textures, ok := technique["textures"].(string); ok {
			desc = append(desc, fmt.Sprintf("Textures: %s", textures))
		}
	}

	if colorApproach, ok := styleData["color_approach"].(map[string]interface{}); ok {
		if paletteType, ok := colorApproach["palette_type"].(string); ok {
			desc = append(desc, fmt.Sprintf("Color palette: %s", paletteType))
		}
	}

	if movement, ok := styleData["artistic_movement"].(string); ok {
		desc = append(desc, fmt.Sprintf("Artistic movement: %s", movement))
	}

	if visual, ok := styleData["visual_characteristics"].(map[string]interface{}); ok {
		if detail, ok := visual["level_of_detail"].(string); ok {
			desc = append(desc, fmt.Sprintf("Detail level: %s", detail))
		}
		if stylization, ok := visual["stylization"].(string); ok {
			desc = append(desc, fmt.Sprintf("Stylization: %s", stylization))
		}
	}

	if distinctive, ok := styleData["distinctive_features"].([]interface{}); ok {
		features := make([]string, 0, len(distinctive))
		for _, f := range distinctive {
			if feature, ok := f.(string); ok {
				features = append(features, feature)
			}
		}
		if len(features) > 0 {
			desc = append(desc, fmt.Sprintf("Distinctive features: %s", strings.Join(features, ", ")))
		}
	}

	return strings.Join(desc, "\n")
}