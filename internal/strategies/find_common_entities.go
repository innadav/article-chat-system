package strategies

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
)

type FindCommonEntitiesStrategy struct {
	BaseStrategy
}

func NewFindCommonEntitiesStrategy() *FindCommonEntitiesStrategy {
	s := &FindCommonEntitiesStrategy{}
	s.doExecute = s.findCommonEntities
	return s
}

func (s *FindCommonEntitiesStrategy) findCommonEntities(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory) (string, error) {
	log.Println("FIND COMMON ENTITIES STRATEGY: Performing entity extraction logic...")

	// Get all articles from database
	allArticles := articleSvc.GetAllArticles(ctx)
	if len(allArticles) == 0 {
		return "No articles found in the database.", nil
	}

	// Collect all topics from all articles
	entityCounts := make(map[string]int)
	var allSummaries []string

	for i, art := range allArticles {
		// Count topics as entities
		for _, topic := range art.Topics {
			if topic != "" {
				entityCounts[topic]++
			}
		}

		// Also collect summaries for LLM analysis
		if art.Summary != "" {
			allSummaries = append(allSummaries, fmt.Sprintf("Article %d: %s\nSummary: %s", i+1, art.Title, art.Summary))
		}
	}

	// Sort entities by frequency
	type entityCount struct {
		entity string
		count  int
	}

	var sortedEntities []entityCount
	for entity, count := range entityCounts {
		sortedEntities = append(sortedEntities, entityCount{entity, count})
	}

	sort.Slice(sortedEntities, func(i, j int) bool {
		return sortedEntities[i].count > sortedEntities[j].count
	})

	// Create response
	result := fmt.Sprintf("Most commonly discussed entities across %d articles:\n\n", len(allArticles))

	// Show top 10 entities
	topCount := 10
	if len(sortedEntities) < topCount {
		topCount = len(sortedEntities)
	}

	for i := 0; i < topCount; i++ {
		result += fmt.Sprintf("%d. %s (mentioned in %d articles)\n", i+1, sortedEntities[i].entity, sortedEntities[i].count)
	}

	// If we have summaries, use LLM to extract additional entities
	if len(allSummaries) > 0 {
		summariesText := strings.Join(allSummaries, "\n\n")
		prompt := fmt.Sprintf("Analyze these article summaries and identify the most commonly discussed entities (people, companies, technologies, concepts, etc.):\n\n%s\n\nProvide a list of the top 10 most frequently mentioned entities with brief descriptions.", summariesText)

		llmResult, err := articleSvc.CallSynthesisLLM(ctx, prompt)
		if err != nil {
			log.Printf("Failed to get LLM entity analysis: %v", err)
		} else {
			result += "\n\nLLM Analysis:\n" + llmResult
		}
	}

	return result, nil
}
