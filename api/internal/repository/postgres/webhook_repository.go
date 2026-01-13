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

// WebhookRepository handles webhook data operations in PostgreSQL
type WebhookRepository struct {
	db *database.PostgresDB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *database.PostgresDB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create creates a new webhook
func (r *WebhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		INSERT INTO webhooks (
			id, project_id, type, name, url, secret, events, is_enabled, headers,
			cost_threshold, latency_threshold, score_threshold, rate_limit_per_hour,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = r.db.Pool.Exec(ctx, query,
		webhook.ID,
		webhook.ProjectID,
		string(webhook.Type),
		webhook.Name,
		webhook.URL,
		webhook.Secret,
		eventsJSON,
		webhook.IsEnabled,
		headersJSON,
		webhook.CostThreshold,
		webhook.LatencyThreshold,
		webhook.ScoreThreshold,
		webhook.RateLimitPerHour,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}

// GetByID retrieves a webhook by ID
func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	query := `
		SELECT id, project_id, type, name, url, secret, events, is_enabled, headers,
			cost_threshold, latency_threshold, score_threshold, rate_limit_per_hour,
			last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE id = $1
	`

	var webhook domain.Webhook
	var eventsJSON, headersJSON []byte
	var webhookType string

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID,
		&webhook.ProjectID,
		&webhookType,
		&webhook.Name,
		&webhook.URL,
		&webhook.Secret,
		&eventsJSON,
		&webhook.IsEnabled,
		&headersJSON,
		&webhook.CostThreshold,
		&webhook.LatencyThreshold,
		&webhook.ScoreThreshold,
		&webhook.RateLimitPerHour,
		&webhook.LastTriggeredAt,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("Webhook")
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Type = domain.WebhookType(webhookType)

	if err := json.Unmarshal(eventsJSON, &webhook.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	return &webhook, nil
}

// GetByProjectID retrieves a webhook by project ID and webhook ID
func (r *WebhookRepository) GetByProjectID(ctx context.Context, projectID, webhookID uuid.UUID) (*domain.Webhook, error) {
	query := `
		SELECT id, project_id, type, name, url, secret, events, is_enabled, headers,
			cost_threshold, latency_threshold, score_threshold, rate_limit_per_hour,
			last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE id = $1 AND project_id = $2
	`

	var webhook domain.Webhook
	var eventsJSON, headersJSON []byte
	var webhookType string

	err := r.db.Pool.QueryRow(ctx, query, webhookID, projectID).Scan(
		&webhook.ID,
		&webhook.ProjectID,
		&webhookType,
		&webhook.Name,
		&webhook.URL,
		&webhook.Secret,
		&eventsJSON,
		&webhook.IsEnabled,
		&headersJSON,
		&webhook.CostThreshold,
		&webhook.LatencyThreshold,
		&webhook.ScoreThreshold,
		&webhook.RateLimitPerHour,
		&webhook.LastTriggeredAt,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("Webhook")
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Type = domain.WebhookType(webhookType)

	if err := json.Unmarshal(eventsJSON, &webhook.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}

	return &webhook, nil
}

// Update updates a webhook
func (r *WebhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		UPDATE webhooks
		SET type = $2, name = $3, url = $4, secret = $5, events = $6, is_enabled = $7,
			headers = $8, cost_threshold = $9, latency_threshold = $10, score_threshold = $11,
			rate_limit_per_hour = $12, updated_at = NOW()
		WHERE id = $1
	`

	_, err = r.db.Pool.Exec(ctx, query,
		webhook.ID,
		string(webhook.Type),
		webhook.Name,
		webhook.URL,
		webhook.Secret,
		eventsJSON,
		webhook.IsEnabled,
		headersJSON,
		webhook.CostThreshold,
		webhook.LatencyThreshold,
		webhook.ScoreThreshold,
		webhook.RateLimitPerHour,
	)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	return nil
}

// Delete deletes a webhook
func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return nil
}

// List retrieves webhooks with filtering
func (r *WebhookRepository) List(ctx context.Context, filter *domain.WebhookFilter, limit, offset int) (*domain.WebhookList, error) {
	// Build query conditions
	conditions := "project_id = $1"
	args := []interface{}{filter.ProjectID}
	argIndex := 2

	if filter.Type != nil {
		conditions += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, string(*filter.Type))
		argIndex++
	}

	if filter.IsEnabled != nil {
		conditions += fmt.Sprintf(" AND is_enabled = $%d", argIndex)
		args = append(args, *filter.IsEnabled)
		argIndex++
	}

	if filter.EventType != nil {
		conditions += fmt.Sprintf(" AND events @> $%d::jsonb", argIndex)
		eventJSON, _ := json.Marshal([]domain.EventType{*filter.EventType})
		args = append(args, eventJSON)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM webhooks WHERE %s", conditions)
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count webhooks: %w", err)
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT id, project_id, type, name, url, secret, events, is_enabled, headers,
			cost_threshold, latency_threshold, score_threshold, rate_limit_per_hour,
			last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, conditions, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []domain.Webhook
	for rows.Next() {
		var webhook domain.Webhook
		var eventsJSON, headersJSON []byte
		var webhookType string

		if err := rows.Scan(
			&webhook.ID,
			&webhook.ProjectID,
			&webhookType,
			&webhook.Name,
			&webhook.URL,
			&webhook.Secret,
			&eventsJSON,
			&webhook.IsEnabled,
			&headersJSON,
			&webhook.CostThreshold,
			&webhook.LatencyThreshold,
			&webhook.ScoreThreshold,
			&webhook.RateLimitPerHour,
			&webhook.LastTriggeredAt,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		webhook.Type = domain.WebhookType(webhookType)

		if err := json.Unmarshal(eventsJSON, &webhook.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
			}
		}

		webhooks = append(webhooks, webhook)
	}

	return &domain.WebhookList{
		Webhooks:   webhooks,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(webhooks)) < totalCount,
	}, nil
}

// ListByProjectID retrieves all webhooks for a project
func (r *WebhookRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]domain.Webhook, error) {
	filter := &domain.WebhookFilter{ProjectID: projectID}
	result, err := r.List(ctx, filter, 1000, 0)
	if err != nil {
		return nil, err
	}
	return result.Webhooks, nil
}

// ListEnabledByEvent retrieves enabled webhooks that subscribe to a specific event
func (r *WebhookRepository) ListEnabledByEvent(ctx context.Context, projectID uuid.UUID, eventType domain.EventType) ([]domain.Webhook, error) {
	eventJSON, _ := json.Marshal([]domain.EventType{eventType})

	query := `
		SELECT id, project_id, type, name, url, secret, events, is_enabled, headers,
			cost_threshold, latency_threshold, score_threshold, rate_limit_per_hour,
			last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE project_id = $1 AND is_enabled = true AND events @> $2::jsonb
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, projectID, eventJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks by event: %w", err)
	}
	defer rows.Close()

	var webhooks []domain.Webhook
	for rows.Next() {
		var webhook domain.Webhook
		var eventsJSON, headersJSON []byte
		var webhookType string

		if err := rows.Scan(
			&webhook.ID,
			&webhook.ProjectID,
			&webhookType,
			&webhook.Name,
			&webhook.URL,
			&webhook.Secret,
			&eventsJSON,
			&webhook.IsEnabled,
			&headersJSON,
			&webhook.CostThreshold,
			&webhook.LatencyThreshold,
			&webhook.ScoreThreshold,
			&webhook.RateLimitPerHour,
			&webhook.LastTriggeredAt,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		webhook.Type = domain.WebhookType(webhookType)

		if err := json.Unmarshal(eventsJSON, &webhook.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
			}
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

// UpdateLastTriggered updates the last triggered timestamp
func (r *WebhookRepository) UpdateLastTriggered(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE webhooks SET last_triggered_at = $2, updated_at = NOW() WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last triggered: %w", err)
	}

	return nil
}

// CreateDelivery creates a webhook delivery record
func (r *WebhookRepository) CreateDelivery(ctx context.Context, delivery *domain.WebhookDelivery) error {
	query := `
		INSERT INTO webhook_deliveries (
			id, webhook_id, event_type, payload, status_code, response,
			success, error, duration_ms, retry_count, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		delivery.ID,
		delivery.WebhookID,
		string(delivery.EventType),
		delivery.Payload,
		delivery.StatusCode,
		delivery.Response,
		delivery.Success,
		delivery.Error,
		delivery.Duration,
		delivery.RetryCount,
		delivery.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create delivery: %w", err)
	}

	return nil
}

// ListDeliveries retrieves deliveries for a webhook
func (r *WebhookRepository) ListDeliveries(ctx context.Context, filter *domain.WebhookDeliveryFilter, limit, offset int) (*domain.WebhookDeliveryList, error) {
	conditions := "webhook_id = $1"
	args := []interface{}{filter.WebhookID}
	argIndex := 2

	if filter.EventType != nil {
		conditions += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, string(*filter.EventType))
		argIndex++
	}

	if filter.Success != nil {
		conditions += fmt.Sprintf(" AND success = $%d", argIndex)
		args = append(args, *filter.Success)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM webhook_deliveries WHERE %s", conditions)
	var totalCount int64
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count deliveries: %w", err)
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT id, webhook_id, event_type, payload, status_code, response,
			success, error, duration_ms, retry_count, created_at
		FROM webhook_deliveries
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, conditions, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []domain.WebhookDelivery
	for rows.Next() {
		var delivery domain.WebhookDelivery
		var eventType string

		if err := rows.Scan(
			&delivery.ID,
			&delivery.WebhookID,
			&eventType,
			&delivery.Payload,
			&delivery.StatusCode,
			&delivery.Response,
			&delivery.Success,
			&delivery.Error,
			&delivery.Duration,
			&delivery.RetryCount,
			&delivery.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delivery: %w", err)
		}

		delivery.EventType = domain.EventType(eventType)
		deliveries = append(deliveries, delivery)
	}

	return &domain.WebhookDeliveryList{
		Deliveries: deliveries,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(deliveries)) < totalCount,
	}, nil
}

// CheckRateLimit checks if webhook can send based on rate limit
func (r *WebhookRepository) CheckRateLimit(ctx context.Context, webhookID uuid.UUID, limitPerHour *int) (bool, error) {
	if limitPerHour == nil {
		return true, nil
	}

	query := `SELECT check_webhook_rate_limit($1, $2)`
	var canSend bool
	err := r.db.Pool.QueryRow(ctx, query, webhookID, *limitPerHour).Scan(&canSend)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return canSend, nil
}

// IncrementDeliveryCount increments the delivery count for rate limiting
func (r *WebhookRepository) IncrementDeliveryCount(ctx context.Context, webhookID uuid.UUID) error {
	query := `SELECT increment_webhook_delivery_count($1)`
	_, err := r.db.Pool.Exec(ctx, query, webhookID)
	if err != nil {
		return fmt.Errorf("failed to increment delivery count: %w", err)
	}

	return nil
}
