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

// ProjectRepository handles project data operations in PostgreSQL
type ProjectRepository struct {
	db *database.PostgresDB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *database.PostgresDB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (id, organization_id, name, slug, description, settings, retention_days, rate_limit_per_minute, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		project.ID,
		project.OrganizationID,
		project.Name,
		project.Slug,
		project.Description,
		project.Settings,
		project.RetentionDays,
		project.RateLimitPerMin,
		project.CreatedAt,
		project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, organization_id, name, slug, description, settings, retention_days, rate_limit_per_minute, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var project domain.Project
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.OrganizationID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Settings,
		&project.RetentionDays,
		&project.RateLimitPerMin,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("project")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetBySlug retrieves a project by organization and slug
func (r *ProjectRepository) GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*domain.Project, error) {
	query := `
		SELECT id, organization_id, name, slug, description, settings, retention_days, rate_limit_per_minute, created_at, updated_at
		FROM projects
		WHERE organization_id = $1 AND slug = $2
	`

	var project domain.Project
	err := r.db.Pool.QueryRow(ctx, query, orgID, slug).Scan(
		&project.ID,
		&project.OrganizationID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Settings,
		&project.RetentionDays,
		&project.RateLimitPerMin,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("project")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// Update updates a project
func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $2, description = $3, settings = $4, retention_days = $5, rate_limit_per_minute = $6, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		project.ID,
		project.Name,
		project.Description,
		project.Settings,
		project.RetentionDays,
		project.RateLimitPerMin,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// Delete deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ListByOrganizationID retrieves projects for an organization
func (r *ProjectRepository) ListByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error) {
	query := `
		SELECT id, organization_id, name, slug, description, settings, retention_days, rate_limit_per_minute, created_at, updated_at
		FROM projects
		WHERE organization_id = $1
		ORDER BY name
	`

	rows, err := r.db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(
			&project.ID,
			&project.OrganizationID,
			&project.Name,
			&project.Slug,
			&project.Description,
			&project.Settings,
			&project.RetentionDays,
			&project.RateLimitPerMin,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// ListByUserID retrieves projects accessible to a user
func (r *ProjectRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.organization_id, p.name, p.slug, p.description, p.settings, p.retention_days, p.rate_limit_per_minute, p.created_at, p.updated_at
		FROM projects p
		JOIN organizations o ON p.organization_id = o.id
		JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY p.name
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(
			&project.ID,
			&project.OrganizationID,
			&project.Name,
			&project.Slug,
			&project.Description,
			&project.Settings,
			&project.RetentionDays,
			&project.RateLimitPerMin,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// AddMember adds a member to a project
func (r *ProjectRepository) AddMember(ctx context.Context, member *domain.ProjectMember) error {
	query := `
		INSERT INTO project_members (id, project_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (project_id, user_id)
		DO UPDATE SET role = $4, updated_at = NOW()
	`

	_, err := r.db.Pool.Exec(ctx, query,
		member.ID,
		member.ProjectID,
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

// GetMember retrieves a member by project and user
func (r *ProjectRepository) GetMember(ctx context.Context, projectID, userID uuid.UUID) (*domain.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, created_at, updated_at
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`

	var member domain.ProjectMember
	err := r.db.Pool.QueryRow(ctx, query, projectID, userID).Scan(
		&member.ID,
		&member.ProjectID,
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

// RemoveMember removes a member from a project
func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`

	_, err := r.db.Pool.Exec(ctx, query, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// GetUserRoleForProject retrieves the effective role for a user in a project
func (r *ProjectRepository) GetUserRoleForProject(ctx context.Context, projectID, userID uuid.UUID) (*domain.OrgRole, error) {
	// First check project-level override
	query := `
		SELECT role FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`

	var role domain.OrgRole
	err := r.db.Pool.QueryRow(ctx, query, projectID, userID).Scan(&role)
	if err == nil {
		return &role, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get project role: %w", err)
	}

	// Fall back to organization-level role
	query = `
		SELECT om.role
		FROM organization_members om
		JOIN projects p ON p.organization_id = om.organization_id
		WHERE p.id = $1 AND om.user_id = $2
	`

	err = r.db.Pool.QueryRow(ctx, query, projectID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization role: %w", err)
	}

	return &role, nil
}

// SlugExists checks if a slug already exists for an organization
func (r *ProjectRepository) SlugExists(ctx context.Context, orgID uuid.UUID, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE organization_id = $1 AND slug = $2)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, orgID, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check slug: %w", err)
	}

	return exists, nil
}
