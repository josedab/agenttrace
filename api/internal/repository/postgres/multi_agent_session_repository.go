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

type MultiAgentSessionRepository struct {
	db *database.PostgresDB
}

func NewMultiAgentSessionRepository(db *database.PostgresDB) *MultiAgentSessionRepository {
	return &MultiAgentSessionRepository{db: db}
}

func (r *MultiAgentSessionRepository) Save(ctx context.Context, session *domain.MultiAgentSession) error {
	agentsJSON, err := json.Marshal(session.Agents)
	if err != nil {
		return fmt.Errorf("failed to marshal agents: %w", err)
	}

	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	bottlenecksJSON, err := json.Marshal(session.Bottlenecks)
	if err != nil {
		return fmt.Errorf("failed to marshal bottlenecks: %w", err)
	}

	query := `
		INSERT INTO multi_agent_sessions (id, project_id, trace_id, name, topology,
			agents, messages, bottlenecks, status, start_time, end_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		session.ID,
		session.ProjectID,
		session.TraceID,
		session.Name,
		session.Topology,
		agentsJSON,
		messagesJSON,
		bottlenecksJSON,
		session.Status,
		session.StartTime,
		session.EndTime,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save multi-agent session: %w", err)
	}

	return nil
}

func (r *MultiAgentSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MultiAgentSession, error) {
	query := `
		SELECT id, project_id, trace_id, name, topology,
			agents, messages, bottlenecks, status, start_time, end_time, created_at
		FROM multi_agent_sessions
		WHERE id = $1
	`

	var s domain.MultiAgentSession
	var agentsJSON, messagesJSON, bottlenecksJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.ProjectID, &s.TraceID, &s.Name, &s.Topology,
		&agentsJSON, &messagesJSON, &bottlenecksJSON,
		&s.Status, &s.StartTime, &s.EndTime, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("multi-agent session")
		}
		return nil, fmt.Errorf("failed to get multi-agent session: %w", err)
	}

	if len(agentsJSON) > 0 {
		if err := json.Unmarshal(agentsJSON, &s.Agents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal agents: %w", err)
		}
	}
	if len(messagesJSON) > 0 {
		if err := json.Unmarshal(messagesJSON, &s.Messages); err != nil {
			return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
		}
	}
	if len(bottlenecksJSON) > 0 {
		if err := json.Unmarshal(bottlenecksJSON, &s.Bottlenecks); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bottlenecks: %w", err)
		}
	}

	return &s, nil
}

func (r *MultiAgentSessionRepository) List(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.MultiAgentSession, error) {
	query := `
		SELECT id, project_id, trace_id, name, topology,
			agents, messages, bottlenecks, status, start_time, end_time, created_at
		FROM multi_agent_sessions
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list multi-agent sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.MultiAgentSession
	for rows.Next() {
		var s domain.MultiAgentSession
		var agentsJSON, messagesJSON, bottlenecksJSON []byte

		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.TraceID, &s.Name, &s.Topology,
			&agentsJSON, &messagesJSON, &bottlenecksJSON,
			&s.Status, &s.StartTime, &s.EndTime, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan multi-agent session: %w", err)
		}

		if len(agentsJSON) > 0 {
			if err := json.Unmarshal(agentsJSON, &s.Agents); err != nil {
				return nil, fmt.Errorf("failed to unmarshal agents: %w", err)
			}
		}
		if len(messagesJSON) > 0 {
			if err := json.Unmarshal(messagesJSON, &s.Messages); err != nil {
				return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
			}
		}
		if len(bottlenecksJSON) > 0 {
			if err := json.Unmarshal(bottlenecksJSON, &s.Bottlenecks); err != nil {
				return nil, fmt.Errorf("failed to unmarshal bottlenecks: %w", err)
			}
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *MultiAgentSessionRepository) Update(ctx context.Context, session *domain.MultiAgentSession) error {
	agentsJSON, err := json.Marshal(session.Agents)
	if err != nil {
		return fmt.Errorf("failed to marshal agents: %w", err)
	}

	messagesJSON, err := json.Marshal(session.Messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	bottlenecksJSON, err := json.Marshal(session.Bottlenecks)
	if err != nil {
		return fmt.Errorf("failed to marshal bottlenecks: %w", err)
	}

	query := `
		UPDATE multi_agent_sessions
		SET name = $2, topology = $3, agents = $4, messages = $5,
			bottlenecks = $6, status = $7, start_time = $8, end_time = $9
		WHERE id = $1
	`

	_, err = r.db.Pool.Exec(ctx, query,
		session.ID,
		session.Name,
		session.Topology,
		agentsJSON,
		messagesJSON,
		bottlenecksJSON,
		session.Status,
		session.StartTime,
		session.EndTime,
	)
	if err != nil {
		return fmt.Errorf("failed to update multi-agent session: %w", err)
	}

	return nil
}

func (r *MultiAgentSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM multi_agent_sessions WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete multi-agent session: %w", err)
	}

	return nil
}

