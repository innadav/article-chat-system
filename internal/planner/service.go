package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/prompts"
)

// plannerService is the concrete implementation of the Service interface.
// The struct name has been changed to avoid a name collision.
type plannerService struct {
	llmClient     llm.Client
	promptFactory *prompts.Factory
	articleSvc    article.Service
}

// NewService is the constructor. It returns the public interface type.
func NewService(llmClient llm.Client, promptFactory *prompts.Factory, articleSvc article.Service) Service {
	// It returns a pointer to the unexported struct, which satisfies the interface.
	return &plannerService{
		llmClient:     llmClient,
		promptFactory: promptFactory,
		articleSvc:    articleSvc,
	}
}

// CreatePlan's receiver is now the concrete struct pointer.
func (s *plannerService) CreatePlan(ctx context.Context, query string) (*QueryPlan, error) {
	articles := s.articleSvc.GetAllArticles(ctx)

	// You need to pass the models.Article slice to the prompt factory
	prompt, err := s.promptFactory.CreatePlannerPrompt(query, articles)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner prompt: %w", err)
	}

	resp, err := s.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("planner LLM call failed: %w", err)
	}

	var plan QueryPlan
	if err := json.Unmarshal([]byte(resp.Text), &plan); err != nil {
		log.Printf("Failed to parse JSON from planner, malformed text: %s", resp.Text)
		return nil, fmt.Errorf("failed to unmarshal plan from LLM response: %w", err)
	}

	log.Printf("Successfully created plan. Intent: %s, Targets: %v", plan.Intent, plan.Targets)
	return &plan, nil
}
