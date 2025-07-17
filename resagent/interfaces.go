package resagent

import (
	"context"

	"github.com/lamlv2305/sentinel/types"
)

// Cache interface for storing and retrieving resources
type Cache interface {
	Get(ctx context.Context, rid string) *types.Resource
	Set(ctx context.Context, data types.Resource) error
	Delete(ctx context.Context, rid string)
	List(ctx context.Context) []types.Resource
}

// Adapter interface for receiving real-time updates
type Adapter interface {
	Connect(ctx context.Context, handler func(ctx context.Context, data types.Resource)) error
}
