package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// WarehouseSyncHandler handles warehouse sync HTTP requests
type WarehouseSyncHandler struct {
	service *service.WarehouseSyncService
	logger  *zap.Logger
}

// NewWarehouseSyncHandler creates a new warehouse sync handler
func NewWarehouseSyncHandler(svc *service.WarehouseSyncService, logger *zap.Logger) *WarehouseSyncHandler {
	return &WarehouseSyncHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateConnection handles POST /api/public/warehouse/connections
func (h *WarehouseSyncHandler) CreateConnection(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.WarehouseConnectionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if input.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Warehouse type is required"})
	}

	conn, err := h.service.CreateConnection(c.Context(), projectID, &input)
	if err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusBadRequest,
			"Invalid request: "+err.Error(),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(conn)
}

// GetConnection handles GET /api/public/warehouse/connections/:connId
func (h *WarehouseSyncHandler) GetConnection(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}
	connID, err := uuid.Parse(c.Params("connId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid connection ID"})
	}

	conn, err := h.service.GetConnection(c.Context(), projectID, connID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Connection not found"})
	}

	return c.JSON(conn)
}

// ListConnections handles GET /api/public/warehouse/connections
func (h *WarehouseSyncHandler) ListConnections(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	conns, err := h.service.ListConnections(c.Context(), projectID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if conns == nil {
		conns = []domain.WarehouseConnection{}
	}

	return c.JSON(fiber.Map{"connections": conns, "count": len(conns)})
}

// DeleteConnection handles DELETE /api/public/warehouse/connections/:connId
func (h *WarehouseSyncHandler) DeleteConnection(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}
	connID, err := uuid.Parse(c.Params("connId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid connection ID"})
	}

	if err := h.service.DeleteConnection(c.Context(), projectID, connID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

// TriggerSync handles POST /api/public/warehouse/connections/:connId/sync
func (h *WarehouseSyncHandler) TriggerSync(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}
	connID, err := uuid.Parse(c.Params("connId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid connection ID"})
	}

	op, err := h.service.TriggerSync(c.Context(), projectID, connID)
	if err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusBadRequest,
			"Invalid request: "+err.Error(),
		)
	}

	return c.JSON(op)
}

// TestConnection handles POST /api/public/warehouse/connections/:connId/test
func (h *WarehouseSyncHandler) TestConnection(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}
	connID, err := uuid.Parse(c.Params("connId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid connection ID"})
	}

	result, err := h.service.TestConnection(c.Context(), projectID, connID)
	if err != nil {
		return respondServiceError(
			c,
			h.logger,
			err,
			fiber.StatusNotFound,
			"Connection not found",
		)
	}

	return c.JSON(result)
}

// GetSyncStatus handles GET /api/public/warehouse/connections/:connId/status
func (h *WarehouseSyncHandler) GetSyncStatus(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}
	connID, err := uuid.Parse(c.Params("connId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid connection ID"})
	}

	ops, err := h.service.GetSyncStatus(c.Context(), projectID, connID)
	if err != nil {
		h.logger.Error("operation failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if ops == nil {
		ops = []domain.SyncOperation{}
	}

	return c.JSON(fiber.Map{"operations": ops, "count": len(ops)})
}

// GetSchemaMapping handles POST /api/public/warehouse/connections/:connId/schema
func (h *WarehouseSyncHandler) GetSchemaMapping(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}
	connID, err := uuid.Parse(c.Params("connId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid connection ID"})
	}

	mapping, err := h.service.GetSchemaMapping(c.Context(), projectID, connID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Resource not found"})
	}

	return c.JSON(fiber.Map{"mapping": mapping})
}
