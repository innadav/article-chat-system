package strategies

import (
	"context"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"

	"github.com/go-shiori/go-readability"
)

type KeywordsStrategy struct {
	BaseStrategy
}

func NewKeywordsStrategy() *KeywordsStrategy {
	s := &KeywordsStrategy{}
	s.doExecute = s.extractKeywords
	return s
}

func (s *KeywordsStrategy) extractKeywords(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("KEYWORDS STRATEGY: Performing specific keyword extraction logic...")
	if len(plan.Targets) == 0 {
		return "Please specify which article you want to extract keywords from.", nil
	}
	art, ok := articleSvc.GetArticle(ctx, plan.Targets[0])
	if !ok {
		return "I couldn't find the requested article.", nil
	}

	// Fetch content on-demand for keyword extraction
	articleData, err := readability.FromURL(art.URL, 30)
	if err != nil {
		return "", err
	}

	prompt, err := promptFactory.CreateKeywordsPrompt(art.Title, articleData.TextContent)
	if err != nil {
		return "", err
	}

	keywords, err := articleSvc.CallSynthesisLLM(ctx, prompt)
	if err != nil {
		return "", err
	}

	return keywords, nil
}
