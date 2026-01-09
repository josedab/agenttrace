package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	logger              *zap.Logger
	notificationService *service.NotificationService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	logger *zap.Logger,
	notificationService *service.NotificationService,
) *WebhookHandler {
	return &WebhookHandler{
		logger:              logger,
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
// @Success 200 {object} domain.WebhookList
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/public/webhooks [get]
func (h *WebhookHandler) ListWebhooks(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
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

	// For now, return empty list - actual implementation would query database
	result := domain.WebhookList{
		Webhooks:   []domain.Webhook{},
		TotalCount: 0,
		HasMore:    false,
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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID", err)
	}

	// For now, return not found - actual implementation would query database
	h.logger.Debug("Get webhook", zap.String("webhookId", webhookID.String()))
	return errorResponse(c, fiber.StatusNotFound, "Webhook not found", nil)
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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	var input domain.WebhookInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate input
	if input.Name == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Name is required", nil)
	}
	if input.URL == "" {
		return errorResponse(c, fiber.StatusBadRequest, "URL is required", nil)
	}
	if len(input.Events) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "At least one event type is required", nil)
	}

	// Create webhook object
	webhook := &domain.Webhook{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Type:             input.Type,
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
	}

	// For now, return the created webhook - actual implementation would persist to database
	h.logger.Info("Created webhook",
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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID", err)
	}

	var input domain.WebhookUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Debug("Update webhook", zap.String("webhookId", webhookID.String()))

	// For now, return not found - actual implementation would update in database
	return errorResponse(c, fiber.StatusNotFound, "Webhook not found", nil)
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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID", err)
	}

	h.logger.Info("Deleted webhook", zap.String("webhookId", webhookID.String()))

	// For now, return not found - actual implementation would delete from database
	return errorResponse(c, fiber.StatusNotFound, "Webhook not found", nil)
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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID", err)
	}

	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	// For testing purposes, create a mock webhook
	// In real implementation, this would fetch from database
	var input struct {
		Type string `json:"type"`
		URL  string `json:"url"`
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	webhook := &domain.Webhook{
		ID:        webhookID,
		ProjectID: projectID,
		Type:      domain.WebhookType(input.Type),
		Name:      input.Name,
		URL:       input.URL,
		IsEnabled: true,
	}

	// Send test notification
	delivery, err := h.notificationService.TestWebhook(c.Context(), webhook)
	if err != nil {
		h.logger.Error("Test webhook failed",
			zap.String("webhookId", webhookID.String()),
			zap.Error(err),
		)
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
// @Success 200 {object} domain.WebhookDeliveryList
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/webhooks/{id}/deliveries [get]
func (h *WebhookHandler) ListWebhookDeliveries(c *fiber.Ctx) error {
	webhookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid webhook ID", err)
	}

	filter := domain.WebhookDeliveryFilter{
		WebhookID: webhookID,
	}

	if success := c.Query("success"); success != "" {
		isSuccess := success == "true"
		filter.Success = &isSuccess
	}

	// For now, return empty list - actual implementation would query database
	result := domain.WebhookDeliveryList{
		Deliveries: []domain.WebhookDelivery{},
		TotalCount: 0,
		HasMore:    false,
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
