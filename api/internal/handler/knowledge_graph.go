package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

type KnowledgeGraphHandler struct {
	service *service.KnowledgeGraphService
	logger  *zap.Logger
}

func NewKnowledgeGraphHandler(svc *service.KnowledgeGraphService, logger *zap.Logger) *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{service: svc, logger: logger}
}

func (h *KnowledgeGraphHandler) Build(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	graph, err := h.service.BuildGraph(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to build knowledge graph", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build graph"})
	}

	return c.JSON(graph)
}

func (h *KnowledgeGraphHandler) Query(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var query domain.KGQuery
	if err := c.BodyParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	graph, err := h.service.QueryGraph(c.Context(), projectID, query)
	if err != nil {
		h.logger.Error("failed to query knowledge graph", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query graph"})
	}

	return c.JSON(graph)
}

func (h *KnowledgeGraphHandler) GetStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get knowledge graph stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get stats"})
	}

	return c.JSON(stats)
}
