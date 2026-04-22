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

// ShareLinkRepository persists hashed, revocable share tokens.
type ShareLinkRepository struct {
	db *database.PostgresDB
}

// NewShareLinkRepository creates a share link repository.
func NewShareLinkRepository(db *database.PostgresDB) *ShareLinkRepository {
	return &ShareLinkRepository{db: db}
}

// Create stores a share link without the raw token.
func (r *ShareLinkRepository) Create(ctx context.Context, link *domain.ShareLink) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO share_links (
			id, project_id, resource_type, resource_id, token_hash,
			redaction_version, expires_at, revoked_at, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		link.ID,
		link.ProjectID,
		link.ResourceType,
		link.ResourceID,
		link.TokenHash,
		link.RedactionVersion,
		link.ExpiresAt,
		link.RevokedAt,
		link.CreatedBy,
		link.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create share link: %w", err)
	}
	return nil
}

// GetByTokenHash resolves an unauthenticated share token.
func (r *ShareLinkRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) (*domain.ShareLink, error) {
	return r.get(ctx, "token_hash = $1", tokenHash)
}

// GetByID retrieves a share link only within its project.
func (r *ShareLinkRepository) GetByID(
	ctx context.Context,
	projectID, linkID uuid.UUID,
) (*domain.ShareLink, error) {
	return r.get(ctx, "project_id = $1 AND id = $2", projectID, linkID)
}

func (r *ShareLinkRepository) get(
	ctx context.Context,
	where string,
	args ...interface{},
) (*domain.ShareLink, error) {
	var link domain.ShareLink
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, project_id, resource_type, resource_id, token_hash,
			redaction_version, expires_at, revoked_at, created_by, created_at
		FROM share_links
		WHERE `+where,
		args...,
	).Scan(
		&link.ID,
		&link.ProjectID,
		&link.ResourceType,
		&link.ResourceID,
		&link.TokenHash,
		&link.RedactionVersion,
		&link.ExpiresAt,
		&link.RevokedAt,
		&link.CreatedBy,
		&link.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("share link")
		}
		return nil, fmt.Errorf("get share link: %w", err)
	}
	return &link, nil
}

// Revoke marks a link revoked within its project.
func (r *ShareLinkRepository) Revoke(
	ctx context.Context,
	projectID, linkID uuid.UUID,
	revokedAt time.Time,
) error {
	result, err := r.db.Pool.Exec(ctx, `
		UPDATE share_links
		SET revoked_at = $3
		WHERE project_id = $1 AND id = $2 AND revoked_at IS NULL
	`, projectID, linkID, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke share link: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("share link")
	}
	return nil
}
