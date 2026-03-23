package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// SyntheticDataHandler handles synthetic data generation HTTP requests
type SyntheticDataHandler struct {
	service *service.SyntheticDataService
	logger  *zap.Logger
}

// NewSyntheticDataHandler creates a new synthetic data handler
func NewSyntheticDataHandler(svc *service.SyntheticDataService, logger *zap.Logger) *SyntheticDataHandler {
	return &SyntheticDataHandler{
		service: svc,
		logger:  logger,
	}
}

// Generate handles POST /api/public/synthetic-data/generate
func (h *SyntheticDataHandler) Generate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.GenerateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" || input.Type == "" || input.Count <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, type, and count (> 0) are required"})
	}

	dataset, err := h.service.Generate(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to generate synthetic data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate synthetic data"})
	}

	return c.Status(fiber.StatusCreated).JSON(dataset)
}

// Get handles GET /api/public/synthetic-data/datasets/:datasetId
func (h *SyntheticDataHandler) Get(c *fiber.Ctx) error {
	datasetID := c.Params("datasetId")
	if datasetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dataset ID is required"})
	}

	dataset, err := h.service.GetDataset(c.Context(), datasetID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(dataset)
}

// List handles GET /api/public/synthetic-data/datasets
func (h *SyntheticDataHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasets, err := h.service.ListDatasets(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to list synthetic datasets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list datasets"})
	}

	return c.JSON(datasets)
}

// GetStats handles GET /api/public/synthetic-data/stats
func (h *SyntheticDataHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetStats(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get synthetic data stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(stats)
}
