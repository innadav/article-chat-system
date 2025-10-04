package strategies

import (
	"context"
	"fmt"
	"log"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
)

type FindTopicStrategy struct {
	BaseStrategy
}

func NewFindTopicStrategy() *FindTopicStrategy {
	s := &FindTopicStrategy{}
	s.doExecute = s.findTopicArticles
	return s
}

func (s *FindTopicStrategy) findTopicArticles(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("FIND TOPIC STRATEGY: Performing topic search logic...")

	if len(plan.Parameters) == 0 {
		return "Please specify a topic to search for.", nil
	}

	topic := strings.Join(plan.Parameters, " ")

	// Get all articles from database
	allArticles := articleSvc.GetAllArticles(ctx)
	if len(allArticles) == 0 {
		return "No articles found in the database.", nil
	}

	// Filter articles that match the topic
	var matchingArticles []string
	var articleSummaries []string

	for _, art := range allArticles {
		// Check if article topics contain the search term or if summary mentions it
		articleText := strings.ToLower(art.Title + " " + art.Summary + " " + strings.Join(art.Topics, " "))
		searchTerm := strings.ToLower(topic)

		if strings.Contains(articleText, searchTerm) {
			matchingArticles = append(matchingArticles, fmt.Sprintf("%d. %s", len(matchingArticles)+1, art.Title))
			articleSummaries = append(articleSummaries, fmt.Sprintf("Article %d: %s\nSummary: %s", len(articleSummaries)+1, art.Title, art.Summary))
		}
	}

	if len(matchingArticles) == 0 {
		return fmt.Sprintf("No articles found discussing '%s'.", topic), nil
	}

	// Create a comprehensive response
	result := fmt.Sprintf("Found %d articles discussing '%s':\n\n", len(matchingArticles), topic)
	result += strings.Join(matchingArticles, "\n")

	// Add detailed analysis if there are matching articles
	if len(articleSummaries) > 0 {
		result += "\n\nDetailed Analysis:\n"
		result += strings.Join(articleSummaries, "\n\n")
	}

	return result, nil
}
