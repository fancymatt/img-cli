package generator

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"os"
	"path/filepath"
	"strings"
)

type StyleGuideGenerator struct {
	BaseGenerator
	client *gemini.Client
}

func NewStyleGuideGenerator(client *gemini.Client) *StyleGuideGenerator {
	return &StyleGuideGenerator{
		BaseGenerator: BaseGenerator{Type: "style_guide"},
		client:        client,
	}
}

func (s *StyleGuideGenerator) Generate(params GenerateParams) (*GenerateResult, error) {
	// StyleAnalysis should contain the JSON analysis from ArtStyleAnalyzer
	if params.StyleAnalysis == nil {
		return nil, fmt.Errorf("style analysis required for style guide generation")
	}

	// Parse the style analysis to get key information
	var styleData map[string]interface{}
	if err := json.Unmarshal(params.StyleAnalysis, &styleData); err != nil {
		return nil, fmt.Errorf("error parsing style analysis: %w", err)
	}

	styleName := "style_guide"
	if name, ok := styleData["style_name"].(string); ok {
		// Sanitize the style name for use as filename
		styleName = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
		styleName = strings.ReplaceAll(styleName, "/", "_")
		styleName = strings.ReplaceAll(styleName, "\\", "_")
	}

	// Create comprehensive prompt for style guide generation
	prompt := s.createStyleGuidePrompt(styleData)

	request := gemini.Request{
		Contents: []gemini.Content{
			{
				Parts: []interface{}{
					gemini.TextPart{Text: prompt},
				},
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.9,
			TopK:        50,
			TopP:        0.95,
		},
	}

	resp, err := s.client.SendRequest(request)
	if err != nil {
		return nil, fmt.Errorf("error generating style guide: %w", err)
	}

	// Extract generated image
	imageData := gemini.ExtractImageFromResponse(resp)
	if imageData == nil {
		textResp := gemini.ExtractTextFromResponse(resp)
		if textResp != "" {
			return &GenerateResult{
				Message:    fmt.Sprintf("Text response: %s", textResp),
				OutputPath: "",
			}, fmt.Errorf("no image generated, only text response")
		}
		return nil, fmt.Errorf("no image generated in response")
	}

	// Ensure styles directory exists
	stylesDir := "styles"
	if params.OutputDir != "" && strings.Contains(params.OutputDir, "styles") {
		stylesDir = params.OutputDir
	}

	if err := os.MkdirAll(stylesDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating styles directory: %w", err)
	}

	// Save the style guide image
	outputPath := filepath.Join(stylesDir, fmt.Sprintf("%s.png", styleName))

	// If file exists, add a number suffix
	if _, err := os.Stat(outputPath); err == nil {
		for i := 2; ; i++ {
			testPath := filepath.Join(stylesDir, fmt.Sprintf("%s_%d.png", styleName, i))
			if _, err := os.Stat(testPath); os.IsNotExist(err) {
				outputPath = testPath
				break
			}
		}
	}

	if err := os.WriteFile(outputPath, imageData.Data, 0644); err != nil {
		return nil, fmt.Errorf("error saving style guide: %w", err)
	}

	// Also save the style analysis JSON alongside the image
	jsonPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".json"
	if err := os.WriteFile(jsonPath, params.StyleAnalysis, 0644); err != nil {
		// Non-fatal error
		fmt.Printf("Warning: Could not save style analysis JSON: %v\n", err)
	}

	return &GenerateResult{
		Message:    fmt.Sprintf("Style guide '%s' generated successfully", styleName),
		OutputPath: outputPath,
	}, nil
}

func (s *StyleGuideGenerator) createStyleGuidePrompt(styleData map[string]interface{}) string {
	// Extract key style information
	styleName := "the style"
	if name, ok := styleData["style_name"].(string); ok {
		styleName = name
	}

	// Build a detailed description from the analysis
	var styleDesc []string

	if medium, ok := styleData["medium"].(string); ok {
		styleDesc = append(styleDesc, fmt.Sprintf("Medium: %s", medium))
	}

	if technique, ok := styleData["technique"].(map[string]interface{}); ok {
		if lineWork, ok := technique["line_work"].(string); ok {
			styleDesc = append(styleDesc, fmt.Sprintf("Line work: %s", lineWork))
		}
		if shading, ok := technique["shading"].(string); ok {
			styleDesc = append(styleDesc, fmt.Sprintf("Shading: %s", shading))
		}
	}

	if colorApproach, ok := styleData["color_approach"].(map[string]interface{}); ok {
		if paletteType, ok := colorApproach["palette_type"].(string); ok {
			styleDesc = append(styleDesc, fmt.Sprintf("Palette: %s", paletteType))
		}
		if colors, ok := colorApproach["dominant_colors"].([]interface{}); ok {
			colorList := make([]string, 0, len(colors))
			for _, c := range colors {
				if colorStr, ok := c.(string); ok {
					colorList = append(colorList, colorStr)
				}
			}
			if len(colorList) > 0 {
				styleDesc = append(styleDesc, fmt.Sprintf("Colors: %s", strings.Join(colorList, ", ")))
			}
		}
	}

	if movement, ok := styleData["artistic_movement"].(string); ok {
		styleDesc = append(styleDesc, fmt.Sprintf("Movement: %s", movement))
	}

	// Create the comprehensive prompt
	prompt := fmt.Sprintf(`Create a comprehensive style guide sample sheet for "%s" art style.

Style Characteristics:
%s

Generate a single image that serves as a complete style guide reference sheet. The image should include:

1. A 3x3 grid layout containing 9 different example illustrations, each demonstrating the style with:
   - Different subjects (mix of: characters, objects, landscapes, abstract patterns)
   - Consistent application of the artistic style
   - Various color combinations from the style's palette
   - Different compositions showing the style's versatility

2. Each panel should clearly showcase:
   - The characteristic line work and shading techniques
   - The color approach and palette
   - The level of detail and stylization
   - Any distinctive features of the style

3. The overall composition should:
   - Be clean and well-organized like a professional style guide
   - Have subtle dividers or spacing between panels
   - Include small style notation marks or artistic flourishes that are characteristic of the style
   - Feel cohesive while showing range

Make this a high-quality reference sheet that an artist could use to understand and reproduce this style. Each panel should be a complete mini-illustration, not just sketches or incomplete samples.

The style must be consistently applied across all 9 panels, creating a unified aesthetic that clearly demonstrates "%s" as a distinct artistic approach.

Generate the image at high resolution with clear, crisp details in each panel.`, styleName, strings.Join(styleDesc, "\n"), styleName)

	return prompt
}

// GenerateBatch creates multiple style guide variations
func (s *StyleGuideGenerator) GenerateBatch(params GenerateParams, count int) ([]*GenerateResult, error) {
	results := make([]*GenerateResult, 0, count)

	for i := 0; i < count; i++ {
		result, err := s.Generate(params)
		if err != nil {
			fmt.Printf("Warning: Failed to generate style guide variation %d: %v\n", i+1, err)
			continue
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("all style guide generations failed")
	}

	return results, nil
}