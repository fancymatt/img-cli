package generator

import (
	"encoding/json"
	"fmt"
	"img-cli/pkg/gemini"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CombinedGenerator struct {
	BaseGenerator
	client *gemini.Client
}

func NewCombinedGenerator(client *gemini.Client) *CombinedGenerator {
	return &CombinedGenerator{
		BaseGenerator: BaseGenerator{Type: "combined"},
		client:        client,
	}
}

func (c *CombinedGenerator) Generate(params GenerateParams) (*GenerateResult, error) {
	imageData, mimeType, err := gemini.LoadImageAsBase64(params.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}

	// Build the combined prompt with outfit and style
	var promptBuilder strings.Builder

	// Check if we're using outfit image instead of text description
	useOutfitImage := params.SendOriginal && params.OutfitReference != "" && params.Prompt == ""

	// Start with base instructions
	promptBuilder.WriteString("Generate an image of this person with EXACT COLOR AND DETAIL ACCURACY.\n")

	if useOutfitImage {
		// Using outfit image reference instead of text description
		promptBuilder.WriteString("The person from the FIRST image should be wearing EXACTLY the outfit shown in the SECOND image.\n")
		promptBuilder.WriteString("Match every detail of the outfit from the reference image precisely.\n")
		promptBuilder.WriteString("IMPORTANT: Any style reference provided is ONLY for photographic style and pose. Do NOT transfer any clothing or accessories from the style reference.\n\n")
	} else {
		// Using text-based outfit description
		promptBuilder.WriteString("IMPORTANT: Any style reference provided is ONLY for photographic style and pose. Do NOT transfer any clothing or accessories from the style reference.\n\n")

		if params.Prompt != "" {
			// Check if the prompt contains leather items and add the leather description if needed
			promptLower := strings.ToLower(params.Prompt)
			enhancedPrompt := params.Prompt
			if strings.Contains(promptLower, "leather") {
				if !strings.Contains(promptLower, "heavy leather") && !strings.Contains(promptLower, "buttery smooth") {
					enhancedPrompt = strings.Replace(params.Prompt, "leather", "heavy leather with folds and wrinkles, puffy, spongy, supple, thick, buttery smooth leather, padded, rugged, sturdy", 1)
				}
			}
			promptBuilder.WriteString("OUTFIT SPECIFICATION (must be followed EXACTLY):\n")
			promptBuilder.WriteString(enhancedPrompt)
			promptBuilder.WriteString("\n\nCRITICAL: Every color, pattern, and detail mentioned must be reproduced PRECISELY as specified.\n")
		} else {
			promptBuilder.WriteString("Generate an image of this person.\n")
		}
	}

	// Add style information if available (always apply style, regardless of outfit mode)
	if params.StyleData != nil {
		var style gemini.VisualStyle
		if err := json.Unmarshal(params.StyleData, &style); err == nil {
			promptBuilder.WriteString("\nCRITICAL STYLE REQUIREMENTS - Apply the following visual style EXACTLY:\n")

			// Pose and body position (most important for matching style)
			if style.Pose != "" {
				promptBuilder.WriteString(fmt.Sprintf("- POSE (MUST MATCH): %s\n", style.Pose))
			}
			if style.BodyPosition != "" {
				promptBuilder.WriteString(fmt.Sprintf("- BODY POSITION (MUST MATCH): %s\n", style.BodyPosition))
			}
			if style.CameraAngle != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Camera angle: %s\n", style.CameraAngle))
			}
			if style.Framing != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Framing: %s\n", style.Framing))
			}

			// Visual quality and era
			if style.FilmGrain != "" {
				promptBuilder.WriteString(fmt.Sprintf("- FILM GRAIN (CRITICAL): %s\n", style.FilmGrain))
			}
			if style.Era != "" {
				promptBuilder.WriteString(fmt.Sprintf("- ERA AESTHETIC (MUST MATCH): %s\n", style.Era))
			}
			if style.ImageQuality != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Image quality: %s\n", style.ImageQuality))
			}

			// Color and lighting
			if style.ColorGrading != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Color grading: %s\n", style.ColorGrading))
			}
			if style.Lighting != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Lighting: %s\n", style.Lighting))
			}
			if len(style.ColorPalette) > 0 {
				promptBuilder.WriteString(fmt.Sprintf("- Color palette: %v\n", style.ColorPalette))
			}

			// Other style elements
			if style.DepthOfField != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Depth of field: %s\n", style.DepthOfField))
			}
			if style.PostProcessing != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Post-processing effects: %s\n", style.PostProcessing))
			}
			if style.Mood != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Mood: %s\n", style.Mood))
			}
			if style.Photographic != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Photographic style: %s\n", style.Photographic))
			}
			if style.ArtisticStyle != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Artistic style: %s\n", style.ArtisticStyle))
			}
			if style.Background != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Background: %s\n", style.Background))
			}

			promptBuilder.WriteString("\nIMPORTANT: The pose, body position, film grain, and era aesthetic MUST be replicated exactly as described.\n")
			promptBuilder.WriteString("\nCRITICAL: DO NOT add ANY clothing, accessories, or outfit elements from the style reference image. NO hats, jewelry, or any other accessories should be added based on the style reference. Glasses/eyewear should ONLY match what the subject originally has - if they have glasses, keep them; if not, don't add them. The style ONLY affects photographic qualities and body pose.\n")
		}
	}

	// Add hair instructions based on HairData (always apply hair modifications if specified)
	if params.HairData != nil {
		var hair gemini.HairDescription
		if err := json.Unmarshal(params.HairData, &hair); err == nil {
			promptBuilder.WriteString("\n\nCRITICAL HAIR REQUIREMENTS (MUST override any other hair instructions):\nApply the following EXACT hair styling from the hair reference image:\n")
			if hair.Color != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Hair color: %s\n", hair.Color))
			}
			if hair.Style != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Hair style: %s\n", hair.Style))
			}
			if hair.Length != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Hair length: %s\n", hair.Length))
			}
			if hair.Texture != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Hair texture: %s\n", hair.Texture))
			}
			if hair.Styling != "" {
				promptBuilder.WriteString(fmt.Sprintf("- Hair styling/finish: %s\n", hair.Styling))
			}
			if len(hair.Details) > 0 {
				promptBuilder.WriteString(fmt.Sprintf("- Hair details: %s\n", strings.Join(hair.Details, ", ")))
			}
			promptBuilder.WriteString("\nIMPORTANT: The subject's hair MUST match the hair reference description above, NOT their original hair.\n")
			if params.DebugPrompt {
				fmt.Printf("[DEBUG] Hair data applied from: %s\n", params.HairSource)
			}
		} else {
			if params.DebugPrompt {
				fmt.Printf("[DEBUG] Failed to parse hair data: %v\n", err)
			}
		}
	} else {
		// Default behavior: keep the subject's original hair
		promptBuilder.WriteString("\nKeep the subject's original hair color and style exactly as it appears in the source image.")
		if params.DebugPrompt {
			fmt.Printf("[DEBUG] No hair data provided - keeping original hair\n")
		}
	}

	// Always add these final instructions
	promptBuilder.WriteString("\nKeep their facial features (eyes, nose, mouth, face shape) exactly the same.")
	promptBuilder.WriteString("\nIMPORTANT: Preserve ALL of the person's original features that are NOT clothing:")
	promptBuilder.WriteString("\n- Keep their exact same makeup (or lack of makeup)")
	promptBuilder.WriteString("\n- Keep any tattoos, birthmarks, or skin markings exactly as they are")
	promptBuilder.WriteString("\n- Keep their same piercings (ears, nose, etc.)")
	promptBuilder.WriteString("\n- Keep their nail polish or natural nails as they are")
	promptBuilder.WriteString("\n- If they're wearing glasses, keep the exact same glasses")
	promptBuilder.WriteString("\nOnly change the CLOTHING items - everything else about the person must remain exactly the same.")
	promptBuilder.WriteString("\nGenerate a realistic photographic image, not an illustration or artwork.")

	if !useOutfitImage {
		// Only add this rule when using text descriptions (not needed when outfit image is provided)
		promptBuilder.WriteString("\n\nABSOLUTE RULE: The generated image must contain ONLY the outfit/clothing specified above. Do NOT add glasses, sunglasses, hats, or any accessories from the style reference image. The style reference is ONLY for photographic style and pose, NOT for any clothing or accessories.")
	}

	// Add variation instructions if generating multiple
	if params.TotalVariations > 1 {
		promptBuilder.WriteString(fmt.Sprintf("\n\nThis is variation %d of %d. Create a subtle variation in pose as if this is part of the same photo shoot. Keep the same outfit, style, and environment, but vary the pose, angle, or expression slightly to create a natural photo shoot variation.", params.VariationIndex, params.TotalVariations))
	}
	
	fullPrompt := promptBuilder.String()

	if params.DebugPrompt {
		fmt.Println("\n[DEBUG] Combined Generation Prompt:")
		fmt.Println("====================================")
		fmt.Printf("Image: %s\n", filepath.Base(params.ImagePath))
		fmt.Printf("Prompt:\n%s\n", fullPrompt)
		if params.StyleData != nil {
			fmt.Printf("Style Data: %s\n", string(params.StyleData))
		}
		fmt.Println("====================================\n")
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
			// Don't modify the prompt - it's already set appropriately above
			if params.DebugPrompt {
				fmt.Printf("[DEBUG] Including outfit reference image: %s (replacing text description: %v)\n",
					filepath.Base(params.OutfitReference), useOutfitImage)
			}
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

	rawResp, err := c.client.SendRequestRaw(request)
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

	// Create filename as outfit_style_subject_timestamp
	subjectName := strings.TrimSuffix(filepath.Base(params.ImagePath), filepath.Ext(params.ImagePath))
	outfitName := params.OutfitSource
	if outfitName == "" {
		outfitName = "outfit"
	}
	styleName := params.StyleSource
	if styleName == "" {
		styleName = outfitName // Default to same as outfit if not specified
	}

	// Generate timestamp in format YYYYMMDDHHMMSS
	timestamp := time.Now().Format("20060102150405")

	outputPath := filepath.Join(params.OutputDir, fmt.Sprintf("%s_%s_%s_%s%s", outfitName, styleName, subjectName, timestamp, extension))

	if err := os.MkdirAll(params.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
		return nil, fmt.Errorf("error saving image: %w", err)
	}

	return &GenerateResult{
		Type:       c.Type,
		OutputPath: outputPath,
		Message:    "Generated transformed image with outfit and style",
	}, nil
}