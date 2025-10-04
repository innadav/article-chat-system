package strategies

import (
	"context"
	"fmt"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
)

type strategyStep func(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error)

type BaseStrategy struct {
	doExecute strategyStep
}

func (s *BaseStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	if err := s.validateInput(plan); err != nil {
		return err.Error(), nil
	}
	result, err := s.doExecute(ctx, plan, articleSvc, promptFactory, vectorSvc)
	if err != nil {
		return "", fmt.Errorf("error during specific execution: %w", err)
	}
	return s.formatResponse(result), nil
}

func (s *BaseStrategy) validateInput(plan *planner.QueryPlan) error {
	if plan == nil || plan.Intent == "" {
		return fmt.Errorf("invalid plan provided")
	}
	return nil
}

func (s *BaseStrategy) formatResponse(response string) string {
	return fmt.Sprintf("🤖 Here is your answer:\n\n%s", response)
}
