package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// EdgeIngestHandler handles edge/mobile device HTTP requests
type EdgeIngestHandler struct {
	service *service.EdgeIngestService
	logger  *zap.Logger
}

// NewEdgeIngestHandler creates a new edge ingest handler
func NewEdgeIngestHandler(svc *service.EdgeIngestService, logger *zap.Logger) *EdgeIngestHandler {
	return &EdgeIngestHandler{
		service: svc,
		logger:  logger,
	}
}

// RegisterDevice handles POST /api/public/edge/devices
func (h *EdgeIngestHandler) RegisterDevice(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.EdgeDeviceInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.DeviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Device ID is required"})
	}
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.Platform == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Platform is required"})
	}

	device, err := h.service.RegisterDevice(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to register edge device", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(device)
}

// IngestBatch handles POST /api/public/edge/ingest
func (h *EdgeIngestHandler) IngestBatch(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var batch domain.EdgeTraceBatch
	if err := c.BodyParser(&batch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if batch.DeviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Device ID is required"})
	}

	result, err := h.service.IngestBatch(c.Context(), projectID, &batch)
	if err != nil {
		h.logger.Error("failed to ingest edge batch", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// ListDevices handles GET /api/public/edge/devices
func (h *EdgeIngestHandler) ListDevices(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	devices, err := h.service.ListDevices(c.Context(), projectID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if devices == nil {
		devices = []domain.EdgeDevice{}
	}

	return c.JSON(fiber.Map{"devices": devices, "count": len(devices)})
}

// GetDeviceStatus handles GET /api/public/edge/devices/:deviceId/status
func (h *EdgeIngestHandler) GetDeviceStatus(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Device ID is required"})
	}

	device, err := h.service.GetDeviceStatus(c.Context(), deviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(device)
}

// SyncOfflineData handles POST /api/public/edge/sync
func (h *EdgeIngestHandler) SyncOfflineData(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var syncReq domain.EdgeSyncRequest
	if err := c.BodyParser(&syncReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	results, err := h.service.SyncOfflineData(c.Context(), projectID, &syncReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"results": results, "batchesSynced": len(results)})
}

// GetStats handles GET /api/public/edge/stats
func (h *EdgeIngestHandler) GetStats(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	stats := h.service.GetStats(c.Context(), projectID)
	return c.JSON(stats)
}
