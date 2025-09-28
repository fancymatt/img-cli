package models

import "encoding/json"

// ModularComponents holds analyzed component data
type ModularComponents struct {
	Outfit      *ComponentData
	OverOutfit  *ComponentData // Base layer outfit that the main outfit is worn over
	Style       *ComponentData
	HairStyle   *ComponentData
	HairColor   *ComponentData
	Makeup      *ComponentData
	Expression  *ComponentData
	Accessories *ComponentData
}

// ComponentData holds analyzed data for a single component
type ComponentData struct {
	Type        string
	Description string
	JSONData    json.RawMessage
	ImagePath   string
}