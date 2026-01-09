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

// PromptRepository handles prompt data operations in PostgreSQL
type PromptRepository struct {
	db *database.PostgresDB
}

// NewPromptRepository creates a new prompt repository
func NewPromptRepository(db *database.PostgresDB) *PromptRepository {
	return &PromptRepository{db: db}
}

// Create creates a new prompt
func (r *PromptRepository) Create(ctx context.Context, prompt *domain.Prompt) error {
	query := `
		INSERT INTO prompts (id, project_id, name, type, description, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		prompt.ID,
		prompt.ProjectID,
		prompt.Name,
		prompt.Type,
		prompt.Description,
		prompt.Tags,
		prompt.CreatedAt,
		prompt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create prompt: %w", err)
	}

	return nil
}

// GetByID retrieves a prompt by ID
func (r *PromptRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Prompt, error) {
	query := `
		SELECT id, project_id, name, type, description, tags, created_at, updated_at
		FROM prompts
		WHERE id = $1
	`

	var prompt domain.Prompt
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&prompt.ID,
		&prompt.ProjectID,
		&prompt.Name,
		&prompt.Type,
		&prompt.Description,
		&prompt.Tags,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt")
		}
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	return &prompt, nil
}

// GetByName retrieves a prompt by name
func (r *PromptRepository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*domain.Prompt, error) {
	query := `
		SELECT id, project_id, name, type, description, tags, created_at, updated_at
		FROM prompts
		WHERE project_id = $1 AND name = $2
	`

	var prompt domain.Prompt
	err := r.db.Pool.QueryRow(ctx, query, projectID, name).Scan(
		&prompt.ID,
		&prompt.ProjectID,
		&prompt.Name,
		&prompt.Type,
		&prompt.Description,
		&prompt.Tags,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt")
		}
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	return &prompt, nil
}

// Update updates a prompt
func (r *PromptRepository) Update(ctx context.Context, prompt *domain.Prompt) error {
	query := `
		UPDATE prompts
		SET name = $2, type = $3, description = $4, tags = $5, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		prompt.ID,
		prompt.Name,
		prompt.Type,
		prompt.Description,
		prompt.Tags,
	)
	if err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	return nil
}

// Delete deletes a prompt
func (r *PromptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prompts WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	return nil
}

// List retrieves prompts with filtering
func (r *PromptRepository) List(ctx context.Context, filter *domain.PromptFilter, limit, offset int) (*domain.PromptList, error) {
	baseQuery := `FROM prompts WHERE project_id = $1`
	args := []interface{}{filter.ProjectID}
	argIndex := 2

	if filter.Name != nil {
		baseQuery += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if len(filter.Tags) > 0 {
		baseQuery += fmt.Sprintf(" AND tags && $%d", argIndex)
		args = append(args, filter.Tags)
		argIndex++
	}

	// Get count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count prompts: %w", err)
	}

	// Get prompts
	query := fmt.Sprintf(`
		SELECT id, project_id, name, type, description, tags, created_at, updated_at
		%s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}
	defer rows.Close()

	var prompts []domain.Prompt
	for rows.Next() {
		var prompt domain.Prompt
		if err := rows.Scan(
			&prompt.ID,
			&prompt.ProjectID,
			&prompt.Name,
			&prompt.Type,
			&prompt.Description,
			&prompt.Tags,
			&prompt.CreatedAt,
			&prompt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	return &domain.PromptList{
		Prompts:    prompts,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(prompts)) < totalCount,
	}, nil
}

// CreateVersion creates a new prompt version
func (r *PromptRepository) CreateVersion(ctx context.Context, version *domain.PromptVersion) error {
	// Get next version number
	var nextVersion int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(version), 0) + 1 FROM prompt_versions WHERE prompt_id = $1`,
		version.PromptID,
	).Scan(&nextVersion)
	if err != nil {
		return fmt.Errorf("failed to get next version: %w", err)
	}
	version.Version = nextVersion

	query := `
		INSERT INTO prompt_versions (id, prompt_id, version, content, config, labels, created_by, commit_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		version.ID,
		version.PromptID,
		version.Version,
		version.Content,
		version.Config,
		version.Labels,
		version.CreatedBy,
		version.CommitMessage,
		version.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create prompt version: %w", err)
	}

	return nil
}

// GetVersion retrieves a specific prompt version
func (r *PromptRepository) GetVersion(ctx context.Context, promptID uuid.UUID, version int) (*domain.PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, content, config, labels, created_by, commit_message, created_at
		FROM prompt_versions
		WHERE prompt_id = $1 AND version = $2
	`

	var v domain.PromptVersion
	err := r.db.Pool.QueryRow(ctx, query, promptID, version).Scan(
		&v.ID,
		&v.PromptID,
		&v.Version,
		&v.Content,
		&v.Config,
		&v.Labels,
		&v.CreatedBy,
		&v.CommitMessage,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt version")
		}
		return nil, fmt.Errorf("failed to get prompt version: %w", err)
	}

	v.Variables = domain.ExtractVariables(v.Content)

	return &v, nil
}

