package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// AgentVersionRepository handles persistence for agent versions
type AgentVersionRepository struct {
	db *database.PostgresDB
}

// NewAgentVersionRepository creates a new agent version repository
func NewAgentVersionRepository(db *database.PostgresDB) *AgentVersionRepository {
	return &AgentVersionRepository{db: db}
}

// Save persists an agent version
func (r *AgentVersionRepository) Save(ctx context.Context, version *domain.AgentVersion) error {
	configJSON, _ := json.Marshal(version.Config)
	query := `INSERT INTO agent_versions (id, project_id, agent_name, version, tag, config, created_by, change_note, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			tag = EXCLUDED.tag, config = EXCLUDED.config, is_active = EXCLUDED.is_active, change_note = EXCLUDED.change_note`
	_, err := r.db.Pool.Exec(ctx, query, version.ID, version.ProjectID, version.AgentName,
		version.Version, version.Tag, configJSON, version.CreatedBy, version.ChangeNote,
		version.IsActive, version.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save agent version: %w", err)
	}
	return nil
}

// GetByID retrieves an agent version by ID
func (r *AgentVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentVersion, error) {
	query := `SELECT id, project_id, agent_name, version, tag, config, created_by, change_note, is_active, created_at
		FROM agent_versions WHERE id = $1`
	var v domain.AgentVersion
	var configJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&v.ID, &v.ProjectID, &v.AgentName,
		&v.Version, &v.Tag, &configJSON, &v.CreatedBy, &v.ChangeNote, &v.IsActive, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent version: %w", err)
	}
	json.Unmarshal(configJSON, &v.Config)
	return &v, nil
}

// ListByAgent returns all versions for a specific agent
func (r *AgentVersionRepository) ListByAgent(ctx context.Context, projectID uuid.UUID, agentName string) ([]domain.AgentVersion, error) {
	query := `SELECT id, project_id, agent_name, version, tag, config, created_by, change_note, is_active, created_at
		FROM agent_versions WHERE project_id = $1 AND agent_name = $2 ORDER BY version DESC LIMIT 100`
	rows, err := r.db.Pool.Query(ctx, query, projectID, agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent versions: %w", err)
	}
	defer rows.Close()
	var versions []domain.AgentVersion
	for rows.Next() {
		var v domain.AgentVersion
		var configJSON []byte
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.AgentName, &v.Version, &v.Tag, &configJSON,
			&v.CreatedBy, &v.ChangeNote, &v.IsActive, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan agent version: %w", err)
		}
		json.Unmarshal(configJSON, &v.Config)
		versions = append(versions, v)
	}
	return versions, nil
}

// GetActive returns the currently active version for an agent
func (r *AgentVersionRepository) GetActive(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.AgentVersion, error) {
	query := `SELECT id, project_id, agent_name, version, tag, config, created_by, change_note, is_active, created_at
		FROM agent_versions WHERE project_id = $1 AND agent_name = $2 AND is_active = true LIMIT 1`
	var v domain.AgentVersion
	var configJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, projectID, agentName).Scan(&v.ID, &v.ProjectID, &v.AgentName,
		&v.Version, &v.Tag, &configJSON, &v.CreatedBy, &v.ChangeNote, &v.IsActive, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get active agent version: %w", err)
	}
	json.Unmarshal(configJSON, &v.Config)
	return &v, nil
}

// SetActive sets a specific version as the active version, deactivating others
func (r *AgentVersionRepository) SetActive(ctx context.Context, projectID uuid.UUID, agentName string, versionID uuid.UUID) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE agent_versions SET is_active = false WHERE project_id = $1 AND agent_name = $2`, projectID, agentName)
	if err != nil {
		return fmt.Errorf("failed to deactivate agent versions: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE agent_versions SET is_active = true WHERE id = $1`, versionID)
	if err != nil {
		return fmt.Errorf("failed to activate agent version: %w", err)
	}
	return tx.Commit(ctx)
}
