package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type PromptLabHandler struct {
	promptLabService *service.PromptLabService
	logger           *zap.Logger
}

func NewPromptLabHandler(svc *service.PromptLabService, logger *zap.Logger) *PromptLabHandler {
	return &PromptLabHandler{promptLabService: svc, logger: logger}
}

func (h *PromptLabHandler) CreateExperiment(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	var input domain.PromptExperimentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if input.Name == "" || len(input.Variants) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name required, minimum 2 variants"})
	}
	exp, err := h.promptLabService.CreateExperiment(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(exp)
}

func (h *PromptLabHandler) GetExperiment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("experimentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid experiment ID"})
	}
	exp, err := h.promptLabService.GetExperiment(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Experiment not found"})
	}
	return c.JSON(exp)
}

func (h *PromptLabHandler) ListExperiments(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	exps, err := h.promptLabService.ListExperiments(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"experiments": exps, "count": len(exps)})
}

func (h *PromptLabHandler) StartExperiment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("experimentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid experiment ID"})
	}
	exp, err := h.promptLabService.StartExperiment(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(exp)
}

func (h *PromptLabHandler) CompleteExperiment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("experimentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid experiment ID"})
	}
	exp, err := h.promptLabService.CompleteExperiment(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(exp)
}

func (h *PromptLabHandler) GetSuggestions(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}
	promptName := c.Query("promptName", "")
	suggestions, err := h.promptLabService.GetOptimizationSuggestions(c.Context(), projectID, promptName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"suggestions": suggestions})
}
