package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PrivacyHandler handles privacy-preserving analytics HTTP requests
type PrivacyHandler struct {
	service *service.PrivacyService
	logger  *zap.Logger
}

// NewPrivacyHandler creates a new privacy handler
func NewPrivacyHandler(svc *service.PrivacyService, logger *zap.Logger) *PrivacyHandler {
	return &PrivacyHandler{
		service: svc,
		logger:  logger,
	}
}

// ScanPII handles POST /api/public/privacy/scan
func (h *PrivacyHandler) ScanPII(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	result, err := h.service.ScanForPII(c.Context(), projectID, body.Content)
	if err != nil {
		h.logger.Error("failed to scan for PII", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan for PII"})
	}

	return c.JSON(result)
}

// GetConfig handles GET /api/public/privacy/config
func (h *PrivacyHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	config, err := h.service.GetConfig(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get PII config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get config"})
	}

	return c.JSON(config)
}

// UpdateConfig handles PUT /api/public/privacy/config
func (h *PrivacyHandler) UpdateConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.PIIConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.UpdateConfig(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to update PII config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update config"})
	}

	return c.JSON(config)
}

// RequestDeletion handles POST /api/public/privacy/deletion-requests
func (h *PrivacyHandler) RequestDeletion(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.DeletionRequestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.RequestType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Request type is required"})
	}

	req, err := h.service.RequestDeletion(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create deletion request", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create deletion request"})
	}

	return c.Status(fiber.StatusCreated).JSON(req)
}

// ListDeletionRequests handles GET /api/public/privacy/deletion-requests
func (h *PrivacyHandler) ListDeletionRequests(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	requests, err := h.service.ListDeletionRequests(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list deletion requests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list deletion requests"})
	}

	if requests == nil {
		requests = []domain.DataDeletionRequest{}
	}
	return c.JSON(requests)
}
