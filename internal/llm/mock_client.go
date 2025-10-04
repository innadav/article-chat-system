package llm

import (
	"context"
	"fmt"
	"strings"
)

// mockClient is a mock implementation for testing purposes.
type mockClient struct{}

// newMockClient creates a mock LLM client for testing.
func newMockClient() Client {
	return &mockClient{}
}

// GenerateContent returns a mock response based on the prompt content.
func (c *mockClient) GenerateContent(ctx context.Context, prompt string) (*Response, error) {
	// Simple mock responses based on prompt keywords
	lowerPrompt := strings.ToLower(prompt)

	// Check if this is a planner prompt (contains "Determine the intent" and "Respond in JSON format")
	if strings.Contains(lowerPrompt, "determine the intent") && strings.Contains(lowerPrompt, "respond in json format") {
		// Extract URL from the prompt if present
		url := ""
		if strings.Contains(prompt, "https://techcrunch.com/2025/10/02/last-chance-alert-founder-and-investor-bundle-savings-for-techcrunch-disrupt-2025-ends-tomorrow/") {
			url = "https://techcrunch.com/2025/10/02/last-chance-alert-founder-and-investor-bundle-savings-for-techcrunch-disrupt-2025-ends-tomorrow/"
		} else if strings.Contains(prompt, "https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/") {
			url = "https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"
		} else if strings.Contains(prompt, "https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/") {
			url = "https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/"
		} else if strings.Contains(prompt, "https://techcrunch.com/2025/07/27/itch-io-is-the-latest-marketplace-to-crack-down-on-adult-games/") {
			url = "https://techcrunch.com/2025/07/27/itch-io-is-the-latest-marketplace-to-crack-down-on-adult-games/"
		} else if strings.Contains(prompt, "https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/") {
			url = "https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/"
		} else if strings.Contains(prompt, "https://edition.cnn.com/2025/07/24/tech/intel-layoffs-15-percent-q2-earnings") {
			url = "https://edition.cnn.com/2025/07/24/tech/intel-layoffs-15-percent-q2-earnings"
		}

		// Check if user is asking about articles in general
		if strings.Contains(lowerPrompt, "what articles") || strings.Contains(lowerPrompt, "list articles") || strings.Contains(lowerPrompt, "articles do you have") {
			return &Response{
				Text: `{"intent": "UNKNOWN", "targets": [], "parameters": []}`,
			}, nil
		}

		if url != "" {
			return &Response{
				Text: fmt.Sprintf(`{"intent": "SUMMARIZE", "targets": ["%s"], "parameters": []}`, url),
			}, nil
		}

		return &Response{
			Text: `{"intent": "SUMMARIZE", "targets": [], "parameters": []}`,
		}, nil
	}

	if strings.Contains(lowerPrompt, "summarize") || strings.Contains(lowerPrompt, "summary") {
		return &Response{
			Text: "This is a mock summary of the article. The article discusses the main topic and provides key insights about the subject matter.",
		}, nil
	}

	if strings.Contains(lowerPrompt, "keywords") || strings.Contains(lowerPrompt, "key topics") {
		return &Response{
			Text: "Key topics: technology, innovation, business strategy, product development",
		}, nil
	}

	if strings.Contains(lowerPrompt, "sentiment") {
		return &Response{
			Text: "The sentiment of this article is generally positive, focusing on opportunities and growth.",
		}, nil
	}

	if strings.Contains(lowerPrompt, "compare") {
		return &Response{
			Text: "When comparing these articles, the main differences lie in their focus areas and the perspectives they present.",
		}, nil
	}

	// Default mock response
	return &Response{
		Text: fmt.Sprintf("Mock response to: %s", prompt[:min(50, len(prompt))]),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
