package resagent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lamlv2305/sentinel/types"
)

var _ Adapter = &SSEAdapter{}

type SSEAdapter struct {
	endpoint   string
	maxRetries int // 0 means infinite retries
	retryDelay time.Duration
	httpClient *http.Client
	logger     *slog.Logger
	handler    func(ctx context.Context, data types.Resource)
}

func NewSSEAdapter(endpoint string, opts ...SSEAdapterOption) *SSEAdapter {
	adapter := &SSEAdapter{
		endpoint:   endpoint,
		maxRetries: 0, // 0 means infinite retries by default
		retryDelay: 5 * time.Second,
		httpClient: &http.Client{
			Timeout: 0, // No timeout for SSE connections
		},
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// Connect implements Adapter.
func (s *SSEAdapter) Connect(ctx context.Context, handler func(ctx context.Context, data types.Resource)) error {
	s.logger.Info("Starting SSE connection", "endpoint", s.endpoint)
	s.handler = handler

	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.retryDelay):
			}
		}

		if err := s.connectOnce(ctx); err != nil {
			if s.maxRetries > 0 && attempt >= s.maxRetries {
				s.logger.Error("Max retries exceeded", "max_retries", s.maxRetries, "error", err.Error())
				return fmt.Errorf("failed to connect after %d attempts: %w", s.maxRetries, err)
			}

			s.logger.Warn("Connection attempt failed, retrying...",
				"attempt", attempt+1,
				"error", err.Error(),
				"next_retry_in", s.retryDelay)
			continue
		}

		s.logger.Info("SSE connection established successfully")
		return nil
	}
}

// connectOnce performs a single connection attempt
func (s *SSEAdapter) connectOnce(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return s.processSSEStream(ctx, resp)
}

// processSSEStream processes the SSE stream
func (s *SSEAdapter) processSSEStream(ctx context.Context, resp *http.Response) error {
	s.logger.Debug("Starting to process SSE stream")
	scanner := bufio.NewScanner(resp.Body)
	var eventData strings.Builder
	var eventType string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			s.logger.Debug("SSE stream processing cancelled")
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// End of event - process accumulated data
		if line == "" {
			if eventData.Len() > 0 {
				if err := s.processEvent(ctx, eventType, eventData.String()); err != nil {
					s.logger.Error("Error processing event", "error", err.Error())
				}
				eventData.Reset()
				eventType = ""
			}
			continue
		}

		// Parse SSE line
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(data)
		} else if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		}
		// Ignore retry:, id:, and other unknown lines
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("Error reading SSE stream", "error", err.Error())
		return fmt.Errorf("error reading SSE stream: %w", err)
	}

	s.logger.Debug("SSE stream processing completed")
	return nil
}

// processEvent processes an SSE event and calls the handler
func (s *SSEAdapter) processEvent(ctx context.Context, eventType, data string) error {
	s.logger.Debug("Processing SSE event", "event_type", eventType, "data_length", len(data))

	// Handle changed events
	if eventType == "changed" {
		var event types.ChangedEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("failed to unmarshal changed event: %w", err)
		}
		s.handler(ctx, event.Resource)
		return nil
	}

	// Handle resource events (including default/empty event type)
	var resource types.Resource
	if err := json.Unmarshal([]byte(data), &resource); err != nil {
		return fmt.Errorf("failed to unmarshal resource event (type: %s): %w", eventType, err)
	}
	s.handler(ctx, resource)
	return nil
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

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) SSEAdapterOption {
	return func(s *SSEAdapter) {
		s.httpClient = client
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *slog.Logger) SSEAdapterOption {
	return func(s *SSEAdapter) {
		s.logger = logger
	}
}
