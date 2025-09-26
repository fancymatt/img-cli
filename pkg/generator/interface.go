package generator

import "encoding/json"

type Generator interface {
	Generate(params GenerateParams) (*GenerateResult, error)
	GetType() string
}

type GenerateParams struct {
	ImagePath       string
	Prompt          string
	StyleData       json.RawMessage
	OutfitData      json.RawMessage
	HairData        json.RawMessage
	StyleAnalysis   json.RawMessage // Analysis data for art style
	StyleReference  string          // Path to style reference image
	OutfitReference string          // Path to outfit reference image (for --send-original)
	OutputDir       string
	Temperature     float64
	DebugPrompt     bool
	OutfitSource    string // Name of outfit source file (without extension)
	StyleSource     string // Name of style source file (without extension)
	HairSource      string // Name of hair source file (without extension)
	VariationIndex  int    // Which variation this is (1, 2, 3, etc.)
	TotalVariations int    // Total number of variations being generated
	SendOriginal    bool   // Whether to include the outfit reference image in the request
}

type GenerateResult struct {
	Type       string `json:"type"`
	OutputPath string `json:"output_path"`
	Message    string `json:"message"`
}

type BaseGenerator struct {
	Type string
}

func (b *BaseGenerator) GetType() string {
	return b.Type
}