package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// OutcomeUseCase defines the transport boundary for outcome analytics.
type OutcomeUseCase interface {
	GetOverview(
		ctx context.Context,
		projectID uuid.UUID,
		from, to time.Time,
	) (*domain.OutcomeOverview, error)
	GetDigest(
		ctx context.Context,
		projectID uuid.UUID,
		from, to time.Time,
	) (*domain.OutcomeDigest, error)
}

// OutcomeHandler transports project outcome analytics requests.
type OutcomeHandler struct {
	service OutcomeUseCase
	logger  *zap.Logger
	clock   func() time.Time
}

// NewOutcomeHandler creates an outcome analytics handler.
func NewOutcomeHandler(outcomeService OutcomeUseCase, logger *zap.Logger) *OutcomeHandler {
	return &OutcomeHandler{
		service: outcomeService,
		logger:  logger,
		clock:   time.Now,
	}
}

// GetOverview handles GET outcome analytics requests.
func (h *OutcomeHandler) GetOverview(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	from, to, err := h.parsePeriod(c)
	if err != nil {
		message := err.Error()
		if appErr := apperrors.GetAppError(err); appErr != nil {
			message = appErr.Message
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": message,
		})
	}

	overview, err := h.service.GetOverview(c.Context(), projectID, from, to)
	if err != nil {
		return h.handleError(c, "get outcome analytics", err)
	}
	return c.JSON(overview)
}

// GetDigest handles GET team digest requests.
func (h *OutcomeHandler) GetDigest(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	from, to, err := h.parsePeriod(c)
	if err != nil {
		message := err.Error()
		if appErr := apperrors.GetAppError(err); appErr != nil {
			message = appErr.Message
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": message,
		})
	}

	digest, err := h.service.GetDigest(c.Context(), projectID, from, to)
	if err != nil {
		return h.handleError(c, "get outcome digest", err)
	}

	format := strings.ToLower(c.Query("format", "json"))
	switch format {
	case "json":
		return c.JSON(digest)
	case "markdown", "md":
		c.Set(fiber.HeaderContentType, "text/markdown; charset=utf-8")
		return c.SendString(service.RenderOutcomeDigestMarkdown(digest))
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "format must be json or markdown",
		})
	}
}

func (h *OutcomeHandler) parsePeriod(
	c *fiber.Ctx,
) (from, to time.Time, resultErr error) {
	now := h.clock().UTC()
	to = now
	from = now.AddDate(0, 0, -30)

	switch c.Query("window") {
	case "", "30d":
	case "24h":
		from = now.Add(-24 * time.Hour)
	case "7d":
		from = now.AddDate(0, 0, -7)
	case "90d":
		from = now.AddDate(0, 0, -90)
	default:
		return time.Time{}, time.Time{}, apperrors.Validation(
			"window must be one of 24h, 7d, 30d, or 90d",
		)
	}

	if value := c.Query("from"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.Validation("from must be RFC3339")
		}
		from = parsed
	}
	if value := c.Query("to"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.Validation("to must be RFC3339")
		}
		to = parsed
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, apperrors.Validation("from must be before to")
	}
	return from.UTC(), to.UTC(), nil
}

func (h *OutcomeHandler) handleError(c *fiber.Ctx, operation string, err error) error {
	if apperrors.IsValidation(err) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": apperrors.GetAppError(err).Message,
		})
	}
	h.logger.Error(operation, zap.Error(err))
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   "Internal Server Error",
		"message": "Failed to load outcome analytics",
	})
}
