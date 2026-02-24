package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

type FileTraceMappingRepository struct {
	db *database.PostgresDB
}

func NewFileTraceMappingRepository(db *database.PostgresDB) *FileTraceMappingRepository {
	return &FileTraceMappingRepository{db: db}
}

func (r *FileTraceMappingRepository) Upsert(ctx context.Context, mapping *domain.FileTraceMapping) error {
	annotationsJSON, err := json.Marshal(mapping.Annotations)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	summaryJSON, err := json.Marshal(mapping.Summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	query := `
		INSERT INTO file_trace_mappings (project_id, file_path, annotations, summary, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, file_path) DO UPDATE
		SET annotations = EXCLUDED.annotations,
			summary = EXCLUDED.summary,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.Pool.Exec(ctx, query,
		mapping.ProjectID,
		mapping.FilePath,
		annotationsJSON,
		summaryJSON,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert file trace mapping: %w", err)
	}

	return nil
}

func (r *FileTraceMappingRepository) GetByFilePath(ctx context.Context, projectID uuid.UUID, filePath string) (*domain.FileTraceMapping, error) {
	query := `
		SELECT project_id, file_path, annotations, summary
		FROM file_trace_mappings
		WHERE project_id = $1 AND file_path = $2
	`

	var m domain.FileTraceMapping
	var annotationsJSON, summaryJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, projectID, filePath).Scan(
		&m.ProjectID, &m.FilePath, &annotationsJSON, &summaryJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("file trace mapping")
		}
		return nil, fmt.Errorf("failed to get file trace mapping: %w", err)
	}

	if len(annotationsJSON) > 0 {
		if err := json.Unmarshal(annotationsJSON, &m.Annotations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}
	if len(summaryJSON) > 0 {
		if err := json.Unmarshal(summaryJSON, &m.Summary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
		}
	}

	return &m, nil
}

func (r *FileTraceMappingRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.FileTraceMapping, error) {
	query := `
		SELECT project_id, file_path, annotations, summary
		FROM file_trace_mappings
		WHERE project_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list file trace mappings: %w", err)
	}
	defer rows.Close()

	var mappings []domain.FileTraceMapping
	for rows.Next() {
		var m domain.FileTraceMapping
		var annotationsJSON, summaryJSON []byte

		if err := rows.Scan(
			&m.ProjectID, &m.FilePath, &annotationsJSON, &summaryJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan file trace mapping: %w", err)
		}

		if len(annotationsJSON) > 0 {
			if err := json.Unmarshal(annotationsJSON, &m.Annotations); err != nil {
				return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
			}
		}
		if len(summaryJSON) > 0 {
			if err := json.Unmarshal(summaryJSON, &m.Summary); err != nil {
				return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
			}
		}

		mappings = append(mappings, m)
	}

	return mappings, nil
}

func (r *FileTraceMappingRepository) GetBatch(ctx context.Context, projectID uuid.UUID, filePaths []string) ([]domain.FileTraceMapping, error) {
	query := `
		SELECT project_id, file_path, annotations, summary
		FROM file_trace_mappings
		WHERE project_id = $1 AND file_path = ANY($2)
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, filePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch file trace mappings: %w", err)
	}
	defer rows.Close()

	var mappings []domain.FileTraceMapping
	for rows.Next() {
		var m domain.FileTraceMapping
		var annotationsJSON, summaryJSON []byte

		if err := rows.Scan(
			&m.ProjectID, &m.FilePath, &annotationsJSON, &summaryJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan file trace mapping: %w", err)
		}

		if len(annotationsJSON) > 0 {
			if err := json.Unmarshal(annotationsJSON, &m.Annotations); err != nil {
				return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
			}
		}
		if len(summaryJSON) > 0 {
			if err := json.Unmarshal(summaryJSON, &m.Summary); err != nil {
				return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
			}
		}

		mappings = append(mappings, m)
	}

	return mappings, nil
}

func (r *FileTraceMappingRepository) Delete(ctx context.Context, projectID uuid.UUID, filePath string) error {
	query := `DELETE FROM file_trace_mappings WHERE project_id = $1 AND file_path = $2`

	_, err := r.db.Pool.Exec(ctx, query, projectID, filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file trace mapping: %w", err)
	}

	return nil
}
