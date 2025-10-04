package article_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/models"
	"article-chat-system/internal/repository"
)

// mockRepository is a mock implementation of the ArticleRepository interface
type mockRepository struct {
	articles map[string]*models.Article
	saveErr  error
	findErr  error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		articles: make(map[string]*models.Article),
	}
}

func (m *mockRepository) Save(ctx context.Context, art *models.Article) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.articles[art.URL] = art
	return nil
}

func (m *mockRepository) FindByURL(ctx context.Context, url string) (*models.Article, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if art, exists := m.articles[url]; exists {
		return art, nil
	}
	return nil, nil // Not found
}

func (m *mockRepository) FindAll(ctx context.Context) ([]*models.Article, error) {
	var articles []*models.Article
	for _, art := range m.articles {
		articles = append(articles, art)
	}
	return articles, nil
}

func (m *mockRepository) FindTopEntities(ctx context.Context, articleURLs []string, limit int) ([]repository.EntityCount, error) {
	// Mock implementation - return some test entities
	return []repository.EntityCount{
		{Entity: "AI", Count: 5},
		{Entity: "Technology", Count: 3},
		{Entity: "Innovation", Count: 2},
	}, nil
}

// mockLLMClient is a mock implementation of the llm.Client interface
type mockLLMClient struct {
	response *llm.Response
	err      error
}

func newMockLLMClient() *mockLLMClient {
	return &mockLLMClient{}
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &llm.Response{Text: "Mock response"}, nil
}

func TestArticleService_GetArticle(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		setup    func(*mockRepository)
		expected bool
	}{
		{
			name: "article exists",
			url:  "https://example.com/test",
			setup: func(mr *mockRepository) {
				mr.articles["https://example.com/test"] = &models.Article{
					URL:         "https://example.com/test",
					Title:       "Test Article",
					ProcessedAt: time.Now(),
				}
			},
			expected: true,
		},
		{
			name:     "article does not exist",
			url:      "https://example.com/nonexistent",
			setup:    func(mr *mockRepository) {},
			expected: false,
		},
		{
			name: "repository error",
			url:  "https://example.com/error",
			setup: func(mr *mockRepository) {
				mr.findErr = errors.New("not found")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			tt.setup(mockRepo)
			mockLLM := newMockLLMClient()
			service := article.NewService(mockLLM, mockRepo, nil)

			article, exists := service.GetArticle(context.Background(), tt.url)

			if exists != tt.expected {
				t.Errorf("Expected exists=%v, got %v", tt.expected, exists)
			}

			if tt.expected && article == nil {
				t.Error("Expected article to be returned when exists=true")
			}
		})
	}
}

func TestArticleService_GetAllArticles(t *testing.T) {
	mockRepo := newMockRepository()
	mockRepo.articles["url1"] = &models.Article{URL: "url1", Title: "Article 1"}
	mockRepo.articles["url2"] = &models.Article{URL: "url2", Title: "Article 2"}

	mockLLM := newMockLLMClient()
	service := article.NewService(mockLLM, mockRepo, nil)

	articles := service.GetAllArticles(context.Background())

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}
}

func TestArticleService_StoreArticle(t *testing.T) {
	mockRepo := newMockRepository()
	mockLLM := newMockLLMClient()
	service := article.NewService(mockLLM, mockRepo, nil)

	article := &models.Article{
		URL:         "https://example.com/test",
		Title:       "Test Article",
		ProcessedAt: time.Now(),
	}

	err := service.StoreArticle(context.Background(), article)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify article was stored
	stored, exists := service.GetArticle(context.Background(), article.URL)
	if !exists {
		t.Error("Article should exist after storing")
	}
	if stored.Title != article.Title {
		t.Errorf("Expected title %s, got %s", article.Title, stored.Title)
	}
}

func TestArticleService_CallSynthesisLLM(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mockLLMClient)
		expected string
		hasError bool
	}{
		{
			name: "successful call",
			setup: func(m *mockLLMClient) {
				m.response = &llm.Response{Text: "Test response"}
			},
			expected: "Test response",
			hasError: false,
		},
		{
			name: "LLM error",
			setup: func(m *mockLLMClient) {
				m.err = errors.New("timeout")
			},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepository()
			mockLLM := newMockLLMClient()
			tt.setup(mockLLM)
			service := article.NewService(mockLLM, mockRepo, nil)

			result, err := service.CallSynthesisLLM(context.Background(), "test prompt")

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

// Add mock error types for testing
type mockErrNotFound struct{}

func (e *mockErrNotFound) Error() string { return "not found" }

type mockErrTimeout struct{}

func (e *mockErrTimeout) Error() string { return "timeout" }
