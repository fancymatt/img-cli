package gemini

import "encoding/json"

type Request struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

type GenerationConfig struct {
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	TopK             int     `json:"topK,omitempty"`
	TopP             float64 `json:"topP,omitempty"`
}

type Content struct {
	Parts []interface{} `json:"parts"`
}

type TextPart struct {
	Text string `json:"text"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type BlobPart struct {
	InlineData InlineData `json:"inlineData"`
}

type Response struct {
	Candidates []Candidate `json:"candidates"`
	Error      *APIError   `json:"error,omitempty"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type AnalysisResult struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OutfitDescription struct {
	Clothing    []interface{}   `json:"clothing"` // Can be strings or ClothingItem objects
	Style       string          `json:"style"`
	Colors      []string        `json:"colors"`
	Accessories []interface{}   `json:"accessories,omitempty"` // Can be strings or AccessoryItem objects
	Overall     string          `json:"overall"`
	Hair        *HairDescription `json:"hair,omitempty"`
}

type ClothingItem struct {
	Item                   string `json:"item"`
	Description            string `json:"description"`
	MainBodyColor          string `json:"main_body_color"`
	CollarColor            string `json:"collar_color"`
	CuffColor              string `json:"cuff_color"`
	ButtonsClosuresColor   string `json:"buttons_closures_color"`
	TrimColor              string `json:"trim_color"`
}

type HairDescription struct {
	Color       string   `json:"color"`
	Style       string   `json:"style"`
	Length      string   `json:"length"`
	Texture     string   `json:"texture"`
	Details     []string `json:"details,omitempty"`
	Styling     string   `json:"styling,omitempty"`
}

type VisualStyle struct {
	Composition      string   `json:"composition"`
	Framing          string   `json:"framing"`
	Pose             string   `json:"pose"`
	BodyPosition     string   `json:"body_position"`
	Lighting         string   `json:"lighting"`
	ColorPalette     []string `json:"color_palette"`
	ColorGrading     string   `json:"color_grading"`
	Mood             string   `json:"mood"`
	Background       string   `json:"background"`
	Photographic     string   `json:"photographic_style"`
	ArtisticStyle    string   `json:"artistic_style,omitempty"`
	FilmGrain        string   `json:"film_grain"`
	ImageQuality     string   `json:"image_quality"`
	Era              string   `json:"era_aesthetic"`
	CameraAngle      string   `json:"camera_angle"`
	DepthOfField     string   `json:"depth_of_field"`
	PostProcessing   string   `json:"post_processing"`
}

type ImageData struct {
	Data     []byte
	MimeType string
}