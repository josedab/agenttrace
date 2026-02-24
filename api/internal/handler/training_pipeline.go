package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// TrainingPipelineHandler handles training pipeline HTTP requests
type TrainingPipelineHandler struct {
	logger  *zap.Logger
	service *service.TrainingPipelineService
}

// NewTrainingPipelineHandler creates a new training pipeline handler
func NewTrainingPipelineHandler(logger *zap.Logger, svc *service.TrainingPipelineService) *TrainingPipelineHandler {
	return &TrainingPipelineHandler{
		logger:  logger,
		service: svc,
	}
}

// CreateDataset handles POST /api/public/training/datasets
func (h *TrainingPipelineHandler) CreateDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.TrainingDatasetInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dataset, err := h.service.CreateDataset(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create training dataset", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(dataset)
}

// ListDatasets handles GET /api/public/training/datasets
func (h *TrainingPipelineHandler) ListDatasets(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasets, err := h.service.ListDatasets(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list training datasets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list datasets"})
	}

	return c.JSON(fiber.Map{"datasets": datasets})
}

// ExportDataset handles POST /api/public/training/datasets/:datasetId/export
func (h *TrainingPipelineHandler) ExportDataset(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dataset ID"})
	}

	format := domain.TrainingDatasetFormat(c.Query("format", string(domain.TrainingFormatJSONL)))

	export, err := h.service.ExportDataset(c.Context(), datasetID, format)
	if err != nil {
		h.logger.Error("failed to export dataset", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(export)
}

// DetectFailures handles GET /api/public/training/failure-patterns
func (h *TrainingPipelineHandler) DetectFailures(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	patterns, err := h.service.DetectFailurePatterns(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to detect failure patterns", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to detect failure patterns"})
	}

	return c.JSON(fiber.Map{"patterns": patterns})
}
