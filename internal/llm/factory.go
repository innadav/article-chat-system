package llm

import (
	"article-chat-system/internal/config"
	"context"
	"fmt"
	"strings"
)

// NewClientFactory reads the config and returns the appropriate LLM client.
func NewClientFactory(ctx context.Context, cfg *config.Config) (Client, error) {
	provider := strings.ToLower(cfg.LLMProvider)
	switch provider {
	case "google", "gemini":
		return newGeminiClient(ctx, cfg.GoogleAPIKey)
	case "openai":
		return newOpenAIClient(ctx, cfg.OpenAIAPIKey)
	case "mock":
		return newMockClient(), nil
	default:
		return nil, fmt.Errorf("unknown or unsupported LLM provider: %s", cfg.LLMProvider)
	}
}
