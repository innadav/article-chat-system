package strategies

import (
	"context"
	"fmt"
	"log"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type CompareToneStrategy struct {
	BaseStrategy
}

func NewCompareToneStrategy() *CompareToneStrategy {
	s := &CompareToneStrategy{}
	s.doExecute = s.compareToneArticles
	return s
}

func (s *CompareToneStrategy) compareToneArticles(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("COMPARE TONE STRATEGY: Performing tone comparison logic...")

	if len(plan.Targets) < 2 {
		return "Please provide at least two articles to compare tone.", nil
	}

	// Get articles from database
	var articles []string
	var summaries []string

	for i, target := range plan.Targets {
		if i >= 2 { // Limit to 2 articles for comparison
			break
		}

		art, ok := articleSvc.GetArticle(ctx, target)
		if !ok {
			return fmt.Sprintf("Article not found: %s", target), nil
		}

		articles = append(articles, fmt.Sprintf("Article %d: %s", i+1, art.Title))
		summaries = append(summaries, fmt.Sprintf("Article %d Summary: %s", i+1, art.Summary))
	}

	// Create comparison prompt
	comparisonContent := strings.Join(summaries, "\n\n")
	prompt := fmt.Sprintf("Compare the tone and writing style of these two articles:\n\n%s\n\nIdentify key differences in tone, formality, perspective, and overall approach. Provide specific examples from the summaries.", comparisonContent)

	// Call LLM for tone comparison
	result, err := articleSvc.CallSynthesisLLM(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate tone comparison: %w", err)
	}

	return result, nil
}
