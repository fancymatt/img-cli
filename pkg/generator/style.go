package generator

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"os"
	"path/filepath"
	"strings"
)

type StyleTransferGenerator struct {
	BaseGenerator
	client *gemini.Client
}

func NewStyleTransferGenerator(client *gemini.Client) *StyleTransferGenerator {
	return &StyleTransferGenerator{
		BaseGenerator: BaseGenerator{Type: "style_transfer"},
		client:        client,
	}
}

func (s *StyleTransferGenerator) Generate(params GenerateParams) (*GenerateResult, error) {
	imageData, mimeType, err := gemini.LoadImageAsBase64(params.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}

	var stylePrompt string
	if params.StyleData != nil {
		var style gemini.VisualStyle
		if err := json.Unmarshal(params.StyleData, &style); err == nil {
			stylePrompt = fmt.Sprintf(`Apply the following visual style:
- Lighting: %s
- Mood: %s
- Color palette: %v
- Photographic style: %s
- Background: %s`,
				style.Lighting, style.Mood, style.ColorPalette,
				style.Photographic, style.Background)
		}
	}

	if stylePrompt == "" && params.Prompt != "" {
		stylePrompt = params.Prompt
	}

	if stylePrompt == "" {
		stylePrompt = "Apply a dramatic, high-contrast visual style"
	}

	fullPrompt := fmt.Sprintf(`Generate a new version of this image with the following requirements:
%s

Keep the subject and composition similar but apply the requested visual style changes.
Maintain high quality and artistic coherence.`, stylePrompt)

	if params.DebugPrompt {
		fmt.Println("\n[DEBUG] Style Transfer Generation Prompt:")
		fmt.Println("=========================================")
		fmt.Printf("Image: %s\n", filepath.Base(params.ImagePath))
		fmt.Printf("Style Prompt:\n%s\n", fullPrompt)
		if params.StyleData != nil {
			fmt.Printf("Style Data: %s\n", string(params.StyleData))
		}
		fmt.Println("=========================================\n")
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
						Text: fullPrompt,
					},
				},
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: params.Temperature,
			TopK:        40,
			TopP:        0.95,
		},
	}

	if params.Temperature == 0 {
		request.GenerationConfig.Temperature = 0.7
	}

	rawResp, err := s.client.SendRequestRaw(request)
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
	outputPath := filepath.Join(params.OutputDir, fmt.Sprintf("%s_styled%s", baseName, extension))

	if err := os.MkdirAll(params.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
		return nil, fmt.Errorf("error saving image: %w", err)
	}

	return &GenerateResult{
		Type:       s.Type,
		OutputPath: outputPath,
		Message:    "Generated styled image",
	}, nil
}