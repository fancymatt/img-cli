package generator

import (
	"fmt"
	"img-cli/pkg/gemini"
	"img-cli/pkg/models"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ModularGenerator struct {
	BaseGenerator
	client *gemini.Client
}

type ModularRequest struct {
	SubjectPath   string
	Prompt        string
	Components    *models.ModularComponents
	SendOriginals bool
	OutputDir     string
}

func NewModularGenerator(client *gemini.Client) *ModularGenerator {
	return &ModularGenerator{
		BaseGenerator: BaseGenerator{Type: "modular"},
		client:        client,
	}
}

func (g *ModularGenerator) Generate(req ModularRequest) (string, error) {
	// Load subject image
	subjectData, subjectMime, err := gemini.LoadImageAsBase64(req.SubjectPath)
	if err != nil {
		return "", fmt.Errorf("error loading subject image: %w", err)
	}

	// Build request parts
	var parts []interface{}

	// Add subject image
	parts = append(parts, gemini.BlobPart{
		InlineData: gemini.InlineData{
			MimeType: subjectMime,
			Data:     subjectData,
		},
	})

	// Optionally add reference images
	if req.SendOriginals && req.Components != nil {
		// Add outfit reference if available
		if req.Components.Outfit != nil && req.Components.Outfit.ImagePath != "" {
			outfitData, outfitMime, err := gemini.LoadImageAsBase64(req.Components.Outfit.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: outfitMime,
						Data:     outfitData,
					},
				})
			}
		}

		// Add style reference if available
		if req.Components.Style != nil && req.Components.Style.ImagePath != "" {
			styleData, styleMime, err := gemini.LoadImageAsBase64(req.Components.Style.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: styleMime,
						Data:     styleData,
					},
				})
			}
		}

		// Add hair style reference if available
		if req.Components.HairStyle != nil && req.Components.HairStyle.ImagePath != "" {
			hairData, hairMime, err := gemini.LoadImageAsBase64(req.Components.HairStyle.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: hairMime,
						Data:     hairData,
					},
				})
			}
		}

		// Add hair color reference if available
		if req.Components.HairColor != nil && req.Components.HairColor.ImagePath != "" {
			colorData, colorMime, err := gemini.LoadImageAsBase64(req.Components.HairColor.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: colorMime,
						Data:     colorData,
					},
				})
			}
		}

		// Add makeup reference if available
		if req.Components.Makeup != nil && req.Components.Makeup.ImagePath != "" {
			makeupData, makeupMime, err := gemini.LoadImageAsBase64(req.Components.Makeup.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: makeupMime,
						Data:     makeupData,
					},
				})
			}
		}

		// Add expression reference if available
		if req.Components.Expression != nil && req.Components.Expression.ImagePath != "" {
			expData, expMime, err := gemini.LoadImageAsBase64(req.Components.Expression.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: expMime,
						Data:     expData,
					},
				})
			}
		}

		// Add accessories reference if available
		if req.Components.Accessories != nil && req.Components.Accessories.ImagePath != "" {
			accData, accMime, err := gemini.LoadImageAsBase64(req.Components.Accessories.ImagePath)
			if err == nil {
				parts = append(parts, gemini.BlobPart{
					InlineData: gemini.InlineData{
						MimeType: accMime,
						Data:     accData,
					},
				})
			}
		}
	}

	// Add the prompt text
	parts = append(parts, gemini.TextPart{
		Text: req.Prompt,
	})

	// Create the API request
	request := gemini.Request{
		Contents: []gemini.Content{
			{
				Parts: parts,
			},
		},
		GenerationConfig: &gemini.GenerationConfig{
			Temperature: 0.8,
			TopP:        0.95,
			TopK:        40,
		},
	}

	// Generate the image
	rawResp, err := g.client.SendRequestRaw(request)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}

	imageBytes, imageMimeType, err := gemini.ExtractGeneratedImage(rawResp)
	if err != nil {
		return "", fmt.Errorf("error extracting image: %w", err)
	}

	extension := ".png"
	if strings.Contains(imageMimeType, "jpeg") || strings.Contains(imageMimeType, "jpg") {
		extension = ".jpg"
	} else if strings.Contains(imageMimeType, "gif") {
		extension = ".gif"
	} else if strings.Contains(imageMimeType, "webp") {
		extension = ".webp"
	}

	// Generate output filename
	timestamp := time.Now().Format("20060102_150405")
	subjectName := filepath.Base(req.SubjectPath)
	subjectName = subjectName[:len(subjectName)-len(filepath.Ext(subjectName))]

	// Build filename parts
	var filenameParts []string

	// Add outfit name if present
	if req.Components != nil && req.Components.Outfit != nil && req.Components.Outfit.ImagePath != "" {
		outfitName := filepath.Base(req.Components.Outfit.ImagePath)
		outfitName = outfitName[:len(outfitName)-len(filepath.Ext(outfitName))]
		filenameParts = append(filenameParts, outfitName)
	}

	// Add style name if present
	if req.Components != nil && req.Components.Style != nil && req.Components.Style.ImagePath != "" {
		styleName := filepath.Base(req.Components.Style.ImagePath)
		styleName = styleName[:len(styleName)-len(filepath.Ext(styleName))]
		filenameParts = append(filenameParts, styleName)
	}

	// Always add subject name
	filenameParts = append(filenameParts, subjectName)

	// Add timestamp
	filenameParts = append(filenameParts, timestamp)

	outputFilename := strings.Join(filenameParts, "_") + extension
	outputPath := filepath.Join(req.OutputDir, outputFilename)

	// Ensure output directory exists
	if err := os.MkdirAll(req.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("error creating output directory: %w", err)
	}

	// Save the image
	if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("error saving image: %w", err)
	}

	return outputPath, nil
}

