package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ShareLinkUseCase defines the transport boundary for share links.
type ShareLinkUseCase interface {
	Create(
		ctx context.Context,
		projectID, actorID uuid.UUID,
		input domain.ShareLinkInput,
	) (*domain.ShareLinkCreated, error)
	Resolve(ctx context.Context, token string) (*domain.SharedResourceView, error)
	Revoke(ctx context.Context, projectID, linkID uuid.UUID) error
}

// ShareLinkHandler transports authenticated creation/revocation and public resolution.
type ShareLinkHandler struct {
	service ShareLinkUseCase
	logger  *zap.Logger
}

// NewShareLinkHandler creates a share link handler.
func NewShareLinkHandler(service ShareLinkUseCase, logger *zap.Logger) *ShareLinkHandler {
	return &ShareLinkHandler{service: service, logger: logger}
}

// Create handles POST /share-links.
func (h *ShareLinkHandler) Create(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return shareLinkError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	actorID, ok := shareLinkActorID(c)
	if !ok {
		return shareLinkError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}

	var input domain.ShareLinkInput
	if err := c.BodyParser(&input); err != nil {
		return shareLinkError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	created, err := h.service.Create(c.Context(), projectID, actorID, input)
	if err != nil {
		return h.handleError(c, err, false)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// CreateTrace handles POST /traces/:traceId/share-links.
func (h *ShareLinkHandler) CreateTrace(c *fiber.Ctx) error {
	return h.createForResource(c, domain.ShareResourceTrace, c.Params("traceId"))
}

// CreateReplayPlan handles POST /replay-plans/:planId/share-links.
func (h *ShareLinkHandler) CreateReplayPlan(c *fiber.Ctx) error {
	return h.createForResource(c, domain.ShareResourceReplayPlan, c.Params("planId"))
}

func (h *ShareLinkHandler) createForResource(
	c *fiber.Ctx,
	resourceType domain.ShareResourceType,
	resourceID string,
) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return shareLinkError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	actorID, ok := shareLinkActorID(c)
	if !ok {
		return shareLinkError(c, fiber.StatusUnauthorized, "Actor ID not found")
	}
	var request struct {
		ExpiresInSeconds int64 `json:"expiresInSeconds,omitempty"`
	}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&request); err != nil {
			return shareLinkError(c, fiber.StatusBadRequest, "Invalid request body")
		}
	}
	created, err := h.service.Create(c.Context(), projectID, actorID, domain.ShareLinkInput{
		ResourceType:     resourceType,
		ResourceID:       resourceID,
		ExpiresInSeconds: request.ExpiresInSeconds,
	})
	if err != nil {
		return h.handleError(c, err, false)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// Revoke handles DELETE /share-links/:linkId.
func (h *ShareLinkHandler) Revoke(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return shareLinkError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	linkID, err := uuid.Parse(c.Params("linkId"))
	if err != nil {
		return shareLinkError(c, fiber.StatusBadRequest, "Invalid share link ID")
	}
	if err := h.service.Revoke(c.Context(), projectID, linkID); err != nil {
		return h.handleError(c, err, false)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Resolve handles unauthenticated GET /api/share/:token.
func (h *ShareLinkHandler) Resolve(c *fiber.Ctx) error {
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	c.Set("X-Robots-Tag", "noindex, nofollow")
	view, err := h.service.Resolve(c.Context(), c.Params("token"))
	if err != nil {
		return h.handleError(c, err, true)
	}
	return c.JSON(view)
}

func (h *ShareLinkHandler) handleError(
	c *fiber.Ctx,
	err error,
	public bool,
) error {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		if public {
			return shareLinkError(c, fiber.StatusNotFound, "Share link not found")
		}
		return shareLinkError(c, appErr.StatusCode, appErr.Message)
	}
	h.logger.Error("share link request failed", zap.Error(err))
	if public {
		return shareLinkError(c, fiber.StatusNotFound, "Share link not found")
	}
	return shareLinkError(c, fiber.StatusInternalServerError, "Share link request failed")
}

// shareLinkActorID attributes the creator of a share link. Share links are not
// constrained by a user foreign key, so machine callers may be attributed by
// their API key identity when no user is present.
func shareLinkActorID(c *fiber.Ctx) (uuid.UUID, bool) {
	return roadmapAttributionID(c)
}

func shareLinkError(c *fiber.Ctx, status int, message string) error {
	return roadmapError(c, status, message)
}
