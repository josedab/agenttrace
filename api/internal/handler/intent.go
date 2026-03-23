package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type IntentHandler struct {
	service *service.IntentService
	logger  *zap.Logger
}

func NewIntentHandler(svc *service.IntentService, logger *zap.Logger) *IntentHandler {
	return &IntentHandler{service: svc, logger: logger}
}

func (h *IntentHandler) Declare(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.IntentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	decl, err := h.service.DeclareIntent(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to declare intent", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to declare intent"})
	}

	return c.Status(fiber.StatusCreated).JSON(decl)
}

func (h *IntentHandler) Verify(c *fiber.Ctx) error {
	intentIDStr := c.Params("intentId")
	intentID, err := uuid.Parse(intentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid intent ID"})
	}

	var body struct {
		ActualActions []string `json:"actualActions"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	decl, err := h.service.VerifyIntent(c.Context(), intentID, body.ActualActions)
	if err != nil {
		h.logger.Error("failed to verify intent", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(decl)
}

func (h *IntentHandler) Get(c *fiber.Ctx) error {
	intentIDStr := c.Params("intentId")
	intentID, err := uuid.Parse(intentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid intent ID"})
	}

	decl, err := h.service.GetVerification(c.Context(), intentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(decl)
}

func (h *IntentHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get intent stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(stats)
}
