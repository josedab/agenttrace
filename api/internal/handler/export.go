package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/worker"
)

// ExportHandler handles export endpoints
type ExportHandler struct {
	asynqClient *asynq.Client
	logger      *zap.Logger
}

// NewExportHandler creates a new export handler
func NewExportHandler(asynqClient *asynq.Client, logger *zap.Logger) *ExportHandler {
	return &ExportHandler{
		asynqClient: asynqClient,
		logger:      logger,
	}
}

// ExportTracesRequest represents a request to export traces
type ExportTracesRequest struct {
	Format      domain.ExportFormat `json:"format"`
	Type        string              `json:"type"` // traces, observations, scores
	Filters     map[string]any      `json:"filters,omitempty"`
	Destination *ExportDestination  `json:"destination,omitempty"`
}

// ExportDatasetRequest represents a request to export a dataset
type ExportDatasetRequest struct {
	DatasetID   string              `json:"datasetId"`
	Format      domain.ExportFormat `json:"format"`
	IncludeRuns bool                `json:"includeRuns,omitempty"`
	Destination *ExportDestination  `json:"destination,omitempty"`
}

// ExportDestination represents export destination configuration
type ExportDestination struct {
	Type   domain.DestinationType `json:"type,omitempty"`
	Bucket string                 `json:"bucket,omitempty"`
	Path   string                 `json:"path,omitempty"`
	Config map[string]string      `json:"config,omitempty"`
}

// ExportResponse represents an export job response
type ExportResponse struct {
	JobID   string `json:"jobId"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ExportData handles POST /v1/export/data
func (h *ExportHandler) ExportData(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request ExportTracesRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if request.Type == "" {
		request.Type = "traces"
	}
	if request.Type != "traces" && request.Type != "observations" && request.Type != "scores" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid export type. Valid types: traces, observations, scores",
		})
	}

	if request.Format == "" {
		request.Format = domain.ExportFormatJSON
	}
	if !request.Format.IsValid() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid format. Valid formats: json, csv",
		})
	}

	// Create export job
	jobID := uuid.New()
	payload := &worker.DataExportPayload{
		JobID:     jobID,
		ProjectID: projectID,
		Type:      request.Type,
		Format:    request.Format,
		Filters:   request.Filters,
	}

	if request.Destination != nil {
		payload.Destination = &worker.ExportDestination{
			Type:   request.Destination.Type,
			Bucket: request.Destination.Bucket,
			Path:   request.Destination.Path,
			Config: request.Destination.Config,
		}
	}

	task, err := worker.NewDataExportTask(payload)
	if err != nil {
		h.logger.Error("failed to create export task", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create export job",
		})
	}

	info, err := h.asynqClient.Enqueue(task)
	if err != nil {
		h.logger.Error("failed to enqueue export task", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to enqueue export job",
		})
	}

	h.logger.Info("export job queued",
		zap.String("job_id", jobID.String()),
		zap.String("task_id", info.ID),
		zap.String("type", request.Type),
	)

	return c.Status(fiber.StatusAccepted).JSON(ExportResponse{
		JobID:   jobID.String(),
		Status:  "queued",
		Message: "Export job has been queued for processing",
	})
}

// ExportDataset handles POST /v1/export/dataset
func (h *ExportHandler) ExportDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request ExportDatasetRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if request.DatasetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "datasetId is required",
		})
	}

	datasetID, err := uuid.Parse(request.DatasetID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid datasetId format",
		})
	}

	if request.Format == "" {
		request.Format = domain.ExportFormatJSON
	}
	if !request.Format.IsValid() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid format. Valid formats: json, csv, openai_finetune",
		})
	}

	// Create export job
	jobID := uuid.New()
	payload := &worker.DatasetExportPayload{
		JobID:       jobID,
		ProjectID:   projectID,
		DatasetID:   datasetID,
		Format:      request.Format,
		IncludeRuns: request.IncludeRuns,
	}

	if request.Destination != nil {
		payload.Destination = &worker.ExportDestination{
			Type:   request.Destination.Type,
			Bucket: request.Destination.Bucket,
			Path:   request.Destination.Path,
			Config: request.Destination.Config,
		}
	}

	task, err := worker.NewDatasetExportTask(payload)
	if err != nil {
		h.logger.Error("failed to create dataset export task", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create export job",
		})
	}

	info, err := h.asynqClient.Enqueue(task)
	if err != nil {
		h.logger.Error("failed to enqueue dataset export task", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to enqueue export job",
		})
	}

	h.logger.Info("dataset export job queued",
		zap.String("job_id", jobID.String()),
		zap.String("task_id", info.ID),
		zap.String("dataset_id", datasetID.String()),
	)

	return c.Status(fiber.StatusAccepted).JSON(ExportResponse{
		JobID:   jobID.String(),
		Status:  "queued",
		Message: "Dataset export job has been queued for processing",
	})
}

// RegisterRoutes registers export routes
func (h *ExportHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Post("/export/data", h.ExportData)
	v1.Post("/export/dataset", h.ExportDataset)
}
