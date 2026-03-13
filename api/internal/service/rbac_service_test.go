package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func newTestRBACService() *RBACService {
	logger, _ := zap.NewDevelopment()
	return NewRBACService(logger)
}

func TestRBACService_CheckPermission(t *testing.T) {
	tests := []struct {
		name       string
		setupRole  *domain.Role
		permission domain.Permission
		expected   bool
	}{
		{
			name:       "admin with write permission returns true",
			setupRole:  rolePtr(domain.RoleAdmin),
			permission: domain.PermTraceWrite,
			expected:   true,
		},
		{
			name:       "viewer requesting write returns false",
			setupRole:  rolePtr(domain.RoleViewer),
			permission: domain.PermTraceWrite,
			expected:   false,
		},
		{
			name:       "user with NO role assignment returns false",
			setupRole:  nil,
			permission: domain.PermTraceRead,
			expected:   false,
		},
		{
			name:       "admin with read permission returns true",
			setupRole:  rolePtr(domain.RoleAdmin),
			permission: domain.PermTraceRead,
			expected:   true,
		},
		{
			name:       "viewer with read permission returns true",
			setupRole:  rolePtr(domain.RoleViewer),
			permission: domain.PermTraceRead,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestRBACService()
			ctx := context.Background()
			userID := uuid.New()
			projectID := uuid.New()

			if tt.setupRole != nil {
				_, err := svc.AssignRole(ctx, domain.RoleAssignmentInput{
					UserID:    userID,
					ProjectID: projectID,
					Role:      *tt.setupRole,
				})
				require.NoError(t, err)
			}

			result, err := svc.CheckPermission(ctx, userID, projectID, tt.permission)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRBACService_UnknownRole(t *testing.T) {
	svc := newTestRBACService()
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()

	_, err := svc.AssignRole(ctx, domain.RoleAssignmentInput{
		UserID:    userID,
		ProjectID: projectID,
		Role:      domain.Role("unknown-role"),
	})
	require.NoError(t, err)

	result, err := svc.CheckPermission(ctx, userID, projectID, domain.PermTraceRead)
	require.NoError(t, err)
	assert.False(t, result, "unknown role should have no permissions")
}

func TestRBACService_MultiProjectIsolation(t *testing.T) {
	svc := newTestRBACService()
	ctx := context.Background()
	userID := uuid.New()
	projectA := uuid.New()
	projectB := uuid.New()

	// Assign admin to project A only
	_, err := svc.AssignRole(ctx, domain.RoleAssignmentInput{
		UserID:    userID,
		ProjectID: projectA,
		Role:      domain.RoleAdmin,
	})
	require.NoError(t, err)

	// Should have access to project A
	result, err := svc.CheckPermission(ctx, userID, projectA, domain.PermTraceWrite)
	require.NoError(t, err)
	assert.True(t, result)

	// Should NOT have access to project B
	result, err = svc.CheckPermission(ctx, userID, projectB, domain.PermTraceWrite)
	require.NoError(t, err)
	assert.False(t, result)
}

func TestRBACService_GetRolePermissions(t *testing.T) {
	svc := newTestRBACService()

	t.Run("known role returns permissions", func(t *testing.T) {
		perms := svc.GetRolePermissions(domain.RoleAdmin)
		assert.NotEmpty(t, perms)
	})

	t.Run("unknown role returns empty", func(t *testing.T) {
		perms := svc.GetRolePermissions(domain.Role("nonexistent"))
		assert.Empty(t, perms)
	})
}

func rolePtr(r domain.Role) *domain.Role {
	return &r
}
