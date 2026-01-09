package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// APIKeyRepository handles API key data operations in PostgreSQL
type APIKeyRepository struct {
	db *database.PostgresDB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *database.PostgresDB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, project_id, name, public_key, secret_key_hash, secret_key_preview, scopes, expires_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		key.ID,
		key.ProjectID,
		key.Name,
		key.PublicKey,
		key.SecretKeyHash,
		key.SecretKeyPreview,
		key.Scopes,
		key.ExpiresAt,
		key.CreatedBy,
		key.CreatedAt,
		key.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	query := `
		SELECT id, project_id, name, public_key, secret_key_hash, secret_key_preview, scopes, expires_at, last_used_at, created_by, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`

	var key domain.APIKey
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&key.ID,
		&key.ProjectID,
		&key.Name,
		&key.PublicKey,
		&key.SecretKeyHash,
		&key.SecretKeyPreview,
		&key.Scopes,
		&key.ExpiresAt,
		&key.LastUsedAt,
		&key.CreatedBy,
		&key.CreatedAt,
		&key.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("API key")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &key, nil
}

// GetByPublicKey retrieves an API key by public key
func (r *APIKeyRepository) GetByPublicKey(ctx context.Context, publicKey string) (*domain.APIKey, error) {
	query := `
		SELECT id, project_id, name, public_key, secret_key_hash, secret_key_preview, scopes, expires_at, last_used_at, created_by, created_at, updated_at
		FROM api_keys
		WHERE public_key = $1
	`

	var key domain.APIKey
	err := r.db.Pool.QueryRow(ctx, query, publicKey).Scan(
		&key.ID,
		&key.ProjectID,
		&key.Name,
		&key.PublicKey,
		&key.SecretKeyHash,
		&key.SecretKeyPreview,
		&key.Scopes,
		&key.ExpiresAt,
		&key.LastUsedAt,
		&key.CreatedBy,
		&key.CreatedAt,
		&key.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("API key")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &key, nil
}

// Update updates an API key
func (r *APIKeyRepository) Update(ctx context.Context, key *domain.APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $2, scopes = $3, expires_at = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query,
		key.ID,
		key.Name,
		key.Scopes,
		key.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// Delete deletes an API key
func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// ListByProjectID retrieves API keys for a project
func (r *APIKeyRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]domain.APIKey, error) {
	query := `
		SELECT id, project_id, name, public_key, secret_key_hash, secret_key_preview, scopes, expires_at, last_used_at, created_by, created_at, updated_at
		FROM api_keys
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var key domain.APIKey
		if err := rows.Scan(
			&key.ID,
			&key.ProjectID,
			&key.Name,
			&key.PublicKey,
			&key.SecretKeyHash,
			&key.SecretKeyPreview,
			&key.Scopes,
			&key.ExpiresAt,
			&key.LastUsedAt,
			&key.CreatedBy,
			&key.CreatedAt,
			&key.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// UpdateLastUsed updates the last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = $2 WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}

// CountByProjectID counts API keys for a project
func (r *APIKeyRepository) CountByProjectID(ctx context.Context, projectID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM api_keys WHERE project_id = $1`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	return count, nil
}

// GetProjectIDByPublicKey retrieves the project ID for a public key
func (r *APIKeyRepository) GetProjectIDByPublicKey(ctx context.Context, publicKey string) (*uuid.UUID, error) {
	query := `
		SELECT project_id
		FROM api_keys
		WHERE public_key = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	var projectID uuid.UUID
	err := r.db.Pool.QueryRow(ctx, query, publicKey).Scan(&projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("API key")
		}
		return nil, fmt.Errorf("failed to get project ID: %w", err)
	}

	return &projectID, nil
}
