package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	category := c.Query("category")
	search := &domain.EvalMarketplaceSearch{
		Query: c.Query("q"),
	}
	if category != "" {
		search.Category = &category
	}
	result, err := h.service.ListDatasets(c.Context(), search)
	if err != nil {
		h.logger.Error("failed to list datasets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list datasets"})
	}

	return c.JSON(result)
}

// GetDataset handles GET /eval-marketplace/datasets/:datasetId
func (h *EvalMarketplaceHandler) GetDataset(c *fiber.Ctx) error {
	datasetUUID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dataset ID"})
	}

	result, err := h.service.GetDataset(c.Context(), datasetUUID)
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

	authorID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found"})
	}
	result, err := h.service.PublishDataset(c.Context(), projectID, authorID, &input)
	if err != nil {
		h.logger.Error("failed to publish dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to publish dataset"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// ImportDataset handles POST /eval-marketplace/datasets/:datasetId/import
func (h *EvalMarketplaceHandler) ImportDataset(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	_ = c.Params("datasetId") // used for routing

	var input domain.EvalDatasetImportInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := h.service.ImportDataset(c.Context(), &input)
	if err != nil {
		h.logger.Error("failed to import dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to import dataset"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "imported"})
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
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	datasetUUID2, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid dataset ID"})
	}

	var input struct {
		Rating int    `json:"rating"`
		Review string `json:"review,omitempty"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found"})
	}
	err = h.service.RateDataset(c.Context(), datasetUUID2, userID, input.Rating, input.Review)
	if err != nil {
		h.logger.Error("failed to rate dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to rate dataset"})
	}

	return c.JSON(fiber.Map{"status": "rated"})
}
