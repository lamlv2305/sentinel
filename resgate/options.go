package resgate

import (
	"log/slog"
	"os"

	"github.com/lamlv2305/sentinel/resgate/adapter"
	"github.com/lamlv2305/sentinel/resgate/resource"
)

type Options struct {
	logger *slog.Logger

	persister  resource.Persister
	authorizer resource.Authorizer
	adapter    adapter.Adapter
}

func defaultOptions() Options {
	return Options{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

type WithOptions func(*Options)

func WithLogger(logger *slog.Logger) WithOptions {
	return func(opts *Options) {
		opts.logger = logger
	}
}

func WithPersister(p resource.Persister) WithOptions {
	return func(opts *Options) {
		opts.persister = p
	}
}

func WithAuthorizer(a resource.Authorizer) WithOptions {
	return func(opts *Options) {
		opts.authorizer = a
	}
}

func WithAdapter(a adapter.Adapter) WithOptions {
	return func(opts *Options) {
		opts.adapter = a
	}
}
