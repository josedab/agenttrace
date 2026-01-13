package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	pgrepo "github.com/agenttrace/agenttrace/api/internal/repository/postgres"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	logger              *zap.Logger
	webhookRepo         *pgrepo.WebhookRepository
	notificationService *service.NotificationService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	logger *zap.Logger,
	webhookRepo *pgrepo.WebhookRepository,
	notificationService *service.NotificationService,
) *WebhookHandler {
	return &WebhookHandler{
		logger:              logger,
		webhookRepo:         webhookRepo,
		notificationService: notificationService,
	}
}

// ListWebhooks returns all webhooks for a project
// @Summary List webhooks
// @Description Get all webhooks for a project
// @Tags webhooks
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param type query string false "Filter by webhook type"
// @Param enabled query bool false "Filter by enabled status"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.WebhookList
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/public/webhooks [get]
func (h *WebhookHandler) ListWebhooks(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	filter := domain.WebhookFilter{
		ProjectID: projectID,
	}

	if webhookType := c.Query("type"); webhookType != "" {
		wt := domain.WebhookType(webhookType)
		filter.Type = &wt
	}

	if enabled := c.Query("enabled"); enabled != "" {
		isEnabled := enabled == "true"
		filter.IsEnabled = &isEnabled
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	result, err := h.webhookRepo.List(c.Context(), &filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list webhooks",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to list webhooks")
	}

	return c.JSON(result)
}

// GetWebhook returns a specific webhook
// @Summary Get webhook
// @Description Get a specific webhook by ID
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} domain.Webhook
// @Failure 404 {object} ErrorResponse
// @Router /api/public/webhooks/{id} [get]
func (h *WebhookHandler) GetWebhook(c *fiber.Ctx) error {
	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	webhook, err := h.webhookRepo.GetByProjectID(c.Context(), projectID, webhookID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return errorResponse(c, fiber.StatusNotFound, "Webhook not found")
		}
		h.logger.Error("failed to get webhook",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get webhook")
	}

	return c.JSON(webhook)
}

