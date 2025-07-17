package adapter

import (
	"context"

	"github.com/lamlv2305/sentinel/types"
)

type Adapter interface {
	OnChanged(ctx context.Context, event types.ChangedEvent) error
	Credential(ctx context.Context, apikey string, project string) error
	Run(ctx context.Context) error
}
