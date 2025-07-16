package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
)

var _ Authorizer = (*CasbinAuthorizer)(nil)

type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinAuthorizer(adapter persist.Adapter, modelPath string) (*CasbinAuthorizer, error) {
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	return &CasbinAuthorizer{
		enforcer: enforcer,
	}, nil
}

func (c *CasbinAuthorizer) Add(ctx context.Context, perm AddPermission) error {
	resource := c.buildResourceID(perm.ProjectId, perm.ResourceId, perm.Group)

	for _, permission := range perm.Permissions {
		action := c.getPermissionAction(permission)
		if action == "" {
			return fmt.Errorf("unknown permission: %d", permission)
		}

		if _, err := c.enforcer.AddPolicy(perm.UserId, resource, action); err != nil {
			return fmt.Errorf("failed to add permission %s for user %s on resource %s: %w", action, perm.UserId, resource, err)
		}
	}

	if err := c.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policies: %w", err)
	}

	return nil
}

func (c *CasbinAuthorizer) Check(ctx context.Context, userId string, resourceId string) ([]Permission, error) {
	var permissions []Permission

	for perm, action := range c.getAllPermissionActions() {
		allowed, err := c.enforcer.Enforce(userId, resourceId, action)
		if err != nil {
			return nil, fmt.Errorf("failed to check permission %s for user %s on resource %s: %w", action, userId, resourceId, err)
		}
		if allowed {
			permissions = append(permissions, perm)
		}
	}

	return permissions, nil
}

func (c *CasbinAuthorizer) Revoke(ctx context.Context, perm RevokePermission) error {
	resource := c.buildResourceID(perm.ProjectId, perm.ResourceId, perm.Group)

	if err := c.removePoliciesForResource(perm.UserId, resource); err != nil {
		return err
	}

	// Handle wildcard removal for project-level revocation
	if perm.ResourceId == "" && perm.Group == "" {
		if err := c.removeProjectPolicies(perm.UserId, perm.ProjectId); err != nil {
			return err
		}
	}

	if err := c.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policies: %w", err)
	}

	return nil
}

func (c *CasbinAuthorizer) removePoliciesForResource(userId, resource string) error {
	policies, err := c.enforcer.GetFilteredPolicy(0, userId, resource)
	if err != nil {
		return fmt.Errorf("failed to get filtered policies: %w", err)
	}

	for _, policy := range policies {
		if len(policy) < 3 || policy[0] != userId || policy[1] != resource {
			continue
		}

		policyInterface := make([]interface{}, len(policy))
		for i, v := range policy {
			policyInterface[i] = v
		}

		if _, err := c.enforcer.RemovePolicy(policyInterface...); err != nil {
			return fmt.Errorf("failed to remove policy %v: %w", policy, err)
		}
	}

	return nil
}

func (c *CasbinAuthorizer) removeProjectPolicies(userId, projectId string) error {
	allPolicies, err := c.enforcer.GetFilteredPolicy(0, userId)
	if err != nil {
		return fmt.Errorf("failed to get all policies for user: %w", err)
	}

	projectPattern := fmt.Sprintf("project:%s", projectId)
	for _, policy := range allPolicies {
		if len(policy) < 3 || policy[0] != userId || !strings.HasPrefix(policy[1], projectPattern) {
			continue
		}

		policyInterface := make([]interface{}, len(policy))
		for i, v := range policy {
			policyInterface[i] = v
		}

		if _, err := c.enforcer.RemovePolicy(policyInterface...); err != nil {
			return fmt.Errorf("failed to remove policy %v: %w", policy, err)
		}
	}

	return nil
}

// Helper methods for cleaner code
func (c *CasbinAuthorizer) buildResourceID(projectId, resourceId, group string) string {
	resource := resourceId
	if resource == "" {
		resource = fmt.Sprintf("project:%s", projectId)
	} else {
		resource = fmt.Sprintf("project:%s:resource:%s", projectId, resourceId)
	}

	if group != "" {
		resource = fmt.Sprintf("%s:group:%s", resource, group)
	}

	return resource
}

func (c *CasbinAuthorizer) getPermissionAction(permission Permission) string {
	actions := map[Permission]string{
		PermissionCreate: "create",
		PermissionRead:   "read",
		PermissionUpdate: "update",
		PermissionDelete: "delete",
		PermissionAdmin:  "admin",
	}
	return actions[permission]
}

func (c *CasbinAuthorizer) getAllPermissionActions() map[Permission]string {
	return map[Permission]string{
		PermissionCreate: "create",
		PermissionRead:   "read",
		PermissionUpdate: "update",
		PermissionDelete: "delete",
		PermissionAdmin:  "admin",
	}
}

// Add implements ResourceAuthorizer.
