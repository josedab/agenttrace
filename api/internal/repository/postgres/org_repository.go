package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// OrgRepository handles organization data operations in PostgreSQL
type OrgRepository struct {
	db *database.PostgresDB
}

// NewOrgRepository creates a new organization repository
func NewOrgRepository(db *database.PostgresDB) *OrgRepository {
	return &OrgRepository{db: db}
}

// Create creates a new organization
func (r *OrgRepository) Create(ctx context.Context, org *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		org.ID,
		org.Name,
		org.Slug,
		org.CreatedAt,
		org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

// GetByID retrieves an organization by ID
func (r *OrgRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	var org domain.Organization
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("organization")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// GetBySlug retrieves an organization by slug
func (r *OrgRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`

	var org domain.Organization
	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("organization")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// Update updates an organization
func (r *OrgRepository) Update(ctx context.Context, org *domain.Organization) error {
	query := `
		UPDATE organizations
		SET name = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, org.ID, org.Name)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	return nil
}

// Delete deletes an organization
func (r *OrgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// ListByUserID retrieves organizations for a user
func (r *OrgRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.created_at, o.updated_at
		FROM organizations o
		JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY o.name
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []domain.Organization
	for rows.Next() {
		var org domain.Organization
		if err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.CreatedAt,
			&org.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// AddMember adds a member to an organization
func (r *OrgRepository) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (organization_id, user_id)
		DO UPDATE SET role = $4, updated_at = NOW()
	`

	_, err := r.db.Pool.Exec(ctx, query,
		member.ID,
		member.OrganizationID,
		member.UserID,
		member.Role,
		member.CreatedAt,
		member.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// GetMember retrieves a member by organization and user
func (r *OrgRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`

	var member domain.OrganizationMember
	err := r.db.Pool.QueryRow(ctx, query, orgID, userID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.CreatedAt,
		&member.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("member")
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return &member, nil
}

// ListMembers retrieves members of an organization
func (r *OrgRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.created_at, om.updated_at,
			   u.id, u.email, u.name, u.image
		FROM organization_members om
		JOIN users u ON om.user_id = u.id
		WHERE om.organization_id = $1
		ORDER BY om.role, u.name
	`

	rows, err := r.db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()

	var members []domain.OrganizationMember
	for rows.Next() {
		var member domain.OrganizationMember
		var user domain.User
		if err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.UserID,
			&member.Role,
			&member.CreatedAt,
			&member.UpdatedAt,
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Image,
		); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		member.User = &user
		members = append(members, member)
	}

	return members, nil
}

// RemoveMember removes a member from an organization
func (r *OrgRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	query := `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`

	_, err := r.db.Pool.Exec(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// UpdateMemberRole updates a member's role
func (r *OrgRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role domain.OrgRole) error {
	query := `
		UPDATE organization_members
		SET role = $3, updated_at = NOW()
		WHERE organization_id = $1 AND user_id = $2
	`

	_, err := r.db.Pool.Exec(ctx, query, orgID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

// CreateInvitation creates an organization invitation
func (r *OrgRepository) CreateInvitation(ctx context.Context, invitation *domain.OrganizationInvitation) error {
	query := `
		INSERT INTO organization_invitations (id, organization_id, email, role, invited_by, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		invitation.ID,
		invitation.OrganizationID,
		invitation.Email,
		invitation.Role,
		invitation.InvitedBy,
		invitation.Token,
		invitation.ExpiresAt,
		invitation.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

// GetInvitationByToken retrieves an invitation by token
func (r *OrgRepository) GetInvitationByToken(ctx context.Context, token string) (*domain.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invited_by, token, expires_at, accepted_at, created_at
		FROM organization_invitations
		WHERE token = $1 AND expires_at > NOW() AND accepted_at IS NULL
	`

	var invitation domain.OrganizationInvitation
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.InvitedBy,
		&invitation.Token,
		&invitation.ExpiresAt,
		&invitation.AcceptedAt,
		&invitation.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("invitation")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &invitation, nil
}

// AcceptInvitation marks an invitation as accepted
func (r *OrgRepository) AcceptInvitation(ctx context.Context, token string) error {
	query := `
		UPDATE organization_invitations
		SET accepted_at = NOW()
		WHERE token = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	return nil
}

// ListPendingInvitations retrieves pending invitations for an organization
func (r *OrgRepository) ListPendingInvitations(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, invited_by, token, expires_at, accepted_at, created_at
		FROM organization_invitations
		WHERE organization_id = $1 AND expires_at > NOW() AND accepted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []domain.OrganizationInvitation
	for rows.Next() {
		var inv domain.OrganizationInvitation
		if err := rows.Scan(
			&inv.ID,
			&inv.OrganizationID,
			&inv.Email,
			&inv.Role,
			&inv.InvitedBy,
			&inv.Token,
			&inv.ExpiresAt,
			&inv.AcceptedAt,
			&inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}

// SlugExists checks if a slug already exists
func (r *OrgRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check slug: %w", err)
	}

	return exists, nil
}
