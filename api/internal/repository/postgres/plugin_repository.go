package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// PluginRepository handles persistence for plugins
type PluginRepository struct {
	db *database.PostgresDB
}

// NewPluginRepository creates a new plugin repository
func NewPluginRepository(db *database.PostgresDB) *PluginRepository {
	return &PluginRepository{db: db}
}

// Save persists a plugin
func (r *PluginRepository) Save(ctx context.Context, plugin *domain.Plugin) error {
	configJSON, _ := json.Marshal(plugin.Config)
	query := `INSERT INTO plugins (id, project_id, name, description, type, version, author, entry_point, config, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, description = EXCLUDED.description, version = EXCLUDED.version,
			config = EXCLUDED.config, status = EXCLUDED.status`
	_, err := r.db.Pool.Exec(ctx, query, plugin.ID, plugin.ProjectID, plugin.Name, plugin.Description,
		plugin.Type, plugin.Version, plugin.Author, plugin.EntryPoint, configJSON, plugin.Status, plugin.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save plugin: %w", err)
	}
	return nil
}

// GetByID retrieves a plugin by ID
func (r *PluginRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Plugin, error) {
	query := `SELECT id, project_id, name, description, type, version, author, entry_point, config, status, created_at
		FROM plugins WHERE id = $1`
	var p domain.Plugin
	var configJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.ProjectID, &p.Name, &p.Description,
		&p.Type, &p.Version, &p.Author, &p.EntryPoint, &configJSON, &p.Status, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	json.Unmarshal(configJSON, &p.Config)
	return &p, nil
}

// ListByProject returns all plugins for a project
func (r *PluginRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Plugin, error) {
	query := `SELECT id, project_id, name, description, type, version, author, entry_point, config, status, created_at
		FROM plugins WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}
	defer rows.Close()
	var plugins []domain.Plugin
	for rows.Next() {
		var p domain.Plugin
		var configJSON []byte
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Name, &p.Description, &p.Type, &p.Version,
			&p.Author, &p.EntryPoint, &configJSON, &p.Status, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		json.Unmarshal(configJSON, &p.Config)
		plugins = append(plugins, p)
	}
	return plugins, nil
}

// UpdateStatus updates the status of a plugin
func (r *PluginRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PluginStatus) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE plugins SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}
	return nil
}

// Delete removes a plugin
func (r *PluginRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM plugins WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}
