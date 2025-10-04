package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"

	// "article-chat-system/internal/models" // Removed unused import
	"article-chat-system/internal/processing"
	"article-chat-system/internal/repository"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Mock LLM for integration testing
type mockLLMClient struct{}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.Response, error) {
	return &llm.Response{Text: "SENTIMENT: Neutral, KEYWORDS: test, keyword"}, nil
}

// setupTestWithDB spins up a real PostgreSQL container for the test.
func setupTestWithDB(t *testing.T) (repository.ArticleRepository, func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5 * time.Minute),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("could not start postgres container: %s", err)
	}

	teardown := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")
	connStr := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("could not connect to test postgres: %s", err)
	}

	// For this test, you'd need a way to run your init.sql schema here.
	initSQL, err := os.ReadFile("../../scripts/init.sql")
	if err != nil {
		t.Fatalf("failed to read init.sql: %v", err)
	}
	_, err = db.ExecContext(ctx, string(initSQL))
	if err != nil {
		t.Fatalf("failed to execute init.sql: %v", err)
	}

	repo := repository.NewPostgresRepositoryWithDB(db)
	return repo, teardown
}

func TestFacade_AddNewArticle_Postgres(t *testing.T) {
	// ARRANGE
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, teardown := setupTestWithDB(t)
	defer teardown()

	mockLLM := &mockLLMClient{}

	articleSvc := article.NewService(mockLLM, repo)
	facade := processing.NewFacade(mockLLM, articleSvc)

	testURL := "https://example.com/integration-test"

	// ACT
	_, err := facade.AddNewArticle(context.Background(), testURL)
	if err != nil {
		t.Fatalf("Facade.AddNewArticle() failed: %v", err)
	}

	// ASSERT
	savedArticle, err := repo.FindByURL(context.Background(), testURL)
	if err != nil {
		t.Fatalf("Repository.FindByURL() failed: %v", err)
	}
	if savedArticle == nil || savedArticle.URL != testURL {
		t.Errorf("expected URL %s, got %s", testURL, savedArticle.URL)
	}
}
