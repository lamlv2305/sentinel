package adapter

import (
	"context"

	"github.com/lamlv2305/sentinel/resgate/resource"
)

type Adapter interface {
	OnChanged(ctx context.Context, event resource.ChangedEvent) error
	Credential(ctx context.Context, apikey string, project string) error
	Run(ctx context.Context) error
}
