package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

type ExpressionAnalyzer struct {
	BaseAnalyzer
	client *gemini.Client
}

func NewExpressionAnalyzer(client *gemini.Client) *ExpressionAnalyzer {
	return &ExpressionAnalyzer{
		BaseAnalyzer: BaseAnalyzer{Type: "expression"},
		client:       client,
	}
}

func (e *ExpressionAnalyzer) Analyze(imagePath string) (json.RawMessage, error) {
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
						Text: `Analyze ONLY the facial expression and emotional state in this image. Ignore all other elements including clothing, hair, makeup, and accessories. Return a JSON object with the following structure:
{
  "primary_emotion": "main emotion displayed (e.g., 'joy', 'serenity', 'confidence', 'contemplation', 'surprise')",
  "intensity": "emotional intensity level (e.g., 'subtle', 'moderate', 'intense', 'restrained')",
  "facial_features": {
    "eyes": "eye expression (e.g., 'bright and alert', 'soft and dreamy', 'focused', 'squinting with joy')",
    "mouth": "mouth expression (e.g., 'gentle smile', 'neutral', 'slight smirk', 'broad grin', 'pursed lips')",
    "brows": "eyebrow position (e.g., 'relaxed', 'slightly raised', 'furrowed', 'arched')",
    "overall_tension": "facial muscle tension (e.g., 'relaxed', 'tense', 'animated')"
  },
  "gaze": {
    "direction": "where the person is looking (e.g., 'direct at camera', 'off to the side', 'downward', 'upward')",
    "quality": "quality of the gaze (e.g., 'piercing', 'soft', 'distant', 'engaged', 'mysterious')"
  },
  "mood": "overall mood conveyed (e.g., 'playful', 'serious', 'romantic', 'professional', 'casual')",
  "energy": "energy level of expression (e.g., 'calm', 'energetic', 'subdued', 'vibrant')",
  "authenticity": "naturalness of expression (e.g., 'genuine', 'posed', 'candid', 'theatrical')",
  "overall": "comprehensive description of the complete facial expression and emotional presentation"
}

IMPORTANT:
- Focus ONLY on facial expression and emotion
- Do not describe physical features, only expressions
- Be specific about subtle emotional nuances
- Describe what emotion/mood is being conveyed, not physical appearance`,
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

	resp, err := e.client.SendRequest(request)
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