// CreateWebhook creates a new webhook
// @Summary Create webhook
// @Description Create a new webhook for notifications
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body domain.WebhookInput true "Webhook configuration"
// @Success 201 {object} domain.Webhook
// @Failure 400 {object} ErrorResponse
// @Router /api/public/webhooks [post]
func (h *WebhookHandler) CreateWebhook(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	var input domain.WebhookInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if input.Name == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Name is required")
	}
	if input.URL == "" {
		return errorResponse(c, fiber.StatusBadRequest, "URL is required")
	}
	if len(input.Events) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "At least one event type is required")
	}

	// Set default type if not specified
	webhookType := input.Type
	if webhookType == "" {
		webhookType = domain.WebhookTypeGeneric
	}

	// Create webhook object
	now := time.Now()
	webhook := &domain.Webhook{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Type:             webhookType,
		Name:             input.Name,
		URL:              input.URL,
		Secret:           input.Secret,
		Events:           input.Events,
		IsEnabled:        input.IsEnabled,
		Headers:          input.Headers,
		CostThreshold:    input.CostThreshold,
		LatencyThreshold: input.LatencyThreshold,
		ScoreThreshold:   input.ScoreThreshold,
		RateLimitPerHour: input.RateLimitPerHour,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.webhookRepo.Create(c.Context(), webhook); err != nil {
		h.logger.Error("failed to create webhook",
			zap.String("projectId", projectID.String()),
			zap.String("name", input.Name),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to create webhook")
	}

	h.logger.Info("created webhook",
		zap.String("webhookId", webhook.ID.String()),
		zap.String("projectId", projectID.String()),
		zap.String("type", string(webhook.Type)),
	)

	return c.Status(fiber.StatusCreated).JSON(webhook)
}

// UpdateWebhook updates an existing webhook
// @Summary Update webhook
// @Description Update an existing webhook configuration
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Param webhook body domain.WebhookUpdateInput true "Updated webhook configuration"
// @Success 200 {object} domain.Webhook
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/webhooks/{id} [patch]
func (h *WebhookHandler) UpdateWebhook(c *fiber.Ctx) error {
	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	// Get existing webhook
	webhook, err := h.webhookRepo.GetByProjectID(c.Context(), projectID, webhookID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return errorResponse(c, fiber.StatusNotFound, "Webhook not found")
		}
		h.logger.Error("failed to get webhook for update",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get webhook")
	}

	var input domain.WebhookUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Apply updates
	if input.Name != nil {
		webhook.Name = *input.Name
	}
	if input.URL != nil {
		webhook.URL = *input.URL
	}
	if input.Secret != nil {
		webhook.Secret = *input.Secret
	}
	if input.Events != nil {
		webhook.Events = input.Events
	}
	if input.IsEnabled != nil {
		webhook.IsEnabled = *input.IsEnabled
	}
	if input.Headers != nil {
		webhook.Headers = input.Headers
	}
	if input.Type != nil {
		webhook.Type = *input.Type
	}
	if input.CostThreshold != nil {
		webhook.CostThreshold = input.CostThreshold
	}
	if input.LatencyThreshold != nil {
		webhook.LatencyThreshold = input.LatencyThreshold
	}
	if input.ScoreThreshold != nil {
		webhook.ScoreThreshold = input.ScoreThreshold
	}
	if input.RateLimitPerHour != nil {
		webhook.RateLimitPerHour = input.RateLimitPerHour
	}

	webhook.UpdatedAt = time.Now()

	if err := h.webhookRepo.Update(c.Context(), webhook); err != nil {
		h.logger.Error("failed to update webhook",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to update webhook")
	}

	h.logger.Info("updated webhook",
		zap.String("webhookId", webhookID.String()),
		zap.String("projectId", projectID.String()),
	)

	return c.JSON(webhook)
}

// DeleteWebhook deletes a webhook
// @Summary Delete webhook
// @Description Delete a webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Router /api/public/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	// Verify webhook exists and belongs to project
	_, err = h.webhookRepo.GetByProjectID(c.Context(), projectID, webhookID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return errorResponse(c, fiber.StatusNotFound, "Webhook not found")
		}
		h.logger.Error("failed to get webhook for deletion",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get webhook")
	}

	if err := h.webhookRepo.Delete(c.Context(), webhookID); err != nil {
		h.logger.Error("failed to delete webhook",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to delete webhook")
	}

	h.logger.Info("deleted webhook",
		zap.String("webhookId", webhookID.String()),
		zap.String("projectId", projectID.String()),
	)

	return c.SendStatus(fiber.StatusNoContent)
}

// TestWebhook sends a test notification to a webhook
// @Summary Test webhook
// @Description Send a test notification to verify webhook configuration
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} domain.WebhookDelivery
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/webhooks/{id}/test [post]
func (h *WebhookHandler) TestWebhook(c *fiber.Ctx) error {
	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	// Get webhook from database
	webhook, err := h.webhookRepo.GetByProjectID(c.Context(), projectID, webhookID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return errorResponse(c, fiber.StatusNotFound, "Webhook not found")
		}
		h.logger.Error("failed to get webhook for test",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get webhook")
	}

	// Send test notification
	delivery, err := h.notificationService.TestWebhook(c.Context(), webhook)
	if err != nil {
		h.logger.Error("test webhook failed",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		// Still return the delivery object to show what happened
	}

	// Store delivery record
	if delivery != nil {
		if err := h.webhookRepo.CreateDelivery(c.Context(), delivery); err != nil {
			h.logger.Warn("failed to store test delivery record",
				zap.String("webhookId", webhookID.String()),
				zap.Error(err),
			)
		}
	}

	return c.JSON(delivery)
}

// ListWebhookDeliveries returns delivery history for a webhook
// @Summary List webhook deliveries
// @Description Get delivery history for a specific webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Param success query bool false "Filter by success status"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.WebhookDeliveryList
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/webhooks/{id}/deliveries [get]
func (h *WebhookHandler) ListWebhookDeliveries(c *fiber.Ctx) error {
	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID")
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	// Verify webhook exists and belongs to project
	_, err = h.webhookRepo.GetByProjectID(c.Context(), projectID, webhookID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return errorResponse(c, fiber.StatusNotFound, "Webhook not found")
		}
		h.logger.Error("failed to get webhook for deliveries",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to get webhook")
	}

	filter := domain.WebhookDeliveryFilter{
		WebhookID: webhookID,
	}

	if success := c.Query("success"); success != "" {
		isSuccess := success == "true"
		filter.Success = &isSuccess
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	result, err := h.webhookRepo.ListDeliveries(c.Context(), &filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list webhook deliveries",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to list deliveries")
	}

	return c.JSON(result)
}

// getProjectIDFromContext extracts project ID from context or API key
func getProjectIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	// Try to get from API key context first
	if apiKey, ok := c.Locals("apiKey").(*domain.APIKey); ok {
		return apiKey.ProjectID, nil
	}

	// Try to get from query parameter
	if projectIDStr := c.Query("projectId"); projectIDStr != "" {
		return uuid.Parse(projectIDStr)
	}

	return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "Project ID required")
}
