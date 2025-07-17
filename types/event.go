package types

import "time"

type ChangedEvent struct {
	Action    ActionType `json:"action"`
	Timestamp time.Time  `json:"timestamp"`
	Resource  Resource   `json:"resource"`
}
