package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type MakeupAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewMakeupAnalyzer(client *gemini.Client) *MakeupAnalyzer {
	return &MakeupAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "makeup"},
		client:       client,
	}
}

func (m *MakeupAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze ONLY the makeup in this image with extreme precision. Ignore all other elements including clothing, hair, and accessories. Return a JSON object with the following structure:
{
  "complexion": {
    "foundation": "coverage level and finish (e.g., 'full coverage matte', 'sheer dewy', 'medium coverage satin')",
    "concealer": "concealer placement and coverage",
    "powder": "powder type and application (e.g., 'translucent setting powder', 'pressed powder')",
    "blush": "blush color and placement (e.g., 'peachy pink on apples of cheeks', 'dusty rose draping')",
    "bronzer": "bronzer shade and placement if visible",
    "highlighter": "highlighter placement and intensity (e.g., 'champagne gold on cheekbones', 'subtle inner corner highlight')",
    "contour": "contour placement and intensity if visible"
  },
  "eyes": {
    "eyeshadow": "detailed eyeshadow colors and placement (e.g., 'warm brown in crease, champagne on lid, dark brown on outer V')",
    "eyeliner": "liner style and color (e.g., 'black winged liner', 'brown smudged liner', 'tightlined upper lash line')",
    "mascara": "mascara effect (e.g., 'volumizing black mascara', 'lengthening brown mascara')",
    "lashes": "false lashes or extensions if visible",
    "brows": "eyebrow styling and color (e.g., 'naturally filled arch', 'bold defined brows', 'feathered brows')"
  },
  "lips": {
    "color": "lip color and finish (e.g., 'nude pink matte', 'berry red gloss', 'mauve satin')",
    "liner": "lip liner if visible",
    "finish": "texture and finish (e.g., 'glossy', 'matte', 'velvet', 'stained')",
    "shape": "lip shape enhancement if any"
  },
  "style": "overall makeup style (e.g., 'natural no-makeup makeup', 'glamorous evening', 'editorial', 'soft romantic')",
  "overall": "comprehensive makeup description including the complete look, techniques used, and aesthetic achieved"
}

IMPORTANT:
- Focus ONLY on makeup elements
- Be extremely specific about colors, techniques, and placement
- Describe actual makeup application, not natural features
- Use professional makeup terminology`,
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

	resp, err := m.client.SendRequest(request)
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