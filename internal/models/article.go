package models

import "time"

// Article is the core data model for an article.
type Article struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Excerpt     string    `json:"excerpt"`
	Summary     string    `json:"summary"`
	Sentiment   string    `json:"sentiment"`
	Topics      []string  `json:"topics"`
	Entities    []string  `json:"entities"`
	ProcessedAt time.Time `json:"processed_at"`
}
