package resagent

import "time"

// Options holds configuration for Resagent
type Options struct {
	resgateURL     string
	timeout        time.Duration
	reconnectDelay time.Duration
	cache          Cache
	adapter        Adapter
}

// Option is a function that configures Options
type Option func(*Options)

// defaultOptions returns default configuration
func defaultOptions() *Options {
	return &Options{
		resgateURL:     "http://localhost:8080",
		timeout:        30 * time.Second,
		reconnectDelay: 5 * time.Second,
		cache:          nil, // Will be set later
		adapter:        nil, // Will be set later
	}
}

// WithResgateURL sets the Resgate URL
func WithResgateURL(url string) Option {
	return func(o *Options) {
		o.resgateURL = url
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

// WithCache sets the cache implementation
func WithCache(cache Cache) Option {
	return func(o *Options) {
		o.cache = cache
	}
}

// WithAdapter sets the adapter implementation
func WithAdapter(adapter Adapter) Option {
	return func(o *Options) {
		o.adapter = adapter
	}
}
