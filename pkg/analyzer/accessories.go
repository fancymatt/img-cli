package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
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
	prompt := `Analyze ONLY the accessories in this image with extreme precision. Ignore clothing items, hair, and makeup. Focus on accessories like jewelry, bags, belts, scarves, hats, watches, etc. Return a JSON object with the following structure:
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
- Do not include weapons or weapon-related items`

	request, err := BuildImageAnalysisRequest(imagePath, prompt, gemini.AnalyzerConfig)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.SendRequest(*request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	textResp := gemini.ExtractTextFromResponse(resp)
	return CleanAndValidateJSONResponse(textResp)
}