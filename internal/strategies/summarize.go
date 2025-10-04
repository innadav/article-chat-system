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

	"github.com/go-shiori/go-readability"
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
	log.Println("SUMMARIZE STRATEGY: Performing specific summarization logic...")
	if len(plan.Targets) == 0 {
		return "Please specify which article you want to summarize.", nil
	}

	targetURL := plan.Targets[0]

	// First try to get the article from the database
	art, ok := articleSvc.GetArticle(ctx, targetURL)
	if !ok {
		// If not found in database, create a temporary article object for on-demand processing
		log.Printf("Article not found in database, processing URL on-demand: %s", targetURL)
		art = &models.Article{
			URL: targetURL,
		}
	}

	// If summary already exists, return it
	if art.Summary != "" {
		return art.Summary, nil
	}

	// Fetch content on-demand for summarization
	articleData, err := readability.FromURL(art.URL, 30)
	if err != nil {
		return "", fmt.Errorf("failed to fetch article content: %w", err)
	}

	// Update article with fetched content
	art.Title = articleData.Title
	art.Excerpt = articleData.Excerpt

	prompt, err := promptFactory.CreateSummarizePrompt(articleData.TextContent)
	if err != nil {
		return "", err
	}
	summary, err := articleSvc.CallSynthesisLLM(ctx, prompt)
	if err != nil {
		return "", err
	}

	// Store the generated summary and article if it wasn't in the database
	art.Summary = summary
	if !ok {
		// Only store if it wasn't already in the database
		_ = articleSvc.StoreArticle(ctx, art)
		log.Printf("Stored new article in database: %s", art.Title)
	} else {
		// Update existing article with summary
		_ = articleSvc.StoreArticle(ctx, art)
	}

	return art.Summary, nil
}
