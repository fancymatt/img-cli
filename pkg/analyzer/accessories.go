package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type AccessoriesAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewAccessoriesAnalyzer(client *gemini.Client) *AccessoriesAnalyzer {
	return &AccessoriesAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "accessories"},
		client:       client,
	}
}

func (a *AccessoriesAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze ONLY the accessories in this image with extreme precision. Ignore clothing items, hair, and makeup. Focus on accessories like jewelry, bags, belts, scarves, hats, watches, etc. Return a JSON object with the following structure:
{
  "jewelry": {
    "earrings": "detailed description (e.g., 'gold hoop earrings with pearl drops', 'diamond studs')",
    "necklaces": "detailed description (e.g., 'layered gold chains', 'pearl choker', 'pendant necklace')",
    "bracelets": "detailed description (e.g., 'silver tennis bracelet', 'leather wrap bracelet')",
    "rings": "detailed description (e.g., 'stacked gold bands', 'statement cocktail ring')",
    "other": "any other jewelry items"
  },
  "bags": "detailed bag description (e.g., 'black leather crossbody with gold hardware', 'canvas tote')",
  "belts": "detailed belt description (e.g., 'brown leather belt with brass buckle', 'chain belt')",
  "scarves": "scarf description if present (e.g., 'silk printed scarf', 'cashmere wrap')",
  "hats": "hat description if present (e.g., 'wide-brim fedora', 'baseball cap', 'beret')",
  "watches": "watch description if present (e.g., 'gold dress watch', 'leather strap chronograph')",
  "eyewear": "glasses or sunglasses if present (e.g., 'tortoiseshell frames', 'aviator sunglasses')",
  "gloves": "glove description if present (e.g., 'black leather gloves', 'lace gloves')",
  "other": [
    "list of any other accessories not covered above"
  ],
  "materials": "primary materials used in accessories (e.g., 'gold-toned metals', 'leather', 'pearls')",
  "style": "overall accessory style (e.g., 'minimalist', 'statement', 'vintage', 'modern')",
  "overall": "comprehensive description of how accessories complement the overall look"
}

IMPORTANT:
- Focus ONLY on accessories, not clothing items
- Do NOT include clothing elements like buttons or zippers on garments
- Be extremely detailed about materials, colors, and styles
- Include all visible accessories, even small ones
- Do not include weapons or weapon-related items`,
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

	// Validate it's JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		// If not valid JSON, return an error
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return json.RawMessage(cleaned), nil
}