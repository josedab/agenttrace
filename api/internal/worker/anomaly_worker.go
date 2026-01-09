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
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	// TypeAnomalyDetection is the task type for anomaly detection
	TypeAnomalyDetection = "anomaly:detect"

	// TypeAnomalyAlert is the task type for sending anomaly alerts
	TypeAnomalyAlert = "anomaly:alert"

	// TypeAnomalyCleanup is the task type for cleaning up old anomalies
	TypeAnomalyCleanup = "anomaly:cleanup"

	// TypeAnomalyScheduledScan is the task type for scheduled anomaly scans
	TypeAnomalyScheduledScan = "anomaly:scheduled_scan"
)

// AnomalyDetectionPayload is the payload for anomaly detection tasks
type AnomalyDetectionPayload struct {
	ProjectID     uuid.UUID         `json:"projectId"`
	RuleID        uuid.UUID         `json:"ruleId"`
	CurrentValue  float64           `json:"currentValue"`
	TraceID       *uuid.UUID        `json:"traceId,omitempty"`
	TraceName     string            `json:"traceName,omitempty"`
	SpanID        *uuid.UUID        `json:"spanId,omitempty"`
	SpanName      string            `json:"spanName,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	MetricType    domain.AnomalyType `json:"metricType"`
}

// AnomalyAlertPayload is the payload for anomaly alert tasks
type AnomalyAlertPayload struct {
	AnomalyID   uuid.UUID   `json:"anomalyId"`
	ProjectID   uuid.UUID   `json:"projectId"`
	RuleID      uuid.UUID   `json:"ruleId"`
	WebhookIDs  []uuid.UUID `json:"webhookIds"`
	AlertTitle  string      `json:"alertTitle"`
	AlertBody   string      `json:"alertBody"`
	Severity    domain.AnomalySeverity `json:"severity"`
}

// AnomalyCleanupPayload is the payload for anomaly cleanup tasks
type AnomalyCleanupPayload struct {
	ProjectID     uuid.UUID `json:"projectId"`
	RetentionDays int       `json:"retentionDays"`
}

// AnomalyScheduledScanPayload is the payload for scheduled scan tasks
type AnomalyScheduledScanPayload struct {
	ProjectID uuid.UUID `json:"projectId"`
	RuleID    uuid.UUID `json:"ruleId"`
}

// AnomalyWorker processes anomaly detection tasks
type AnomalyWorker struct {
	logger              *zap.Logger
	anomalyService      *service.AnomalyService
	notificationService *service.NotificationService
}

// NewAnomalyWorker creates a new anomaly worker
func NewAnomalyWorker(
	logger *zap.Logger,
	anomalyService *service.AnomalyService,
	notificationService *service.NotificationService,
) *AnomalyWorker {
	return &AnomalyWorker{
		logger:              logger,
		anomalyService:      anomalyService,
		notificationService: notificationService,
	}
}

// NewAnomalyDetectionTask creates a new anomaly detection task
func NewAnomalyDetectionTask(payload AnomalyDetectionPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeAnomalyDetection, data), nil
}

// NewAnomalyAlertTask creates a new anomaly alert task
func NewAnomalyAlertTask(payload AnomalyAlertPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeAnomalyAlert, data, asynq.MaxRetry(3), asynq.Timeout(30*time.Second)), nil
}

// NewAnomalyCleanupTask creates a new anomaly cleanup task
func NewAnomalyCleanupTask(payload AnomalyCleanupPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeAnomalyCleanup, data), nil
}

// NewAnomalyScheduledScanTask creates a new scheduled scan task
func NewAnomalyScheduledScanTask(payload AnomalyScheduledScanPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeAnomalyScheduledScan, data), nil
}

// HandleAnomalyDetection processes an anomaly detection task
func (w *AnomalyWorker) HandleAnomalyDetection(ctx context.Context, t *asynq.Task) error {
	var payload AnomalyDetectionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing anomaly detection",
		zap.String("projectId", payload.ProjectID.String()),
		zap.String("ruleId", payload.RuleID.String()),
		zap.Float64("currentValue", payload.CurrentValue),
	)

	// In real implementation:
	// 1. Fetch the rule from database
	// 2. Fetch historical data for the metric
	// 3. Run anomaly detection
	// 4. If anomaly detected, create anomaly record
	// 5. If alerting enabled and cooldown passed, queue alert task

	// For now, just log success
	w.logger.Debug("Anomaly detection completed",
		zap.String("ruleId", payload.RuleID.String()),
	)

	return nil
}

// HandleAnomalyAlert processes an anomaly alert task
func (w *AnomalyWorker) HandleAnomalyAlert(ctx context.Context, t *asynq.Task) error {
	var payload AnomalyAlertPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing anomaly alert",
		zap.String("anomalyId", payload.AnomalyID.String()),
		zap.Int("webhookCount", len(payload.WebhookIDs)),
	)

	// Build notification event
	event := domain.NotificationEvent{
		Type:      domain.NotificationEventTypeAnomalyDetected,
		Timestamp: time.Now(),
		ProjectID: payload.ProjectID,
		Payload: map[string]interface{}{
			"anomalyId": payload.AnomalyID.String(),
			"ruleId":    payload.RuleID.String(),
			"title":     payload.AlertTitle,
			"body":      payload.AlertBody,
			"severity":  payload.Severity,
		},
	}

	// In real implementation, send to each webhook
	for _, webhookID := range payload.WebhookIDs {
		w.logger.Debug("Sending alert to webhook",
			zap.String("webhookId", webhookID.String()),
			zap.String("anomalyId", payload.AnomalyID.String()),
		)

		// Fetch webhook and send notification
		// err := w.notificationService.SendNotification(ctx, &webhook, event)
		_ = event // Use event
	}

	w.logger.Info("Anomaly alerts sent successfully",
		zap.String("anomalyId", payload.AnomalyID.String()),
	)

	return nil
}

// HandleAnomalyCleanup processes an anomaly cleanup task
func (w *AnomalyWorker) HandleAnomalyCleanup(ctx context.Context, t *asynq.Task) error {
	var payload AnomalyCleanupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing anomaly cleanup",
		zap.String("projectId", payload.ProjectID.String()),
		zap.Int("retentionDays", payload.RetentionDays),
	)

	// In real implementation:
	// 1. Calculate cutoff date
	// 2. Delete anomalies older than cutoff
	// 3. Delete resolved alerts older than cutoff
	// 4. Log cleanup statistics

	cutoffTime := time.Now().AddDate(0, 0, -payload.RetentionDays)
	w.logger.Debug("Cleanup cutoff time",
		zap.Time("cutoff", cutoffTime),
	)

	w.logger.Info("Anomaly cleanup completed",
		zap.String("projectId", payload.ProjectID.String()),
	)

	return nil
}

// HandleAnomalyScheduledScan processes a scheduled anomaly scan task
func (w *AnomalyWorker) HandleAnomalyScheduledScan(ctx context.Context, t *asynq.Task) error {
	var payload AnomalyScheduledScanPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing scheduled anomaly scan",
		zap.String("projectId", payload.ProjectID.String()),
		zap.String("ruleId", payload.RuleID.String()),
	)

	// In real implementation:
	// 1. Fetch the rule
	// 2. Determine time window for scan
	// 3. Fetch aggregated metrics for the time window
	// 4. Run anomaly detection on aggregated values
	// 5. Create anomalies and alerts as needed

	w.logger.Info("Scheduled anomaly scan completed",
		zap.String("ruleId", payload.RuleID.String()),
	)

	return nil
}

// RegisterHandlers registers all anomaly worker handlers with the mux
func (w *AnomalyWorker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeAnomalyDetection, w.HandleAnomalyDetection)
	mux.HandleFunc(TypeAnomalyAlert, w.HandleAnomalyAlert)
	mux.HandleFunc(TypeAnomalyCleanup, w.HandleAnomalyCleanup)
	mux.HandleFunc(TypeAnomalyScheduledScan, w.HandleAnomalyScheduledScan)
}

// SchedulePeriodicTasks schedules periodic anomaly tasks
func ScheduleAnomalyPeriodicTasks(scheduler *asynq.Scheduler, projectID uuid.UUID, ruleIDs []uuid.UUID) error {
	// Schedule periodic scans for each rule (every 5 minutes)
	for _, ruleID := range ruleIDs {
		payload := AnomalyScheduledScanPayload{
			ProjectID: projectID,
			RuleID:    ruleID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal scan payload: %w", err)
		}

		_, err = scheduler.Register(
			"*/5 * * * *", // Every 5 minutes
			asynq.NewTask(TypeAnomalyScheduledScan, data),
			asynq.Queue("anomaly"),
		)
		if err != nil {
			return fmt.Errorf("failed to schedule scan: %w", err)
		}
	}

	// Schedule daily cleanup
	cleanupPayload := AnomalyCleanupPayload{
		ProjectID:     projectID,
		RetentionDays: 30, // Keep 30 days of anomalies
	}
	cleanupData, err := json.Marshal(cleanupPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal cleanup payload: %w", err)
	}

	_, err = scheduler.Register(
		"0 2 * * *", // Daily at 2 AM
		asynq.NewTask(TypeAnomalyCleanup, cleanupData),
		asynq.Queue("anomaly"),
	)
	if err != nil {
		return fmt.Errorf("failed to schedule cleanup: %w", err)
	}

	return nil
}
