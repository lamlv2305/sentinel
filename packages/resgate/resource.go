package resgate

import (
	"time"
)

type ActionType string

const (
	ActionTypeCreate ActionType = "create"
	ActionTypeUpdate ActionType = "update"
	ActionTypeDelete ActionType = "delete"
)

type ResourceType string

const (
	ResourceTypeText       ResourceType = "text"
	ResourceTypeJsonObject ResourceType = "json_object"
	ResourceTypeJsonArray  ResourceType = "json_array"
	ResourceTypeBinary     ResourceType = "binary"
	ResourceTypeImage      ResourceType = "image"
)

type Resource struct {
	ResourceId   string       `json:"resource_id"`
	ProjectId    string       `json:"project_id"`
	Group        string       `json:"group,omitempty"`
	ResourceType ResourceType `json:"resource_type"`
	Data         []byte       `json:"data,omitempty"`
}

type ResourceChangedEvent struct {
	Action    ActionType `json:"action"`
	Timestamp time.Time  `json:"timestamp"`
	Resource  Resource   `json:"resource"`
}
