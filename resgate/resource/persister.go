package resource

import (
	"context"

	"github.com/lamlv2305/sentinel/types"
)

type ResourceQuery struct {
	ProjectId  string `json:"project_id"`
	Group      string `json:"group,omitempty"`
	ResourceId string `json:"resource_id,omitempty"`
}

type Persister interface {
	Save(ctx context.Context, resource types.Resource) error
	Delete(ctx context.Context, query ResourceQuery) error
	Get(ctx context.Context, resourceId string) (types.Resource, error)
	GetByQuery(ctx context.Context, query ResourceQuery) (types.Resource, error) // New method
	List(ctx context.Context, query ResourceQuery) ([]types.Resource, error)
}
