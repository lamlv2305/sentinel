package agent

import (
	"context"
	"log/slog"

	"github.com/lamlv2305/sentinel/types"
)

// Agent represents the main agent for interacting with resgate
type Agent struct {
	*Options
}

// New creates a new Resagent instance
func New(opts ...Option) *Agent {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	ra := &Agent{
		Options: options,
	}

	if ra.adapter == nil {
		slog.Warn("Adapter is not set")
	}

	if ra.persister == nil {
		slog.Warn("Persister is not set")
	}

	return ra
}

// Start begins the resagent operations
func (ra *Agent) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	defer close(errCh)
	defer close(ra.subscriber)

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
func (ra *Agent) handleDataChange(ctx context.Context, data types.Resource) {
	// Update cache
	ra.persister.Save(ctx, data)

	select {
	case ra.subscriber <- data:
	default:
		// If the subscriber channel is full, we skip sending the update
	}
}
