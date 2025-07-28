package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/lamlv2305/sentinel/types"
	"github.com/r3labs/sse/v2"
)

var _ Adapter = &SSEAdapter{}

// SSEEvent represents a parsed server-sent event
type SSEEvent struct {
	Type string
	Data string
}

type SSEAdapter struct {
	endpoint   string
	maxRetries int // 0 means infinite retries
	retryDelay time.Duration
	client     *sse.Client
	logger     *slog.Logger
}

func NewSSEAdapter(endpoint string, opts ...SSEAdapterOption) *SSEAdapter {
	adapter := &SSEAdapter{
		endpoint:   endpoint,
		maxRetries: 0, // 0 means infinite retries by default
		retryDelay: 5 * time.Second,
		client:     sse.NewClient(endpoint),
		logger:     slog.Default(),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// Connect implements Adapter.
func (s *SSEAdapter) Connect(ctx context.Context, handler func(ctx context.Context, data types.Resource)) error {
	return s.client.SubscribeRaw(func(msg *sse.Event) {
		bytes, err := base64.StdEncoding.DecodeString(string(msg.Data))
		if err != nil {
			return
		}

		var ce types.ChangedEvent
		if err := json.Unmarshal(bytes, &ce); err != nil {
			s.logger.Error("Failed to unmarshal SSE event", "error", err)
			return
		}

		s.logger.Debug("Received SSE event",
			"resource", ce.Resource,
			"data", string(ce.Resource.Data))
	})
}

// SSEAdapterOption configures the SSE adapter
type SSEAdapterOption func(*SSEAdapter)

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(maxRetries int) SSEAdapterOption {
	return func(s *SSEAdapter) {
		s.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the delay between retry attempts
func WithRetryDelay(delay time.Duration) SSEAdapterOption {
	return func(s *SSEAdapter) {
		s.retryDelay = delay
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *slog.Logger) SSEAdapterOption {
	return func(s *SSEAdapter) {
		s.logger = logger
	}
}
