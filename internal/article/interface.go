package article

import (
	"article-chat-system/internal/models"
	"context"
)

// Service defines the contract for article-related operations.
type Service interface {
	GetArticle(ctx context.Context, url string) (*models.Article, bool)
	GetAllArticles(ctx context.Context) []*models.Article
	StoreArticle(ctx context.Context, article *models.Article) error
	CallSynthesisLLM(ctx context.Context, prompt string) (string, error)
	FindCommonEntities(ctx context.Context, articleURLs []string) ([]EntityCount, error)
}

// EntityCount represents an entity with its frequency count
type EntityCount struct {
	Entity string `json:"entity"`
	Count  int    `json:"count"`
}
