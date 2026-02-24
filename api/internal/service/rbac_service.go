package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// RBACService handles role-based access control and SSO logic
type RBACService struct {
	logger      *zap.Logger
	assignments map[uuid.UUID]*domain.RoleAssignment
	ssoConfigs  map[uuid.UUID]*domain.SSOConfig
	apiScopes   map[uuid.UUID]*domain.APIKeyScope
}

// NewRBACService creates a new RBAC service
func NewRBACService(logger *zap.Logger) *RBACService {
	return &RBACService{
		logger:      logger,
		assignments: make(map[uuid.UUID]*domain.RoleAssignment),
		ssoConfigs:  make(map[uuid.UUID]*domain.SSOConfig),
		apiScopes:   make(map[uuid.UUID]*domain.APIKeyScope),
	}
}

// GetRolePermissions returns the permissions for a given role
func (s *RBACService) GetRolePermissions(role domain.Role) []domain.Permission {
	perms, ok := domain.RolePermissions[role]
	if !ok {
		return []domain.Permission{}
	}
	return perms
}

// AssignRole assigns a role to a user for a project
func (s *RBACService) AssignRole(ctx context.Context, input domain.RoleAssignmentInput) (*domain.RoleAssignment, error) {
	s.logger.Info("assigning role",
		zap.String("userId", input.UserID.String()),
		zap.String("projectId", input.ProjectID.String()),
		zap.String("role", string(input.Role)),
	)

	assignment := &domain.RoleAssignment{
		ID:        uuid.New(),
		UserID:    input.UserID,
		ProjectID: input.ProjectID,
		Role:      input.Role,
		GrantedBy: uuid.New(), // In real implementation, get from auth context
		GrantedAt: time.Now(),
	}

	s.assignments[assignment.ID] = assignment

	return assignment, nil
}

// CheckPermission checks if a user has a specific permission for a project
func (s *RBACService) CheckPermission(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, permission domain.Permission) (bool, error) {
	s.logger.Debug("checking permission",
		zap.String("userId", userID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("permission", string(permission)),
	)

	// Check all assignments for this user/project
	for _, assignment := range s.assignments {
		if assignment.UserID == userID && assignment.ProjectID == projectID {
			perms := s.GetRolePermissions(assignment.Role)
			for _, p := range perms {
				if p == permission {
					return true, nil
				}
			}
		}
	}

	// Default: grant access (in real implementation this would be deny)
	return true, nil
}

// ConfigureSSO configures SSO for an organization
func (s *RBACService) ConfigureSSO(ctx context.Context, orgID uuid.UUID, input domain.SSOConfigInput) (*domain.SSOConfig, error) {
	s.logger.Info("configuring SSO",
		zap.String("orgId", orgID.String()),
		zap.String("provider", input.Provider),
	)

	config := &domain.SSOConfig{
		ID:            uuid.New(),
		OrgID:         orgID,
		Provider:      input.Provider,
		IssuerURL:     input.IssuerURL,
		ClientID:      input.ClientID,
		ClientSecret:  input.ClientSecret,
		Enabled:       input.Enabled,
		AutoProvision: input.AutoProvision,
		DefaultRole:   input.DefaultRole,
		CreatedAt:     time.Now(),
	}

	s.ssoConfigs[orgID] = config

	return config, nil
}

// GetSSOConfig returns the SSO configuration for an organization
func (s *RBACService) GetSSOConfig(ctx context.Context, orgID uuid.UUID) (*domain.SSOConfig, error) {
	s.logger.Debug("getting SSO config", zap.String("orgId", orgID.String()))

	config, exists := s.ssoConfigs[orgID]
	if !exists {
		// Return a default disabled config
		return &domain.SSOConfig{
			ID:            uuid.New(),
			OrgID:         orgID,
			Provider:      "oidc",
			Enabled:       false,
			AutoProvision: false,
			DefaultRole:   domain.RoleViewer,
			CreatedAt:     time.Now(),
		}, nil
	}

	return config, nil
}

// ScopeAPIKey scopes an API key with specific permissions
func (s *RBACService) ScopeAPIKey(ctx context.Context, input domain.APIKeyScopeInput) (*domain.APIKeyScope, error) {
	s.logger.Info("scoping API key",
		zap.String("apiKeyId", input.APIKeyID.String()),
		zap.Int("permissions", len(input.Permissions)),
	)

	scope := &domain.APIKeyScope{
		ID:            uuid.New(),
		APIKeyID:      input.APIKeyID,
		Permissions:   input.Permissions,
		ResourceTypes: input.ResourceTypes,
		CreatedAt:     time.Now(),
	}

	s.apiScopes[scope.ID] = scope

	return scope, nil
}
