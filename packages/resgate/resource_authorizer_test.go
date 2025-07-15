package resgate

import (
	"context"
	"os"
	"testing"

	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

func TestCasbinAuthorizer(t *testing.T) {
	// Setup test files
	testPolicyFile := "./test_policies.csv"
	testModelFile := "./model/rbac_model.conf"

	// Create empty policy file for file adapter
	file, err := os.Create(testPolicyFile)
	if err != nil {
		t.Fatalf("Failed to create test policy file: %v", err)
	}
	file.Close()

	// Clean up test file after test
	defer os.Remove(testPolicyFile)

	// Create test adapter and authorizer
	adapter := fileadapter.NewAdapter(testPolicyFile)
	authorizer, err := NewCasbinAuthorizer(adapter, testModelFile)
	if err != nil {
		t.Fatalf("Failed to create authorizer: %v", err)
	}

	ctx := context.Background()

	t.Run("Add and Check Permissions", func(t *testing.T) {
		// Add permissions for a user
		addPerm := AddPermission{
			UserId:     "user123",
			ProjectId:  "project1",
			ResourceId: "resource1",
			Permissions: []Permission{
				PermissionRead,
				PermissionUpdate,
			},
		}

		if err := authorizer.Add(ctx, addPerm); err != nil {
			t.Fatalf("Failed to add permissions: %v", err)
		}

		// Check permissions
		permissions, err := authorizer.Check(ctx, "user123", "project:project1:resource:resource1")
		if err != nil {
			t.Fatalf("Failed to check permissions: %v", err)
		}

		expectedPerms := []Permission{PermissionRead, PermissionUpdate}
		if !containsAllPermissions(permissions, expectedPerms) {
			t.Errorf("Expected permissions %v, got %v", expectedPerms, permissions)
		}
	})

	t.Run("Add Project-wide Admin Permissions", func(t *testing.T) {
		// Add admin permissions for all resources in a project
		adminPerm := AddPermission{
			UserId:     "admin456",
			ProjectId:  "project1",
			ResourceId: "", // empty means all resources in the project
			Permissions: []Permission{
				PermissionAdmin,
			},
		}

		if err := authorizer.Add(ctx, adminPerm); err != nil {
			t.Fatalf("Failed to add admin permissions: %v", err)
		}

		// Check admin permissions on project level
		permissions, err := authorizer.Check(ctx, "admin456", "project:project1")
		if err != nil {
			t.Fatalf("Failed to check admin permissions: %v", err)
		}

		if !contains(permissions, PermissionAdmin) {
			t.Errorf("Expected admin permission, got %v", permissions)
		}
	})

	t.Run("Revoke Specific Resource Permissions", func(t *testing.T) {
		// Revoke permissions for a specific resource
		revokePerm := RevokePermission{
			UserId:     "user123",
			ProjectId:  "project1",
			ResourceId: "resource1",
		}

		if err := authorizer.Revoke(ctx, revokePerm); err != nil {
			t.Fatalf("Failed to revoke permissions: %v", err)
		}

		// Check permissions after revocation
		permissions, err := authorizer.Check(ctx, "user123", "project:project1:resource:resource1")
		if err != nil {
			t.Fatalf("Failed to check permissions after revocation: %v", err)
		}

		if len(permissions) > 0 {
			t.Errorf("Expected no permissions after revocation, got %v", permissions)
		}
	})

	t.Run("Revoke All Project Permissions", func(t *testing.T) {
		// First add some permissions
		addPerm := AddPermission{
			UserId:      "user789",
			ProjectId:   "project2",
			ResourceId:  "resource1",
			Permissions: []Permission{PermissionRead},
		}
		if err := authorizer.Add(ctx, addPerm); err != nil {
			t.Fatalf("Failed to add permissions: %v", err)
		}

		// Revoke all permissions in project
		revokeAllPerm := RevokePermission{
			UserId:     "user789",
			ProjectId:  "project2",
			ResourceId: "", // empty means all resources
		}

		if err := authorizer.Revoke(ctx, revokeAllPerm); err != nil {
			t.Fatalf("Failed to revoke all project permissions: %v", err)
		}

		// Check that permissions are gone
		permissions, err := authorizer.Check(ctx, "user789", "project:project2:resource:resource1")
		if err != nil {
			t.Fatalf("Failed to check permissions: %v", err)
		}

		if len(permissions) > 0 {
			t.Errorf("Expected no permissions after project-wide revocation, got %v", permissions)
		}
	})

	t.Run("Invalid Permission", func(t *testing.T) {
		// Test with invalid permission
		addPerm := AddPermission{
			UserId:      "user999",
			ProjectId:   "project1",
			ResourceId:  "resource1",
			Permissions: []Permission{Permission(999)}, // Invalid permission
		}

		err := authorizer.Add(ctx, addPerm)
		if err == nil {
			t.Error("Expected error for invalid permission, got nil")
		}
	})
}

func TestBuildResourceID(t *testing.T) {
	// Create empty policy file for file adapter
	testFile := "./test_build.csv"
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()
	defer os.Remove(testFile)

	adapter := fileadapter.NewAdapter(testFile)
	authorizer, err := NewCasbinAuthorizer(adapter, "./model/rbac_model.conf")
	if err != nil {
		t.Fatalf("Failed to create authorizer: %v", err)
	}

	tests := []struct {
		name       string
		projectId  string
		resourceId string
		group      string
		expected   string
	}{
		{
			name:       "Project only",
			projectId:  "proj1",
			resourceId: "",
			group:      "",
			expected:   "project:proj1",
		},
		{
			name:       "Project and resource",
			projectId:  "proj1",
			resourceId: "res1",
			group:      "",
			expected:   "project:proj1:resource:res1",
		},
		{
			name:       "Project, resource, and group",
			projectId:  "proj1",
			resourceId: "res1",
			group:      "group1",
			expected:   "project:proj1:resource:res1:group:group1",
		},
		{
			name:       "Project and group",
			projectId:  "proj1",
			resourceId: "",
			group:      "group1",
			expected:   "project:proj1:group:group1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authorizer.buildResourceID(tt.projectId, tt.resourceId, tt.group)
			if result != tt.expected {
				t.Errorf("buildResourceID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetPermissionAction(t *testing.T) {
	// Create empty policy file for file adapter
	testFile := "./test_action.csv"
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()
	defer os.Remove(testFile)

	adapter := fileadapter.NewAdapter(testFile)
	authorizer, err := NewCasbinAuthorizer(adapter, "./model/rbac_model.conf")
	if err != nil {
		t.Fatalf("Failed to create authorizer: %v", err)
	}

	tests := []struct {
		permission Permission
		expected   string
	}{
		{PermissionCreate, "create"},
		{PermissionRead, "read"},
		{PermissionUpdate, "update"},
		{PermissionDelete, "delete"},
		{PermissionAdmin, "admin"},
		{Permission(999), ""}, // Invalid permission
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := authorizer.getPermissionAction(tt.permission)
			if result != tt.expected {
				t.Errorf("getPermissionAction(%v) = %v, want %v", tt.permission, result, tt.expected)
			}
		})
	}
}

// Helper functions for tests
func contains(slice []Permission, item Permission) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func containsAllPermissions(got, expected []Permission) bool {
	if len(got) != len(expected) {
		return false
	}

	gotMap := make(map[Permission]bool)
	for _, p := range got {
		gotMap[p] = true
	}

	for _, p := range expected {
		if !gotMap[p] {
			return false
		}
	}

	return true
}
