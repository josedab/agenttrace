package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// DatasetsHandler handles dataset endpoints
type DatasetsHandler struct {
	datasetService *service.DatasetService
	logger         *zap.Logger
}

// NewDatasetsHandler creates a new datasets handler
func NewDatasetsHandler(datasetService *service.DatasetService, logger *zap.Logger) *DatasetsHandler {
	return &DatasetsHandler{
		datasetService: datasetService,
		logger:         logger,
	}
}

// ListDatasets handles GET /v1/datasets
func (h *DatasetsHandler) ListDatasets(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := &domain.DatasetFilter{
		ProjectID: projectID,
	}

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.datasetService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list datasets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list datasets",
		})
	}

	return c.JSON(list)
}

// GetDataset handles GET /v1/datasets/:datasetId
func (h *DatasetsHandler) GetDataset(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	dataset, err := h.datasetService.Get(c.Context(), datasetID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset not found",
			})
		}
		h.logger.Error("failed to get dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get dataset",
		})
	}

	return c.JSON(dataset)
}

// GetDatasetByName handles GET /v1/datasets/name/:name
func (h *DatasetsHandler) GetDatasetByName(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Dataset name required",
		})
	}

	dataset, err := h.datasetService.GetByName(c.Context(), projectID, name)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset not found",
			})
		}
		h.logger.Error("failed to get dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get dataset",
		})
	}

	return c.JSON(dataset)
}

// CreateDataset handles POST /v1/datasets
func (h *DatasetsHandler) CreateDataset(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.DatasetInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	dataset, err := h.datasetService.Create(c.Context(), projectID, &input)
	if err != nil {
		if apperrors.IsValidation(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}
		h.logger.Error("failed to create dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create dataset",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dataset)
}

// UpdateDataset handles PATCH /v1/datasets/:datasetId
func (h *DatasetsHandler) UpdateDataset(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	var input domain.DatasetInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	dataset, err := h.datasetService.Update(c.Context(), datasetID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset not found",
			})
		}
		h.logger.Error("failed to update dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update dataset",
		})
	}

	return c.JSON(dataset)
}

// DeleteDataset handles DELETE /v1/datasets/:datasetId
func (h *DatasetsHandler) DeleteDataset(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	if err := h.datasetService.Delete(c.Context(), datasetID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset not found",
			})
		}
		h.logger.Error("failed to delete dataset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete dataset",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListItems handles GET /v1/datasets/:datasetId/items
func (h *DatasetsHandler) ListItems(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	filter := &domain.DatasetItemFilter{
		DatasetID: datasetID,
	}

	if status := c.Query("status"); status != "" {
		s := domain.DatasetItemStatus(status)
		filter.Status = &s
	}

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	items, totalCount, err := h.datasetService.ListItems(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list items", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list items",
		})
	}

	return c.JSON(fiber.Map{
		"data":       items,
		"totalCount": totalCount,
	})
}

