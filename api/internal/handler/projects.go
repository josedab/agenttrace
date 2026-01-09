package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ProjectsHandler handles project endpoints
type ProjectsHandler struct {
	projectService *service.ProjectService
	logger         *zap.Logger
}

// NewProjectsHandler creates a new projects handler
func NewProjectsHandler(projectService *service.ProjectService, logger *zap.Logger) *ProjectsHandler {
	return &ProjectsHandler{
		projectService: projectService,
		logger:         logger,
	}
}

// ListProjects handles GET /v1/projects
func (h *ProjectsHandler) ListProjects(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}

	// Check for organization filter
	orgIDStr := c.Query("organizationId")
	var projects []domain.Project
	var err error

	if orgIDStr != "" {
		orgID, parseErr := uuid.Parse(orgIDStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid organization ID",
			})
		}
		projects, err = h.projectService.ListByOrganization(c.Context(), orgID)
	} else {
		projects, err = h.projectService.ListByUser(c.Context(), userID)
	}

	if err != nil {
		h.logger.Error("failed to list projects", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list projects",
		})
	}

	return c.JSON(fiber.Map{
		"data": projects,
	})
}

// GetProject handles GET /v1/projects/:projectId
func (h *ProjectsHandler) GetProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid project ID",
		})
	}

	project, err := h.projectService.Get(c.Context(), projectID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Project not found",
			})
		}
		h.logger.Error("failed to get project", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get project",
		})
	}

	return c.JSON(project)
}

// CreateProject handles POST /v1/projects
func (h *ProjectsHandler) CreateProject(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}

	var input struct {
		OrganizationID  string                  `json:"organizationId"`
		Name            string                  `json:"name"`
		Description     string                  `json:"description,omitempty"`
		Settings        *domain.ProjectSettings `json:"settings,omitempty"`
		RetentionDays   *int                    `json:"retentionDays,omitempty"`
		RateLimitPerMin *int                    `json:"rateLimitPerMin,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.OrganizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "organizationId is required",
		})
	}

	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid organization ID",
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "name is required",
		})
	}

	projectInput := &service.ProjectInput{
		Name:            input.Name,
		Description:     input.Description,
		Settings:        input.Settings,
		RetentionDays:   input.RetentionDays,
		RateLimitPerMin: input.RateLimitPerMin,
	}

	project, err := h.projectService.Create(c.Context(), orgID, projectInput, userID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Organization not found",
			})
		}
		h.logger.Error("failed to create project", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create project",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(project)
}

// UpdateProject handles PATCH /v1/projects/:projectId
func (h *ProjectsHandler) UpdateProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid project ID",
		})
	}

	var input service.ProjectInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	project, err := h.projectService.Update(c.Context(), projectID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Project not found",
			})
		}
		h.logger.Error("failed to update project", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to update project",
		})
	}

	return c.JSON(project)
}

// DeleteProject handles DELETE /v1/projects/:projectId
func (h *ProjectsHandler) DeleteProject(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid project ID",
		})
	}

	if err := h.projectService.Delete(c.Context(), projectID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Project not found",
			})
		}
		h.logger.Error("failed to delete project", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to delete project",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddMember handles POST /v1/projects/:projectId/members
func (h *ProjectsHandler) AddMember(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid project ID",
		})
	}

	var input struct {
		UserID string         `json:"userId"`
		Role   domain.OrgRole `json:"role"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid user ID",
		})
	}

	if err := h.projectService.AddMember(c.Context(), projectID, userID, input.Role); err != nil {
		h.logger.Error("failed to add member", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to add member",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Member added successfully",
	})
}

// RemoveMember handles DELETE /v1/projects/:projectId/members/:userId
func (h *ProjectsHandler) RemoveMember(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid project ID",
		})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid user ID",
		})
	}

	if err := h.projectService.RemoveMember(c.Context(), projectID, userID); err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Member not found",
			})
		}
		h.logger.Error("failed to remove member", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to remove member",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetUserRole handles GET /v1/projects/:projectId/role
func (h *ProjectsHandler) GetUserRole(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid project ID",
		})
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "User ID not found",
		})
	}

	role, err := h.projectService.GetUserRole(c.Context(), projectID, userID)
	if err != nil {
		h.logger.Error("failed to get user role", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get user role",
		})
	}

	if role == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "User not a member of this project",
		})
	}

	return c.JSON(fiber.Map{
		"role": role,
	})
}

// RegisterRoutes registers project routes
func (h *ProjectsHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireJWT())

	v1.Get("/projects", h.ListProjects)
	v1.Get("/projects/:projectId", h.GetProject)
	v1.Post("/projects", h.CreateProject)
	v1.Patch("/projects/:projectId", h.UpdateProject)
	v1.Delete("/projects/:projectId", h.DeleteProject)

	v1.Post("/projects/:projectId/members", h.AddMember)
	v1.Delete("/projects/:projectId/members/:userId", h.RemoveMember)
	v1.Get("/projects/:projectId/role", h.GetUserRole)
}
