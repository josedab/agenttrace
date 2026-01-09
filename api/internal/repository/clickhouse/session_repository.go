package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// SessionRepository handles session data operations in ClickHouse
// Sessions are aggregated views from traces grouped by session_id
type SessionRepository struct {
	db *database.ClickHouseDB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *database.ClickHouseDB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Upsert is a no-op for sessions as they are derived from traces
func (r *SessionRepository) Upsert(ctx context.Context, session *domain.Session) error {
	// Sessions are aggregated from traces, no direct insert needed
	return nil
}

// GetByID retrieves a session by ID with aggregated metrics
func (r *SessionRepository) GetByID(ctx context.Context, projectID uuid.UUID, sessionID string) (*domain.Session, error) {
	query := `
		SELECT
			session_id as id,
			project_id,
			any(user_id) as user_id,
			max(bookmarked) as bookmarked,
			max(public) as public,
			min(created_at) as created_at,
			max(updated_at) as updated_at,
			count() as trace_count,
			sum(total_cost) as total_cost,
			sum(total_tokens) as total_tokens,
			min(start_time) as first_trace_time,
			max(start_time) as last_trace_time
		FROM traces FINAL
		WHERE project_id = ? AND session_id = ? AND session_id != ''
		GROUP BY session_id, project_id
	`

	var session domain.Session
	row := r.db.QueryRow(ctx, query, projectID, sessionID)
	err := row.Scan(
		&session.ID,
		&session.ProjectID,
		&session.UserID,
		&session.Bookmarked,
		&session.Public,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.TraceCount,
		&session.TotalCost,
		&session.TotalTokens,
		&session.FirstTraceTime,
		&session.LastTraceTime,
	)
	if err != nil {
		return nil, apperrors.NotFound("session not found")
	}

	return &session, nil
}

// List retrieves sessions with filtering and pagination
func (r *SessionRepository) List(ctx context.Context, filter *domain.SessionFilter, limit, offset int) (*domain.SessionList, error) {
	// Build WHERE clause
	conditions := []string{"project_id = ?", "session_id != ''"}
	args := []interface{}{filter.ProjectID}

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.FromTime != nil {
		conditions = append(conditions, "start_time >= ?")
		args = append(args, *filter.FromTime)
	}

	if filter.ToTime != nil {
		conditions = append(conditions, "start_time <= ?")
		args = append(args, *filter.ToTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get total count of distinct sessions
	countQuery := fmt.Sprintf(`
		SELECT count(DISTINCT session_id)
		FROM traces FINAL
		WHERE %s
	`, whereClause)

	var totalCount int64
	row := r.db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Get aggregated sessions
	query := fmt.Sprintf(`
		SELECT
			session_id as id,
			project_id,
			any(user_id) as user_id,
			max(bookmarked) as bookmarked,
			max(public) as public,
			min(created_at) as created_at,
			max(updated_at) as updated_at,
			count() as trace_count,
			sum(total_cost) as total_cost,
			sum(total_tokens) as total_tokens,
			min(start_time) as first_trace_time,
			max(start_time) as last_trace_time
		FROM traces FINAL
		WHERE %s
		GROUP BY session_id, project_id
		ORDER BY last_trace_time DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit+1, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var session domain.Session
		if err := rows.Scan(
			&session.ID,
			&session.ProjectID,
			&session.UserID,
			&session.Bookmarked,
			&session.Public,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.TraceCount,
			&session.TotalCost,
			&session.TotalTokens,
			&session.FirstTraceTime,
			&session.LastTraceTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	hasMore := len(sessions) > limit
	if hasMore {
		sessions = sessions[:limit]
	}

	return &domain.SessionList{
		Sessions:   sessions,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// GetDistinctUserIDs returns distinct user IDs from sessions for a project
func (r *SessionRepository) GetDistinctUserIDs(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT user_id
		FROM traces FINAL
		WHERE project_id = ? AND session_id != '' AND user_id != ''
		ORDER BY user_id
		LIMIT 1000
	`

	var userIDs []string
	if err := r.db.Select(ctx, &userIDs, query, projectID); err != nil {
		return nil, fmt.Errorf("failed to get distinct user IDs: %w", err)
	}

	return userIDs, nil
}
