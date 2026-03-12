package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CostForecastHandler handles cost forecast and budget simulator HTTP requests
type CostForecastHandler struct {
	service *service.CostForecastService
	logger  *zap.Logger
}

// NewCostForecastHandler creates a new cost forecast handler
func NewCostForecastHandler(svc *service.CostForecastService, logger *zap.Logger) *CostForecastHandler {
	return &CostForecastHandler{
		service: svc,
		logger:  logger,
	}
}

// GetForecast handles GET /cost-forecast
func (h *CostForecastHandler) GetForecast(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	period := c.Query("period", "monthly")
	days := c.QueryInt("days", 30)
	result, err := h.service.GetForecast(c.Context(), projectID, period, days)
	if err != nil {
		h.logger.Error("failed to get cost forecast", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get cost forecast"})
	}

	return c.JSON(result)
}

// Simulate handles POST /cost-forecast/simulate
func (h *CostForecastHandler) Simulate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.WhatIfInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	result, err := h.service.Simulate(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to simulate cost forecast", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to simulate cost forecast"})
	}

	return c.JSON(result)
}

// GetHistory handles GET /cost-forecast/history
func (h *CostForecastHandler) GetHistory(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	period := c.Query("period", "monthly")
	result, err := h.service.GetHistory(c.Context(), projectID, period)
	if err != nil {
		h.logger.Error("failed to get cost history", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get cost history"})
	}

	return c.JSON(result)
}

// CreateBudgetPlan handles POST /cost-forecast/budget-plan
func (h *CostForecastHandler) CreateBudgetPlan(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.BudgetPlanInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	result, err := h.service.CreateBudgetPlan(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create budget plan", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create budget plan"})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}
