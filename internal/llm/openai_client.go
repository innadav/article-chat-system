package llm

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// openaiClient is the concrete implementation for the OpenAI provider.
type openaiClient struct {
	client *openai.Client
	model  string
}

var tracer = otel.Tracer("llm-openai-client") // Create a tracer for this package

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
	// 1. Start a new span. The 'ctx' carries the parent span's context.
	ctx, span := tracer.Start(ctx, "LLM.GenerateContent")
	defer span.End() // Ensure the span is ended when the function returns.

	// 2. Add the prompt to the span as an attribute.
	span.SetAttributes(
		attribute.String("llm.provider", "openai"),
		attribute.String("llm.model", c.model),
		attribute.String("llm.prompt", prompt),
	)

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
		span.RecordError(err) // Record any errors that occur.
		return nil, fmt.Errorf("openai API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		responseText := "Received an empty response from the model."
		span.SetAttributes(attribute.String("llm.response", responseText))
		return &Response{Text: responseText}, nil
	}

	responseText := resp.Choices[0].Message.Content
	// 3. Add the response to the span as another attribute.
	span.SetAttributes(attribute.String("llm.response", responseText))

	// 4. Add token usage information to the trace and log
	// Log usage information for debugging
	slog.Info("LLM API response received",
		"provider", "openai",
		"model", c.model,
		"usage_total_tokens", resp.Usage.TotalTokens,
		"usage_prompt_tokens", resp.Usage.PromptTokens,
		"usage_completion_tokens", resp.Usage.CompletionTokens,
	)

	if resp.Usage.TotalTokens > 0 {
		span.SetAttributes(
			attribute.Int("llm.usage.prompt_tokens", resp.Usage.PromptTokens),
			attribute.Int("llm.usage.completion_tokens", resp.Usage.CompletionTokens),
			attribute.Int("llm.usage.total_tokens", resp.Usage.TotalTokens),
		)
	}

	return &Response{
		Text: responseText,
	}, nil
}
