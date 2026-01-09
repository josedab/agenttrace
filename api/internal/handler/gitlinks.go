package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// GitLinksHandler handles git link endpoints
type GitLinksHandler struct {
	gitLinkService *service.GitLinkService
	logger         *zap.Logger
}

// NewGitLinksHandler creates a new git links handler
func NewGitLinksHandler(gitLinkService *service.GitLinkService, logger *zap.Logger) *GitLinksHandler {
	return &GitLinksHandler{
		gitLinkService: gitLinkService,
		logger:         logger,
	}
}

// ListGitLinks handles GET /v1/git-links
func (h *GitLinksHandler) ListGitLinks(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	filter := h.parseGitLinkFilter(c)
	filter.ProjectID = projectID
	limit := parseIntParam(c, "limit", 50)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	list, err := h.gitLinkService.List(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list git links", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list git links",
		})
	}

	return c.JSON(fiber.Map{
		"data":       list.GitLinks,
		"totalCount": list.TotalCount,
		"hasMore":    list.HasMore,
	})
}

// GetGitLink handles GET /v1/git-links/:gitLinkId
func (h *GitLinksHandler) GetGitLink(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	gitLinkIDStr := c.Params("gitLinkId")
	if gitLinkIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Git link ID required",
		})
	}

	gitLinkID, err := uuid.Parse(gitLinkIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid git link ID format",
		})
	}

	gitLink, err := h.gitLinkService.Get(c.Context(), projectID, gitLinkID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Git link not found",
			})
		}
		h.logger.Error("failed to get git link", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get git link",
		})
	}

	return c.JSON(gitLink)
}

// CreateGitLink handles POST /v1/git-links
func (h *GitLinksHandler) CreateGitLink(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.GitLinkInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if input.TraceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "traceId is required",
		})
	}

	if input.CommitSha == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "commitSha is required",
		})
	}

	gitLink, err := h.gitLinkService.Create(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create git link", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create git link",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(gitLink)
}

// GetByCommit handles GET /v1/git-links/commit/:commitSha
func (h *GitLinksHandler) GetByCommit(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	commitSha := c.Params("commitSha")
	if commitSha == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Commit SHA required",
		})
	}

	gitLinks, err := h.gitLinkService.GetByCommitSha(c.Context(), projectID, commitSha)
	if err != nil {
		h.logger.Error("failed to get git links by commit", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get git links",
		})
	}

	return c.JSON(fiber.Map{
		"data": gitLinks,
	})
}

// GetTimeline handles GET /v1/git-links/timeline
func (h *GitLinksHandler) GetTimeline(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	branch := c.Query("branch")
	limit := parseIntParam(c, "limit", 50)

	if limit > 100 {
		limit = 100
	}

	timeline, err := h.gitLinkService.GetTimeline(c.Context(), projectID, branch, limit)
	if err != nil {
		h.logger.Error("failed to get git timeline", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get git timeline",
		})
	}

	return c.JSON(timeline)
}

// GetTraceGitLinks handles GET /v1/traces/:traceId/git-links
func (h *GitLinksHandler) GetTraceGitLinks(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	traceID := c.Params("traceId")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Trace ID required",
		})
	}

	gitLinks, err := h.gitLinkService.GetByTraceID(c.Context(), projectID, traceID)
	if err != nil {
		h.logger.Error("failed to get trace git links", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get trace git links",
		})
	}

	return c.JSON(fiber.Map{
		"data": gitLinks,
	})
}

// parseGitLinkFilter parses git link filter from query params
func (h *GitLinksHandler) parseGitLinkFilter(c *fiber.Ctx) *domain.GitLinkFilter {
	filter := &domain.GitLinkFilter{}

	if traceID := c.Query("traceId"); traceID != "" {
		filter.TraceID = &traceID
	}

	if commitSha := c.Query("commitSha"); commitSha != "" {
		filter.CommitSha = &commitSha
	}

	if branch := c.Query("branch"); branch != "" {
		filter.Branch = &branch
	}

	if linkType := c.Query("linkType"); linkType != "" {
		lt := domain.GitLinkType(linkType)
		filter.LinkType = &lt
	}

	if from := c.Query("fromTimestamp"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.FromTime = &t
		}
	}

	if to := c.Query("toTimestamp"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.ToTime = &t
		}
	}

	return filter
}

// RegisterRoutes registers git link routes
func (h *GitLinksHandler) RegisterRoutes(app *fiber.App, authMiddleware *middleware.AuthMiddleware) {
	v1 := app.Group("/v1", authMiddleware.RequireAPIKey())

	v1.Get("/git-links", h.ListGitLinks)
	v1.Get("/git-links/timeline", h.GetTimeline)
	v1.Get("/git-links/commit/:commitSha", h.GetByCommit)
	v1.Get("/git-links/:gitLinkId", h.GetGitLink)
	v1.Post("/git-links", h.CreateGitLink)

	v1.Get("/traces/:traceId/git-links", h.GetTraceGitLinks)
}
