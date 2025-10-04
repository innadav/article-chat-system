package llm_test

import (
	"context"
	"testing"

	"article-chat-system/internal/config"
	"article-chat-system/internal/llm"
)

func TestNewClientFactory(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid google provider",
			config: &config.Config{
				LLMProvider:  "google",
				GoogleAPIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid gemini provider",
			config: &config.Config{
				LLMProvider:  "gemini",
				GoogleAPIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid GOOGLE provider (case insensitive)",
			config: &config.Config{
				LLMProvider:  "GOOGLE",
				GoogleAPIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "valid GEMINI provider (case insensitive)",
			config: &config.Config{
				LLMProvider:  "GEMINI",
				GoogleAPIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "unknown provider",
			config: &config.Config{
				LLMProvider:  "unknown",
				GoogleAPIKey: "test-api-key",
			},
			expectError: true,
			errorMsg:    "unknown or unsupported LLM provider: unknown",
		},
		{
			name: "empty provider",
			config: &config.Config{
				LLMProvider:  "",
				GoogleAPIKey: "test-api-key",
			},
			expectError: true,
			errorMsg:    "unknown or unsupported LLM provider: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := llm.NewClientFactory(context.Background(), tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if client != nil {
					t.Error("Expected client to be nil when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected client to be non-nil")
				}
			}
		})
	}
}

func TestNewClientFactory_GoogleProvider(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		expectError bool
	}{
		{
			name:        "valid api key",
			apiKey:      "test-api-key",
			expectError: false,
		},
		{
			name:        "empty api key",
			apiKey:      "",
			expectError: true,
		},
		{
			name:        "long api key",
			apiKey:      "very-long-api-key-that-should-be-valid-for-testing-purposes",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				LLMProvider:  "google",
				GoogleAPIKey: tt.apiKey,
			}

			client, err := llm.NewClientFactory(context.Background(), cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected client to be non-nil")
				}
			}
		})
	}
}
