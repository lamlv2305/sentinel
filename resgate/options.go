package resgate

import (
	"log/slog"
	"net/http"
	"os"
)

type Options struct {
	port       int
	logger     *slog.Logger
	mux        *http.ServeMux
	hub        *hub
	resolver   Resolver
	persister  ResourcePersister
	authorizer ResourceAuthorizer
}

func defaultOptions() Options {
	return Options{
		port:     8080, // Default port
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		mux:      http.NewServeMux(),
		hub:      defaultHub(),
		resolver: defaultResolver(),
	}
}

type WithOptions func(*Options)

func WithPort(port int) WithOptions {
	return func(opts *Options) {
		opts.port = port
	}
}

func WithLogger(logger *slog.Logger) WithOptions {
	return func(opts *Options) {
		opts.logger = logger
	}
}

func WithServer(mux *http.ServeMux) WithOptions {
	return func(opts *Options) {
		opts.mux = mux
	}
}

func WithDispatcher(d *hub) WithOptions {
	return func(opts *Options) {
		opts.hub = d
	}
}

func WithResolver(r Resolver) WithOptions {
	return func(opts *Options) {
		opts.resolver = r
	}
}

func WithPersister(p ResourcePersister) WithOptions {
	return func(opts *Options) {
		opts.persister = p
	}
}

func WithAuthorizer(a ResourceAuthorizer) WithOptions {
	return func(opts *Options) {
		opts.authorizer = a
	}
}
