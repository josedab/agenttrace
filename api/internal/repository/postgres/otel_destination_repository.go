package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// OTelDestinationRepository handles OTel export destination data operations in PostgreSQL
type OTelDestinationRepository struct {
	db *database.PostgresDB
}

// NewOTelDestinationRepository creates a new OTel destination repository
func NewOTelDestinationRepository(db *database.PostgresDB) *OTelDestinationRepository {
	return &OTelDestinationRepository{db: db}
}

// Save creates a new export destination
func (r *OTelDestinationRepository) Save(ctx context.Context, dest *domain.ExportDestination) error {
	headersJSON, err := json.Marshal(dest.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		INSERT INTO otel_export_destinations (id, project_id, name, type, endpoint, protocol, headers, enabled, sampling, batch_size, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		dest.ID,
		dest.ProjectID,
		dest.Name,
		dest.Type,
		dest.Endpoint,
		dest.Protocol,
		headersJSON,
		dest.Enabled,
		dest.Sampling,
		dest.BatchSize,
		dest.Status,
		dest.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save export destination: %w", err)
	}

	return nil
}

// GetByID retrieves an export destination by ID
func (r *OTelDestinationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ExportDestination, error) {
	query := `
		SELECT id, project_id, name, type, endpoint, protocol, headers, enabled, sampling, batch_size, status, created_at
		FROM otel_export_destinations
		WHERE id = $1
	`

	var dest domain.ExportDestination
	var headersJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&dest.ID,
		&dest.ProjectID,
		&dest.Name,
		&dest.Type,
		&dest.Endpoint,
		&dest.Protocol,
		&headersJSON,
		&dest.Enabled,
		&dest.Sampling,
		&dest.BatchSize,
		&dest.Status,
		&dest.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("export destination")
		}
		return nil, fmt.Errorf("failed to get export destination: %w", err)
	}

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &dest.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	return &dest, nil
}

// ListByProject retrieves export destinations for a project
func (r *OTelDestinationRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.ExportDestination, error) {
	query := `
		SELECT id, project_id, name, type, endpoint, protocol, headers, enabled, sampling, batch_size, status, created_at
		FROM otel_export_destinations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list export destinations: %w", err)
	}
	defer rows.Close()

	var destinations []domain.ExportDestination
	for rows.Next() {
		var dest domain.ExportDestination
		var headersJSON []byte

		if err := rows.Scan(
			&dest.ID,
			&dest.ProjectID,
			&dest.Name,
			&dest.Type,
			&dest.Endpoint,
			&dest.Protocol,
			&headersJSON,
			&dest.Enabled,
			&dest.Sampling,
			&dest.BatchSize,
			&dest.Status,
			&dest.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan export destination: %w", err)
		}

		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &dest.Headers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
			}
		}

		destinations = append(destinations, dest)
	}

	return destinations, nil
}

// Delete removes an export destination by ID
func (r *OTelDestinationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM otel_export_destinations WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete export destination: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("export destination")
	}

	return nil
}

// UpdateExportStats updates export statistics for a destination
func (r *OTelDestinationRepository) UpdateExportStats(ctx context.Context, id uuid.UUID, count int64, errorCount int64) error {
	query := `
		UPDATE otel_export_destinations
		SET status = CASE WHEN $3 > 0 THEN 'error' ELSE 'active' END
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, id, count, errorCount)
	if err != nil {
		return fmt.Errorf("failed to update export stats: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("export destination")
	}

	return nil
}
