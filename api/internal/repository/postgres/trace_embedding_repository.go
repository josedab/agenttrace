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

// TraceEmbeddingRecord represents a trace embedding stored in the database
type TraceEmbeddingRecord struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"projectId"`
	TraceID        uuid.UUID  `json:"traceId"`
	ObservationID  *uuid.UUID `json:"observationId,omitempty"`
	ContentType    string     `json:"contentType"`
	ContentHash    string     `json:"contentHash"`
	EmbeddingModel string     `json:"embeddingModel"`
	IndexedAt      time.Time  `json:"indexedAt"`
}

type TraceEmbeddingRepository struct {
	db *database.PostgresDB
}

func NewTraceEmbeddingRepository(db *database.PostgresDB) *TraceEmbeddingRepository {
	return &TraceEmbeddingRepository{db: db}
}

func (r *TraceEmbeddingRepository) SaveEmbedding(ctx context.Context, record *TraceEmbeddingRecord) error {
	query := `
		INSERT INTO trace_embeddings (id, project_id, trace_id, observation_id,
			content_type, content_hash, embedding_model, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		record.ID,
		record.ProjectID,
		record.TraceID,
		record.ObservationID,
		record.ContentType,
		record.ContentHash,
		record.EmbeddingModel,
		record.IndexedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save trace embedding: %w", err)
	}

	return nil
}

func (r *TraceEmbeddingRepository) ListEmbeddings(ctx context.Context, projectID uuid.UUID) ([]TraceEmbeddingRecord, error) {
	query := `
		SELECT id, project_id, trace_id, observation_id,
			content_type, content_hash, embedding_model, indexed_at
		FROM trace_embeddings
		WHERE project_id = $1
		ORDER BY indexed_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list trace embeddings: %w", err)
	}
	defer rows.Close()

	var records []TraceEmbeddingRecord
	for rows.Next() {
		var rec TraceEmbeddingRecord
		if err := rows.Scan(
			&rec.ID, &rec.ProjectID, &rec.TraceID, &rec.ObservationID,
			&rec.ContentType, &rec.ContentHash, &rec.EmbeddingModel, &rec.IndexedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trace embedding: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}

func (r *TraceEmbeddingRepository) SaveCluster(ctx context.Context, cluster *domain.TraceCluster) error {
	commonPatternsJSON, err := json.Marshal(cluster.CommonPatterns)
	if err != nil {
		return fmt.Errorf("failed to marshal common patterns: %w", err)
	}

	representativeIDsJSON, err := json.Marshal(cluster.RepresentativeTraceIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal representative trace IDs: %w", err)
	}

	query := `
		INSERT INTO trace_clusters (id, label, description, trace_count,
			common_patterns, avg_score, representative_trace_ids)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		cluster.ID,
		cluster.Label,
		cluster.Description,
		cluster.TraceCount,
		commonPatternsJSON,
		cluster.AvgScore,
		representativeIDsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save trace cluster: %w", err)
	}

	return nil
}

func (r *TraceEmbeddingRepository) ListClusters(ctx context.Context, projectID uuid.UUID) ([]domain.TraceCluster, error) {
	query := `
		SELECT id, label, description, trace_count,
			common_patterns, avg_score, representative_trace_ids
		FROM trace_clusters
		WHERE project_id = $1
		ORDER BY trace_count DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list trace clusters: %w", err)
	}
	defer rows.Close()

	var clusters []domain.TraceCluster
	for rows.Next() {
		var c domain.TraceCluster
		var commonPatternsJSON, representativeIDsJSON []byte

		if err := rows.Scan(
			&c.ID, &c.Label, &c.Description, &c.TraceCount,
			&commonPatternsJSON, &c.AvgScore, &representativeIDsJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trace cluster: %w", err)
		}

		if len(commonPatternsJSON) > 0 {
			if err := json.Unmarshal(commonPatternsJSON, &c.CommonPatterns); err != nil {
				return nil, fmt.Errorf("failed to unmarshal common patterns: %w", err)
			}
		}
		if len(representativeIDsJSON) > 0 {
			if err := json.Unmarshal(representativeIDsJSON, &c.RepresentativeTraceIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal representative trace IDs: %w", err)
			}
		}

		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (r *TraceEmbeddingRepository) GetClusterByID(ctx context.Context, id uuid.UUID) (*domain.TraceCluster, error) {
	query := `
		SELECT id, label, description, trace_count,
			common_patterns, avg_score, representative_trace_ids
		FROM trace_clusters
		WHERE id = $1
	`

	var c domain.TraceCluster
	var commonPatternsJSON, representativeIDsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Label, &c.Description, &c.TraceCount,
		&commonPatternsJSON, &c.AvgScore, &representativeIDsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("trace cluster")
		}
		return nil, fmt.Errorf("failed to get trace cluster: %w", err)
	}

	if len(commonPatternsJSON) > 0 {
		if err := json.Unmarshal(commonPatternsJSON, &c.CommonPatterns); err != nil {
			return nil, fmt.Errorf("failed to unmarshal common patterns: %w", err)
		}
	}
	if len(representativeIDsJSON) > 0 {
		if err := json.Unmarshal(representativeIDsJSON, &c.RepresentativeTraceIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal representative trace IDs: %w", err)
		}
	}

	return &c, nil
}

func (r *TraceEmbeddingRepository) CountEmbeddings(ctx context.Context, projectID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM trace_embeddings WHERE project_id = $1`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count embeddings: %w", err)
	}

	return count, nil
}
