package resgate

import (
	"context"
)

type ResGate struct {
	Options
}

func NewResGate(opts ...WithOptions) *ResGate {
	rg := &ResGate{
		Options: defaultOptions(),
	}

	for _, opt := range opts {
		opt(&rg.Options)
	}

	return rg
}

func (rg *ResGate) Run(ctx context.Context) error {
	return rg.adapter.Run(ctx)
}
