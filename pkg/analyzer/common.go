package analyzer

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"strings"
)

// CleanAndValidateJSONResponse removes markdown code blocks and validates JSON
func CleanAndValidateJSONResponse(textResp string) (json.RawMessage, error) {
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
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return json.RawMessage(cleaned), nil
}

// BuildImageAnalysisRequest creates a standard Gemini request for image analysis
func BuildImageAnalysisRequest(imagePath string, prompt string, config *gemini.GenerationConfig) (*gemini.Request, error) {
	imageData, mimeType, err := gemini.LoadImageAsBase64(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}

	request := &gemini.Request{
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
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: config,
	}

	return request, nil
}