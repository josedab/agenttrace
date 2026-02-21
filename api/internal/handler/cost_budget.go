package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CostBudgetHandler handles cost budget HTTP requests
type CostBudgetHandler struct {
	costBudgetService *service.CostBudgetService
	logger            *zap.Logger
}

// NewCostBudgetHandler creates a new cost budget handler
func NewCostBudgetHandler(costBudgetService *service.CostBudgetService, logger *zap.Logger) *CostBudgetHandler {
	return &CostBudgetHandler{
		costBudgetService: costBudgetService,
		logger:            logger,
	}
}

// CreateBudget handles POST /budgets
func (h *CostBudgetHandler) CreateBudget(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.CostBudgetInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	budget, err := h.costBudgetService.CreateBudget(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create budget",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create budget",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(budget)
}

// ListBudgets handles GET /budgets
func (h *CostBudgetHandler) ListBudgets(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	budgets, err := h.costBudgetService.ListBudgets(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list budgets",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list budgets",
		})
	}

	return c.JSON(budgets)
}

// GetBudget handles GET /budgets/:id
func (h *CostBudgetHandler) GetBudget(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	budgetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid budget ID",
		})
	}

	budget, err := h.costBudgetService.GetBudget(c.Context(), budgetID)
	if err != nil {
		h.logger.Error("failed to get budget",
			zap.String("budgetId", budgetID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get budget",
		})
	}

	return c.JSON(budget)
}

// UpdateBudget handles PUT /budgets/:id
func (h *CostBudgetHandler) UpdateBudget(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	budgetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid budget ID",
		})
	}

	var input domain.CostBudgetInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	budget, err := h.costBudgetService.UpdateBudget(c.Context(), budgetID, input)
	if err != nil {
		h.logger.Error("failed to update budget",
			zap.String("budgetId", budgetID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update budget",
		})
	}

	return c.JSON(budget)
}

// DeleteBudget handles DELETE /budgets/:id
func (h *CostBudgetHandler) DeleteBudget(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	budgetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid budget ID",
		})
	}

	if err := h.costBudgetService.DeleteBudget(c.Context(), budgetID); err != nil {
		h.logger.Error("failed to delete budget",
			zap.String("budgetId", budgetID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete budget",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetForecast handles GET /budgets/forecast
func (h *CostBudgetHandler) GetForecast(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	forecast, err := h.costBudgetService.GetForecast(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get forecast",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get forecast",
		})
	}

	return c.JSON(forecast)
}

// CheckBudgetRequest represents the request body for budget check
type CheckBudgetRequest struct {
	AdditionalCostCents int64 `json:"additionalCostCents"`
}

// CheckBudget handles POST /budgets/check
func (h *CostBudgetHandler) CheckBudget(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var req CheckBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	allowed, autoAction, err := h.costBudgetService.CheckBudget(c.Context(), projectID, req.AdditionalCostCents)
	if err != nil {
		h.logger.Error("failed to check budget",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to check budget",
		})
	}

	return c.JSON(fiber.Map{
		"allowed":    allowed,
		"autoAction": autoAction,
	})
}
