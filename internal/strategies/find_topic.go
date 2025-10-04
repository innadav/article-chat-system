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

type FindTopicStrategy struct {
	BaseStrategy
}

func NewFindTopicStrategy() *FindTopicStrategy {
	s := &FindTopicStrategy{}
	s.doExecute = s.findTopicArticles
	return s
}

func (s *FindTopicStrategy) findTopicArticles(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("FIND TOPIC STRATEGY: Performing vector search and synthesis...")

	if len(plan.Parameters) == 0 {
		return "Please specify a topic to search for.", nil
	}
	topic := plan.Parameters[0]

	// 1. Perform Semantic Search (Vector Search)
	// Instead of getting all articles, we ask the article service (which uses Weaviate)
	// to find the top 3 most relevant articles for the topic.
	relevantArticles, err := articleSvc.SearchSimilarArticles(ctx, topic, 3)
	if err != nil {
		return "", fmt.Errorf("vector search failed: %w", err)
	}

	// 2. Handle No Results
	if len(relevantArticles) == 0 {
		return fmt.Sprintf("I could not find any articles discussing '%s'.", topic), nil
	}

	// 3. Craft a Synthesis Prompt for the LLM
	// We now have a small, highly relevant list of articles. We'll ask the LLM
	// to create a final answer based on their content.
	prompt, err := promptFactory.CreateFindTopicPrompt(topic, relevantArticles)
	if err != nil {
		return "", err
	}

	// 4. Call the LLM for the Final Answer
	// The LLM will generate a natural language response explaining what it found.
	return articleSvc.CallSynthesisLLM(ctx, prompt)
}
