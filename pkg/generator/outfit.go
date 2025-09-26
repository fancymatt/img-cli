package generator

import (
	"encoding/base64"
	"fmt"
	"img-cli/pkg/gemini"
	"os"
	"path/filepath"
	"strings"
)

type OutfitGenerator struct {
	BaseGenerator
	client *gemini.Client
}

func NewOutfitGenerator(client *gemini.Client) *OutfitGenerator {
	return &OutfitGenerator{
		BaseGenerator: BaseGenerator{Type: "outfit"},
		client:        client,
	}
}

func (o *OutfitGenerator) Generate(params GenerateParams) (*GenerateResult, error) {
	imageData, mimeType, err := gemini.LoadImageAsBase64(params.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}

	prompt := params.Prompt
	if prompt == "" {
		prompt = "a formal business suit"
	}

	// Check if the prompt contains leather and enhance the description
	enhancedPrompt := prompt
	if strings.Contains(strings.ToLower(prompt), "leather") {
		if !strings.Contains(strings.ToLower(prompt), "heavy leather") && !strings.Contains(strings.ToLower(prompt), "buttery smooth") {
			enhancedPrompt = strings.Replace(prompt, "leather", "heavy leather with folds and wrinkles, puffy, spongy, supple, thick, buttery smooth leather, padded, rugged, sturdy", 1)
		}
	}

	fullPrompt := fmt.Sprintf(`Generate a 9:16 portrait format image of this person wearing EXACTLY the following outfit with PRECISE COLOR ACCURACY:
%s

CRITICAL REQUIREMENTS:
- Every color mentioned must be reproduced EXACTLY as specified (e.g., if a white collar is mentioned, it MUST be white, not black or any other color)
- All garment details, trims, patterns, and color combinations must match the description precisely
- Keep their face and features exactly the same
- IMPORTANT: If the person is wearing glasses in the original image, they MUST keep wearing the exact same glasses. If they're not wearing glasses, they should not have glasses in the generated image
- Glasses are NOT part of the outfit - preserve the subject's original eyewear status
- Show them from the waist up against a pure black background
- Put them in a different, natural pose from the source image
- Image must be in 9:16 aspect ratio (portrait/vertical format)

The outfit details provided are from a fashion designer's specification and MUST be followed exactly.`, enhancedPrompt)

	if params.DebugPrompt {
		fmt.Println("\n[DEBUG] Outfit Generation Prompt:")
		fmt.Println("================================")
		fmt.Printf("Image: %s\n", filepath.Base(params.ImagePath))
		fmt.Printf("Prompt:\n%s\n", fullPrompt)
		fmt.Println("================================\n")
	}

	// Build parts for the request
	parts := []interface{}{
		gemini.BlobPart{
			InlineData: gemini.InlineData{
				MimeType: mimeType,
				Data:     imageData,
			},
		},
	}

	// If SendOriginal is true and we have an outfit reference image, include it
	if params.SendOriginal && params.OutfitReference != "" {
		outfitData, outfitMimeType, err := gemini.LoadImageAsBase64(params.OutfitReference)
		if err != nil {
			fmt.Printf("Warning: Could not load outfit reference image: %v\n", err)
		} else {
			parts = append(parts, gemini.BlobPart{
				InlineData: gemini.InlineData{
					MimeType: outfitMimeType,
					Data:     outfitData,
				},
			})
			// Modify prompt to reference the outfit image
			fullPrompt = fmt.Sprintf(`Generate a 9:16 portrait format image of the person from the first image wearing the outfit shown in the reference image(s).

Outfit description: %s

CRITICAL REQUIREMENTS:
- Match the outfit from the reference image as closely as possible
- Every color and detail from the reference must be reproduced accurately
- Keep the person's face and features exactly the same as the first image
- IMPORTANT: If the person is wearing glasses in the original image, they MUST keep wearing the exact same glasses. If they're not wearing glasses, they should not have glasses in the generated image
- Glasses are NOT part of the outfit - preserve the subject's original eyewear status
- Show them from the waist up against a pure black background
- Put them in a different, natural pose from the source image
- Image must be in 9:16 aspect ratio (portrait/vertical format)

The outfit details provided are from a fashion designer's specification and MUST be followed exactly.`, enhancedPrompt)
		}
	}

	// Add the text prompt
	parts = append(parts, gemini.TextPart{
		Text: fullPrompt,
	})

	request := gemini.Request{
		Contents: []gemini.Content{
			{
				Parts: parts,
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: params.Temperature,
			TopK:        40,
			TopP:        0.95,
		},
	}

	if params.Temperature == 0 {
		request.GenerationConfig.Temperature = 0.8
	}

	rawResp, err := o.client.SendRequestRaw(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	imageBytes, imageMimeType, err := gemini.ExtractGeneratedImage(rawResp)
	if err != nil {
		return nil, fmt.Errorf("error extracting image: %w", err)
	}

	extension := ".png"
	if strings.Contains(imageMimeType, "jpeg") || strings.Contains(imageMimeType, "jpg") {
		extension = ".jpg"
	} else if strings.Contains(imageMimeType, "gif") {
		extension = ".gif"
	} else if strings.Contains(imageMimeType, "webp") {
		extension = ".webp"
	}

	baseName := strings.TrimSuffix(filepath.Base(params.ImagePath), filepath.Ext(params.ImagePath))
	outputPath := filepath.Join(params.OutputDir, fmt.Sprintf("%s_outfit%s", baseName, extension))

	if err := os.MkdirAll(params.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
		return nil, fmt.Errorf("error saving image: %w", err)
	}

	return &GenerateResult{
		Type:       o.Type,
		OutputPath: outputPath,
		Message:    fmt.Sprintf("Generated outfit image with: %s", prompt),
	}, nil
}

func ExtractImageFromRawResponse(rawResp map[string]interface{}) ([]byte, string, error) {
	if candidates, ok := rawResp["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok {
					for _, part := range parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							if inlineData, ok := partMap["inlineData"].(map[string]interface{}); ok {
								if mimeType, ok := inlineData["mimeType"].(string); ok {
									if data, ok := inlineData["data"].(string); ok {
										imageData, err := base64.StdEncoding.DecodeString(data)
										if err != nil {
											return nil, "", fmt.Errorf("error decoding image: %w", err)
										}
										return imageData, mimeType, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil, "", fmt.Errorf("no image found in response")
}