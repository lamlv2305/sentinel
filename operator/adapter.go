package operator

import (
	"context"

	"github.com/lamlv2305/sentinel/types"
)

type Adapter interface {
	Broadcast(ctx context.Context, event types.ChangedEvent) error
	Run(ctx context.Context) error
}

type Hook struct {
	OnConnected    []func(ctx context.Context, client *Client)
	OnDisconnected []func(ctx context.Context, client *Client)
}

type CredentialVerifier func(ctx context.Context, apikey string, project string) error
