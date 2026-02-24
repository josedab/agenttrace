package handler

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// CloudOnboardingHandler handles cloud onboarding HTTP requests
type CloudOnboardingHandler struct {
	logger  *zap.Logger
	service *service.CloudOnboardingService
}

// NewCloudOnboardingHandler creates a new cloud onboarding handler
func NewCloudOnboardingHandler(svc *service.CloudOnboardingService, logger *zap.Logger) *CloudOnboardingHandler {
	return &CloudOnboardingHandler{
		logger:  logger,
		service: svc,
	}
}

// GetOnboarding handles GET /api/public/cloud-onboarding
// @Summary Get onboarding status
// @Description Get the current onboarding status for a tenant
// @Tags cloud-onboarding
// @Accept json
// @Produce json
// @Success 200 {object} domain.CloudOnboarding
// @Failure 401 {object} map[string]string
// @Router /api/public/cloud-onboarding [get]
func (h *CloudOnboardingHandler) GetOnboarding(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	onboarding, err := h.service.GetOnboarding(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get onboarding", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get onboarding"})
	}

	return c.JSON(onboarding)
}

// CloudOnboardingCompleteStepRequest represents the request to complete an onboarding step
type CloudOnboardingCompleteStepRequest struct {
	Step domain.OnboardingStep `json:"step"`
}

// CompleteStep handles POST /api/public/cloud-onboarding/steps/complete
// @Summary Complete onboarding step
// @Description Mark an onboarding step as completed
// @Tags cloud-onboarding
// @Accept json
// @Produce json
// @Param body body CloudOnboardingCompleteStepRequest true "Step to complete"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/cloud-onboarding/steps/complete [post]
func (h *CloudOnboardingHandler) CompleteStep(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req CloudOnboardingCompleteStepRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.CompleteStep(c.Context(), projectID, req.Step); err != nil {
		h.logger.Error("failed to complete step", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to complete step"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// CloudOnboardingQuickstartRequest represents the request to generate a quickstart
type CloudOnboardingQuickstartRequest struct {
	Language  string `json:"language"`
	Framework string `json:"framework"`
}

// GenerateQuickstart handles POST /api/public/cloud-onboarding/quickstart
// @Summary Generate quickstart
// @Description Generate a quickstart configuration for a language and framework
// @Tags cloud-onboarding
// @Accept json
// @Produce json
// @Param body body CloudOnboardingQuickstartRequest true "Quickstart parameters"
// @Success 200 {object} domain.QuickstartConfig
// @Failure 400 {object} map[string]string
// @Router /api/public/cloud-onboarding/quickstart [post]
func (h *CloudOnboardingHandler) GenerateQuickstart(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req CloudOnboardingQuickstartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Language == "" || req.Framework == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Language and framework are required"})
	}

	config, err := h.service.GenerateQuickstart(c.Context(), req.Language, req.Framework, "", projectID.String(), "")
	if err != nil {
		h.logger.Error("failed to generate quickstart", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate quickstart"})
	}

	return c.JSON(config)
}

// GetUsage handles GET /api/public/cloud-onboarding/usage
// @Summary Get usage
// @Description Get usage metrics for the tenant
// @Tags cloud-onboarding
// @Accept json
// @Produce json
// @Success 200 {object} domain.UsageMeter
// @Failure 401 {object} map[string]string
// @Router /api/public/cloud-onboarding/usage [get]
func (h *CloudOnboardingHandler) GetUsage(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	usage, err := h.service.GetUsage(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get usage", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get usage"})
	}

	return c.JSON(usage)
}

// CloudOnboardingCheckQuotaRequest represents the request to check quota
type CloudOnboardingCheckQuotaRequest struct {
	Metric   string `json:"metric"`
	Quantity int64  `json:"quantity"`
}

// CheckQuota handles POST /api/public/cloud-onboarding/quota/check
// @Summary Check quota
// @Description Check if a quota allows for the specified quantity
// @Tags cloud-onboarding
// @Accept json
// @Produce json
// @Param body body CloudOnboardingCheckQuotaRequest true "Quota check parameters"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/public/cloud-onboarding/quota/check [post]
func (h *CloudOnboardingHandler) CheckQuota(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req CloudOnboardingCheckQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Metric == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Metric is required"})
	}

	allowed, err := h.service.CheckQuota(c.Context(), projectID, req.Metric, req.Quantity)
	if err != nil {
		h.logger.Error("failed to check quota", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check quota"})
	}

	return c.JSON(fiber.Map{"allowed": allowed})
}
