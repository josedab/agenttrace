package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// roadmapError writes the standard {error, message} envelope used by the
// outcome, replay, Eval Hub, share link, digest, and import endpoints.
func roadmapError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error":   http.StatusText(status),
		"message": message,
	})
}

// roadmapAppError maps an application error onto its own status and keeps
// unexpected failures generic. Callers stay in control of the fallback so the
// flow of each endpoint remains explicit.
func roadmapAppError(
	c *fiber.Ctx,
	logger *zap.Logger,
	err error,
	fallbackStatus int,
	fallbackMessage string,
) error {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		return roadmapError(c, appErr.StatusCode, appErr.Message)
	}
	if logger != nil {
		logger.Error(fallbackMessage, zap.Error(err))
	}
	return roadmapError(c, fallbackStatus, fallbackMessage)
}

// roadmapActorID resolves the acting user for user-foreign-key constrained
// operations (outcomes, replay, Eval Hub, migration import). It resolves to the
// authenticated user, which for owned API keys is the key's creator. It never
// falls back to the API key UUID, because that identifier is not a valid user
// foreign key and would surface as an opaque 500 during persistence. Unowned
// legacy keys therefore resolve to (uuid.Nil, false) so callers return an
// explicit authentication error instead.
func roadmapActorID(c *fiber.Ctx) (uuid.UUID, bool) {
	return middleware.GetUserID(c)
}

// roadmapAttributionID resolves an actor for records that are NOT constrained by
// a user foreign key (e.g. share links). It prefers the authenticated user but
// falls back to the API key identity so machine callers remain attributed.
func roadmapAttributionID(c *fiber.Ctx) (uuid.UUID, bool) {
	if userID, ok := middleware.GetUserID(c); ok {
		return userID, true
	}
	return middleware.GetAPIKeyID(c)
}
