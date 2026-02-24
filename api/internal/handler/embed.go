package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// EmbedHandler handles white-label embed HTTP requests
type EmbedHandler struct {
	service *service.EmbedService
	logger  *zap.Logger
}

// NewEmbedHandler creates a new embed handler
func NewEmbedHandler(svc *service.EmbedService, logger *zap.Logger) *EmbedHandler {
	return &EmbedHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateConfig handles POST /api/public/embed/config
func (h *EmbedHandler) CreateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EmbedConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.CreateConfig(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create embed config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create config"})
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

// GetConfig handles GET /api/public/embed/config
func (h *EmbedHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Embed config not found"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /api/public/embed/config
func (h *EmbedHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.EmbedConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(config)
}

// GenerateToken handles POST /api/public/embed/token
func (h *EmbedHandler) GenerateToken(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	token, err := h.service.GenerateToken(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(token)
}
