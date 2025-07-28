package agent

import (
	"context"

	"github.com/lamlv2305/sentinel/types"
)

// Adapter interface for receiving real-time updates
type Adapter interface {
	Connect(ctx context.Context, handler func(ctx context.Context, data types.Resource)) error
}
