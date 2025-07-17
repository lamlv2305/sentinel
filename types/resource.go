package types

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
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

var (
	_ encoding.TextMarshaler   = (*Resource)(nil)
	_ encoding.TextUnmarshaler = (*Resource)(nil)
)

type Resource struct {
	ResourceId   string       `json:"resource_id"`
	ProjectId    string       `json:"project_id"`
	Group        string       `json:"group,omitempty"`
	ResourceType ResourceType `json:"resource_type"`
	Data         []byte       `json:"data,omitempty"`
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (r *Resource) UnmarshalText(text []byte) error {
	bytes, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, r); err != nil {
		return err
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (r Resource) MarshalText() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return []byte(base64.StdEncoding.EncodeToString(bytes)), nil
}
