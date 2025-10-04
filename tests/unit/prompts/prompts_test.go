package prompts_test

import (
	"html/template"
	"os"
	"path/filepath"
	"testing"

	"article-chat-system/internal/models"
	"article-chat-system/internal/prompts"
)

func TestNewLoader(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "valid version",
			version:     "v1",
			expectError: false,
		},
		{
			name:        "empty version",
			version:     "",
			expectError: false,
		},
		{
			name:        "version with path",
			version:     "v2/test",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader, err := prompts.NewLoader(tt.version)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if loader == nil {
					t.Error("Expected loader to be non-nil")
				}
				if loader.PromptDir != filepath.Join("configs", "prompts", tt.version) {
					t.Errorf("Expected PromptDir %s, got %s",
						filepath.Join("configs", "prompts", tt.version), loader.PromptDir)
				}
			}
		})
	}
}

func TestLoader_LoadPrompt(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test prompt file
	testPromptFile := filepath.Join(tempDir, "test.yaml")
	testPromptContent := `template: "Hello {{.Name}}, this is a test prompt"`
	err := os.WriteFile(testPromptFile, []byte(testPromptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	tests := []struct {
		name        string
		promptName  string
		setupLoader func() *prompts.Loader
		expectError bool
	}{
		{
			name:       "load existing prompt",
			promptName: "test",
			setupLoader: func() *prompts.Loader {
				loader := &prompts.Loader{
					PromptDir: tempDir,
					Cache:     make(map[string]*template.Template),
				}
				return loader
			},
			expectError: false,
		},
		{
			name:       "load non-existing prompt",
			promptName: "nonexistent",
			setupLoader: func() *prompts.Loader {
				loader := &prompts.Loader{
					PromptDir: tempDir,
					Cache:     make(map[string]*template.Template),
				}
				return loader
			},
			expectError: true,
		},
		{
			name:       "load from cache",
			promptName: "test",
			setupLoader: func() *prompts.Loader {
				loader := &prompts.Loader{
					PromptDir: tempDir,
					Cache:     make(map[string]*template.Template),
				}
				// Pre-load the template into cache
				template, _ := template.New("test").Parse("Hello {{.Name}}, this is a test prompt")
				loader.Cache["test"] = template
				return loader
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := tt.setupLoader()
			tmpl, err := loader.LoadPrompt(tt.promptName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tmpl != nil {
					t.Error("Expected template to be nil when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tmpl == nil {
					t.Error("Expected template to be non-nil")
				}
			}
		})
	}
}

func TestNewFactory(t *testing.T) {
	loader := &prompts.Loader{
		PromptDir: "test/prompts",
		Cache:     make(map[string]*template.Template),
	}

	tests := []struct {
		name        string
		model       string
		loader      *prompts.Loader
		expectError bool
	}{
		{
			name:        "valid factory",
			model:       "gemini-1.5-flash",
			loader:      loader,
			expectError: false,
		},
		{
			name:        "nil loader",
			model:       "gemini-1.5-flash",
			loader:      nil,
			expectError: false, // Factory doesn't validate loader
		},
		{
			name:        "empty model",
			model:       "",
			loader:      loader,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, err := prompts.NewFactory(tt.model, tt.loader)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if factory != nil {
					t.Error("Expected factory to be nil when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if factory == nil {
					t.Error("Expected factory to be non-nil")
				}
				if factory.Model != tt.model {
					t.Errorf("Expected model %s, got %s", tt.model, factory.Model)
				}
				if factory.Loader != tt.loader {
					t.Error("Expected loader to match")
				}
			}
		})
	}
}

func TestFactory_CreateSummarizePrompt(t *testing.T) {
	// Create a temporary directory and test prompt file
	tempDir := t.TempDir()
	testPromptFile := filepath.Join(tempDir, "summarize.yaml")
	testPromptContent := `template: "Summarize this content: {{.Content}}"`
	err := os.WriteFile(testPromptFile, []byte(testPromptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	loader := &prompts.Loader{
		PromptDir: tempDir,
		Cache:     make(map[string]*template.Template),
	}

	factory, err := prompts.NewFactory("test-model", loader)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	prompt, err := factory.CreateSummarizePrompt("This is test content")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Summarize this content: This is test content"
	if prompt != expected {
		t.Errorf("Expected prompt '%s', got '%s'", expected, prompt)
	}
}

func TestFactory_CreateKeywordsPrompt(t *testing.T) {
	// Create a temporary directory and test prompt file
	tempDir := t.TempDir()
	testPromptFile := filepath.Join(tempDir, "keywords.yaml")
	testPromptContent := `template: "Extract keywords from '{{.Title}}': {{.Content}}"`
	err := os.WriteFile(testPromptFile, []byte(testPromptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	loader := &prompts.Loader{
		PromptDir: tempDir,
		Cache:     make(map[string]*template.Template),
	}

	factory, err := prompts.NewFactory("test-model", loader)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	prompt, err := factory.CreateKeywordsPrompt("Test Title", "Test content")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Extract keywords from 'Test Title': Test content"
	if prompt != expected {
		t.Errorf("Expected prompt '%s', got '%s'", expected, prompt)
	}
}

func TestFactory_CreateSentimentPrompt(t *testing.T) {
	// Create a temporary directory and test prompt file
	tempDir := t.TempDir()
	testPromptFile := filepath.Join(tempDir, "sentiment.yaml")
	testPromptContent := `template: "Analyze sentiment of '{{.Title}}': {{.Content}}"`
	err := os.WriteFile(testPromptFile, []byte(testPromptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	loader := &prompts.Loader{
		PromptDir: tempDir,
		Cache:     make(map[string]*template.Template),
	}

	factory, err := prompts.NewFactory("test-model", loader)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	prompt, err := factory.CreateSentimentPrompt("Test Title", "Test content")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Analyze sentiment of 'Test Title': Test content"
	if prompt != expected {
		t.Errorf("Expected prompt '%s', got '%s'", expected, prompt)
	}
}

func TestFactory_CreatePlannerPrompt(t *testing.T) {
	// Create a temporary directory and test prompt file
	tempDir := t.TempDir()
	testPromptFile := filepath.Join(tempDir, "planner.yaml")
	testPromptContent := `template: "Query: {{.Query}}\nArticles:\n- {{.Articles}}"`
	err := os.WriteFile(testPromptFile, []byte(testPromptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	loader := &prompts.Loader{
		PromptDir: tempDir,
		Cache:     make(map[string]*template.Template),
	}

	factory, err := prompts.NewFactory("test-model", loader)
	if err != nil {
		t.Fatalf("Failed to create factory: %v", err)
	}

	articles := []*models.Article{
		{Title: "Article 1"},
		{Title: "Article 2"},
	}

	prompt, err := factory.CreatePlannerPrompt("test query", articles)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "Query: test query\nArticles:\n- Article 1\n- Article 2"
	if prompt != expected {
		t.Errorf("Expected prompt '%s', got '%s'", expected, prompt)
	}
}
