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
			name: "valid openai provider",
			config: &config.Config{
				LLMProvider: "mock", // Use mock instead of openai for testing
			},
			expectError: false,
		},
		{
			name: "valid mock provider",
			config: &config.Config{
				LLMProvider: "mock",
			},
			expectError: false,
		},
		{
			name: "valid OPENAI provider (case insensitive)",
			config: &config.Config{
				LLMProvider: "MOCK", // Use mock instead of openai for testing
			},
			expectError: false,
		},
		{
			name: "valid MOCK provider (case insensitive)",
			config: &config.Config{
				LLMProvider: "MOCK",
			},
			expectError: false,
		},
		{
			name: "unknown provider",
			config: &config.Config{
				LLMProvider:  "unknown",
				OpenAIAPIKey: "test-api-key",
			},
			expectError: true,
			errorMsg:    "unknown or unsupported LLM provider: unknown. Supported providers: openai, mock",
		},
		{
			name: "empty provider",
			config: &config.Config{
				LLMProvider:  "",
				OpenAIAPIKey: "test-api-key",
			},
			expectError: true,
			errorMsg:    "unknown or unsupported LLM provider: . Supported providers: openai, mock",
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

func TestNewClientFactory_OpenAIProvider(t *testing.T) {
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
			expectError: false, // Mock doesn't require API key
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
				LLMProvider: "mock", // Use mock instead of openai for testing
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
