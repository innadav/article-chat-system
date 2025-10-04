package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// geminiClient is the concrete implementation for the Google Gemini provider.
type geminiClient struct {
	model *genai.GenerativeModel
}

// newGeminiClient creates a client for interacting with Gemini.
func newGeminiClient(ctx context.Context, apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Google Gemini API key is missing. Please set the GEMINI_API_KEY environment variable")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google genai client: %w", err)
	}
	log.Println("Successfully initialized Google Gemini client.")
	model := client.GenerativeModel("gemini-pro")
	return &geminiClient{model: model}, nil
}

// GenerateContent calls the Gemini API and adapts its response to our universal format.
func (c *geminiClient) GenerateContent(ctx context.Context, prompt string) (*Response, error) {
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini API call failed: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &Response{Text: "Received an empty response from the model."}, nil
	}
	return &Response{
		Text: string(resp.Candidates[0].Content.Parts[0].(genai.Text)),
	}, nil
}
