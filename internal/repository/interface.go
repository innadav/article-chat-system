package repository

import (
	"article-chat-system/internal/models"
	"context"
)

// ArticleRepository defines the interface for article persistence.
type ArticleRepository interface {
	Save(ctx context.Context, art *models.Article) error
	FindByURL(ctx context.Context, url string) (*models.Article, error)
	FindAll(ctx context.Context) ([]*models.Article, error)
}
