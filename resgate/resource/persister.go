package resource

import (
	"context"
)

type ResourceQuery struct {
	ProjectId  string `json:"project_id"`
	Group      string `json:"group,omitempty"`
	ResourceId string `json:"resource_id,omitempty"`
}

type Persister interface {
	Save(ctx context.Context, resource Resource) error
	Delete(ctx context.Context, query ResourceQuery) error
	Get(ctx context.Context, resourceId string) (Resource, error)
	GetByQuery(ctx context.Context, query ResourceQuery) (Resource, error) // New method
	List(ctx context.Context, query ResourceQuery) ([]Resource, error)
}
