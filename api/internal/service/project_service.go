package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// FullProjectRepository defines all project repository operations
type FullProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Project, error)
	ListAll(ctx context.Context, limit, offset int) ([]domain.Project, error)
	AddMember(ctx context.Context, member *domain.ProjectMember) error
	GetMember(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	GetUserRoleForProject(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error)
	SlugExists(ctx context.Context, orgID uuid.UUID, slug string) (bool, error)
}

// ProjectInput represents input for creating/updating a project
type ProjectInput struct {
	Name           string              `json:"name" validate:"required"`
	Description    string              `json:"description,omitempty"`
	Settings       *domain.ProjectSettings `json:"settings,omitempty"`
	RetentionDays  *int                `json:"retentionDays,omitempty"`
	RateLimitPerMin *int               `json:"rateLimitPerMin,omitempty"`
}

// ProjectService handles project operations
type ProjectService struct {
	projectRepo FullProjectRepository
	orgRepo     OrgRepository
}

// NewProjectService creates a new project service
func NewProjectService(projectRepo FullProjectRepository, orgRepo OrgRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		orgRepo:     orgRepo,
	}
}

// Create creates a new project
func (s *ProjectService) Create(ctx context.Context, orgID uuid.UUID, input *ProjectInput, userID uuid.UUID) (*domain.Project, error) {
	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	slug := domain.GenerateSlug(input.Name)

	// Check if slug exists
	exists, err := s.projectRepo.SlugExists(ctx, orgID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}

	// Generate unique slug if exists
	if exists {
		slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	}

	now := time.Now()

	var settings string
	if input.Settings != nil {
		settingsBytes, _ := json.Marshal(input.Settings)
		settings = string(settingsBytes)
	}

	retentionDays := 30
	if input.RetentionDays != nil {
		retentionDays = *input.RetentionDays
	}

	defaultRateLimit := 1000
	rateLimitPerMin := &defaultRateLimit
	if input.RateLimitPerMin != nil {
		rateLimitPerMin = input.RateLimitPerMin
	}

	project := &domain.Project{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		Name:            input.Name,
		Slug:            slug,
		Description:     input.Description,
		Settings:        settings,
		RetentionDays:   retentionDays,
		RateLimitPerMin: rateLimitPerMin,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// Get retrieves a project by ID
func (s *ProjectService) Get(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// GetBySlug retrieves a project by organization and slug
func (s *ProjectService) GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*domain.Project, error) {
	return s.projectRepo.GetBySlug(ctx, orgID, slug)
}

// Update updates a project
func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, input *ProjectInput) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}
	if input.Settings != nil {
		settingsBytes, _ := json.Marshal(input.Settings)
		project.Settings = string(settingsBytes)
	}
	if input.RetentionDays != nil {
		project.RetentionDays = *input.RetentionDays
	}
	if input.RateLimitPerMin != nil {
		project.RateLimitPerMin = input.RateLimitPerMin
	}

	project.UpdatedAt = time.Now()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// Delete deletes a project
func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.projectRepo.Delete(ctx, id)
}

// ListByOrganization retrieves projects for an organization
func (s *ProjectService) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error) {
	return s.projectRepo.ListByOrganizationID(ctx, orgID)
}

// ListByUser retrieves projects accessible to a user
func (s *ProjectService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	return s.projectRepo.ListByUserID(ctx, userID)
}

// AddMember adds a member to a project
func (s *ProjectService) AddMember(ctx context.Context, projectID, userID uuid.UUID, role domain.OrgRole) error {
	now := time.Now()
	member := &domain.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.projectRepo.AddMember(ctx, member)
}

// RemoveMember removes a member from a project
func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

// CheckAccess checks if a user has access to a project
func (s *ProjectService) CheckAccess(ctx context.Context, projectID, userID uuid.UUID, requiredRole domain.OrgRole) error {
	role, err := s.projectRepo.GetUserRoleForProject(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role == nil {
		return apperrors.Forbidden("no access to project")
	}

	roleLevel := map[domain.OrgRole]int{
		domain.OrgRoleViewer: 1,
		domain.OrgRoleMember: 2,
		domain.OrgRoleAdmin:  3,
		domain.OrgRoleOwner:  4,
	}

	if roleLevel[*role] < roleLevel[requiredRole] {
		return apperrors.Forbidden("insufficient permissions")
	}

	return nil
}

// GetUserRole retrieves the user's role for a project
func (s *ProjectService) GetUserRole(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error) {
	return s.projectRepo.GetUserRoleForProject(ctx, projectID, userID)
}

// ListAll retrieves all projects with pagination (for system tasks like cleanup)
func (s *ProjectService) ListAll(ctx context.Context, limit, offset int) ([]domain.Project, error) {
	return s.projectRepo.ListAll(ctx, limit, offset)
}
