package resgate

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"
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

func (rg *ResGate) Subscribe(client *Client) {
	rg.hub.add(client)
}

func (rg *ResGate) Unsubscribe(client *Client) {
	rg.hub.remove(client.ProjectId, client.Id)
}

func (rg *ResGate) Changed(ctx context.Context, event ResourceChangedEvent) {
	bytes, err := json.Marshal(event)
	if err != nil {
		rg.logger.ErrorContext(ctx, "Failed to marshal event", "error", err)
		return
	}

	rg.hub.broadcast(event.Resource.ProjectId, base64.StdEncoding.EncodeToString(bytes))
}

func (rg *ResGate) Run(ctx context.Context) error {
	healthCheckTicker := time.NewTicker(5 * time.Second)
	defer healthCheckTicker.Stop()

	reportTicker := time.NewTicker(60 * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-healthCheckTicker.C:
			rg.hub.healthCheck()

		case <-reportTicker.C:
			rg.logger.DebugContext(ctx, "Running periodic report",
				"clients", rg.hub.getTotalClientCount(),
				"projects", len(rg.hub.getProjectIds()))
		}
	}
}
