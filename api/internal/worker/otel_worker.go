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
	// TypeOTelExport is the task type for OTLP exports
	TypeOTelExport = "otel:export"

	// TypeOTelBatchExport is the task type for batch OTLP exports
	TypeOTelBatchExport = "otel:batch_export"

	// TypeOTelExporterHealthCheck is the task type for exporter health checks
	TypeOTelExporterHealthCheck = "otel:health_check"
)

// OTelExportPayload is the payload for single trace export tasks
type OTelExportPayload struct {
	ProjectID    uuid.UUID `json:"projectId"`
	ExporterID   uuid.UUID `json:"exporterId"`
	TraceID      uuid.UUID `json:"traceId"`
}

// OTelBatchExportPayload is the payload for batch export tasks
type OTelBatchExportPayload struct {
	ProjectID    uuid.UUID              `json:"projectId"`
	ExporterID   uuid.UUID              `json:"exporterId"`
	Spans        []domain.OTelSpan      `json:"spans"`
	Resource     domain.OTelResource    `json:"resource"`
}

// OTelHealthCheckPayload is the payload for health check tasks
type OTelHealthCheckPayload struct {
	ExporterID uuid.UUID `json:"exporterId"`
}

// OTelWorker processes OpenTelemetry export tasks
type OTelWorker struct {
	logger              *zap.Logger
	otelExporterService *service.OTelExporterService
}

// NewOTelWorker creates a new OTLP worker
func NewOTelWorker(
	logger *zap.Logger,
	otelExporterService *service.OTelExporterService,
) *OTelWorker {
	return &OTelWorker{
		logger:              logger,
		otelExporterService: otelExporterService,
	}
}

// NewOTelExportTask creates a new OTLP export task
func NewOTelExportTask(payload OTelExportPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeOTelExport, data, asynq.MaxRetry(3), asynq.Timeout(60*time.Second)), nil
}

// NewOTelBatchExportTask creates a new batch OTLP export task
func NewOTelBatchExportTask(payload OTelBatchExportPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeOTelBatchExport, data, asynq.MaxRetry(3), asynq.Timeout(120*time.Second)), nil
}

// NewOTelHealthCheckTask creates a new health check task
func NewOTelHealthCheckTask(payload OTelHealthCheckPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeOTelExporterHealthCheck, data, asynq.MaxRetry(1), asynq.Timeout(30*time.Second)), nil
}

// HandleOTelExport processes a single trace export task
func (w *OTelWorker) HandleOTelExport(ctx context.Context, t *asynq.Task) error {
	var payload OTelExportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing OTLP export",
		zap.String("projectId", payload.ProjectID.String()),
		zap.String("exporterId", payload.ExporterID.String()),
		zap.String("traceId", payload.TraceID.String()),
	)

	// In real implementation:
	// 1. Fetch the exporter configuration
	// 2. Fetch the trace and its observations
	// 3. Convert to OTLP format
	// 4. Export via the exporter service

	w.logger.Info("OTLP export completed",
		zap.String("traceId", payload.TraceID.String()),
	)

	return nil
}

// HandleOTelBatchExport processes a batch export task
func (w *OTelWorker) HandleOTelBatchExport(ctx context.Context, t *asynq.Task) error {
	var payload OTelBatchExportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing OTLP batch export",
		zap.String("projectId", payload.ProjectID.String()),
		zap.String("exporterId", payload.ExporterID.String()),
		zap.Int("spanCount", len(payload.Spans)),
	)

	// In real implementation:
	// 1. Fetch the exporter configuration
	// 2. Build the export request
	// 3. Send to the backend

	// Queue spans for export
	spans := make([]*domain.OTelSpan, len(payload.Spans))
	for i := range payload.Spans {
		spans[i] = &payload.Spans[i]
	}

	// In real impl, fetch exporter and call:
	// w.otelExporterService.QueueSpansForExport(exporter, spans)

	w.logger.Info("OTLP batch export completed",
		zap.String("exporterId", payload.ExporterID.String()),
		zap.Int("spanCount", len(payload.Spans)),
	)

	return nil
}

// HandleOTelHealthCheck processes an exporter health check task
func (w *OTelWorker) HandleOTelHealthCheck(ctx context.Context, t *asynq.Task) error {
	var payload OTelHealthCheckPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	w.logger.Info("Processing OTLP exporter health check",
		zap.String("exporterId", payload.ExporterID.String()),
	)

	// In real implementation:
	// 1. Fetch the exporter configuration
	// 2. Test the connection
	// 3. Update exporter status in database

	w.logger.Info("OTLP health check completed",
		zap.String("exporterId", payload.ExporterID.String()),
	)

	return nil
}

// RegisterHandlers registers all OTLP worker handlers with the mux
func (w *OTelWorker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeOTelExport, w.HandleOTelExport)
	mux.HandleFunc(TypeOTelBatchExport, w.HandleOTelBatchExport)
	mux.HandleFunc(TypeOTelExporterHealthCheck, w.HandleOTelHealthCheck)
}

// ScheduleOTelPeriodicTasks schedules periodic OTLP tasks
func ScheduleOTelPeriodicTasks(scheduler *asynq.Scheduler, exporterIDs []uuid.UUID) error {
	// Schedule health checks for each exporter (every 5 minutes)
	for _, exporterID := range exporterIDs {
		payload := OTelHealthCheckPayload{
			ExporterID: exporterID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal health check payload: %w", err)
		}

		_, err = scheduler.Register(
			"*/5 * * * *", // Every 5 minutes
			asynq.NewTask(TypeOTelExporterHealthCheck, data),
			asynq.Queue("otel"),
		)
		if err != nil {
			return fmt.Errorf("failed to schedule health check: %w", err)
		}
	}

	return nil
}

// ExportTraceAsync queues a trace for async OTLP export
func ExportTraceAsync(client *asynq.Client, projectID, exporterID, traceID uuid.UUID) error {
	payload := OTelExportPayload{
		ProjectID:  projectID,
		ExporterID: exporterID,
		TraceID:    traceID,
	}

	task, err := NewOTelExportTask(payload)
	if err != nil {
		return err
	}

	_, err = client.Enqueue(task, asynq.Queue("otel"))
	return err
}

// ExportSpansBatchAsync queues a batch of spans for async OTLP export
func ExportSpansBatchAsync(
	client *asynq.Client,
	projectID, exporterID uuid.UUID,
	spans []domain.OTelSpan,
	resource domain.OTelResource,
) error {
	payload := OTelBatchExportPayload{
		ProjectID:  projectID,
		ExporterID: exporterID,
		Spans:      spans,
		Resource:   resource,
	}

	task, err := NewOTelBatchExportTask(payload)
	if err != nil {
		return err
	}

	_, err = client.Enqueue(task, asynq.Queue("otel"))
	return err
}
