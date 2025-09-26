// Package models provides data structures and types used throughout the image generation application.
// It defines concrete types for analysis results, generation parameters, and other domain objects
// to ensure type safety and clarity.
package models

import (
	"time"
)

// OutfitAnalysis represents the result of analyzing an outfit image
type OutfitAnalysis struct {
	ClothingItems []ClothingItem `json:"clothing_items"`
	Overall       OverallStyle   `json:"overall"`
	Hair          HairDetails    `json:"hair"`
	Accessories   []string       `json:"accessories"`
	Timestamp     time.Time      `json:"timestamp"`
}

// ClothingItem represents a single piece of clothing
type ClothingItem struct {
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Color        string   `json:"color"`
	Material     string   `json:"material,omitempty"`
	Pattern      string   `json:"pattern,omitempty"`
	Fit          string   `json:"fit,omitempty"`
	Details      []string `json:"details,omitempty"`
	Construction string   `json:"construction,omitempty"`
	Hardware     string   `json:"hardware,omitempty"`
}

// OverallStyle represents the overall style analysis
type OverallStyle struct {
	Genre           string   `json:"genre"`
	Formality       string   `json:"formality"`
	Aesthetic       string   `json:"aesthetic"`
	Influences      []string `json:"influences,omitempty"`
	Era             string   `json:"era,omitempty"`
	Season          string   `json:"season,omitempty"`
	ColorPalette    []string `json:"color_palette"`
	Description     string   `json:"description"`
	StyleNotes      []string `json:"style_notes,omitempty"`
	KeyCharacteristics []string `json:"key_characteristics,omitempty"`
}

// HairDetails represents hair analysis details
type HairDetails struct {
	Color   string   `json:"color"`
	Length  string   `json:"length"`
	Style   string   `json:"style"`
	Texture string   `json:"texture"`
	Details []string `json:"details,omitempty"`
}

// VisualStyleAnalysis represents the result of analyzing visual/photographic style
type VisualStyleAnalysis struct {
	Photography   PhotographyStyle   `json:"photography"`
	Lighting      LightingDetails    `json:"lighting"`
	ColorGrading  ColorGrading       `json:"color_grading"`
	Composition   CompositionDetails `json:"composition"`
	PostProcessing []string          `json:"post_processing,omitempty"`
	Mood          string            `json:"mood"`
	Timestamp     time.Time         `json:"timestamp"`
}

// PhotographyStyle represents photographic style details
type PhotographyStyle struct {
	Type        string   `json:"type"`
	Perspective string   `json:"perspective"`
	FocalLength string   `json:"focal_length,omitempty"`
	Aperture    string   `json:"aperture,omitempty"`
	Style       string   `json:"style"`
	Techniques  []string `json:"techniques,omitempty"`
}

// LightingDetails represents lighting analysis
type LightingDetails struct {
	Type       string   `json:"type"`
	Direction  string   `json:"direction"`
	Quality    string   `json:"quality"`
	Intensity  string   `json:"intensity"`
	ColorTemp  string   `json:"color_temperature,omitempty"`
	Modifiers  []string `json:"modifiers,omitempty"`
}

// ColorGrading represents color grading details
type ColorGrading struct {
	Overall     string   `json:"overall"`
	Shadows     string   `json:"shadows,omitempty"`
	Midtones    string   `json:"midtones,omitempty"`
	Highlights  string   `json:"highlights,omitempty"`
	Saturation  string   `json:"saturation"`
	Contrast    string   `json:"contrast"`
	Temperature string   `json:"temperature"`
}

// CompositionDetails represents composition analysis
type CompositionDetails struct {
	Framing    string   `json:"framing"`
	Balance    string   `json:"balance"`
	Background string   `json:"background"`
	Foreground string   `json:"foreground,omitempty"`
	Depth      string   `json:"depth,omitempty"`
	Rules      []string `json:"rules,omitempty"`
}

// ArtStyleAnalysis represents artistic style analysis
type ArtStyleAnalysis struct {
	Style       ArtisticStyle `json:"style"`
	Technique   Technique     `json:"technique"`
	Elements    ArtElements   `json:"elements"`
	Influences  []string      `json:"influences,omitempty"`
	Description string        `json:"description"`
	Timestamp   time.Time     `json:"timestamp"`
}

// ArtisticStyle represents the artistic style characteristics
type ArtisticStyle struct {
	Movement    string   `json:"movement,omitempty"`
	Genre       string   `json:"genre"`
	Period      string   `json:"period,omitempty"`
	Aesthetic   string   `json:"aesthetic"`
	Mood        string   `json:"mood"`
	Attributes  []string `json:"attributes,omitempty"`
}

// Technique represents artistic technique details
type Technique struct {
	Medium      string   `json:"medium"`
	Brushwork   string   `json:"brushwork,omitempty"`
	Texture     string   `json:"texture"`
	Linework    string   `json:"linework,omitempty"`
	Shading     string   `json:"shading,omitempty"`
	Details     []string `json:"details,omitempty"`
}

// ArtElements represents artistic elements
type ArtElements struct {
	ColorPalette []string `json:"color_palette"`
	Composition  string   `json:"composition"`
	Lighting     string   `json:"lighting"`
	Perspective  string   `json:"perspective,omitempty"`
	Focus        string   `json:"focus,omitempty"`
}

// GenerationRequest represents a request to generate an image
type GenerationRequest struct {
	SubjectImage    string          `json:"subject_image"`
	Prompt          string          `json:"prompt"`
	OutfitAnalysis  *OutfitAnalysis `json:"outfit_analysis,omitempty"`
	StyleAnalysis   interface{}     `json:"style_analysis,omitempty"` // Can be VisualStyleAnalysis or ArtStyleAnalysis
	ReferenceImage  string          `json:"reference_image,omitempty"`
	Temperature     float64         `json:"temperature,omitempty"`
	OutputDir       string          `json:"output_dir"`
	SendOriginal    bool            `json:"send_original"`
}

// GenerationResult represents the result of image generation
type GenerationResult struct {
	OutputPath   string    `json:"output_path"`
	Message      string    `json:"message"`
	GeneratedAt  time.Time `json:"generated_at"`
	ProcessingTime float64 `json:"processing_time_seconds"`
}

// WorkflowConfig represents configuration for a workflow
type WorkflowConfig struct {
	Type            string            `json:"type"`
	InputPath       string            `json:"input_path"`
	OutputDir       string            `json:"output_dir"`
	Options         map[string]string `json:"options,omitempty"`
	StyleReference  string            `json:"style_reference,omitempty"`
	OutfitReference string            `json:"outfit_reference,omitempty"`
	TestMode        bool              `json:"test_mode"`
	TestSubject     string            `json:"test_subject,omitempty"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries   int            `json:"total_entries"`
	EntriesByType  map[string]int `json:"entries_by_type"`
	TotalSize      int64          `json:"total_size_bytes"`
	OldestEntry    time.Time      `json:"oldest_entry,omitempty"`
	NewestEntry    time.Time      `json:"newest_entry,omitempty"`
}