// CreateItem handles POST /v1/datasets/:datasetId/items
func (h *DatasetsHandler) CreateItem(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	var input domain.DatasetItemInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	item, err := h.datasetService.AddItem(c.Context(), datasetID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset not found",
			})
		}
		h.logger.Error("failed to create item", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// CreateItemFromTrace handles POST /v1/datasets/:datasetId/items/from-trace
func (h *DatasetsHandler) CreateItemFromTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	var input struct {
		TraceID       string  `json:"traceId"`
		ObservationID *string `json:"observationId,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	item, err := h.datasetService.AddItemFromTrace(c.Context(), datasetID, projectID, input.TraceID, input.ObservationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset or trace not found",
			})
		}
		h.logger.Error("failed to create item from trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create item from trace",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// UpdateItem handles PATCH /v1/datasets/:datasetId/items/:itemId
func (h *DatasetsHandler) UpdateItem(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid item ID",
		})
	}

	var input domain.DatasetItemUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	item, err := h.datasetService.UpdateItem(c.Context(), itemID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Item not found",
			})
		}
		h.logger.Error("failed to update item", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update item",
		})
	}

	return c.JSON(item)
}

// DeleteItem handles DELETE /v1/datasets/:datasetId/items/:itemId
func (h *DatasetsHandler) DeleteItem(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid item ID",
		})
	}

	if err := h.datasetService.DeleteItem(c.Context(), itemID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Item not found",
			})
		}
		h.logger.Error("failed to delete item", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete item",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListRuns handles GET /v1/datasets/:datasetId/runs
func (h *DatasetsHandler) ListRuns(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	runs, totalCount, err := h.datasetService.ListRuns(c.Context(), datasetID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list runs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list runs",
		})
	}

	return c.JSON(fiber.Map{
		"data":       runs,
		"totalCount": totalCount,
	})
}

// GetRun handles GET /v1/datasets/:datasetId/runs/:runId
func (h *DatasetsHandler) GetRun(c *fiber.Ctx) error {
	runID, err := uuid.Parse(c.Params("runId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid run ID",
		})
	}

	run, err := h.datasetService.GetRun(c.Context(), runID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Run not found",
			})
		}
		h.logger.Error("failed to get run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get run",
		})
	}

	return c.JSON(run)
}

// CreateRun handles POST /v1/datasets/:datasetId/runs
func (h *DatasetsHandler) CreateRun(c *fiber.Ctx) error {
	datasetID, err := uuid.Parse(c.Params("datasetId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid dataset ID",
		})
	}

	var input domain.DatasetRunInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	run, err := h.datasetService.CreateRun(c.Context(), datasetID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Dataset not found",
			})
		}
		h.logger.Error("failed to create run", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create run",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(run)
}

// AddRunItem handles POST /v1/datasets/:datasetId/runs/:runId/items
func (h *DatasetsHandler) AddRunItem(c *fiber.Ctx) error {
	runID, err := uuid.Parse(c.Params("runId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid run ID",
		})
	}

	var input domain.DatasetRunItemInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.DatasetItemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "datasetItemId is required",
		})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	item, err := h.datasetService.AddRunItem(c.Context(), runID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Run or dataset item not found",
			})
		}
		h.logger.Error("failed to add run item", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to add run item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// BulkAddRunItems handles POST /v1/datasets/:datasetId/runs/:runId/items/batch
func (h *DatasetsHandler) BulkAddRunItems(c *fiber.Ctx) error {
	runID, err := uuid.Parse(c.Params("runId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid run ID",
		})
	}

	var request struct {
		Items []*domain.DatasetRunItemInput `json:"items"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if len(request.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "At least one item is required",
		})
	}

	if len(request.Items) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Maximum 100 items per batch",
		})
	}

	// Validate each item
	for i, item := range request.Items {
		if item.DatasetItemID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": fmt.Sprintf("items[%d]: datasetItemId is required", i),
			})
		}
		if item.TraceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": fmt.Sprintf("items[%d]: traceId is required", i),
			})
		}
	}

	items, err := h.datasetService.AddRunItemsBatch(c.Context(), runID, request.Items)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Run or dataset item not found",
			})
		}
		if apperrors.IsValidation(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}
		h.logger.Error("failed to bulk add run items", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to add run items",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": items,
	})
}

// GetRunResults handles GET /v1/datasets/:datasetId/runs/:runId/results
func (h *DatasetsHandler) GetRunResults(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	runID, err := uuid.Parse(c.Params("runId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid run ID",
		})
	}

	results, err := h.datasetService.GetRunResults(c.Context(), projectID, runID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Run not found",
			})
		}
		h.logger.Error("failed to get run results", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get run results",
		})
	}

	return c.JSON(results)
}

// RegisterRoutes registers dataset routes
func (h *DatasetsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	// Dataset endpoints
	v1.Get("/datasets", h.ListDatasets)
	v1.Get("/datasets/name/:name", h.GetDatasetByName)
	v1.Get("/datasets/:datasetId", h.GetDataset)
	v1.Post("/datasets", h.CreateDataset)
	v1.Patch("/datasets/:datasetId", h.UpdateDataset)
	v1.Delete("/datasets/:datasetId", h.DeleteDataset)

	// Item endpoints
	v1.Get("/datasets/:datasetId/items", h.ListItems)
	v1.Post("/datasets/:datasetId/items", h.CreateItem)
	v1.Post("/datasets/:datasetId/items/from-trace", h.CreateItemFromTrace)
	v1.Patch("/datasets/:datasetId/items/:itemId", h.UpdateItem)
	v1.Delete("/datasets/:datasetId/items/:itemId", h.DeleteItem)

	// Run endpoints
	v1.Get("/datasets/:datasetId/runs", h.ListRuns)
	v1.Get("/datasets/:datasetId/runs/:runId", h.GetRun)
	v1.Post("/datasets/:datasetId/runs", h.CreateRun)
	v1.Post("/datasets/:datasetId/runs/:runId/items", h.AddRunItem)
	v1.Post("/datasets/:datasetId/runs/:runId/items/batch", h.BulkAddRunItems)
	v1.Get("/datasets/:datasetId/runs/:runId/results", h.GetRunResults)
}
