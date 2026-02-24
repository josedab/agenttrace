package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// SecurityScannerHandler handles security scanner HTTP requests
type SecurityScannerHandler struct {
	logger  *zap.Logger
	service *service.SecurityScannerService
}

// NewSecurityScannerHandler creates a new security scanner handler
func NewSecurityScannerHandler(svc *service.SecurityScannerService, logger *zap.Logger) *SecurityScannerHandler {
	return &SecurityScannerHandler{
		logger:  logger,
		service: svc,
	}
}

// SecurityScannerScanRequest represents the request to scan a trace
type SecurityScannerScanRequest struct {
	TraceID string `json:"traceId"`
}

// ScanTrace handles POST /api/public/security-scanner/scan
// @Summary Scan trace
// @Description Scan a trace for security issues
// @Tags security-scanner
// @Accept json
// @Produce json
// @Param body body SecurityScannerScanRequest true "Scan request"
// @Success 200 {object} domain.SecurityScanResult
// @Failure 400 {object} map[string]string
// @Router /api/public/security-scanner/scan [post]
func (h *SecurityScannerHandler) ScanTrace(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req SecurityScannerScanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	traceID, err := uuid.Parse(req.TraceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	result, err := h.service.ScanTrace(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to scan trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan trace"})
	}

	return c.JSON(result)
}

// CreateSecurityPolicy handles POST /api/public/security-scanner/policies
// @Summary Create security policy
// @Description Create a new security policy
// @Tags security-scanner
// @Accept json
// @Produce json
// @Param policy body domain.SecurityPolicy true "Security policy"
// @Success 201 {object} domain.SecurityPolicy
// @Failure 400 {object} map[string]string
// @Router /api/public/security-scanner/policies [post]
func (h *SecurityScannerHandler) CreateSecurityPolicy(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.SecurityPolicy
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	policy, err := h.service.CreatePolicy(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create security policy", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create security policy"})
	}

	return c.Status(fiber.StatusCreated).JSON(policy)
}

// ListSecurityPolicies handles GET /api/public/security-scanner/policies
// @Summary List security policies
// @Description List all security policies for a project
// @Tags security-scanner
// @Accept json
// @Produce json
// @Success 200 {array} domain.SecurityPolicy
// @Failure 401 {object} map[string]string
// @Router /api/public/security-scanner/policies [get]
func (h *SecurityScannerHandler) ListSecurityPolicies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	policies, err := h.service.ListPolicies(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list security policies", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list security policies"})
	}

	return c.JSON(policies)
}

// GetSecurityDashboard handles GET /api/public/security-scanner/dashboard
// @Summary Get security dashboard
// @Description Get the security dashboard for a project
// @Tags security-scanner
// @Accept json
// @Produce json
// @Success 200 {object} domain.SecurityDashboard
// @Failure 401 {object} map[string]string
// @Router /api/public/security-scanner/dashboard [get]
func (h *SecurityScannerHandler) GetSecurityDashboard(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	dashboard, err := h.service.GetDashboard(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get security dashboard", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get security dashboard"})
	}

	return c.JSON(dashboard)
}

// AcknowledgeSecurityFinding handles POST /api/public/security-scanner/findings/:findingId/acknowledge
// @Summary Acknowledge security finding
// @Description Acknowledge a security finding
// @Tags security-scanner
// @Accept json
// @Produce json
// @Param findingId path string true "Finding ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/public/security-scanner/findings/{findingId}/acknowledge [post]
func (h *SecurityScannerHandler) AcknowledgeSecurityFinding(c *fiber.Ctx) error {
	findingID, err := uuid.Parse(c.Params("findingId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid finding ID"})
	}

	if err := h.service.AcknowledgeFinding(c.Context(), findingID); err != nil {
		h.logger.Error("failed to acknowledge security finding", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to acknowledge finding"})
	}

	return c.JSON(fiber.Map{"status": "acknowledged"})
}
