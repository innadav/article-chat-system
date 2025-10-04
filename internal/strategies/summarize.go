package strategies

import (
	"context"
	"fmt"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
)

type SummarizeStrategy struct {
	BaseStrategy
}

func NewSummarizeStrategy() *SummarizeStrategy {
	s := &SummarizeStrategy{}
	s.doExecute = s.summarizeArticle
	return s
}

func (s *SummarizeStrategy) summarizeArticle(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("SUMMARIZE STRATEGY: Retrieving cached summary...")
	if len(plan.Targets) == 0 {
		return "Please specify which article you want to summarize.", nil
	}

	targetURL := plan.Targets[0]

	// Get the article from the database (summary should already be cached from initial analysis)
	art, ok := articleSvc.GetArticle(ctx, targetURL)
	if !ok {
		return "", fmt.Errorf("article not found in database: %s", targetURL)
	}

	if art.Summary == "" {
		return "", fmt.Errorf("summary not available for article: %s", art.Title)
	}

	return art.Summary, nil
}
