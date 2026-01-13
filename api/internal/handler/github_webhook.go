package handler

import (
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// GitHubWebhookHandler handles GitHub App webhooks
type GitHubWebhookHandler struct {
	githubService *service.GitHubAppService
	logger        *zap.Logger
}

// NewGitHubWebhookHandler creates a new GitHub webhook handler
func NewGitHubWebhookHandler(githubService *service.GitHubAppService, logger *zap.Logger) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{
		githubService: githubService,
		logger:        logger,
	}
}

// HandleWebhook handles POST /webhooks/github
func (h *GitHubWebhookHandler) HandleWebhook(c *fiber.Ctx) error {
	// Get event type and delivery ID from headers
	eventType := c.Get("X-GitHub-Event")
	deliveryID := c.Get("X-GitHub-Delivery")
	signature := c.Get("X-Hub-Signature-256")

	if eventType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Missing X-GitHub-Event header",
		})
	}

	// Get raw body
	body := c.Body()

	// Verify signature
	if !h.githubService.VerifyWebhookSignature(body, signature) {
		h.logger.Warn("invalid webhook signature",
			zap.String("event", eventType),
			zap.String("delivery_id", deliveryID),
		)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Invalid signature",
		})
	}

	// Handle ping event
	if eventType == "ping" {
		h.logger.Info("received GitHub ping",
			zap.String("delivery_id", deliveryID),
		)
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "pong",
		})
	}

	// Process the webhook
	if err := h.githubService.HandleWebhook(c.Context(), eventType, deliveryID, body); err != nil {
		h.logger.Error("failed to process webhook",
			zap.String("event", eventType),
			zap.String("delivery_id", deliveryID),
			zap.Error(err),
		)
		// Return 200 to prevent GitHub retries, but log the error
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to process webhook",
		})
	}

	h.logger.Info("processed GitHub webhook",
		zap.String("event", eventType),
		zap.String("delivery_id", deliveryID),
	)

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Webhook processed",
	})
}

// GitHubAppHandler handles GitHub App management endpoints
type GitHubAppHandler struct {
	githubService *service.GitHubAppService
	logger        *zap.Logger
}

// NewGitHubAppHandler creates a new GitHub App management handler
func NewGitHubAppHandler(githubService *service.GitHubAppService, logger *zap.Logger) *GitHubAppHandler {
	return &GitHubAppHandler{
		githubService: githubService,
		logger:        logger,
	}
}

// ListInstallations handles GET /v1/integrations/github/installations
func (h *GitHubAppHandler) ListInstallations(c *fiber.Ctx) error {
	organizationID, ok := middleware.GetOrganizationID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Organization ID not found",
		})
	}

	installations, err := h.githubService.GetInstallations(c.Context(), organizationID)
	if err != nil {
		h.logger.Error("failed to list installations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list installations",
		})
	}

	return c.JSON(fiber.Map{
		"data": installations,
	})
}

// ListRepositories handles GET /v1/integrations/github/installations/:installationId/repositories
func (h *GitHubAppHandler) ListRepositories(c *fiber.Ctx) error {
	installationIDStr := c.Params("installationId")
	if installationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Installation ID required",
		})
	}

	installationID, err := uuid.Parse(installationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid installation ID format",
		})
	}

	repos, err := h.githubService.GetRepositories(c.Context(), installationID)
	if err != nil {
		h.logger.Error("failed to list repositories", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list repositories",
		})
	}

	return c.JSON(fiber.Map{
		"data": repos,
	})
}

// ListProjectRepositories handles GET /v1/projects/:projectId/github/repositories
func (h *GitHubAppHandler) ListProjectRepositories(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	repos, err := h.githubService.GetProjectRepositories(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list project repositories", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list repositories",
		})
	}

	return c.JSON(fiber.Map{
		"data": repos,
	})
}

// LinkRepository handles POST /v1/integrations/github/repositories/link
func (h *GitHubAppHandler) LinkRepository(c *fiber.Ctx) error {
	var input domain.LinkRepoToProjectInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	if input.RepositoryID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "repositoryId is required",
		})
	}

	if input.ProjectID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "projectId is required",
		})
	}

	if err := h.githubService.LinkRepositoryToProject(c.Context(), &input); err != nil {
		h.logger.Error("failed to link repository", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to link repository",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Repository linked successfully",
	})
}

// ListUnlinkedInstallations handles GET /v1/admin/github/installations/unlinked
// Returns installations that were created but not auto-linked to any organization
func (h *GitHubAppHandler) ListUnlinkedInstallations(c *fiber.Ctx) error {
	installations, err := h.githubService.GetUnlinkedInstallations(c.Context())
	if err != nil {
		h.logger.Error("failed to list unlinked installations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list unlinked installations",
		})
	}

	return c.JSON(fiber.Map{
		"data":  installations,
		"count": len(installations),
	})
}

// LinkInstallation handles POST /v1/integrations/github/installations/:installationId/link
func (h *GitHubAppHandler) LinkInstallation(c *fiber.Ctx) error {
	organizationID, ok := middleware.GetOrganizationID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Organization ID not found",
		})
	}

	installationIDStr := c.Params("installationId")
	installationID, err := parseInstallationID(installationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid installation ID",
		})
	}

	if err := h.githubService.LinkInstallationToOrganization(c.Context(), installationID, organizationID); err != nil {
		h.logger.Error("failed to link installation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to link installation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Installation linked successfully",
	})
}

// parseInstallationID parses installation ID from string
func parseInstallationID(s string) (int64, error) {
	var id int64
	_, err := io.WriteString(io.Discard, s)
	if err != nil {
		return 0, err
	}
	// Parse as int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fiber.ErrBadRequest
		}
		id = id*10 + int64(c-'0')
	}
	return id, nil
}

// RegisterWebhookRoutes registers GitHub webhook routes
func (h *GitHubWebhookHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/webhooks/github", h.HandleWebhook)
}

// RegisterGitHubAppRoutes registers GitHub App management routes
func (h *GitHubAppHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAuth())

	// Installation management
	github := v1.Group("/integrations/github")
	github.Get("/installations", h.ListInstallations)
	github.Get("/installations/:installationId/repositories", h.ListRepositories)
	github.Post("/installations/:installationId/link", h.LinkInstallation)
	github.Post("/repositories/link", h.LinkRepository)

	// Admin routes for managing GitHub installations
	admin := v1.Group("/admin/github")
	admin.Get("/installations/unlinked", h.ListUnlinkedInstallations)

	// Project-scoped GitHub routes
	projects := v1.Group("/projects/:projectId/github", authMiddleware.RequireAPIKey())
	projects.Get("/repositories", h.ListProjectRepositories)
}
