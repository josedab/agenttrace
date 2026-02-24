package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// DistributedTraceHandler handles distributed tracing HTTP requests
type DistributedTraceHandler struct {
	service *service.DistributedTraceService
	logger  *zap.Logger
}

// NewDistributedTraceHandler creates a new distributed trace handler
func NewDistributedTraceHandler(svc *service.DistributedTraceService, logger *zap.Logger) *DistributedTraceHandler {
	return &DistributedTraceHandler{
		service: svc,
		logger:  logger,
	}
}

// GetTrace handles GET /api/public/distributed/traces/:traceId
func (h *DistributedTraceHandler) GetTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	trace, err := h.service.GetDistributedTrace(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get distributed trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get distributed trace"})
	}

	return c.JSON(trace)
}

// GetServiceMap handles GET /api/public/distributed/service-map
func (h *DistributedTraceHandler) GetServiceMap(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	serviceMap, err := h.service.GetServiceMap(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get service map", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get service map"})
	}

	return c.JSON(serviceMap)
}

// Correlate handles POST /api/public/distributed/correlate
func (h *DistributedTraceHandler) Correlate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.TraceCorrelationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Trace ID is required"})
	}

	trace, err := h.service.CorrelateTraces(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to correlate traces", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to correlate traces"})
	}

	return c.JSON(trace)
}
