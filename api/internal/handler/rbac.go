package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RBACHandler handles RBAC and SSO HTTP requests
type RBACHandler struct {
	logger  *zap.Logger
	service *service.RBACService
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(logger *zap.Logger, svc *service.RBACService) *RBACHandler {
	return &RBACHandler{
		logger:  logger,
		service: svc,
	}
}

// GetPermissions handles GET /api/public/rbac/permissions
func (h *RBACHandler) GetPermissions(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	roleStr := c.Query("role", string(domain.RoleViewer))
	role := domain.Role(roleStr)

	perms := h.service.GetRolePermissions(role)

	return c.JSON(fiber.Map{
		"role":        role,
		"permissions": perms,
	})
}

// AssignRole handles POST /api/public/rbac/roles
func (h *RBACHandler) AssignRole(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.RoleAssignmentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	assignment, err := h.service.AssignRole(c.Context(), input)
	if err != nil {
		h.logger.Error("failed to assign role", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign role"})
	}

	return c.Status(fiber.StatusCreated).JSON(assignment)
}

// CheckPermission handles POST /api/public/rbac/check
func (h *RBACHandler) CheckPermission(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var req struct {
		UserID     uuid.UUID         `json:"userId"`
		ProjectID  uuid.UUID         `json:"projectId"`
		Permission domain.Permission `json:"permission"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	allowed, err := h.service.CheckPermission(c.Context(), req.UserID, req.ProjectID, req.Permission)
	if err != nil {
		h.logger.Error("failed to check permission", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check permission"})
	}

	return c.JSON(fiber.Map{
		"allowed":    allowed,
		"userId":     req.UserID,
		"projectId":  req.ProjectID,
		"permission": req.Permission,
	})
}

// ConfigureSSO handles POST /api/public/rbac/sso
func (h *RBACHandler) ConfigureSSO(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.SSOConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or missing orgId query parameter"})
	}

	config, err := h.service.ConfigureSSO(c.Context(), orgID, input)
	if err != nil {
		h.logger.Error("failed to configure SSO", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to configure SSO"})
	}

	return c.Status(fiber.StatusCreated).JSON(config)
}

// GetSSOConfig handles GET /api/public/rbac/sso
func (h *RBACHandler) GetSSOConfig(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid or missing orgId query parameter"})
	}

	config, err := h.service.GetSSOConfig(c.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get SSO config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get SSO config"})
	}

	return c.JSON(config)
}

// ScopeAPIKey handles POST /api/public/rbac/api-key-scope
func (h *RBACHandler) ScopeAPIKey(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.APIKeyScopeInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	scope, err := h.service.ScopeAPIKey(c.Context(), input)
	if err != nil {
		h.logger.Error("failed to scope API key", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scope API key"})
	}

	return c.Status(fiber.StatusCreated).JSON(scope)
}
