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

// WorkflowDefinitionRepository handles workflow definition data operations in PostgreSQL
type WorkflowDefinitionRepository struct {
	db *database.PostgresDB
}

// NewWorkflowDefinitionRepository creates a new workflow definition repository
func NewWorkflowDefinitionRepository(db *database.PostgresDB) *WorkflowDefinitionRepository {
	return &WorkflowDefinitionRepository{db: db}
}

// SaveDefinition creates a new workflow definition
func (r *WorkflowDefinitionRepository) SaveDefinition(ctx context.Context, def *domain.WorkflowDefinition) error {
	nodesJSON, err := json.Marshal(def.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(def.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	query := `
		INSERT INTO workflow_definitions (id, project_id, name, description, status, nodes, edges, version, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		def.ID,
		def.ProjectID,
		def.Name,
		def.Description,
		def.Status,
		nodesJSON,
		edgesJSON,
		def.Version,
		def.CreatedAt,
		def.UpdatedAt,
		def.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save workflow definition: %w", err)
	}

	return nil
}

// GetDefinitionByID retrieves a workflow definition by ID
func (r *WorkflowDefinitionRepository) GetDefinitionByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowDefinition, error) {
	query := `
		SELECT id, project_id, name, description, status, nodes, edges, version, created_at, updated_at, created_by
		FROM workflow_definitions
		WHERE id = $1
	`

	var def domain.WorkflowDefinition
	var nodesJSON, edgesJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&def.ID,
		&def.ProjectID,
		&def.Name,
		&def.Description,
		&def.Status,
		&nodesJSON,
		&edgesJSON,
		&def.Version,
		&def.CreatedAt,
		&def.UpdatedAt,
		&def.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("workflow definition")
		}
		return nil, fmt.Errorf("failed to get workflow definition: %w", err)
	}

	if len(nodesJSON) > 0 {
		if err := json.Unmarshal(nodesJSON, &def.Nodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
		}
	}
	if len(edgesJSON) > 0 {
		if err := json.Unmarshal(edgesJSON, &def.Edges); err != nil {
			return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
		}
	}

	return &def, nil
}

// ListDefinitions retrieves workflow definitions for a project
func (r *WorkflowDefinitionRepository) ListDefinitions(ctx context.Context, projectID uuid.UUID) ([]domain.WorkflowDefinition, error) {
	query := `
		SELECT id, project_id, name, description, status, nodes, edges, version, created_at, updated_at, created_by
		FROM workflow_definitions
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow definitions: %w", err)
	}
	defer rows.Close()

	var definitions []domain.WorkflowDefinition
	for rows.Next() {
		var def domain.WorkflowDefinition
		var nodesJSON, edgesJSON []byte

		if err := rows.Scan(
			&def.ID,
			&def.ProjectID,
			&def.Name,
			&def.Description,
			&def.Status,
			&nodesJSON,
			&edgesJSON,
			&def.Version,
			&def.CreatedAt,
			&def.UpdatedAt,
			&def.CreatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workflow definition: %w", err)
		}

		if len(nodesJSON) > 0 {
			if err := json.Unmarshal(nodesJSON, &def.Nodes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
			}
		}
		if len(edgesJSON) > 0 {
			if err := json.Unmarshal(edgesJSON, &def.Edges); err != nil {
				return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
			}
		}

		definitions = append(definitions, def)
	}

	return definitions, nil
}

// UpdateDefinition updates an existing workflow definition
func (r *WorkflowDefinitionRepository) UpdateDefinition(ctx context.Context, def *domain.WorkflowDefinition) error {
	nodesJSON, err := json.Marshal(def.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(def.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	query := `
		UPDATE workflow_definitions
		SET name = $2, description = $3, status = $4, nodes = $5, edges = $6, version = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		def.ID,
		def.Name,
		def.Description,
		def.Status,
		nodesJSON,
		edgesJSON,
		def.Version,
		def.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow definition: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("workflow definition")
	}

	return nil
}

// DeleteDefinition deletes a workflow definition by ID
func (r *WorkflowDefinitionRepository) DeleteDefinition(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workflow_definitions WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow definition: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("workflow definition")
	}

	return nil
}

// SaveSimulation creates a new workflow simulation
func (r *WorkflowDefinitionRepository) SaveSimulation(ctx context.Context, sim *domain.WorkflowSimulation) error {
	resultsJSON, err := json.Marshal(sim.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	query := `
		INSERT INTO workflow_simulations (id, workflow_id, project_id, name, status, predicted_cost_usd, predicted_latency_ms, predicted_quality_score, trace_data_used, results, started_at, completed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		sim.ID,
		sim.WorkflowID,
		sim.ProjectID,
		sim.Name,
		sim.Status,
		sim.PredictedCostUSD,
		sim.PredictedLatencyMs,
		sim.PredictedQualityScore,
		sim.TraceDataUsed,
		resultsJSON,
		sim.StartedAt,
		sim.CompletedAt,
		sim.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save workflow simulation: %w", err)
	}

	return nil
}

// GetSimulationByID retrieves a workflow simulation by ID
func (r *WorkflowDefinitionRepository) GetSimulationByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowSimulation, error) {
	query := `
		SELECT id, workflow_id, project_id, name, status, predicted_cost_usd, predicted_latency_ms, predicted_quality_score, trace_data_used, results, started_at, completed_at, created_at
		FROM workflow_simulations
		WHERE id = $1
	`

	var sim domain.WorkflowSimulation
	var resultsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&sim.ID,
		&sim.WorkflowID,
		&sim.ProjectID,
		&sim.Name,
		&sim.Status,
		&sim.PredictedCostUSD,
		&sim.PredictedLatencyMs,
		&sim.PredictedQualityScore,
		&sim.TraceDataUsed,
		&resultsJSON,
		&sim.StartedAt,
		&sim.CompletedAt,
		&sim.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("workflow simulation")
		}
		return nil, fmt.Errorf("failed to get workflow simulation: %w", err)
	}

	if len(resultsJSON) > 0 {
		if err := json.Unmarshal(resultsJSON, &sim.Results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}
	}

	return &sim, nil
}

// ListSimulations retrieves workflow simulations for a workflow
func (r *WorkflowDefinitionRepository) ListSimulations(ctx context.Context, workflowID uuid.UUID) ([]domain.WorkflowSimulation, error) {
	query := `
		SELECT id, workflow_id, project_id, name, status, predicted_cost_usd, predicted_latency_ms, predicted_quality_score, trace_data_used, results, started_at, completed_at, created_at
		FROM workflow_simulations
		WHERE workflow_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow simulations: %w", err)
	}
	defer rows.Close()

	var simulations []domain.WorkflowSimulation
	for rows.Next() {
		var sim domain.WorkflowSimulation
		var resultsJSON []byte

		if err := rows.Scan(
			&sim.ID,
			&sim.WorkflowID,
			&sim.ProjectID,
			&sim.Name,
			&sim.Status,
			&sim.PredictedCostUSD,
			&sim.PredictedLatencyMs,
			&sim.PredictedQualityScore,
			&sim.TraceDataUsed,
			&resultsJSON,
			&sim.StartedAt,
			&sim.CompletedAt,
			&sim.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workflow simulation: %w", err)
		}

		if len(resultsJSON) > 0 {
			if err := json.Unmarshal(resultsJSON, &sim.Results); err != nil {
				return nil, fmt.Errorf("failed to unmarshal results: %w", err)
			}
		}

		simulations = append(simulations, sim)
	}

	return simulations, nil
}
