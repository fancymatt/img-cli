package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type OutfitAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewOutfitAnalyzer(client *gemini.Client) *OutfitAnalyzer {
	return &OutfitAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "outfit"},
		client:       client,
	}
}

func (o *OutfitAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze the outfit, personal style, and hair in this image with extreme precision and detail. You are analyzing for fashion designers who need comprehensive information about every element. Return a JSON object with the following structure:
{
  "clothing": [extremely detailed list of each clothing item with comprehensive descriptions like "fitted charcoal gray merino wool blazer with notch lapels, two-button closure, functional buttonholes, ticket pocket, and subtle pick-stitching along the edges"],
  "style": "clothing style ONLY - fashion genre, formality level, and garment styling techniques. DO NOT include environmental descriptions, lighting, or background elements",
  "colors": [ONLY colors of the actual CLOTHING and ACCESSORIES - use fashion terminology like "midnight navy", "winter white", "camel beige", "oxblood red". DO NOT include lighting colors, background colors, or environmental colors],
  "accessories": [exhaustive list with detailed descriptions of watches, jewelry, belts, bags, scarves, hats, etc. but NOT glasses, weapons, or weapon-related items],
  "overall": "thorough outfit analysis covering garment interaction, proportions, styling choices, layering techniques, fabric interplay, and overall aesthetic impact",
  "hair": {
    "color": "precise hair color description (e.g., 'ash blonde with platinum highlights', 'jet black', 'chestnut brown with caramel balayage')",
    "style": "detailed hairstyle name and description (e.g., 'sleek low bun with face-framing tendrils', 'tousled beach waves', 'slicked-back pompadour')",
    "length": "specific length description (e.g., 'shoulder-length', 'pixie cut', 'waist-length', 'chin-length bob')",
    "texture": "hair texture analysis (e.g., 'fine and straight', 'thick and wavy', 'coily', 'kinky')",
    "details": ["specific styling details like 'side part', 'undercut', 'baby hairs styled', 'hair accessories'],
    "styling": "products or techniques visible (e.g., 'high-gloss gel finish', 'matte texture paste', 'heat-styled curls', 'natural air-dried')"
  }
}

For CLOTHING items, provide exhaustive detail including:
- Exact garment types with all variations (e.g., "cropped bomber jacket", "midi wrap dress", "palazzo pants")
- CRITICAL: Collar details - type, color, contrast (e.g., "white Peter Pan collar", "black notched collar", "contrast white shirt collar visible beneath")
- Cuffs and trim colors - specify if different from main garment (e.g., "white cuffs on black blazer", "red piping on navy jacket")
- Fabric composition and quality indicators (e.g., "heavyweight denim", "silk charmeuse", "boiled wool")
- IMPORTANT: NEVER use "faux" or similar qualifiers - describe all materials as genuine (e.g., say "fur" not "faux fur", "leather" not "faux leather", "suede" not "faux suede")
- Construction details (e.g., "French seams", "raw hem", "topstitching", "darted")
- Hardware and notions - specify color/finish (e.g., "gold-tone buttons", "silver zipper", "black rivets")
- Pattern details (stripes, checks, prints) with exact colors and scale
- Pockets, plackets, and other details with color specifications
- Condition and wear (pristine, distressed, vintage)
- Styling details (cuffed, tucked, belted, layered)
- Brand indicators or designer signatures if visible

ESSENTIAL: For EVERY garment, explicitly state the color of:
- Main body fabric
- Collar (if present)
- Cuffs (if visible)
- Buttons/closures
- Any trim, piping, or contrast elements

For LEATHER items specifically, ALWAYS describe as:
"heavy leather with pronounced folds and wrinkles, puffy and spongy texture, supple and thick, buttery smooth finish, padded construction, rugged and sturdy appearance"

For HAIR analysis, capture:
- Natural vs. colored appearance
- Styling products evident (wet look, matte, glossy)
- Volume and body
- Part placement and styling
- Any hair accessories or embellishments
- Professional styling vs. natural
- Health and condition indicators

Do NOT mention or include:
- Glasses or eyewear
- Weapons of any kind (guns, knives, etc.)
- Weapon accessories (holsters, sheaths, ammunition, etc.)
- Military or tactical equipment that could be weapon-related
- Makeup, cosmetics, or facial features
- Tattoos, body art, or skin markings
- Piercings (except earrings as accessories)
- Nail polish or nail art
- Environmental lighting (neon lights, street lights, etc.)
- Background elements or settings
- Atmospheric descriptions (dark, moody, bright, etc.)
- Location or scene descriptions

CRITICAL MATERIAL DESCRIPTION RULES:
- ALWAYS describe materials as genuine, NEVER as "faux", "vegan", "synthetic", or "imitation"
- If something looks like leather, describe it as "leather"
- If something looks like fur, describe it as "fur"
- If something looks like suede, describe it as "suede"
- This applies to ALL materials - always use the genuine material name

Remember: Fashion designers need this level of detail for accurate recreation and styling decisions.`,
					},
				},
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature:      0.3,
			TopK:             20,
			TopP:             0.8,
			// Note: Gemini 2.5 Flash Image doesn't support JSON mode
			// ResponseMimeType: "application/json",
		},
	}

	resp, err := o.client.SendRequest(request)
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

	var outfit gemini.OutfitDescription
	if err := json.Unmarshal([]byte(cleaned), &outfit); err != nil {
		// Return the cleaned JSON even if we can't parse it into the struct
		return json.RawMessage(cleaned), nil
	}

	// Filter out any weapon-related items from the analysis
	outfit = o.filterWeaponReferences(outfit)

	return json.Marshal(outfit)
}

// filterWeaponReferences removes any weapon-related items from the outfit analysis
func (o *OutfitAnalyzer) filterWeaponReferences(outfit gemini.OutfitDescription) gemini.OutfitDescription {
	// List of weapon-related terms to filter out
	weaponTerms := []string{
		"gun", "pistol", "rifle", "firearm", "weapon", "holster",
		"ammunition", "ammo", "bullet", "cartridge", "magazine",
		"revolver", "shotgun", "carbine", "assault", "tactical",
		"knife", "blade", "dagger", "sword", "machete",
	}

	// List of makeup and body modification terms to filter out
	beautyTerms := []string{
		"makeup", "lipstick", "eyeshadow", "mascara", "foundation",
		"blush", "concealer", "eyeliner", "bronzer", "highlighter",
		"tattoo", "tattoos", "ink", "body art", "piercing",
		"nail polish", "nail art", "manicure", "pedicure",
	}

	// List of environmental/lighting terms to filter out
	environmentTerms := []string{
		"neon", "lighting", "backdrop", "background", "environment",
		"atmosphere", "moody", "dark room", "bright room", "urban",
		"street", "nightlife", "cyberpunk", "synthwave", "noir",
		"futuristic", "retro-futurism", "rave", "club",
	}

	// Helper function to check if a string contains excluded terms
	containsExcludedTerm := func(s string) bool {
		lower := strings.ToLower(s)
		// Check weapon terms
		for _, term := range weaponTerms {
			if strings.Contains(lower, term) {
				return true
			}
		}
		// Check beauty/makeup terms
		for _, term := range beautyTerms {
			if strings.Contains(lower, term) {
				return true
			}
		}
		// Check environmental terms
		for _, term := range environmentTerms {
			if strings.Contains(lower, term) {
				return true
			}
		}
		return false
	}

	// Filter clothing items
	var filteredClothing []interface{}
	for _, item := range outfit.Clothing {
		// Check if item is a string
		if str, ok := item.(string); ok {
			if !containsExcludedTerm(str) {
				filteredClothing = append(filteredClothing, item)
			}
		} else {
			// For non-string items (like ClothingItem structs), keep them
			// You might want to add more sophisticated filtering here
			filteredClothing = append(filteredClothing, item)
		}
	}
	outfit.Clothing = filteredClothing

	// Filter accessories (but allow earrings even if they contain "piercing")
	var filteredAccessories []interface{}
	for _, item := range outfit.Accessories {
		// Check if item is a string
		if str, ok := item.(string); ok {
			// Special case: allow earrings even if they mention piercing
			if strings.Contains(strings.ToLower(str), "earring") {
				filteredAccessories = append(filteredAccessories, item)
			} else if !containsExcludedTerm(str) {
				filteredAccessories = append(filteredAccessories, item)
			}
		} else {
			// For non-string items, keep them
			filteredAccessories = append(filteredAccessories, item)
		}
	}
	outfit.Accessories = filteredAccessories

	// Filter colors to remove environmental/lighting colors
	var filteredColors []string
	for _, color := range outfit.Colors {
		if !containsExcludedTerm(color) {
			// Additional check for parenthetical lighting descriptions
			if idx := strings.Index(color, "("); idx > 0 {
				// Keep only the color part before parentheses
				color = strings.TrimSpace(color[:idx])
			}
			filteredColors = append(filteredColors, color)
		}
	}
	outfit.Colors = filteredColors

	// Filter the overall description
	if containsExcludedTerm(outfit.Overall) {
		// Remove sentences that contain excluded terms
		sentences := strings.Split(outfit.Overall, ". ")
		var filteredSentences []string
		for _, sentence := range sentences {
			if !containsExcludedTerm(sentence) {
				filteredSentences = append(filteredSentences, sentence)
			}
		}
		outfit.Overall = strings.Join(filteredSentences, ". ")
		// Clean up any trailing period issues
		outfit.Overall = strings.TrimSuffix(outfit.Overall, "..")
		if !strings.HasSuffix(outfit.Overall, ".") && outfit.Overall != "" {
			outfit.Overall += "."
		}
	}

	// Filter the style description
	if containsExcludedTerm(outfit.Style) {
		// Remove excluded term references
		sentences := strings.Split(outfit.Style, ". ")
		var filteredSentences []string
		for _, sentence := range sentences {
			if !containsExcludedTerm(sentence) {
				filteredSentences = append(filteredSentences, sentence)
			}
		}
		outfit.Style = strings.Join(filteredSentences, ". ")
		outfit.Style = strings.TrimSuffix(outfit.Style, "..")
		if !strings.HasSuffix(outfit.Style, ".") && outfit.Style != "" {
			outfit.Style += "."
		}
	}

	return outfit
}