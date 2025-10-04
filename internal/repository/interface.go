package repository

import (
	"article-chat-system/internal/models"
	"context"
)

// EntityCount represents an entity with its frequency count
type EntityCount struct {
	Entity string `json:"entity"`
	Count  int    `json:"count"`
}

// ArticleRepository defines the interface for article persistence.
type ArticleRepository interface {
	Save(ctx context.Context, art *models.Article) error
	FindByURL(ctx context.Context, url string) (*models.Article, error)
	FindAll(ctx context.Context) ([]*models.Article, error)
	FindTopEntities(ctx context.Context, articleURLs []string, limit int) ([]EntityCount, error)
}
