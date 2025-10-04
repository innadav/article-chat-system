package vector

import (
	"article-chat-system/internal/models"
	"context"
)

// Service defines the contract for vector database operations
type Service interface {
	// SearchByTopics searches for articles that contain the specified topics/parameters
	SearchByTopics(ctx context.Context, topics []string, limit int) ([]*models.Article, error)
	
	// SearchBySemanticSimilarity searches for articles semantically similar to the query
	SearchBySemanticSimilarity(ctx context.Context, query string, limit int) ([]*models.Article, error)
	
	// IndexArticle adds an article to the vector database
	IndexArticle(ctx context.Context, article *models.Article) error
	
	// RemoveArticle removes an article from the vector database
	RemoveArticle(ctx context.Context, url string) error
}

// VectorSearchResult represents a search result with similarity score
type VectorSearchResult struct {
	Article  *models.Article `json:"article"`
	Score    float64        `json:"score"`
	Distance float64        `json:"distance"`
}
