package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// FederationRepository handles persistence for federation peers and export destinations
type FederationRepository struct {
	db *database.PostgresDB
}

// NewFederationRepository creates a new federation repository
func NewFederationRepository(db *database.PostgresDB) *FederationRepository {
	return &FederationRepository{db: db}
}

// SavePeer saves a federation peer
func (r *FederationRepository) SavePeer(ctx context.Context, peer *domain.OTelFederationPeer) error {
	query := `
		INSERT INTO federation_peers (id, project_id, name, url, api_key, status, last_seen, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, url = EXCLUDED.url, api_key = EXCLUDED.api_key,
			status = EXCLUDED.status, last_seen = EXCLUDED.last_seen
	`
	_, err := r.db.Pool.Exec(ctx, query,
		peer.ID, peer.ProjectID, peer.Name, peer.URL, peer.APIKey,
		peer.Status, peer.LastSeen, peer.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save federation peer: %w", err)
	}
	return nil
}

// ListPeers returns all peers for a project
func (r *FederationRepository) ListPeers(ctx context.Context, projectID uuid.UUID) ([]domain.OTelFederationPeer, error) {
	query := `
		SELECT id, project_id, name, url, api_key, status, last_seen,
			   traces_exported, spans_exported, error_count, avg_latency_ms, last_export_at, created_at
		FROM federation_peers WHERE project_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list federation peers: %w", err)
	}
	defer rows.Close()

	var peers []domain.OTelFederationPeer
	for rows.Next() {
		var p domain.OTelFederationPeer
		if err := rows.Scan(
			&p.ID, &p.ProjectID, &p.Name, &p.URL, &p.APIKey, &p.Status, &p.LastSeen,
			&p.Metrics.TracesExported, &p.Metrics.SpansExported, &p.Metrics.ErrorCount,
			&p.Metrics.AvgLatencyMs, &p.Metrics.LastExportAt, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan federation peer: %w", err)
		}
		peers = append(peers, p)
	}
	return peers, nil
}

// DeletePeer removes a federation peer
func (r *FederationRepository) DeletePeer(ctx context.Context, peerID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM federation_peers WHERE id = $1`, peerID)
	if err != nil {
		return fmt.Errorf("failed to delete federation peer: %w", err)
	}
	return nil
}

// SaveDestination saves an export destination
func (r *FederationRepository) SaveDestination(ctx context.Context, dest *domain.ExportDestination) error {
	query := `
		INSERT INTO export_destinations (
			id, project_id, name, type, endpoint, protocol, headers,
			enabled, sampling, batch_size, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, endpoint = EXCLUDED.endpoint, enabled = EXCLUDED.enabled,
			sampling = EXCLUDED.sampling, batch_size = EXCLUDED.batch_size, status = EXCLUDED.status
	`
	_, err := r.db.Pool.Exec(ctx, query,
		dest.ID, dest.ProjectID, dest.Name, dest.Type, dest.Endpoint,
		dest.Protocol, dest.Headers, dest.Enabled, dest.Sampling,
		dest.BatchSize, dest.Status, dest.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save export destination: %w", err)
	}
	return nil
}

// ListDestinations returns all destinations for a project
func (r *FederationRepository) ListDestinations(ctx context.Context, projectID uuid.UUID) ([]domain.ExportDestination, error) {
	query := `
		SELECT id, project_id, name, type, endpoint, protocol, headers,
			   enabled, sampling, batch_size, status,
			   total_exported, total_failed, last_export_at, last_error, avg_batch_ms, created_at
		FROM export_destinations WHERE project_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list export destinations: %w", err)
	}
	defer rows.Close()

	var dests []domain.ExportDestination
	for rows.Next() {
		var d domain.ExportDestination
		if err := rows.Scan(
			&d.ID, &d.ProjectID, &d.Name, &d.Type, &d.Endpoint, &d.Protocol, &d.Headers,
			&d.Enabled, &d.Sampling, &d.BatchSize, &d.Status,
			&d.Stats.TotalExported, &d.Stats.TotalFailed, &d.Stats.LastExportAt,
			&d.Stats.LastError, &d.Stats.AvgBatchMs, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan export destination: %w", err)
		}
		dests = append(dests, d)
	}
	return dests, nil
}
