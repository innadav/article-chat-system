package planner

import (
	"context"
)

// Service defines the contract for the planner.
type Service interface {
	CreatePlan(ctx context.Context, query string) (*QueryPlan, error)
}
