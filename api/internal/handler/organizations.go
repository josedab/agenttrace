package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OrganizationsHandler handles organization endpoints
type OrganizationsHandler struct {
	orgService *service.OrgService
	logger     *zap.Logger
}

// NewOrganizationsHandler creates a new organizations handler
func NewOrganizationsHandler(orgService *service.OrgService, logger *zap.Logger) *OrganizationsHandler {
	return &OrganizationsHandler{
		orgService: orgService,
		logger:     logger,
	}
}

// ListOrganizations handles GET /v1/organizations
func (h *OrganizationsHandler) ListOrganizations(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}

	orgs, err := h.orgService.ListByUser(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list organizations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list organizations",
		})
	}

	return c.JSON(fiber.Map{
		"data": orgs,
	})
}

// GetOrganization handles GET /v1/organizations/:orgId
func (h *OrganizationsHandler) GetOrganization(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid organization ID",
		})
	}

	org, err := h.orgService.Get(c.Context(), orgID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Organization not found",
			})
		}
		h.logger.Error("failed to get organization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get organization",
		})
	}

	return c.JSON(org)
}

// GetOrganizationBySlug handles GET /v1/organizations/slug/:slug
func (h *OrganizationsHandler) GetOrganizationBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Organization slug required",
		})
	}

	org, err := h.orgService.GetBySlug(c.Context(), slug)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Organization not found",
			})
		}
		h.logger.Error("failed to get organization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get organization",
		})
	}

	return c.JSON(org)
}

// CreateOrganization handles POST /v1/organizations
func (h *OrganizationsHandler) CreateOrganization(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	org, err := h.orgService.Create(c.Context(), input.Name, userID)
	if err != nil {
		h.logger.Error("failed to create organization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create organization",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(org)
}

// UpdateOrganization handles PATCH /v1/organizations/:orgId
func (h *OrganizationsHandler) UpdateOrganization(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid organization ID",
		})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	org, err := h.orgService.Update(c.Context(), orgID, input.Name)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Organization not found",
			})
		}
		h.logger.Error("failed to update organization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update organization",
		})
	}

	return c.JSON(org)
}

// DeleteOrganization handles DELETE /v1/organizations/:orgId
func (h *OrganizationsHandler) DeleteOrganization(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid organization ID",
		})
	}

	if err := h.orgService.Delete(c.Context(), orgID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Organization not found",
			})
		}
		h.logger.Error("failed to delete organization", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete organization",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetMember handles GET /v1/organizations/:orgId/members/:userId
func (h *OrganizationsHandler) GetMember(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid organization ID",
		})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid user ID",
		})
	}

	member, err := h.orgService.GetMember(c.Context(), orgID, userID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Member not found",
			})
		}
		h.logger.Error("failed to get member", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get member",
		})
	}

	return c.JSON(member)
}

// RegisterRoutes registers organization routes
func (h *OrganizationsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireJWT())

	v1.Get("/organizations", h.ListOrganizations)
	v1.Get("/organizations/slug/:slug", h.GetOrganizationBySlug)
	v1.Get("/organizations/:orgId", h.GetOrganization)
	v1.Post("/organizations", h.CreateOrganization)
	v1.Patch("/organizations/:orgId", h.UpdateOrganization)
	v1.Delete("/organizations/:orgId", h.DeleteOrganization)

	v1.Get("/organizations/:orgId/members/:userId", h.GetMember)
}
