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
func (s *ArticleService) FindCommonEntities(ctx context.Context, articleURLs []string) ([]EntityCount, error) {
	var articles []*models.Article

	if len(articleURLs) == 0 {
		// If no URLs provided, get all articles
		articles = s.GetAllArticles(ctx)
	} else {
		// Get specific articles by URL
		for _, url := range articleURLs {
			if article, found := s.GetArticle(ctx, url); found {
				articles = append(articles, article)
			}
		}
	}

	if len(articles) == 0 {
		return []EntityCount{}, nil
	}

	// Count entities (topics) from articles
	entityCounts := make(map[string]int)
	for _, article := range articles {
		for _, topic := range article.Topics {
			if topic != "" {
				entityCounts[topic]++
			}
		}
	}

	// Convert to slice and sort by count
	result := make([]EntityCount, 0, len(entityCounts))
	for entity, count := range entityCounts {
		result = append(result, EntityCount{Entity: entity, Count: count})
	}

	// Sort by count (descending)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Count < result[j].Count {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Return top 10
	topCount := 10
	if len(result) < topCount {
		topCount = len(result)
	}

	return result[:topCount], nil
}
