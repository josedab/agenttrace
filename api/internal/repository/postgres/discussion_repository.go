package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// DiscussionRepository handles persistence for discussions and eval queues
type DiscussionRepository struct {
	db *database.PostgresDB
}

// NewDiscussionRepository creates a new repository
func NewDiscussionRepository(db *database.PostgresDB) *DiscussionRepository {
	return &DiscussionRepository{db: db}
}

// SaveThread saves a discussion thread
func (r *DiscussionRepository) SaveThread(ctx context.Context, projectID uuid.UUID, thread *domain.DiscussionThread) error {
	query := `
		INSERT INTO discussion_threads (id, project_id, trace_id, observation_id, title, status, created_by, created_by_name, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		thread.ID, projectID, thread.TraceID.String(), thread.ObservationID,
		thread.Title, thread.Status, thread.CreatedBy, thread.CreatedByName,
		thread.Tags, thread.CreatedAt, thread.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save discussion thread: %w", err)
	}

	for _, msg := range thread.Messages {
		msg.ThreadID = thread.ID
		if err := r.SaveMessage(ctx, &msg); err != nil {
			return err
		}
	}
	return nil
}

// SaveMessage saves a thread message
func (r *DiscussionRepository) SaveMessage(ctx context.Context, msg *domain.ThreadMessage) error {
	query := `
		INSERT INTO thread_messages (id, thread_id, user_id, user_name, content, mentions, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		msg.ID, msg.ThreadID, msg.UserID, msg.UserName, msg.Content, msg.Mentions, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save thread message: %w", err)
	}

	r.db.Pool.Exec(ctx, `UPDATE discussion_threads SET updated_at = $1 WHERE id = $2`, time.Now(), msg.ThreadID)
	return nil
}

// SaveEvalQueue saves an evaluation queue
func (r *DiscussionRepository) SaveEvalQueue(ctx context.Context, queue *domain.EvalQueue) error {
	query := `
		INSERT INTO eval_queues (id, project_id, name, description, assignees, trace_ids, status, total, completed, in_progress, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		queue.ID, queue.ProjectID, queue.Name, queue.Description,
		queue.Assignees, queue.TraceIDs, queue.Status,
		queue.Progress.Total, queue.Progress.Completed, queue.Progress.InProgress,
		queue.CreatedAt, queue.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save eval queue: %w", err)
	}
	return nil
}
