package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

// openaiClient is the concrete implementation for the OpenAI provider.
type openaiClient struct {
	client *openai.Client
	model  string
}

// newOpenAIClient creates a client for interacting with OpenAI.
func newOpenAIClient(ctx context.Context, apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is missing. Please set the OPENAI_API_KEY environment variable")
	}

	client := openai.NewClient(apiKey)

	// Test the connection by making a simple request
	_, err := client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to OpenAI API: %w", err)
	}

	log.Println("Successfully initialized OpenAI client.")
	return &openaiClient{
		client: client,
		model:  "gpt-3.5-turbo", // Default model
	}, nil
}

// GenerateContent calls the OpenAI API and adapts its response to our universal format.
func (c *openaiClient) GenerateContent(ctx context.Context, prompt string) (*Response, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("openai API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return &Response{Text: "Received an empty response from the model."}, nil
	}

	return &Response{
		Text: resp.Choices[0].Message.Content,
	}, nil
}
