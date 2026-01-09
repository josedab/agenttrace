package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
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
	WebhookID   string            `json:"webhookId"`
	EventType   domain.EventType  `json:"eventType"`
	Data        map[string]any    `json:"data"`
	RetryCount  int               `json:"retryCount"`
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
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(
	logger *zap.Logger,
	notificationService *service.NotificationService,
) *NotificationWorker {
	return &NotificationWorker{
		logger:              logger,
		notificationService: notificationService,
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
	)

	// In a real implementation, this would:
	// 1. Fetch the webhook configuration from database
	// 2. Check rate limits
	// 3. Send the notification
	// 4. Record the delivery in database

	// For now, just log that we would send the notification
	w.logger.Info("Would send notification",
		zap.String("webhookId", payload.WebhookID),
		zap.String("eventType", string(payload.EventType)),
		zap.Any("data", payload.Data),
	)

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

	// In a real implementation, this would:
	// 1. Query cost data for the project for the given date
	// 2. Find all webhooks subscribed to daily_cost_report events
	// 3. Send notifications to each webhook

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

	// In a real implementation, this would:
	// 1. Fetch the trace details
	// 2. Fetch all webhooks for the project with threshold events
	// 3. Check each threshold (cost, latency, score)
	// 4. Queue notifications for any exceeded thresholds

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
	)

	_, err = client.Enqueue(task)
	return err
}
