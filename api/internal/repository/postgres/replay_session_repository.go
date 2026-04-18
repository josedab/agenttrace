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

type ReplaySessionRepository struct {
	db *database.PostgresDB
}

func NewReplaySessionRepository(db *database.PostgresDB) *ReplaySessionRepository {
	return &ReplaySessionRepository{db: db}
}

func (r *ReplaySessionRepository) Save(ctx context.Context, session *domain.AgentReplaySession) error {
	query := `
		INSERT INTO replay_sessions (id, project_id, trace_id, name, description, status, recording_fidelity,
			total_events, total_duration_ms, files_tracked, checkpoint_count,
			parent_session_id, branch_point, is_public, share_url,
			created_at, updated_at, created_by, ended_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		session.ID,
		session.ProjectID,
		session.TraceID,
		session.Name,
		session.Description,
		session.Status,
		session.RecordingFidelity,
		session.TotalEvents,
		session.TotalDurationMs,
		session.FilesTracked,
		session.CheckpointCount,
		session.ParentSessionID,
		session.BranchPoint,
		session.IsPublic,
		session.ShareURL,
		session.CreatedAt,
		session.UpdatedAt,
		session.CreatedBy,
		session.EndedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save replay session: %w", err)
	}

	return nil
}

// GetByID retrieves a replay session within a project.
func (r *ReplaySessionRepository) GetByID(
	ctx context.Context,
	projectID, id uuid.UUID,
) (*domain.AgentReplaySession, error) {
	query := `
		SELECT id, project_id, trace_id, name, description, status, recording_fidelity,
			total_events, total_duration_ms, files_tracked, checkpoint_count,
			parent_session_id, branch_point, is_public, share_url,
			created_at, updated_at, created_by, ended_at
		FROM replay_sessions
		WHERE project_id = $1 AND id = $2
	`

	var s domain.AgentReplaySession
	err := r.db.Pool.QueryRow(ctx, query, projectID, id).Scan(
		&s.ID, &s.ProjectID, &s.TraceID, &s.Name, &s.Description,
		&s.Status, &s.RecordingFidelity,
		&s.TotalEvents, &s.TotalDurationMs, &s.FilesTracked, &s.CheckpointCount,
		&s.ParentSessionID, &s.BranchPoint, &s.IsPublic, &s.ShareURL,
		&s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.EndedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("replay session")
		}
		return nil, fmt.Errorf("failed to get replay session: %w", err)
	}

	return &s, nil
}

// Update persists a project-scoped replay session.
func (r *ReplaySessionRepository) Update(
	ctx context.Context,
	session *domain.AgentReplaySession,
) error {
	query := `
		UPDATE replay_sessions
		SET name = $3, description = $4, status = $5, recording_fidelity = $6,
			total_events = $7, total_duration_ms = $8, files_tracked = $9,
			checkpoint_count = $10, parent_session_id = $11, branch_point = $12,
			is_public = $13, share_url = $14, updated_at = $15, ended_at = $16
		WHERE project_id = $1 AND id = $2
	`

	result, err := r.db.Pool.Exec(ctx, query,
		session.ProjectID,
		session.ID,
		session.Name,
		session.Description,
		session.Status,
		session.RecordingFidelity,
		session.TotalEvents,
		session.TotalDurationMs,
		session.FilesTracked,
		session.CheckpointCount,
		session.ParentSessionID,
		session.BranchPoint,
		session.IsPublic,
		session.ShareURL,
		session.UpdatedAt,
		session.EndedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update replay session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.NotFound("replay session")
	}
	return nil
}

func (r *ReplaySessionRepository) List(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.AgentReplaySession, error) {
	query := `
		SELECT id, project_id, trace_id, name, description, status, recording_fidelity,
			total_events, total_duration_ms, files_tracked, checkpoint_count,
			parent_session_id, branch_point, is_public, share_url,
			created_at, updated_at, created_by, ended_at
		FROM replay_sessions
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list replay sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.AgentReplaySession
	for rows.Next() {
		var s domain.AgentReplaySession
		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.TraceID, &s.Name, &s.Description,
			&s.Status, &s.RecordingFidelity,
			&s.TotalEvents, &s.TotalDurationMs, &s.FilesTracked, &s.CheckpointCount,
			&s.ParentSessionID, &s.BranchPoint, &s.IsPublic, &s.ShareURL,
			&s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.EndedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan replay session: %w", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *ReplaySessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE replay_sessions SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update replay session status: %w", err)
	}

	return nil
}

func (r *ReplaySessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM replay_sessions WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete replay session: %w", err)
	}

	return nil
}

func (r *ReplaySessionRepository) SaveEvent(ctx context.Context, event *domain.AgentReplayTimelineEvent) error {
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	inputJSON, err := json.Marshal(event.Input)
	if err != nil {
		return fmt.Errorf("failed to marshal event input: %w", err)
	}

	outputJSON, err := json.Marshal(event.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal event output: %w", err)
	}

	var fileDeltaJSON []byte
	if event.FileDelta != nil {
		fileDeltaJSON, err = json.Marshal(event.FileDelta)
		if err != nil {
			return fmt.Errorf("failed to marshal file delta: %w", err)
		}
	}

	query := `
		INSERT INTO replay_events (id, session_id, event_index, event_type, timestamp,
			data, input, output, duration_ms, observation_id, file_delta)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		event.ID,
		event.SessionID,
		event.Index,
		event.Type,
		event.Timestamp,
		dataJSON,
		inputJSON,
		outputJSON,
		event.DurationMs,
		event.ObservationID,
		fileDeltaJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save replay event: %w", err)
	}

	return nil
}

func (r *ReplaySessionRepository) ListEvents(ctx context.Context, sessionID uuid.UUID) ([]domain.AgentReplayTimelineEvent, error) {
	query := `
		SELECT id, session_id, event_index, event_type, timestamp,
			data, input, output, duration_ms, observation_id, file_delta
		FROM replay_events
		WHERE session_id = $1
		ORDER BY event_index ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list replay events: %w", err)
	}
	defer rows.Close()

	var events []domain.AgentReplayTimelineEvent
	for rows.Next() {
		var e domain.AgentReplayTimelineEvent
		var dataJSON, inputJSON, outputJSON, fileDeltaJSON []byte

		if err := rows.Scan(
			&e.ID, &e.SessionID, &e.Index, &e.Type, &e.Timestamp,
			&dataJSON, &inputJSON, &outputJSON,
			&e.DurationMs, &e.ObservationID, &fileDeltaJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan replay event: %w", err)
		}

		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &e.Data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
			}
		}
		if len(inputJSON) > 0 {
			if err := json.Unmarshal(inputJSON, &e.Input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event input: %w", err)
			}
		}
		if len(outputJSON) > 0 {
			if err := json.Unmarshal(outputJSON, &e.Output); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event output: %w", err)
			}
		}
		if len(fileDeltaJSON) > 0 {
			e.FileDelta = &domain.ReplayFileDelta{}
			if err := json.Unmarshal(fileDeltaJSON, e.FileDelta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal file delta: %w", err)
			}
		}

		events = append(events, e)
	}

	return events, nil
}
