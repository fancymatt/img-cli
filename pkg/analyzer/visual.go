package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type VisualStyleAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewVisualStyleAnalyzer(client *gemini.Client) *VisualStyleAnalyzer {
	return &VisualStyleAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "visual_style"},
		client:       client,
	}
}

func (v *VisualStyleAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze the complete visual style, aesthetics, and technical qualities of this image with extreme detail. Return a JSON object with the following structure:
{
  "composition": "detailed description of composition, rule of thirds, visual balance, leading lines, etc.",
  "framing": "precise framing details (e.g., extreme close-up, close-up, medium shot, full body, waist-up, 3/4 shot, wide shot, etc.)",
  "pose": "exact BODY POSE description - hands position relative to body, arms position, head tilt, shoulders, stance. DO NOT mention any clothing items or accessories like sunglasses, hats, jewelry",
  "body_position": "body position and orientation (e.g., standing, sitting, lying down, leaning, profile view, three-quarter view, facing camera, looking away, etc.)",
  "lighting": "comprehensive lighting analysis including type, direction, quality, shadows, highlights, contrast",
  "color_palette": [list of all dominant and accent colors],
  "color_grading": "color grading and toning (e.g., warm tones, cool tones, desaturated, high contrast, vintage color cast, sepia, etc.)",
  "mood": "overall mood, atmosphere, and emotional tone",
  "background": "detailed background description including depth, bokeh, environmental elements",
  "photographic_style": "specific photographic style (e.g., fashion editorial, candid snapshot, formal portrait, street photography, studio shot, etc.)",
  "artistic_style": "artistic and aesthetic style (e.g., retro 80s, film noir, minimalist, grunge, glamour, etc.)",
  "film_grain": "presence and intensity of film grain or noise (e.g., heavy grain, subtle grain, clean/no grain, digital noise)",
  "image_quality": "image quality characteristics (e.g., sharp, soft focus, motion blur, lens flare, chromatic aberration, vignetting)",
  "era_aesthetic": "time period aesthetic if apparent (e.g., 1980s, 1990s, modern, vintage, retro-futuristic, timeless)",
  "camera_angle": "camera angle and perspective (e.g., eye level, low angle, high angle, dutch angle, bird's eye view)",
  "depth_of_field": "depth of field characteristics (e.g., shallow DOF with bokeh, deep DOF, selective focus, tilt-shift)",
  "post_processing": "apparent post-processing effects (e.g., HDR, cross-processing, split-toning, filters, overlays, light leaks)"
}

CRITICAL INSTRUCTIONS:
- DO NOT include ANY clothing, accessories, or outfit elements in your analysis
- DO NOT mention sunglasses, hats, jewelry, watches, or any worn items
- Focus ONLY on photographic style, body positioning, and visual aesthetics
- The "pose" field should describe ONLY body position (arms, hands, head angle, stance)
- Clothing/accessories will be handled separately - you must IGNORE them completely

Be EXTREMELY detailed and specific about every visual element, especially:
- The exact body pose and position (without mentioning any clothing/accessories)
- Film grain, noise, and image quality characteristics
- Era-specific photographic aesthetics (not fashion/clothing)
- Color grading and processing effects
- Any distinctive visual treatments or filters

IMPORTANT: Even if the image appears to be an illustration or artwork, describe all qualities as photographic elements that can be recreated in a photograph.`,
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

	resp, err := v.client.SendRequest(request)
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

	var style gemini.VisualStyle
	if err := json.Unmarshal([]byte(cleaned), &style); err != nil {
		// Return the cleaned JSON even if we can't parse it into the struct
		return json.RawMessage(cleaned), nil
	}

	return json.Marshal(style)
}