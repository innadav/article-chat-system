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

type SentimentStrategy struct {
	BaseStrategy
}

func NewSentimentStrategy() *SentimentStrategy {
	s := &SentimentStrategy{}
	s.doExecute = s.analyzeSentiment
	return s
}

func (s *SentimentStrategy) analyzeSentiment(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("SENTIMENT STRATEGY: Performing specific sentiment analysis logic...")
	if len(plan.Targets) == 0 {
		return "Please specify which article you want to analyze sentiment for.", nil
	}

	var results []string
	var foundArticles int

	for i, target := range plan.Targets {
		art, ok := articleSvc.GetArticle(ctx, target)
		if !ok {
			results = append(results, fmt.Sprintf("Article %d: Could not find article at %s", i+1, target))
			continue
		}

		foundArticles++
		if art.Sentiment != "" {
			sentimentDesc := s.getSentimentDescription(art.Sentiment)
			log.Printf("DEBUG: Original sentiment: '%s', Parsed description: '%s'", art.Sentiment, sentimentDesc)
			results = append(results, fmt.Sprintf("Article %d: '%s' - Sentiment: %s (%s)", i+1, art.Title, art.Sentiment, sentimentDesc))
		} else {
			results = append(results, fmt.Sprintf("Article %d: '%s' - Sentiment analysis not available", i+1, art.Title))
		}
	}

	if foundArticles == 0 {
		return "No articles found for sentiment analysis.", nil
	}

	return fmt.Sprintf("Sentiment Analysis Results:\n\n%s", strings.Join(results, "\n")), nil
}

// getSentimentDescription converts numeric sentiment to descriptive text
func (s *SentimentStrategy) getSentimentDescription(sentiment string) string {
	if sentiment == "" {
		return "Unknown"
	}

	// Handle format like "0.40 (positive)" - extract the numeric score
	if strings.Contains(sentiment, "(") && strings.Contains(sentiment, ")") {
		// Extract the numeric part before the parenthesis
		parts := strings.Split(sentiment, " (")
		if len(parts) >= 1 {
			scoreStr := strings.TrimSpace(parts[0])
			return s.getNumericSentimentDescription(scoreStr)
		}
	}

	// Handle common sentiment descriptions
	switch strings.ToLower(sentiment) {
	case "positive", "pos":
		return "Positive"
	case "negative", "neg":
		return "Negative"
	case "neutral", "neu":
		return "Neutral"
	default:
		// Try to parse as numeric
		return s.getNumericSentimentDescription(sentiment)
	}
}

// getNumericSentimentDescription converts numeric sentiment score to descriptive text
func (s *SentimentStrategy) getNumericSentimentDescription(scoreStr string) string {
	switch scoreStr {
	case "0.7", "0.8", "0.9", "1.0":
		return "Very Positive"
	case "0.5", "0.6":
		return "Somewhat Positive"
	case "0.1", "0.2", "0.3", "0.4":
		return "Slightly Positive"
	case "0.0":
		return "Neutral"
	case "-0.1", "-0.2":
		return "Slightly Negative"
	case "-0.3", "-0.4":
		return "Somewhat Negative"
	case "-0.5", "-0.6":
		return "Somewhat Negative"
	case "-0.7", "-0.8", "-0.9", "-1.0":
		return "Very Negative"
	default:
		// If it's not a recognized value, return as-is
		return scoreStr
	}
}
