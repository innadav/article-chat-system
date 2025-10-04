package vector

import (
	"article-chat-system/internal/article"
	"article-chat-system/internal/models"
	"context"
	"strings"
)

// SimpleVectorService implements vector operations using the existing article service
// This is a placeholder implementation that can be replaced with a real vector DB
type SimpleVectorService struct {
	articleSvc article.Service
}

// NewSimpleVectorService creates a new simple vector service
func NewSimpleVectorService(articleSvc article.Service) *SimpleVectorService {
	return &SimpleVectorService{
		articleSvc: articleSvc,
	}
}

// SearchByTopics searches for articles that contain the specified topics/parameters
func (s *SimpleVectorService) SearchByTopics(ctx context.Context, topics []string, limit int) ([]*models.Article, error) {
	// Get all articles from the article service
	allArticles := s.articleSvc.GetAllArticles(ctx)

	var matchingArticles []*models.Article

	for _, article := range allArticles {
		// Check if article contains any of the specified topics
		articleText := strings.ToLower(article.Title + " " + article.Summary + " " + strings.Join(article.Topics, " "))

		for _, topic := range topics {
			if strings.Contains(articleText, strings.ToLower(topic)) {
				matchingArticles = append(matchingArticles, article)
				break // Found a match, no need to check other topics for this article
			}
		}

		// Limit results
		if len(matchingArticles) >= limit {
			break
		}
	}

	return matchingArticles, nil
}

// SearchBySemanticSimilarity searches for articles semantically similar to the query
func (s *SimpleVectorService) SearchBySemanticSimilarity(ctx context.Context, query string, limit int) ([]*models.Article, error) {
	// For now, use topic-based search as a fallback
	// In a real implementation, this would use vector embeddings
	queryWords := strings.Fields(strings.ToLower(query))
	return s.SearchByTopics(ctx, queryWords, limit)
}

// IndexArticle adds an article to the vector database
func (s *SimpleVectorService) IndexArticle(ctx context.Context, article *models.Article) error {
	// For now, articles are already indexed in the article service
	// In a real implementation, this would add the article to the vector DB
	return nil
}

// RemoveArticle removes an article from the vector database
func (s *SimpleVectorService) RemoveArticle(ctx context.Context, url string) error {
	// For now, articles are managed by the article service
	// In a real implementation, this would remove the article from the vector DB
	return nil
}
