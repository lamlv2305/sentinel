package resagent

import "context"

// Cache interface for storing and retrieving resources
type Cache interface {
	Get(rid string) (any, bool)
	Set(rid string, data any)
	Delete(rid string)
	List() []any
	Clear()
}

// Client interface for fetching resources from resgate
type Client interface {
	Get(rid string) (any, error)
}

// Adapter interface for receiving real-time updates
type Adapter interface {
	Connect(ctx context.Context, handler func(rid string, data any)) error
}
