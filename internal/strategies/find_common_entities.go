package strategies

import (
	"context"
	"fmt"
	"log"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/repository"
	"article-chat-system/internal/vector"
)

type FindCommonEntitiesStrategy struct {
	BaseStrategy
}

func NewFindCommonEntitiesStrategy() *FindCommonEntitiesStrategy {
	s := &FindCommonEntitiesStrategy{}
	s.doExecute = s.findCommonEntities
	return s
}

func (s *FindCommonEntitiesStrategy) findCommonEntities(ctx context.Context, plan *planner.QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error) {
	log.Println("FIND COMMON ENTITIES STRATEGY: Performing entity extraction logic...")

	// Get common entities from database using efficient PostgreSQL query
	// If no targets specified, get entities from all articles
	var entityCounts []repository.EntityCount
	var err error

	if len(plan.Targets) == 0 {
		entityCounts, err = articleSvc.FindCommonEntities(ctx, []string{})
	} else {
		entityCounts, err = articleSvc.FindCommonEntities(ctx, plan.Targets)
	}

	if err != nil {
		return "", fmt.Errorf("failed to find common entities: %w", err)
	}

	if len(entityCounts) == 0 {
		return "No entities found in the database.", nil
	}

	// Return the raw entity counts as JSON-like string
	result := fmt.Sprintf("Found %d entities:\n", len(entityCounts))
	for _, entity := range entityCounts {
		result += fmt.Sprintf("- %s: %d occurrences\n", entity.Entity, entity.Count)
	}

	return result, nil
}
