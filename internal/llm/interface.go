package llm

import "context"

// Response is a standardized struct for any LLM's output.
type Response struct {
	Text string
}

// Client is a universal interface for any generative AI model.
type Client interface {
	GenerateContent(ctx context.Context, prompt string) (*Response, error)
}
