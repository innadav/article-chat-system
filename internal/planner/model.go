package planner

import (
	"context"

	"article-chat-system/internal/article"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/vector"
)

// QueryIntent represents the user's goal.
type QueryIntent string

const (
	IntentSummarize           QueryIntent = "SUMMARIZE"
	IntentKeywords            QueryIntent = "KEYWORDS"
	IntentSentiment           QueryIntent = "SENTIMENT"
	IntentCompareTone         QueryIntent = "COMPARE_TONE"
	IntentFindTopic           QueryIntent = "FIND_BY_TOPIC"
	IntentComparePositive     QueryIntent = "COMPARE_POSITIVITY"
	IntentFindCommonEntities  QueryIntent = "FIND_COMMON_ENTITIES"
	IntentCompareAllSentiment QueryIntent = "COMPARE_ALL_SENTIMENT"
	IntentCompareMultiple     QueryIntent = "COMPARE_MULTIPLE"
	IntentUnknown             QueryIntent = "UNKNOWN"
)

// QueryPlan is the structured representation of a user's request.
type QueryPlan struct {
	Intent     QueryIntent `json:"intent"`
	Targets    []string    `json:"targets"`
	Parameters []string    `json:"parameters"`
	Question   string      `json:"question"`
}

// IntentStrategy defines the interface for executing a query based on its intent.
type IntentStrategy interface {
	Execute(ctx context.Context, plan *QueryPlan, articleSvc article.Service, promptFactory *prompts.Factory, vectorSvc vector.Service) (string, error)
}
