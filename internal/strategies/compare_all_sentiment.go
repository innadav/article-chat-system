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

type CompareAllSentimentStrategy struct {
	BaseStrategy
}

func NewCompareAllSentimentStrategy() *CompareAllSentimentStrategy {
	s := &CompareAllSentimentStrategy{}
	s.doExecute = s.compareAllSentiment
	return s
}

func (s *CompareAllSentimentStrategy) compareAllSentiment(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("COMPARE ALL SENTIMENT STRATEGY: Executing...")

	if len(plan.Parameters) == 0 {
		return "Please specify the topic for sentiment comparison.", nil
	}
	topic := plan.Parameters[0]

	// 1. Find relevant articles using vector search
	relevantArticles, err := articleSvc.SearchSimilarArticles(ctx, topic, 5)
	if err != nil {
		return "", fmt.Errorf("failed to find articles for sentiment comparison: %w", err)
	}

	if len(relevantArticles) == 0 {
		return fmt.Sprintf("I could not find any articles discussing '%s' to compare sentiment.", topic), nil
	}

	// 2. Analyze sentiment for each article
	var sentimentResults []string
	for i, article := range relevantArticles {
		// Get sentiment from database if available
		if article.Sentiment != "" {
			sentimentResults = append(sentimentResults, fmt.Sprintf("Article %d: %s - Sentiment: %s", i+1, article.Title, article.Sentiment))
		} else {
			// Fallback: analyze sentiment using LLM
			prompt := fmt.Sprintf("Analyze the sentiment of this article about %s. Return only the sentiment (positive/negative/neutral) and a brief explanation:\n\nTitle: %s\nSummary: %s", topic, article.Title, article.Summary)
			sentiment, err := articleSvc.CallSynthesisLLM(ctx, prompt)
			if err != nil {
				sentimentResults = append(sentimentResults, fmt.Sprintf("Article %d: %s - Sentiment: Unable to analyze", i+1, article.Title))
			} else {
				sentimentResults = append(sentimentResults, fmt.Sprintf("Article %d: %s - Sentiment: %s", i+1, article.Title, sentiment))
			}
		}
	}

	// 3. Create comprehensive comparison
	result := fmt.Sprintf("Sentiment Analysis for Articles about '%s':\n\n", topic)
	for _, sentiment := range sentimentResults {
		result += sentiment + "\n"
	}

	result += fmt.Sprintf("\nFound %d articles discussing '%s'. ", len(relevantArticles), topic)
	if len(relevantArticles) > 1 {
		result += "Compare the sentiment patterns above to identify trends and differences."
	}

	return result, nil
}
