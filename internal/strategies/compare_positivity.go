package strategies

import (
	"context"
	"fmt"
	"log"
	"strings"

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
	s.doExecute = s.comparePositivityArticles
	return s
}

func (s *ComparePositivityStrategy) comparePositivityArticles(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("COMPARE POSITIVITY STRATEGY: Performing positivity comparison logic...")
	log.Printf("COMPARE POSITIVITY: Plan targets: %v", plan.Targets)
	log.Printf("COMPARE POSITIVITY: Plan parameters: %v", plan.Parameters)

	var articles []*models.Article
	var err error

	// If we have parameters, search by topics in vector DB
	if len(plan.Parameters) > 0 {
		log.Printf("COMPARE POSITIVITY: Searching by parameters: %v", plan.Parameters)
		articles, err = vectorSvc.SearchByTopics(ctx, plan.Parameters, 2)
		if err != nil {
			return "", fmt.Errorf("failed to search articles by topics: %w", err)
		}
		log.Printf("COMPARE POSITIVITY: Found %d articles by topic search", len(articles))
	} else if len(plan.Targets) >= 2 {
		// Fallback to target-based search if no parameters
		log.Printf("COMPARE POSITIVITY: Using target-based search")
		for i, target := range plan.Targets {
			if i >= 2 { // Limit to 2 articles for comparison
				break
			}

			log.Printf("COMPARE POSITIVITY: Looking for article with target: %s", target)
			art, ok := articleSvc.GetArticle(ctx, target)
			if !ok {
				log.Printf("COMPARE POSITIVITY: Article not found for target: %s", target)
				return fmt.Sprintf("Article not found: %s", target), nil
			}
			articles = append(articles, art)
		}
	} else {
		return "Please provide topics/parameters to search for articles to compare, or specify at least two article URLs.", nil
	}

	if len(articles) < 2 {
		return fmt.Sprintf("Found only %d article(s) matching the criteria. Need at least 2 articles to compare positivity.", len(articles)), nil
	}

	// Prepare articles for comparison
	var summaries []string
	for i, art := range articles[:2] { // Limit to first 2 articles
		log.Printf("COMPARE POSITIVITY: Found article: %s, Summary length: %d", art.Title, len(art.Summary))
		log.Printf("COMPARE POSITIVITY: Summary preview: %.100s...", art.Summary)

		summaries = append(summaries, fmt.Sprintf("Article %d Summary: %s", i+1, art.Summary))
	}

	// Create comparison prompt focused on positivity
	comparisonContent := strings.Join(summaries, "\n\n")
	prompt := fmt.Sprintf("Compare the positivity and sentiment of these two articles:\n\n%s\n\nAnalyze which article is more positive, optimistic, or favorable in its tone and content. Consider the language used, overall sentiment, and perspective. Provide specific examples and explain your reasoning.", comparisonContent)

	// Call LLM for positivity comparison
	result, err := articleSvc.CallSynthesisLLM(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate positivity comparison: %w", err)
	}

	return result, nil
}
