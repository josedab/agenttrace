package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// EvalMarketplaceHandler handles evaluation dataset marketplace HTTP requests
type EvalMarketplaceHandler struct {
	service *service.EvalMarketplaceService
	logger  *zap.Logger
}

// NewEvalMarketplaceHandler creates a new eval marketplace handler
func NewEvalMarketplaceHandler(svc *service.EvalMarketplaceService, logger *zap.Logger) *EvalMarketplaceHandler {
	return &EvalMarketplaceHandler{
		service: svc,
		logger:  logger,
	}
}

// ListDatasets handles GET /eval-marketplace/datasets
func (h *EvalMarketplaceHandler) ListDatasets(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.ListDatasets(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to list datasets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list datasets"})
	}

	return c.JSON(result)
}

// GetDataset handles GET /eval-marketplace/datasets/:datasetId
func (h *EvalMarketplaceHandler) GetDataset(c *fiber.Ctx) error {
	datasetID := c.Params("datasetId")
	if datasetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dataset ID is required"})
	}

	result, err := h.service.GetDataset(c.Context(), datasetID)
	if err != nil {
		h.logger.Error("failed to get dataset", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Dataset not found"})
	}

	return c.JSON(result)
}

// PublishDataset handles POST /eval-marketplace/datasets
func (h *EvalMarketplaceHandler) PublishDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EvalDatasetPublishInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	result, err := h.service.PublishDataset(c.Context(), projectID.String(), &input)
	if err != nil {
		h.logger.Error("failed to publish dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to publish dataset"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ImportDataset handles POST /eval-marketplace/datasets/:datasetId/import
func (h *EvalMarketplaceHandler) ImportDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasetID := c.Params("datasetId")
	if datasetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dataset ID is required"})
	}

	var input domain.EvalDatasetImportInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.ImportDataset(c.Context(), projectID.String(), datasetID, &input)
	if err != nil {
		h.logger.Error("failed to import dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to import dataset"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ListCategories handles GET /eval-marketplace/categories
func (h *EvalMarketplaceHandler) ListCategories(c *fiber.Ctx) error {
	result, err := h.service.ListCategories(c.Context())
	if err != nil {
		h.logger.Error("failed to list categories", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list categories"})
	}

	return c.JSON(result)
}

// RateDataset handles POST /eval-marketplace/datasets/:datasetId/rate
func (h *EvalMarketplaceHandler) RateDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasetID := c.Params("datasetId")
	if datasetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dataset ID is required"})
	}

	var input struct {
		Rating float64 `json:"rating"`
		Review string  `json:"review,omitempty"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	result, err := h.service.RateDataset(c.Context(), projectID.String(), datasetID, input.Rating, input.Review)
	if err != nil {
		h.logger.Error("failed to rate dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to rate dataset"})
	}

	return c.JSON(result)
}
