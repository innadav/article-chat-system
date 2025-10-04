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

type ComparePositivityStrategy struct {
	BaseStrategy
}

func NewComparePositivityStrategy() *ComparePositivityStrategy {
	s := &ComparePositivityStrategy{}
	s.doExecute = s.comparePositivity
	return s
}

func (s *ComparePositivityStrategy) comparePositivity(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("COMPARE POSITIVITY STRATEGY: Executing...")
	if len(plan.Parameters) == 0 {
		return "Please specify the topic for comparison.", nil
	}
	topic := plan.Parameters[0]

	// --- LOGIC FOR THE NEW "FIND AND COMPARE" WORKFLOW ---
	if len(plan.Targets) == 0 {
		log.Println("No targets specified. Finding relevant articles first...")

		// 1. Find candidate articles using vector search.
		candidateArticles, err := articleSvc.SearchSimilarArticles(ctx, topic, 3) // Find top 3
		if err != nil {
			return "", fmt.Errorf("failed to find articles for comparison: %w", err)
		}
		if len(candidateArticles) < 2 {
			return "I could not find enough relevant articles to perform a comparison.", nil
		}

		// 2. Create the advanced prompt.
		prompt, err := promptFactory.CreateComparePositivityPrompt(topic, candidateArticles)
		if err != nil {
			return "", err
		}

		// 3. Call the LLM for the final analysis.
		return articleSvc.CallSynthesisLLM(ctx, prompt)
	}

	// --- ORIGINAL LOGIC FOR COMPARING TWO SPECIFIC ARTICLES ---
	if len(plan.Targets) < 2 {
		return "Please specify at least two articles to compare.", nil
	}
	art1, ok1 := articleSvc.GetArticle(ctx, plan.Targets[0])
	art2, ok2 := articleSvc.GetArticle(ctx, plan.Targets[1])
	if !ok1 || !ok2 {
		return "Could not find one or both of the specified articles.", nil
	}

	// For a simple 2-article comparison, we reuse the advanced prompt.
	prompt, err := promptFactory.CreateComparePositivityPrompt(topic, []*models.Article{art1, art2})
	if err != nil {
		return "", err
	}
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
