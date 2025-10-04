package article

import (
	"article-chat-system/internal/models"
	"article-chat-system/internal/repository"
	"context"
)

// Service defines the contract for article-related operations.
type Service interface {
	GetArticle(ctx context.Context, url string) (*models.Article, bool)
	GetAllArticles(ctx context.Context) []*models.Article
	StoreArticle(ctx context.Context, article *models.Article) error
	CallSynthesisLLM(ctx context.Context, prompt string) (string, error)
	FindCommonEntities(ctx context.Context, articleURLs []string) ([]repository.EntityCount, error)
}
