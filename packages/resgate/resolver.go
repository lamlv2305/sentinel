package resgate

import (
	"context"
	"errors"
)

type Resolver interface {
	Credential(ctx context.Context, apikey string, project string) error
}

func defaultResolver() Resolver {
	return &defaultResolverImpl{}
}

type defaultResolverImpl struct{}

func (r *defaultResolverImpl) Credential(ctx context.Context, apikey string, project string) error {
	return errors.New("default resolver does not implement Credential method")
}