// GetLatestVersion retrieves the latest prompt version
func (r *PromptRepository) GetLatestVersion(ctx context.Context, promptID uuid.UUID) (*domain.PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, content, config, labels, created_by, commit_message, created_at
		FROM prompt_versions
		WHERE prompt_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	var v domain.PromptVersion
	err := r.db.Pool.QueryRow(ctx, query, promptID).Scan(
		&v.ID,
		&v.PromptID,
		&v.Version,
		&v.Content,
		&v.Config,
		&v.Labels,
		&v.CreatedBy,
		&v.CommitMessage,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt version")
		}
		return nil, fmt.Errorf("failed to get latest prompt version: %w", err)
	}

	v.Variables = domain.ExtractVariables(v.Content)

	return &v, nil
}

// GetVersionByLabel retrieves the latest prompt version with a specific label
func (r *PromptRepository) GetVersionByLabel(ctx context.Context, promptID uuid.UUID, label string) (*domain.PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, content, config, labels, created_by, commit_message, created_at
		FROM prompt_versions
		WHERE prompt_id = $1 AND $2 = ANY(labels)
		ORDER BY version DESC
		LIMIT 1
	`

	var v domain.PromptVersion
	err := r.db.Pool.QueryRow(ctx, query, promptID, label).Scan(
		&v.ID,
		&v.PromptID,
		&v.Version,
		&v.Content,
		&v.Config,
		&v.Labels,
		&v.CreatedBy,
		&v.CommitMessage,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("prompt version")
		}
		return nil, fmt.Errorf("failed to get prompt version by label: %w", err)
	}

	v.Variables = domain.ExtractVariables(v.Content)

	return &v, nil
}

// ListVersions retrieves all versions of a prompt
func (r *PromptRepository) ListVersions(ctx context.Context, promptID uuid.UUID) ([]domain.PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, content, config, labels, created_by, commit_message, created_at
		FROM prompt_versions
		WHERE prompt_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, promptID)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompt versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.PromptVersion
	for rows.Next() {
		var v domain.PromptVersion
		if err := rows.Scan(
			&v.ID,
			&v.PromptID,
			&v.Version,
			&v.Content,
			&v.Config,
			&v.Labels,
			&v.CreatedBy,
			&v.CommitMessage,
			&v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan prompt version: %w", err)
		}
		v.Variables = domain.ExtractVariables(v.Content)
		versions = append(versions, v)
	}

	return versions, nil
}

// UpdateVersionLabels updates the labels of a prompt version
func (r *PromptRepository) UpdateVersionLabels(ctx context.Context, versionID uuid.UUID, labels []string) error {
	query := `UPDATE prompt_versions SET labels = $2 WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, versionID, labels)
	if err != nil {
		return fmt.Errorf("failed to update version labels: %w", err)
	}

	return nil
}

// NameExists checks if a prompt name already exists
func (r *PromptRepository) NameExists(ctx context.Context, projectID uuid.UUID, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM prompts WHERE project_id = $1 AND name = $2)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, projectID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check name: %w", err)
	}

	return exists, nil
}
