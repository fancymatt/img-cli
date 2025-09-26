// Package gemini provides a client for interacting with Google's Gemini API.
// It handles image generation requests, base64 encoding of images, and API communication.
package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	APIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image-preview:generateContent"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 3 minutes for image generation
		},
	}
}

func LoadImageAsBase64(imagePath string) (string, string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", "", err
	}

	ext := strings.ToLower(filepath.Ext(imagePath))
	mimeType := "image/jpeg"
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	}

	encodedData := base64.StdEncoding.EncodeToString(imageData)
	return encodedData, mimeType, nil
}

func (c *Client) SendRequest(request Request) (*Response, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", APIURL+"?key="+c.apiKey, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var geminiResp Response
		if err := json.Unmarshal(body, &geminiResp); err == nil && geminiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", geminiResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var geminiResp Response
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &geminiResp, nil
}

func (c *Client) SendRequestRaw(request Request) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", APIURL+"?key="+c.apiKey, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var geminiResp Response
		if err := json.Unmarshal(body, &geminiResp); err == nil && geminiResp.Error != nil {
			return nil, fmt.Errorf("API error: %s", geminiResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return rawResp, nil
}

func ExtractTextFromResponse(resp *Response) string {
	if len(resp.Candidates) == 0 {
		return ""
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(map[string]interface{}); ok {
			if text, ok := textPart["text"].(string); ok {
				return text
			}
		}
	}

	return ""
}

func ExtractGeneratedImage(rawResp map[string]interface{}) ([]byte, string, error) {
	if candidates, ok := rawResp["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			// Check for finish reason first
			if finishReason, ok := candidate["finishReason"].(string); ok && finishReason != "" {
				// Only show finish reason for non-STOP cases
				if finishReason != "STOP" {
					fmt.Printf("\n[API] Finish Reason: %s\n", finishReason)
				}
			}

			if content, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok {
					// First, look for text parts to capture any error messages
					var textContent string
					for _, part := range parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							if text, ok := partMap["text"].(string); ok {
								textContent = text
							}
						}
					}

					// Now look for image parts
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

					// If we got here, no image was found but we might have text
					if textContent != "" {
						// Print the text content for debugging
						fmt.Println("\n=== API Response (Text Instead of Image) ===")
						fmt.Println(textContent)
						fmt.Println("===========================================\n")
						return nil, "", fmt.Errorf("no image found in response, received text instead (see above)")
					}
				}
			}
		}
	}

	return nil, "", fmt.Errorf("no image found in response")
}

// LoadFile loads a file as bytes
func LoadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// SaveFile saves data to a file
func SaveFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// GetFileInfo returns file info
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// GetImagesFromDirectory returns all image files from a directory
func GetImagesFromDirectory(dirPath string) ([]string, error) {
	supportedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	var imageFiles []string

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		for _, supportedExt := range supportedExtensions {
			if ext == supportedExt {
				imageFiles = append(imageFiles, filepath.Join(dirPath, file.Name()))
				break
			}
		}
	}

	return imageFiles, nil
}

// ExtractImageFromResponse extracts generated image data from a Response struct
func ExtractImageFromResponse(resp *Response) *ImageData {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil
	}

	candidate := resp.Candidates[0]
	for _, part := range candidate.Content.Parts {
		if blobPart, ok := part.(map[string]interface{}); ok {
			if inlineDataMap, ok := blobPart["inlineData"].(map[string]interface{}); ok {
				mimeType, _ := inlineDataMap["mimeType"].(string)
				dataStr, _ := inlineDataMap["data"].(string)

				if dataStr != "" {
					imageBytes, err := base64.StdEncoding.DecodeString(dataStr)
					if err == nil {
						return &ImageData{
							Data:     imageBytes,
							MimeType: mimeType,
						}
					}
				}
			}
		}
		// Also check if the part is already a BlobPart struct
		if blob, ok := part.(BlobPart); ok {
			imageBytes, err := base64.StdEncoding.DecodeString(blob.InlineData.Data)
			if err == nil {
				return &ImageData{
					Data:     imageBytes,
					MimeType: blob.InlineData.MimeType,
				}
			}
		}
	}

	return nil
}