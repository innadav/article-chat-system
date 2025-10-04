package models

import "time"

// Article is the core data model for an article.
type Article struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Excerpt     string    `json:"excerpt"`
	Topics      []string  `json:"topics"`
	Sentiment   string    `json:"sentiment"`
	Summary     string    `json:"summary"`
	ProcessedAt time.Time `json:"processed_at"`
}
