package workflow

import (
	"encoding/json"
	"time"
)

type WorkflowOptions struct {
	OutputDir       string
	Outfits         []string
	StyleReference  string
	StylePrompt     string
	NewOutfit       string
	OutfitReference string
	OutfitText      string // Text description of outfit (alternative to OutfitReference)
	HairReference   string
	TargetImage     string
	DebugPrompt     bool
	SendOriginal    bool   // Include outfit reference image in generation request
	Variations      int
	Prompt          string // For text-to-image generation and naming
}

type WorkflowResult struct {
	Workflow  string       `json:"workflow"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Steps     []StepResult `json:"steps"`
}

type StepResult struct {
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Data       json.RawMessage `json:"data,omitempty"`
	OutputPath string          `json:"output_path,omitempty"`
	Message    string          `json:"message,omitempty"`
}