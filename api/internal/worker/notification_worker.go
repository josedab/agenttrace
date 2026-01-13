package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	pgrepo "github.com/agenttrace/agenttrace/api/internal/repository/postgres"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	// TypeNotificationSend is the task type for sending notifications
	TypeNotificationSend = "notification:send"

	// TypeDailyCostReport is the task type for daily cost report notifications
	TypeDailyCostReport = "notification:daily_cost_report"

	// TypeCheckThresholds is the task type for checking thresholds
	TypeCheckThresholds = "notification:check_thresholds"
)

// NotificationPayload represents the payload for notification tasks
type NotificationPayload struct {
	WebhookID  string           `json:"webhookId"`
	EventType  domain.EventType `json:"eventType"`
	Data       map[string]any   `json:"data"`
	RetryCount int              `json:"retryCount"`
}

// DailyCostReportPayload represents the payload for daily cost reports
type DailyCostReportPayload struct {
	ProjectID string `json:"projectId"`
	Date      string `json:"date"`
}

// ThresholdCheckPayload represents the payload for threshold checks
type ThresholdCheckPayload struct {
	TraceID   string `json:"traceId"`
	ProjectID string `json:"projectId"`
}

// NotificationWorker handles notification-related background tasks
type NotificationWorker struct {
	logger              *zap.Logger
	notificationService *service.NotificationService
	webhookRepo         *pgrepo.WebhookRepository
	queryService        *service.QueryService
	asynqClient         *asynq.Client
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(
	logger *zap.Logger,
	notificationService *service.NotificationService,
	webhookRepo *pgrepo.WebhookRepository,
	queryService *service.QueryService,
	asynqClient *asynq.Client,
) *NotificationWorker {
	return &NotificationWorker{
		logger:              logger.Named("notification_worker"),
		notificationService: notificationService,
		webhookRepo:         webhookRepo,
		queryService:        queryService,
		asynqClient:         asynqClient,
	}
}

// RegisterHandlers registers all notification task handlers
func (w *NotificationWorker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeNotificationSend, w.HandleNotificationSend)
	mux.HandleFunc(TypeDailyCostReport, w.HandleDailyCostReport)
	mux.HandleFunc(TypeCheckThresholds, w.HandleCheckThresholds)
}

// HandleNotificationSend handles sending a notification to a webhook
func (w *NotificationWorker) HandleNotificationSend(ctx context.Context, t *asynq.Task) error {
	var payload NotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing notification task",
		zap.String("webhookId", payload.WebhookID),
		zap.String("eventType", string(payload.EventType)),
		zap.Int("retryCount", payload.RetryCount),
	)

	// Parse webhook ID
	webhookID, err := uuid.Parse(payload.WebhookID)
	if err != nil {
		w.logger.Error("invalid webhook ID", zap.String("webhookId", payload.WebhookID), zap.Error(err))
		return fmt.Errorf("invalid webhook ID: %w", err)
	}

	// Check if webhook repository is configured
	if w.webhookRepo == nil {
		w.logger.Error("webhook repository not configured")
		return fmt.Errorf("webhook repository not configured")
	}

	// Fetch webhook from database
	webhook, err := w.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		w.logger.Error("failed to fetch webhook", zap.String("webhookId", payload.WebhookID), zap.Error(err))
		return fmt.Errorf("failed to fetch webhook: %w", err)
	}

	// Check if webhook is enabled
	if !webhook.IsEnabled {
		w.logger.Info("webhook is disabled, skipping", zap.String("webhookId", payload.WebhookID))
		return nil
	}

	// Check rate limit
	canSend, err := w.webhookRepo.CheckRateLimit(ctx, webhookID, webhook.RateLimitPerHour)
	if err != nil {
		w.logger.Warn("failed to check rate limit", zap.String("webhookId", payload.WebhookID), zap.Error(err))
		// Continue anyway - don't block on rate limit check failure
	} else if !canSend {
		w.logger.Info("rate limit exceeded for webhook", zap.String("webhookId", payload.WebhookID))
		return nil // Don't retry - rate limited
	}

	// Send the notification
	delivery, err := w.notificationService.SendNotification(ctx, webhook, payload.EventType, payload.Data)
	if err != nil {
		w.logger.Error("failed to send notification",
			zap.String("webhookId", payload.WebhookID),
			zap.Error(err),
		)
	} else {
		w.logger.Info("notification sent successfully",
			zap.String("webhookId", payload.WebhookID),
			zap.String("eventType", string(payload.EventType)),
			zap.Int("statusCode", delivery.StatusCode),
			zap.Int64("durationMs", delivery.Duration),
		)
	}

	// Store delivery record
	if delivery != nil {
		delivery.RetryCount = payload.RetryCount
		if storeErr := w.webhookRepo.CreateDelivery(ctx, delivery); storeErr != nil {
			w.logger.Warn("failed to store delivery record",
				zap.String("webhookId", payload.WebhookID),
				zap.Error(storeErr),
			)
		}
	}

	// Increment rate limit counter on successful send
	if delivery != nil && delivery.Success {
		if incErr := w.webhookRepo.IncrementDeliveryCount(ctx, webhookID); incErr != nil {
			w.logger.Warn("failed to increment delivery count", zap.Error(incErr))
		}
		// Update last triggered timestamp
		if updateErr := w.webhookRepo.UpdateLastTriggered(ctx, webhookID); updateErr != nil {
			w.logger.Warn("failed to update last triggered", zap.Error(updateErr))
		}
	}

	// Return error to trigger retry if delivery failed
	if err != nil {
		return err
	}

	return nil
}

