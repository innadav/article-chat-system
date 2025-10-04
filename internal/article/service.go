package article

import (
	"context"
	"time"

	"article-chat-system/internal/llm"
	"article-chat-system/internal/models" // Import 'models'
	"article-chat-system/internal/repository"
)

// ArticleService orchestrates calls to the repository and LLM.
type ArticleService struct {
	repo      repository.ArticleRepository
	llmClient llm.Client
}

// NewService is the constructor for the article service.
func NewService(llmClient llm.Client, repo repository.ArticleRepository) *ArticleService {
	return &ArticleService{
		repo:      repo,
		llmClient: llmClient,
	}
}

// GetArticle retrieves a single article from the repository.
func (s *ArticleService) GetArticle(ctx context.Context, url string) (*models.Article, bool) {
	art, err := s.repo.FindByURL(ctx, url)
	if err != nil || art == nil {
		return nil, false
	}
	return art, true
}

// GetAllArticles retrieves all articles from the repository.
// Deprecated: Use FindTopEntities or other specific repository methods instead for better performance.
func (s *ArticleService) GetAllArticles(ctx context.Context) []*models.Article {
	articles, err := s.repo.FindAll(ctx)
	if err != nil {
		return []*models.Article{}
	}
	return articles
}

// StoreArticle saves an article to the repository.
func (s *ArticleService) StoreArticle(ctx context.Context, article *models.Article) error {
	return s.repo.Save(ctx, article)
}

// CallSynthesisLLM is a helper method for strategies to generate text.
func (s *ArticleService) CallSynthesisLLM(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	resp, err := s.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// FindCommonEntities finds the top 10 most common entities from specified articles or all articles if no URLs provided
func (s *ArticleService) FindCommonEntities(ctx context.Context, articleURLs []string) ([]repository.EntityCount, error) {
	// Use efficient PostgreSQL query instead of loading all articles into memory
	return s.repo.FindTopEntities(ctx, articleURLs, 10)
}
