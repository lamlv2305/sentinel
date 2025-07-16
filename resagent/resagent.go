package resagent

import (
	"context"
)

// ResAgent represents the main agent for interacting with resgate
type ResAgent struct {
	*Options
}

// New creates a new Resagent instance
func New(opts ...Option) *ResAgent {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	ra := &ResAgent{
		Options: options,
	}

	return ra
}

// Start begins the resagent operations
func (ra *ResAgent) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- ra.adapter.Connect(ctx, ra.handleDataChange)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-errCh:
			return err
		}
	}
}

// handleDataChange processes incoming data changes from resgate
func (ra *ResAgent) handleDataChange(rid string, data any) {
	// Update cache
	ra.cache.Set(rid, data)
}
