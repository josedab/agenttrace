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

// KnowledgeSnapshotRecord represents a knowledge graph snapshot stored in the database
type KnowledgeSnapshotRecord struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	Nodes       []domain.KGNode `json:"nodes"`
	Edges       []domain.KGEdge `json:"edges"`
	Stats       domain.KGStats  `json:"stats"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

type KnowledgeSnapshotRepository struct {
	db *database.PostgresDB
}

func NewKnowledgeSnapshotRepository(db *database.PostgresDB) *KnowledgeSnapshotRepository {
	return &KnowledgeSnapshotRepository{db: db}
}

func (r *KnowledgeSnapshotRepository) Save(ctx context.Context, snapshot *KnowledgeSnapshotRecord) error {
	nodesJSON, err := json.Marshal(snapshot.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(snapshot.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	statsJSON, err := json.Marshal(snapshot.Stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	query := `
		INSERT INTO agent_knowledge_snapshots (id, project_id, nodes, edges, stats, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		snapshot.ID,
		snapshot.ProjectID,
		nodesJSON,
		edgesJSON,
		statsJSON,
		snapshot.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save knowledge snapshot: %w", err)
	}

	return nil
}

func (r *KnowledgeSnapshotRepository) GetLatest(ctx context.Context, projectID uuid.UUID) (*KnowledgeSnapshotRecord, error) {
	query := `
		SELECT id, project_id, nodes, edges, stats, generated_at
		FROM agent_knowledge_snapshots
		WHERE project_id = $1
		ORDER BY generated_at DESC
		LIMIT 1
	`

	var s KnowledgeSnapshotRecord
	var nodesJSON, edgesJSON, statsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(
		&s.ID, &s.ProjectID, &nodesJSON, &edgesJSON, &statsJSON, &s.GeneratedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("knowledge snapshot")
		}
		return nil, fmt.Errorf("failed to get latest knowledge snapshot: %w", err)
	}

	if len(nodesJSON) > 0 {
		if err := json.Unmarshal(nodesJSON, &s.Nodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
		}
	}
	if len(edgesJSON) > 0 {
		if err := json.Unmarshal(edgesJSON, &s.Edges); err != nil {
			return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
		}
	}
	if len(statsJSON) > 0 {
		if err := json.Unmarshal(statsJSON, &s.Stats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
		}
	}

	return &s, nil
}

func (r *KnowledgeSnapshotRepository) List(ctx context.Context, projectID uuid.UUID, limit int) ([]KnowledgeSnapshotRecord, error) {
	query := `
		SELECT id, project_id, nodes, edges, stats, generated_at
		FROM agent_knowledge_snapshots
		WHERE project_id = $1
		ORDER BY generated_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list knowledge snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []KnowledgeSnapshotRecord
	for rows.Next() {
		var s KnowledgeSnapshotRecord
		var nodesJSON, edgesJSON, statsJSON []byte

		if err := rows.Scan(
			&s.ID, &s.ProjectID, &nodesJSON, &edgesJSON, &statsJSON, &s.GeneratedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan knowledge snapshot: %w", err)
		}

		if len(nodesJSON) > 0 {
			if err := json.Unmarshal(nodesJSON, &s.Nodes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
			}
		}
		if len(edgesJSON) > 0 {
			if err := json.Unmarshal(edgesJSON, &s.Edges); err != nil {
				return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
			}
		}
		if len(statsJSON) > 0 {
			if err := json.Unmarshal(statsJSON, &s.Stats); err != nil {
				return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
			}
		}

		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}
