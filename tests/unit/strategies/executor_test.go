package strategies_test

import (
	"context"
	"errors"
	"testing"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/strategies"
	"article-chat-system/internal/vector"
)

// mockStrategy is a mock implementation of the IntentStrategy interface.
type mockStrategy struct {
	ExecuteFunc func(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error)
}

func (m *mockStrategy) Execute(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, plan, articleSvc, promptFactory, vectorSvc)
	}
	return "", errors.New("ExecuteFunc not implemented")
}

func TestExecutor_ExecutePlan(t *testing.T) {
	// ARRANGE
	mockSummarizeStrategy := &mockStrategy{}
	wasCalled := false

	mockSummarizeStrategy.ExecuteFunc = func(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
		wasCalled = true
		return "Mocked summary response", nil
	}

	executor := &strategies.Executor{
		Strategies: map[planner.QueryIntent]planner.IntentStrategy{
			planner.IntentSummarize: mockSummarizeStrategy,
		},
	}

	plan := &planner.QueryPlan{
		Intent: planner.IntentSummarize,
	}

	// ACT
	response, err := executor.ExecutePlan(context.Background(), plan, nil, nil, nil)

	// ASSERT
	if err != nil {
		t.Fatalf("ExecutePlan() returned an unexpected error: %v", err)
	}
	if !wasCalled {
		t.Error("Expected the SummarizeStrategy's Execute method to be called, but it was not.")
	}
	if response != "Mocked summary response" {
		t.Errorf("Expected response 'Mocked summary response', but got '%s'", response)
	}
}