// HandleDailyCostReport generates and sends daily cost reports
func (w *NotificationWorker) HandleDailyCostReport(ctx context.Context, t *asynq.Task) error {
	var payload DailyCostReportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing daily cost report",
		zap.String("projectId", payload.ProjectID),
		zap.String("date", payload.Date),
	)

	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Find all webhooks subscribed to daily cost report events
	webhooks, err := w.webhookRepo.ListEnabledByEvent(ctx, projectID, domain.EventTypeDailyCostReport)
	if err != nil {
		w.logger.Error("failed to list webhooks for daily cost report", zap.Error(err))
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		w.logger.Debug("no webhooks subscribed to daily cost report", zap.String("projectId", payload.ProjectID))
		return nil
	}

	// Build cost report data
	// In a real implementation, this would query cost data from ClickHouse
	// For now, send placeholder data
	reportData := map[string]any{
		"projectId":  payload.ProjectID,
		"date":       payload.Date,
		"totalCost":  0.0,
		"traceCount": 0,
		"topModels":  []map[string]any{},
	}

	// Queue notifications for each webhook
	for _, webhook := range webhooks {
		if err := EnqueueNotification(w.asynqClient, webhook.ID.String(), domain.EventTypeDailyCostReport, reportData); err != nil {
			w.logger.Warn("failed to queue daily cost report notification",
				zap.String("webhookId", webhook.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// HandleCheckThresholds checks if a trace exceeded any thresholds
func (w *NotificationWorker) HandleCheckThresholds(ctx context.Context, t *asynq.Task) error {
	var payload ThresholdCheckPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Checking thresholds for trace",
		zap.String("traceId", payload.TraceID),
		zap.String("projectId", payload.ProjectID),
	)

	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Fetch the trace from the query service
	if w.queryService == nil {
		w.logger.Warn("query service not configured, skipping threshold check")
		return nil
	}

	trace, err := w.queryService.GetTrace(ctx, projectID, payload.TraceID)
	if err != nil {
		w.logger.Error("failed to fetch trace for threshold check",
			zap.String("traceId", payload.TraceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to fetch trace: %w", err)
	}

	if trace == nil {
		w.logger.Warn("trace not found for threshold check",
			zap.String("traceId", payload.TraceID),
		)
		return nil
	}

	// Get all webhooks for the project with threshold-based events
	costWebhooks, err := w.webhookRepo.ListEnabledByEvent(ctx, projectID, domain.EventTypeTraceCostThreshold)
	if err != nil {
		w.logger.Warn("failed to list cost threshold webhooks", zap.Error(err))
		costWebhooks = nil
	}

	latencyWebhooks, err := w.webhookRepo.ListEnabledByEvent(ctx, projectID, domain.EventTypeTraceLatencyThreshold)
	if err != nil {
		w.logger.Warn("failed to list latency threshold webhooks", zap.Error(err))
		latencyWebhooks = nil
	}

	w.logger.Debug("found threshold webhooks",
		zap.String("traceId", payload.TraceID),
		zap.Int("costWebhooks", len(costWebhooks)),
		zap.Int("latencyWebhooks", len(latencyWebhooks)),
		zap.Float64("traceCost", trace.TotalCost),
		zap.Float64("traceDurationMs", trace.DurationMs),
	)

	var notificationCount int

	// Check cost thresholds
	for _, webhook := range costWebhooks {
		if webhook.CostThreshold != nil && trace.TotalCost > *webhook.CostThreshold {
			w.logger.Info("trace exceeded cost threshold",
				zap.String("traceId", payload.TraceID),
				zap.Float64("traceCost", trace.TotalCost),
				zap.Float64("threshold", *webhook.CostThreshold),
				zap.String("webhookId", webhook.ID.String()),
			)

			if err := EnqueueNotification(w.asynqClient, webhook.ID.String(), domain.EventTypeTraceCostThreshold, map[string]any{
				"traceId":   payload.TraceID,
				"traceName": trace.Name,
				"cost":      trace.TotalCost,
				"threshold": *webhook.CostThreshold,
				"projectId": payload.ProjectID,
				"timestamp": trace.StartTime.Format(time.RFC3339),
			}); err != nil {
				w.logger.Warn("failed to enqueue cost threshold notification",
					zap.String("webhookId", webhook.ID.String()),
					zap.Error(err),
				)
			} else {
				notificationCount++
			}
		}
	}

	// Check latency thresholds
	for _, webhook := range latencyWebhooks {
		if webhook.LatencyThreshold != nil && trace.DurationMs > float64(*webhook.LatencyThreshold) {
			w.logger.Info("trace exceeded latency threshold",
				zap.String("traceId", payload.TraceID),
				zap.Float64("traceDurationMs", trace.DurationMs),
				zap.Int64("threshold", *webhook.LatencyThreshold),
				zap.String("webhookId", webhook.ID.String()),
			)

			if err := EnqueueNotification(w.asynqClient, webhook.ID.String(), domain.EventTypeTraceLatencyThreshold, map[string]any{
				"traceId":    payload.TraceID,
				"traceName":  trace.Name,
				"latencyMs":  trace.DurationMs,
				"threshold":  *webhook.LatencyThreshold,
				"projectId":  payload.ProjectID,
				"timestamp":  trace.StartTime.Format(time.RFC3339),
			}); err != nil {
				w.logger.Warn("failed to enqueue latency threshold notification",
					zap.String("webhookId", webhook.ID.String()),
					zap.Error(err),
				)
			} else {
				notificationCount++
			}
		}
	}

	w.logger.Info("threshold check completed",
		zap.String("traceId", payload.TraceID),
		zap.Int("notificationsQueued", notificationCount),
	)

	return nil
}

// EnqueueNotification creates a task to send a notification
func EnqueueNotification(
	client *asynq.Client,
	webhookID string,
	eventType domain.EventType,
	data map[string]any,
) error {
	payload := NotificationPayload{
		WebhookID:  webhookID,
		EventType:  eventType,
		Data:       data,
		RetryCount: 0,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeNotificationSend, payloadBytes,
		asynq.MaxRetry(3),
		asynq.Queue("notifications"),
		asynq.Timeout(30*time.Second),
	)

	_, err = client.Enqueue(task)
	return err
}

// EnqueueDailyCostReport creates a task to generate daily cost reports
func EnqueueDailyCostReport(client *asynq.Client, projectID string, date string) error {
	payload := DailyCostReportPayload{
		ProjectID: projectID,
		Date:      date,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeDailyCostReport, payloadBytes,
		asynq.Queue("default"),
		asynq.Timeout(5*time.Minute),
	)

	_, err = client.Enqueue(task)
	return err
}

// EnqueueThresholdCheck creates a task to check thresholds for a trace
func EnqueueThresholdCheck(client *asynq.Client, traceID string, projectID string) error {
	payload := ThresholdCheckPayload{
		TraceID:   traceID,
		ProjectID: projectID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeCheckThresholds, payloadBytes,
		asynq.Queue("default"),
		asynq.Timeout(30*time.Second),
	)

	_, err = client.Enqueue(task)
	return err
}

// EnqueueNotificationForProject sends a notification to all webhooks subscribed to an event for a project
func EnqueueNotificationForProject(
	client *asynq.Client,
	webhookRepo *pgrepo.WebhookRepository,
	ctx context.Context,
	projectID uuid.UUID,
	eventType domain.EventType,
	data map[string]any,
) error {
	// Find all enabled webhooks subscribed to this event
	webhooks, err := webhookRepo.ListEnabledByEvent(ctx, projectID, eventType)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	// Queue notification for each webhook
	for _, webhook := range webhooks {
		if err := EnqueueNotification(client, webhook.ID.String(), eventType, data); err != nil {
			// Log but don't fail - we want to try all webhooks
			continue
		}
	}

	return nil
}
