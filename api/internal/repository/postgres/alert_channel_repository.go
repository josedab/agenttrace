package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// AlertChannelRepository handles persistence for alert channels
type AlertChannelRepository struct {
	db *database.PostgresDB
}

// NewAlertChannelRepository creates a new repository
func NewAlertChannelRepository(db *database.PostgresDB) *AlertChannelRepository {
	return &AlertChannelRepository{db: db}
}

// Save saves an alert channel
func (r *AlertChannelRepository) Save(ctx context.Context, channel *domain.AlertChannel) error {
	configJSON, err := json.Marshal(channel.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal channel config: %w", err)
	}

	query := `
		INSERT INTO alert_channels (id, project_id, name, type, config, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, config = EXCLUDED.config, enabled = EXCLUDED.enabled
	`
	_, err = r.db.Pool.Exec(ctx, query,
		channel.ID, channel.ProjectID, channel.Name, channel.Type,
		configJSON, channel.Enabled, channel.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save alert channel: %w", err)
	}
	return nil
}

// ListByProject returns all channels for a project
func (r *AlertChannelRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.AlertChannel, error) {
	query := `
		SELECT id, project_id, name, type, config, enabled, created_at
		FROM alert_channels WHERE project_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.AlertChannel
	for rows.Next() {
		var ch domain.AlertChannel
		var configJSON []byte
		if err := rows.Scan(&ch.ID, &ch.ProjectID, &ch.Name, &ch.Type, &configJSON, &ch.Enabled, &ch.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(configJSON, &ch.Config)
		channels = append(channels, ch)
	}
	return channels, nil
}
