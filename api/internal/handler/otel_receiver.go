package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OTelReceiverHandler handles OTLP receiver HTTP requests
type OTelReceiverHandler struct {
	otelReceiverService *service.OTelReceiverService
	logger              *zap.Logger
}

// NewOTelReceiverHandler creates a new OTel receiver handler
func NewOTelReceiverHandler(otelReceiverService *service.OTelReceiverService, logger *zap.Logger) *OTelReceiverHandler {
	return &OTelReceiverHandler{
		otelReceiverService: otelReceiverService,
		logger:              logger,
	}
}

// ReceiveTraces handles POST /v1/traces (OTLP HTTP)
func (h *OTelReceiverHandler) ReceiveTraces(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request domain.OTLPTraceRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid OTLP trace request",
		})
	}

	resp, err := h.otelReceiverService.ReceiveTraces(c.Context(), projectID, &request)
	if err != nil {
		h.logger.Error("failed to receive OTLP traces", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to process traces",
		})
	}

	return c.JSON(resp)
}

// ReceiveMetrics handles POST /v1/metrics
func (h *OTelReceiverHandler) ReceiveMetrics(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request any
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid OTLP metrics request",
		})
	}

	resp, err := h.otelReceiverService.ReceiveMetrics(c.Context(), projectID, request)
	if err != nil {
		h.logger.Error("failed to receive OTLP metrics", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to process metrics",
		})
	}

	return c.JSON(resp)
}

// ReceiveLogs handles POST /v1/logs
func (h *OTelReceiverHandler) ReceiveLogs(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var request any
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid OTLP logs request",
		})
	}

	resp, err := h.otelReceiverService.ReceiveLogs(c.Context(), projectID, request)
	if err != nil {
		h.logger.Error("failed to receive OTLP logs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to process logs",
		})
	}

	return c.JSON(resp)
}
