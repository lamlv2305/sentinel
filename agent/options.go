package agent

import (
	"time"

	"github.com/lamlv2305/sentinel/persister"
	"github.com/lamlv2305/sentinel/types"
)

// Options holds configuration for Resagent
type Options struct {
	timeout        time.Duration
	reconnectDelay time.Duration
	persister      persister.Persister[types.Resource]
	adapter        Adapter
	subscriber     chan types.Resource // Channel for receiving updates
}

// Option is a function that configures Options
type Option func(*Options)

// defaultOptions returns default configuration
func defaultOptions() *Options {
	return &Options{
		timeout:        30 * time.Second,
		reconnectDelay: 5 * time.Second,
		persister:      nil,                          // Will be set later
		adapter:        nil,                          // Will be set later
		subscriber:     make(chan types.Resource, 1), // Buffered channel for updates
	}
}

// WithTimeout sets the HTTP timeout
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.timeout = timeout
	}
}

// WithReconnectDelay sets the SSE reconnect delay
func WithReconnectDelay(delay time.Duration) Option {
	return func(o *Options) {
		o.reconnectDelay = delay
	}
}

// WithPersister sets the cache implementation
func WithPersister(cache persister.Persister[types.Resource]) Option {
	return func(o *Options) {
		o.persister = cache
	}
}

// WithAdapter sets the adapter implementation
func WithAdapter(adapter Adapter) Option {
	return func(o *Options) {
		o.adapter = adapter
	}
}

func WithSubscriber(subscriber chan types.Resource) Option {
	return func(o *Options) {
		o.subscriber = subscriber
	}
}
