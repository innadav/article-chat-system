package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
)

// plannerService is the concrete implementation of the Service interface.
// The struct name has been changed to avoid a name collision.
type plannerService struct {
	llmClient     llm.Client
	promptFactory *prompts.Factory
	articleSvc    article.Service
	vecRepo       *repository.VectorRepository
}

// NewService is the constructor. It returns the public interface type.
func NewService(llmClient llm.Client, promptFactory *prompts.Factory, articleSvc article.Service, vecRepo *repository.VectorRepository) Service {
	// It returns a pointer to the unexported struct, which satisfies the interface.
	return &plannerService{
		llmClient:     llmClient,
		promptFactory: promptFactory,
		articleSvc:    articleSvc,
		vecRepo:       vecRepo,
	}
}

// CreatePlan's receiver is now the concrete struct pointer.
func (s *plannerService) CreatePlan(ctx context.Context, query string) (*QueryPlan, error) {
	// 1. Find the top 5 most relevant articles using vector search.
	relevantArticles, err := s.articleSvc.SearchSimilarArticles(ctx, query, 5)
	if err != nil {
		log.Printf("WARNING: Vector search failed, falling back to all articles: %v", err)
		// Fallback to using all articles if vector search fails
		relevantArticles = s.articleSvc.GetAllArticles(ctx)
	}

	// 2. Build the prompt using ONLY the relevant articles as context.
	prompt, err := s.promptFactory.CreatePlannerPrompt(query, relevantArticles)
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
