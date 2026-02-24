package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ReviewQueueRepository handles review queue data operations in PostgreSQL
type ReviewQueueRepository struct {
	db *database.PostgresDB
}

// NewReviewQueueRepository creates a new review queue repository
func NewReviewQueueRepository(db *database.PostgresDB) *ReviewQueueRepository {
	return &ReviewQueueRepository{db: db}
}

// SaveQueue creates a new review queue
func (r *ReviewQueueRepository) SaveQueue(ctx context.Context, queue *domain.ReviewQueue) error {
	filtersJSON, err := json.Marshal(queue.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	assignedToJSON, err := json.Marshal(queue.AssignedTo)
	if err != nil {
		return fmt.Errorf("failed to marshal assigned_to: %w", err)
	}

	query := `
		INSERT INTO review_queues (id, project_id, name, description, filters, assigned_to, pending_count, completed_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		queue.ID,
		queue.ProjectID,
		queue.Name,
		queue.Description,
		filtersJSON,
		assignedToJSON,
		queue.PendingCount,
		queue.CompletedCount,
		queue.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save review queue: %w", err)
	}

	return nil
}

// ListQueues retrieves review queues for a project
func (r *ReviewQueueRepository) ListQueues(ctx context.Context, projectID uuid.UUID) ([]domain.ReviewQueue, error) {
	query := `
		SELECT id, project_id, name, description, filters, assigned_to, pending_count, completed_count, created_at
		FROM review_queues
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list review queues: %w", err)
	}
	defer rows.Close()

	var queues []domain.ReviewQueue
	for rows.Next() {
		var q domain.ReviewQueue
		var filtersJSON, assignedToJSON []byte

		if err := rows.Scan(
			&q.ID,
			&q.ProjectID,
			&q.Name,
			&q.Description,
			&filtersJSON,
			&assignedToJSON,
			&q.PendingCount,
			&q.CompletedCount,
			&q.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan review queue: %w", err)
		}

		if len(filtersJSON) > 0 {
			if err := json.Unmarshal(filtersJSON, &q.Filters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
			}
		}
		if len(assignedToJSON) > 0 {
			if err := json.Unmarshal(assignedToJSON, &q.AssignedTo); err != nil {
				return nil, fmt.Errorf("failed to unmarshal assigned_to: %w", err)
			}
		}

		queues = append(queues, q)
	}

	return queues, nil
}

// SaveAssignment creates a new review assignment
func (r *ReviewQueueRepository) SaveAssignment(ctx context.Context, assignment *domain.ReviewAssignment) error {
	query := `
		INSERT INTO review_assignments (id, queue_id, trace_id, assigned_to, status, feedback, score, assigned_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		assignment.ID,
		assignment.QueueID,
		assignment.TraceID,
		assignment.AssignedTo,
		assignment.Status,
		assignment.Feedback,
		assignment.Score,
		assignment.AssignedAt,
		assignment.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save review assignment: %w", err)
	}

	return nil
}

// UpdateAssignment updates an existing review assignment
func (r *ReviewQueueRepository) UpdateAssignment(ctx context.Context, assignment *domain.ReviewAssignment) error {
	query := `
		UPDATE review_assignments
		SET status = $2, feedback = $3, score = $4, completed_at = $5
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		assignment.ID,
		assignment.Status,
		assignment.Feedback,
		assignment.Score,
		assignment.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update review assignment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("review assignment")
	}

	return nil
}

// SaveStandard creates a new quality standard
func (r *ReviewQueueRepository) SaveStandard(ctx context.Context, standard *domain.QualityStandard) error {
	rulesJSON, err := json.Marshal(standard.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		INSERT INTO quality_standards (id, project_id, name, enabled, rules, enforce_on_deploy, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		standard.ID,
		standard.ProjectID,
		standard.Name,
		standard.Enabled,
		rulesJSON,
		standard.EnforceOnDeploy,
		standard.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save quality standard: %w", err)
	}

	return nil
}

// ListStandards retrieves quality standards for a project
func (r *ReviewQueueRepository) ListStandards(ctx context.Context, projectID uuid.UUID) ([]domain.QualityStandard, error) {
	query := `
		SELECT id, project_id, name, enabled, rules, enforce_on_deploy, created_at
		FROM quality_standards
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list quality standards: %w", err)
	}
	defer rows.Close()

	var standards []domain.QualityStandard
	for rows.Next() {
		var s domain.QualityStandard
		var rulesJSON []byte

		if err := rows.Scan(
			&s.ID,
			&s.ProjectID,
			&s.Name,
			&s.Enabled,
			&rulesJSON,
			&s.EnforceOnDeploy,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan quality standard: %w", err)
		}

		if len(rulesJSON) > 0 {
			if err := json.Unmarshal(rulesJSON, &s.Rules); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
			}
		}

		standards = append(standards, s)
	}

	return standards, nil
}

// SaveActivity creates a new activity feed item
func (r *ReviewQueueRepository) SaveActivity(ctx context.Context, activity *domain.ActivityFeedItem) error {
	metadataJSON, err := json.Marshal(activity.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO activity_feed (id, project_id, type, user_id, user_name, description, resource_id, resource_type, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		activity.ID,
		activity.ProjectID,
		activity.Type,
		activity.UserID,
		activity.UserName,
		activity.Description,
		activity.ResourceID,
		activity.ResourceType,
		activity.Timestamp,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save activity feed item: %w", err)
	}

	return nil
}

// ListActivities retrieves activity feed items for a project
func (r *ReviewQueueRepository) ListActivities(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.ActivityFeedItem, error) {
	query := `
		SELECT id, project_id, type, user_id, user_name, description, resource_id, resource_type, timestamp, metadata
		FROM activity_feed
		WHERE project_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list activity feed items: %w", err)
	}
	defer rows.Close()

	var activities []domain.ActivityFeedItem
	for rows.Next() {
		var a domain.ActivityFeedItem
		var metadataJSON []byte

		if err := rows.Scan(
			&a.ID,
			&a.ProjectID,
			&a.Type,
			&a.UserID,
			&a.UserName,
			&a.Description,
			&a.ResourceID,
			&a.ResourceType,
			&a.Timestamp,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity feed item: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &a.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, a)
	}

	return activities, nil
}
