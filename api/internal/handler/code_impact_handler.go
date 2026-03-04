package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CodeImpactHandler handles trace-linked code impact map HTTP requests
type CodeImpactHandler struct {
	service *service.CodeImpactService
	logger  *zap.Logger
}

// NewCodeImpactHandler creates a new code impact handler
func NewCodeImpactHandler(svc *service.CodeImpactService, logger *zap.Logger) *CodeImpactHandler {
	return &CodeImpactHandler{
		service: svc,
		logger:  logger,
	}
}

// GetCodeImpact handles GET /traces/:traceId/code-impact
func (h *CodeImpactHandler) GetCodeImpact(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	result, err := h.service.GetCodeImpact(c.Context(), projectID.String(), traceID)
	if err != nil {
		h.logger.Error("failed to get code impact", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get code impact"})
	}

	return c.JSON(result)
}

// GetProjectImpactSummary handles GET /code-impact/summary
func (h *CodeImpactHandler) GetProjectImpactSummary(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.GetProjectImpactSummary(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get project impact summary", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get project impact summary"})
	}

	return c.JSON(result)
}

// GetFileTree handles GET /code-impact/file-tree
func (h *CodeImpactHandler) GetFileTree(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	result, err := h.service.GetFileTree(c.Context(), projectID.String())
	if err != nil {
		h.logger.Error("failed to get file tree", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get file tree"})
	}

	return c.JSON(result)
}
