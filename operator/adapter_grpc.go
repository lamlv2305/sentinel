package operator

import (
	"context"

	"github.com/lamlv2305/sentinel/types"
)

// TODO: future implementation of gRPC adapter
var _ Adapter = &AdapterGRPC{}

type AdapterGRPC struct{}

// Credential implements Adapter.
func (a *AdapterGRPC) Credential(ctx context.Context, apikey string, project string) error {
	panic("unimplemented")
}

// OnChanged implements Adapter.
func (a *AdapterGRPC) Broadcast(ctx context.Context, event types.ChangedEvent) error {
	panic("unimplemented")
}

// Run implements Adapter.
func (a *AdapterGRPC) Run(ctx context.Context) error {
	panic("unimplemented")
}

func NewGRPC() *AdapterGRPC {
	return &AdapterGRPC{}
}
