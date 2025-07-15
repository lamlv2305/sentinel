package resgate

import "context"

type Permission int

const (
	PermissionNone Permission = iota
	PermissionCreate
	PermissionRead
	PermissionUpdate
	PermissionDelete
	PermissionAdmin
)

type AddPermission struct {
	UserId    string // Unable to empty
	ProjectId string // Unable to empty

	ResourceId string // Empty means all resources in the project
	Group      string // Empty means all groups

	// Admin is a special permission that grants all other permissions.
	Permissions []Permission // At least one permission must be specified
}

type ResourceAuthorizer interface {
	Check(ctx context.Context, userId string, resourceId string) ([]Permission, error)
	Add(ctx context.Context, perm AddPermission) error
	Revoke(ctx context.Context, userId string, resourceId string, group string) error
}

var _ ResourceAuthorizer = &CasbinAuthorizer{}

type CasbinAuthorizer struct {
	// TODO: this adapter can be any available casbin adapter, such as FileAdapter, MemoryAdapter, PostgresAdapter, etc.
	// Change it later to fit with casbin's adapter interface.
	adapter any
}

func NewCasbinAuthorizer(adapter any) *CasbinAuthorizer {
	return &CasbinAuthorizer{
		adapter: adapter,
	}
}
