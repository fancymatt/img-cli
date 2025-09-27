package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type HairStyleAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewHairStyleAnalyzer(client *gemini.Client) *HairStyleAnalyzer {
	return &HairStyleAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "hair_style"},
		client:       client,
	}
}

func (h *HairStyleAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze ONLY the hairstyle structure and styling in this image. COMPLETELY IGNORE hair color - focus exclusively on the cut, shape, and styling. Return a JSON object with the following structure:
{
  "style": "detailed hairstyle name and description (e.g., 'sleek low bun with face-framing tendrils', 'tousled beach waves', 'slicked-back pompadour')",
  "length": "specific length description (e.g., 'shoulder-length', 'pixie cut', 'waist-length', 'chin-length bob')",
  "texture": "hair texture and treatment (e.g., 'straightened smooth', 'natural waves', 'tight curls', 'crimped')",
  "volume": "volume and body description (e.g., 'voluminous with teased crown', 'sleek and flat', 'full-bodied')",
  "layers": "layering and cut details (e.g., 'long layers', 'blunt cut', 'feathered', 'graduated bob')",
  "parting": "part style if visible (e.g., 'deep side part', 'center part', 'zigzag part', 'no visible part')",
  "styling_technique": "how the hair is styled (e.g., 'blow-dried smooth', 'air-dried natural', 'heat-styled curls', 'braided', 'twisted')",
  "front_styling": "how front/bangs are styled (e.g., 'side-swept bangs', 'curtain bangs', 'pulled back', 'face-framing layers')",
  "accessories": "hair accessories only if they affect the style (e.g., 'held with pearl clips', 'secured with elastic', 'decorated with flowers')",
  "overall": "comprehensive description of the complete hairstyle focusing on cut, shape, and styling techniques"
}

IMPORTANT:
- Focus ONLY on hairstyle structure, NOT color
- Describe the cut, shape, and styling method
- Do not mention hair color at all
- Include styling techniques and how the hair is arranged`,
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