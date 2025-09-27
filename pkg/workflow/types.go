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
	TargetImage     string   // Single target (for backward compatibility)
	TargetImages    []string // Multiple targets for outfit-swap workflow
	DebugPrompt     bool
	SendOriginal    bool   // Include outfit reference image in generation request
	Variations      int
	Prompt          string // For text-to-image generation and naming
	SkipCostConfirm bool   // Skip cost confirmation prompts (for automation)
	// Modular component references
	HairStyleRef   string
	HairColorRef   string
	MakeupRef      string
	ExpressionRef  string
	AccessoriesRef string
}

type WorkflowResult struct {
	Workflow       string       `json:"workflow"`
	StartTime      time.Time    `json:"start_time"`
	EndTime        time.Time    `json:"end_time"`
	Steps          []StepResult `json:"steps"`
	SubjectCount   int          `json:"subject_count,omitempty"`
	OutfitCount    int          `json:"outfit_count,omitempty"`
	StyleCount     int          `json:"style_count,omitempty"`
	VariationCount int          `json:"variation_count,omitempty"`
}

type StepResult struct {
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Data       json.RawMessage `json:"data,omitempty"`
	OutputPath string          `json:"output_path,omitempty"`
	Message    string          `json:"message,omitempty"`
}