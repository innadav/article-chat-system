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
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
	"article-chat-system/internal/vector"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// mockLLMClient provides a predictable response for the analyzer.
type mockLLMClient struct{}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.Response, error) {
	// Return a structured analysis response for the Facade's analyzer to parse.
	return &llm.Response{Text: `{
		"headline": "Test Article: Integration Testing with PostgreSQL",
		"key_points": ["Database integration works", "LLM analysis successful", "Data persistence verified"],
		"sentiment": "Positive",
		"entities": ["PostgreSQL", "Integration", "Testing", "Database", "Analysis"]
	}`}, nil
}

// setupTestWithDB spins up a real PostgreSQL container for the test.
func setupTestWithDB(t *testing.T) (repository.ArticleRepository, func()) {
	ctx := context.Background()
	// Define the PostgreSQL container request
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

	// Teardown function to terminate the container after the test.
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

	// Run the schema initialization script.
	initSQL, err := os.ReadFile("scripts/init.sql")
	if err != nil {
		t.Fatalf("could not read init.sql: %v", err)
	}
	_, err = db.ExecContext(ctx, string(initSQL))
	if err != nil {
		t.Fatalf("could not run init.sql: %v", err)
	}

	repo := repository.NewPostgresRepositoryWithDB(db)
	return repo, teardown
}

func TestFacade_AddNewArticle_WithDatabase(t *testing.T) {
	// ARRANGE
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Change to project root directory for prompt loading
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to project root (two levels up from tests/integration/)
	if err := os.Chdir("../../"); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	repo, teardown := setupTestWithDB(t)
	defer teardown()

	mockLLM := &mockLLMClient{}
	articleSvc := article.NewService(mockLLM, repo, nil)

	// Create a prompt factory for the integration test
	promptLoader, err := prompts.NewLoader("v1")
	if err != nil {
		t.Fatalf("Failed to create prompt loader: %v", err)
	}
	promptFactory, err := prompts.NewFactory("test-model", promptLoader)
	if err != nil {
		t.Fatalf("Failed to create prompt factory: %v", err)
	}

	// For integration test, we don't need vector service
	var vectorSvc vector.Service = nil

	// For integration test, we don't need Weaviate, so we pass nil for vecRepo
	facade := processing.NewFacade(mockLLM, articleSvc, promptFactory, vectorSvc, nil)

	testURL := "https://example.com/integration-test"

	// ACT
	_, err = facade.AddNewArticle(context.Background(), testURL)
	if err != nil {
		t.Fatalf("Facade.AddNewArticle() failed unexpectedly: %v", err)
	}

	// ASSERT
	// Directly query the database to verify the data was saved correctly.
	savedArticle, err := repo.FindByURL(context.Background(), testURL)
	if err != nil {
		t.Fatalf("Repository.FindByURL() failed: %v", err)
	}
	if savedArticle == nil {
		t.Fatal("Article was not saved to the database")
	}

	// Verify the data from the mock LLM was correctly parsed and saved.
	if savedArticle.URL != testURL {
		t.Errorf("expected URL '%s', got '%s'", testURL, savedArticle.URL)
	}
	if savedArticle.Sentiment != "Positive" {
		t.Errorf("expected sentiment 'Positive', got '%s'", savedArticle.Sentiment)
	}
	if len(savedArticle.Topics) != 5 {
		t.Errorf("expected 5 topics, got %d", len(savedArticle.Topics))
	}
	if len(savedArticle.Entities) != 5 {
		t.Errorf("expected 5 entities, got %d", len(savedArticle.Entities))
	}

	// Verify the summary contains the headline and key points
	expectedSummary := "Test Article: Integration Testing with PostgreSQL\n- Database integration works\n- LLM analysis successful\n- Data persistence verified"
	if savedArticle.Summary != expectedSummary {
		t.Errorf("expected summary '%s', got '%s'", expectedSummary, savedArticle.Summary)
	}
}
