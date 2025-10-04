package models

// EntityExtraction represents the structured response from LLM entity extraction
type EntityExtraction struct {
	Summary   string            `json:"summary"`
	Entities  []Entity          `json:"entities"`
	Keywords  []Keyword         `json:"keywords"`
	Topics    []Topic           `json:"topics"`
	Sentiment SentimentAnalysis `json:"sentiment"`
	Tone      ToneAnalysis      `json:"tone"`
}

// Entity represents a named entity with category and confidence
type Entity struct {
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

// Keyword represents a keyword with relevance and context
type Keyword struct {
	Term      string  `json:"term"`
	Relevance float64 `json:"relevance"`
	Context   string  `json:"context"`
}

// Topic represents a topic with score and description
type Topic struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Score      float64 `json:"score"`
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

// ToneAnalysis represents tone analysis results
type ToneAnalysis struct {
	Style      string  `json:"style"`
	Mood       string  `json:"mood"`
	Confidence float64 `json:"confidence"`
}
