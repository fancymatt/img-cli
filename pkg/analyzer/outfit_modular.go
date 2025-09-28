package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type ModularOutfitAnalyzer struct {
	BaseAnalyzer
	client           *gemini.Client
	excludeHair      bool
	excludeMakeup    bool
	excludeAccessories bool
}

type ExcludeOptions struct {
	Hair       bool
	Makeup     bool
	Accessories bool
}

func NewModularOutfitAnalyzer(client *gemini.Client, excludeOpts ExcludeOptions) *ModularOutfitAnalyzer {
	return &ModularOutfitAnalyzer{
		BaseAnalyzer:       BaseAnalyzer{Type: "outfit"},
		client:            client,
		excludeHair:       excludeOpts.Hair,
		excludeMakeup:     excludeOpts.Makeup,
		excludeAccessories: excludeOpts.Accessories,
	}
}

func (o *ModularOutfitAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
	imageData, mimeType, err := gemini.LoadImageAsBase64(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}

	// Build the prompt based on what should be excluded
	var promptParts []string

	promptParts = append(promptParts, "Analyze the outfit in this image with extreme precision and detail. You are analyzing for fashion designers who need comprehensive information.")

	if o.excludeHair {
		promptParts = append(promptParts, "IMPORTANT: DO NOT include any hair information in your analysis.")
	}
	if o.excludeMakeup {
		promptParts = append(promptParts, "IMPORTANT: DO NOT include any makeup information in your analysis.")
	}
	if o.excludeAccessories {
		promptParts = append(promptParts, "IMPORTANT: DO NOT include any accessories (jewelry, bags, belts, watches, etc.) in your analysis.")
	}

	promptParts = append(promptParts, `Return a JSON object with the following structure:
{
  "clothing": [extremely detailed list of each clothing item with comprehensive descriptions like "fitted charcoal gray merino wool blazer with notch lapels, two-button closure, functional buttonholes, ticket pocket, and subtle pick-stitching along the edges"],
  "style": "clothing style ONLY - fashion genre, formality level, and garment styling techniques. DO NOT include environmental descriptions, lighting, or background elements",
  "colors": [ONLY colors of the actual CLOTHING - use fashion terminology like "midnight navy", "winter white", "camel beige", "oxblood red". DO NOT include lighting colors, background colors, or environmental colors],`)

	// Only include accessories field if not excluded
	if !o.excludeAccessories {
		promptParts = append(promptParts, `  "accessories": [exhaustive list with detailed descriptions of watches, jewelry, belts, bags, scarves, hats, etc. but NOT glasses, weapons, or weapon-related items],`)
	}

	promptParts = append(promptParts, `  "overall": "thorough outfit analysis covering garment interaction, proportions, styling choices, layering techniques, fabric interplay, and overall aesthetic impact"`)

	// Only include hair field if not excluded
	if !o.excludeHair {
		promptParts = append(promptParts, `,
  "hair": {
    "color": "precise hair color description (e.g., 'ash blonde with platinum highlights', 'jet black', 'chestnut brown with caramel balayage')",
    "style": "detailed hairstyle name and description (e.g., 'sleek low bun with face-framing tendrils', 'tousled beach waves', 'slicked-back pompadour')",
    "length": "specific length description (e.g., 'shoulder-length', 'pixie cut', 'waist-length', 'chin-length bob')",
    "texture": "hair texture analysis (e.g., 'fine and straight', 'thick and wavy', 'coily', 'kinky')",
    "details": "any additional hair styling details"
  }`)
	}

	promptParts = append(promptParts, `
}`)

	// Add exclusion reminders
	if o.excludeHair || o.excludeMakeup || o.excludeAccessories {
		promptParts = append(promptParts, "\nREMINDER:")
		if o.excludeHair {
			promptParts = append(promptParts, "- Do NOT include hair information")
		}
		if o.excludeMakeup {
			promptParts = append(promptParts, "- Do NOT analyze or mention makeup")
		}
		if o.excludeAccessories {
			promptParts = append(promptParts, "- Do NOT include accessories")
		}
	}

	promptParts = append(promptParts, `

CRITICAL REQUIREMENTS:
- Focus on actual clothing construction, materials, and styling
- Use professional fashion terminology
- Be extremely specific about garment details
- Describe materials accurately (use "leather" not "faux leather", "fur" not "faux fur")
- Never include glasses in accessories
- Never describe environmental elements or lighting as part of the outfit`)

	fullPrompt := strings.Join(promptParts, "\n")

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
						Text: fullPrompt,
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

	resp, err := o.client.SendRequest(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	textResp := gemini.ExtractTextFromResponse(resp)
	return CleanAndValidateJSONResponse(textResp)
}