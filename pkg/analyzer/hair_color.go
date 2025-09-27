package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type HairColorAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewHairColorAnalyzer(client *gemini.Client) *HairColorAnalyzer {
	return &HairColorAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "hair_color"},
		client:       client,
	}
}

func (h *HairColorAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze ONLY the hair color and coloring in this image. IGNORE hairstyle, cut, and shape completely - focus only on the color, tones, and coloring technique. Return a JSON object with the following structure:
{
  "base_color": "primary hair color (e.g., 'dark brown', 'platinum blonde', 'jet black', 'auburn', 'strawberry blonde')",
  "undertones": "color undertones (e.g., 'ash', 'warm golden', 'cool', 'neutral', 'red undertones')",
  "highlights": "highlight colors and placement if present (e.g., 'caramel highlights throughout', 'face-framing blonde highlights', 'subtle sun-kissed streaks')",
  "lowlights": "lowlight colors if present (e.g., 'chocolate brown lowlights', 'deeper auburn strands')",
  "technique": "coloring technique if apparent (e.g., 'balayage', 'ombre', 'solid color', 'foiled highlights', 'color melt', 'babylights')",
  "dimension": "color dimension and variation (e.g., 'multi-dimensional', 'solid uniform color', 'natural variation')",
  "roots": "root color if different (e.g., 'darker roots', 'grown-out roots', 'shadow root', 'matching roots')",
  "shine": "hair shine and luster (e.g., 'glossy', 'matte', 'silky sheen', 'vibrant shine')",
  "special_effects": "any special color effects (e.g., 'pearlescent sheen', 'metallic tones', 'fashion colors', 'rainbow highlights')",
  "overall": "comprehensive description of the complete hair color including all tones, techniques, and effects"
}

IMPORTANT:
- Focus ONLY on hair color, NOT style or cut
- Describe colors, tones, and coloring techniques
- Do not mention hairstyle, length, or texture
- Be specific about color placement and technique`,
					},
				},
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.1,
			TopP:        0.95,
			TopK:        20,
		},
	}

	resp, err := h.client.SendRequest(request)
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

	// Validate it's JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		// If not valid JSON, return an error
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return json.RawMessage(cleaned), nil
}