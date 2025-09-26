package analyzer

import (
	"encoding/json"
	"fmt"
)

type Analyzer interface {
	Analyze(imagePath string) (json.RawMessage, error)
	GetType() string
}

type BaseAnalyzer struct {
	Type string
}

func (b *BaseAnalyzer) GetType() string {
	return b.Type
}

type Result struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewResult(analyzerType string, data interface{}) (*Result, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	return &Result{
		Type: analyzerType,
		Data: jsonData,
	}, nil
}