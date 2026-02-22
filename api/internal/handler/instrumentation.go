package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// InstrumentationHandler handles instrumentation HTTP requests
type InstrumentationHandler struct {
	instrumentationService *service.InstrumentationService
	logger                 *zap.Logger
}

// NewInstrumentationHandler creates a new instrumentation handler
func NewInstrumentationHandler(instrumentationService *service.InstrumentationService, logger *zap.Logger) *InstrumentationHandler {
	return &InstrumentationHandler{
		instrumentationService: instrumentationService,
		logger:                 logger,
	}
}

// ListFrameworks handles GET /instrumentation/frameworks
func (h *InstrumentationHandler) ListFrameworks(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	frameworks := h.instrumentationService.ListFrameworks()
	return c.JSON(frameworks)
}

// GetSetup handles GET /instrumentation/setup/:framework?language=python
func (h *InstrumentationHandler) GetSetup(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	framework := c.Params("framework")
	if framework == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Framework is required",
		})
	}

	language := c.Query("language", "python")

	setup, err := h.instrumentationService.GetSetupInstructions(framework, language)
	if err != nil {
		h.logger.Error("failed to get setup instructions",
			zap.String("framework", framework),
			zap.String("language", language),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": err.Error(),
		})
	}

	return c.JSON(setup)
}
