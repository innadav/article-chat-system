package strategies

import (
	"context"
	"fmt"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/models"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
)

type CompareMultipleStrategy struct {
	BaseStrategy
}

func NewCompareMultipleStrategy() *CompareMultipleStrategy {
	s := &CompareMultipleStrategy{}
	s.doExecute = s.compareMultipleArticles
	return s
}

func (s *CompareMultipleStrategy) compareMultipleArticles(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("COMPARE MULTIPLE STRATEGY: Executing...")
	if len(plan.Targets) < 2 {
		return "Please specify at least two articles to compare.", nil
	}

	// 1. Fetch all target articles from the database.
	var articlesToCompare []*models.Article
	for _, url := range plan.Targets {
		art, ok := articleSvc.GetArticle(ctx, url)
		if !ok {
			return "", fmt.Errorf("could not find article with URL: %s", url)
		}
		articlesToCompare = append(articlesToCompare, art)
	}

	// 2. Create the advanced comparison prompt.
	prompt, err := promptFactory.CreateCompareMultiplePrompt(articlesToCompare)
	if err != nil {
		return "", err
	}

	// 3. Call the LLM for the final, synthesized analysis.
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
