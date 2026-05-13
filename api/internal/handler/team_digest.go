package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// TeamDigestUseCase defines digest delivery behavior.
type TeamDigestUseCase interface {
	Deliver(
		ctx context.Context,
		projectID uuid.UUID,
		input domain.TeamDigestDeliveryInput,
	) (*domain.TeamDigestDeliveryResult, error)
}

// TeamDigestHandler transports team digest delivery requests.
type TeamDigestHandler struct {
	service TeamDigestUseCase
	logger  *zap.Logger
}

// NewTeamDigestHandler creates a digest handler.
func NewTeamDigestHandler(service TeamDigestUseCase, logger *zap.Logger) *TeamDigestHandler {
	return &TeamDigestHandler{service: service, logger: logger}
}

// Deliver handles POST /outcomes/digest/deliver.
func (h *TeamDigestHandler) Deliver(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return teamDigestError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	var input domain.TeamDigestDeliveryInput
	if err := c.BodyParser(&input); err != nil {
		return teamDigestError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	result, err := h.service.Deliver(c.Context(), projectID, input)
	if err != nil {
		return roadmapAppError(
			c,
			h.logger,
			err,
			fiber.StatusBadGateway,
			"Team digest delivery failed",
		)
	}
	return c.JSON(result)
}

func teamDigestError(c *fiber.Ctx, status int, message string) error {
	return roadmapError(c, status, message)
}
