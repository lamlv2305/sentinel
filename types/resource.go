package types

import (
	"encoding/base64"
	"encoding/json"
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

func (r Resource) Encode() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// Id implements persister.Element.
func (r Resource) Id() string {
	return r.ResourceId
}
