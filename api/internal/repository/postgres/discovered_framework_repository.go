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

// DiscoveredFrameworkRepository handles discovered framework data operations in PostgreSQL
type DiscoveredFrameworkRepository struct {
	db *database.PostgresDB
}

// NewDiscoveredFrameworkRepository creates a new discovered framework repository
func NewDiscoveredFrameworkRepository(db *database.PostgresDB) *DiscoveredFrameworkRepository {
	return &DiscoveredFrameworkRepository{db: db}
}

// Save creates or updates a discovered framework using UPSERT
func (r *DiscoveredFrameworkRepository) Save(ctx context.Context, framework *domain.DiscoveredFramework) error {
	componentsJSON, err := json.Marshal(framework.Components)
	if err != nil {
		return fmt.Errorf("failed to marshal components: %w", err)
	}

	configJSON, err := json.Marshal(framework.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO discovered_frameworks (id, project_id, framework, version, status, detected_at, components, auto_instrumented, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (project_id, framework) DO UPDATE SET
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			components = EXCLUDED.components,
			auto_instrumented = EXCLUDED.auto_instrumented,
			config = EXCLUDED.config
	`

	_, err = r.db.Pool.Exec(ctx, query,
		framework.ID,
		framework.ProjectID,
		framework.Framework,
		framework.Version,
		framework.Status,
		framework.DetectedAt,
		componentsJSON,
		framework.AutoInstrumented,
		configJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save discovered framework: %w", err)
	}

	return nil
}

// GetByID retrieves a discovered framework by ID
func (r *DiscoveredFrameworkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DiscoveredFramework, error) {
	query := `
		SELECT id, project_id, framework, version, status, detected_at, components, auto_instrumented, config
		FROM discovered_frameworks
		WHERE id = $1
	`

	var f domain.DiscoveredFramework
	var componentsJSON, configJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&f.ID,
		&f.ProjectID,
		&f.Framework,
		&f.Version,
		&f.Status,
		&f.DetectedAt,
		&componentsJSON,
		&f.AutoInstrumented,
		&configJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("discovered framework")
		}
		return nil, fmt.Errorf("failed to get discovered framework: %w", err)
	}

	if len(componentsJSON) > 0 {
		if err := json.Unmarshal(componentsJSON, &f.Components); err != nil {
			return nil, fmt.Errorf("failed to unmarshal components: %w", err)
		}
	}
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &f.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	return &f, nil
}

// ListByProject retrieves discovered frameworks for a project
func (r *DiscoveredFrameworkRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.DiscoveredFramework, error) {
	query := `
		SELECT id, project_id, framework, version, status, detected_at, components, auto_instrumented, config
		FROM discovered_frameworks
		WHERE project_id = $1
		ORDER BY detected_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list discovered frameworks: %w", err)
	}
	defer rows.Close()

	var frameworks []domain.DiscoveredFramework
	for rows.Next() {
		var f domain.DiscoveredFramework
		var componentsJSON, configJSON []byte

		if err := rows.Scan(
			&f.ID,
			&f.ProjectID,
			&f.Framework,
			&f.Version,
			&f.Status,
			&f.DetectedAt,
			&componentsJSON,
			&f.AutoInstrumented,
			&configJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan discovered framework: %w", err)
		}

		if len(componentsJSON) > 0 {
			if err := json.Unmarshal(componentsJSON, &f.Components); err != nil {
				return nil, fmt.Errorf("failed to unmarshal components: %w", err)
			}
		}
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &f.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}

		frameworks = append(frameworks, f)
	}

	return frameworks, nil
}

// UpdateStatus updates the status of a discovered framework
func (r *DiscoveredFrameworkRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DiscoveryStatus) error {
	query := `UPDATE discovered_frameworks SET status = $2 WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update discovered framework status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("discovered framework")
	}

	return nil
}

// Delete removes a discovered framework by ID
func (r *DiscoveredFrameworkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM discovered_frameworks WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete discovered framework: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("discovered framework")
	}

	return nil
}
