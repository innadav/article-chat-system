package strategies

import (
	"context"
	"fmt"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
)

// Executor holds the map of all available strategies.
type Executor struct {
	Strategies map[planner.QueryIntent]planner.IntentStrategy
}

// NewExecutor creates and initializes the map of all strategies.
// This function acts as a central registry for your system's capabilities.
func NewExecutor() *Executor {
	return &Executor{
		Strategies: map[planner.QueryIntent]planner.IntentStrategy{
			planner.IntentSummarize:           NewSummarizeStrategy(),
			planner.IntentKeywords:            NewKeywordsStrategy(),
			planner.IntentSentiment:           NewSentimentStrategy(),
			planner.IntentCompareTone:         NewCompareToneStrategy(),
			planner.IntentFindTopic:           NewFindTopicStrategy(),
			planner.IntentComparePositive:     NewComparePositivityStrategy(),
			planner.IntentFindCommonEntities:  NewFindCommonEntitiesStrategy(),
			planner.IntentCompareAllSentiment: NewCompareAllSentimentStrategy(),
		},
	}
}

// ExecutePlan finds the correct strategy for the plan's intent and executes it.
// This method acts as a smart dispatcher, delegating the work.
func (e *Executor) ExecutePlan(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	strategy, ok := e.Strategies[plan.Intent]
	if !ok {
		// Fallback for any intent that isn't registered.
		return fmt.Sprintf("I'm sorry, I don't know how to handle the intent: %s", plan.Intent), nil
	}

	// The call to strategy.Execute will trigger the BaseStrategy's template method.
	return strategy.Execute(ctx, plan, articleSvc, promptFactory, vectorSvc)
}
