package strategies

import (
	"context"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
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

	// Use existing topics and entities from the database instead of fetching content
	if len(art.Topics) > 0 || len(art.Entities) > 0 {
		result := "KEYWORDS extracted from the article:\n\n"

		if len(art.Topics) > 0 {
			result += "Main Topics:\n"
			for _, topic := range art.Topics {
				result += "- " + topic + "\n"
			}
		}

		if len(art.Entities) > 0 {
			result += "\nKey Entities:\n"
			for _, entity := range art.Entities {
				result += "- " + entity + "\n"
			}
		}

		return result, nil
	}

	// Fallback: if no topics/entities available, try to extract from summary
	prompt, err := promptFactory.CreateKeywordsPrompt(art.Title, art.Summary)
	if err != nil {
		return "", err
	}

	keywords, err := articleSvc.CallSynthesisLLM(ctx, prompt)
	if err != nil {
		return "", err
	}

	return keywords, nil
}
