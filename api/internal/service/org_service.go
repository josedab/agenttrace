package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// OrgService handles organization operations
type OrgService struct {
	orgRepo OrgRepository
}

// NewOrgService creates a new organization service
func NewOrgService(orgRepo OrgRepository) *OrgService {
	return &OrgService{
		orgRepo: orgRepo,
	}
}

// Create creates a new organization
func (s *OrgService) Create(ctx context.Context, name string, ownerID uuid.UUID) (*domain.Organization, error) {
	slug := domain.GenerateSlug(name)

	// Check if slug exists
	exists, err := s.orgRepo.SlugExists(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}

	// Generate unique slug if exists
	if exists {
		slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	}

	now := time.Now()
	org := &domain.Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add owner
	member := &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         ownerID,
		Role:           domain.OrgRoleOwner,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add owner: %w", err)
	}

	return org, nil
}

// Get retrieves an organization by ID
func (s *OrgService) Get(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return s.orgRepo.GetByID(ctx, id)
}

// GetBySlug retrieves an organization by slug
func (s *OrgService) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	return s.orgRepo.GetBySlug(ctx, slug)
}

// Update updates an organization
func (s *OrgService) Update(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	org.Name = name
	org.UpdatedAt = time.Now()

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return org, nil
}

// Delete deletes an organization
func (s *OrgService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.orgRepo.Delete(ctx, id)
}

// ListByUser retrieves organizations for a user
func (s *OrgService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error) {
	return s.orgRepo.ListByUserID(ctx, userID)
}

// GetMember retrieves a member by organization and user
func (s *OrgService) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	return s.orgRepo.GetMember(ctx, orgID, userID)
}

// CheckAccess checks if a user has access to an organization
func (s *OrgService) CheckAccess(ctx context.Context, orgID, userID uuid.UUID, requiredRole domain.OrgRole) error {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.Forbidden("no access to organization")
		}
		return err
	}

	roleLevel := map[domain.OrgRole]int{
		domain.OrgRoleViewer: 1,
		domain.OrgRoleMember: 2,
		domain.OrgRoleAdmin:  3,
		domain.OrgRoleOwner:  4,
	}

	if roleLevel[member.Role] < roleLevel[requiredRole] {
		return apperrors.Forbidden("insufficient permissions")
	}

	return nil
}
