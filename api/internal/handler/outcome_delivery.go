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

// GitHubOutcomeDelivery defines the optional delivery use case.
type GitHubOutcomeDelivery interface {
	Deliver(
		ctx context.Context,
		projectID uuid.UUID,
		input domain.GitHubOutcomeReportInput,
	) (*domain.GitHubOutcomeReportResult, error)
}

// OutcomeDeliveryHandler transports optional report delivery requests.
type OutcomeDeliveryHandler struct {
	github GitHubOutcomeDelivery
	logger *zap.Logger
}

// NewOutcomeDeliveryHandler creates an outcome delivery handler.
func NewOutcomeDeliveryHandler(
	github GitHubOutcomeDelivery,
	logger *zap.Logger,
) *OutcomeDeliveryHandler {
	return &OutcomeDeliveryHandler{github: github, logger: logger}
}

// DeliverGitHub handles POST /outcomes/github-report.
func (h *OutcomeDeliveryHandler) DeliverGitHub(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return outcomeDeliveryError(c, fiber.StatusUnauthorized, "Project ID not found")
	}
	var input domain.GitHubOutcomeReportInput
	if err := c.BodyParser(&input); err != nil {
		return outcomeDeliveryError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	result, err := h.github.Deliver(c.Context(), projectID, input)
	if err != nil {
		if appErr := apperrors.GetAppError(err); appErr != nil {
			return outcomeDeliveryError(c, appErr.StatusCode, appErr.Message)
		}
		h.logger.Error("GitHub outcome report delivery failed", zap.Error(err))
		return outcomeDeliveryError(
			c,
			fiber.StatusBadGateway,
			"GitHub outcome report delivery failed",
		)
	}
	return c.JSON(result)
}

func outcomeDeliveryError(c *fiber.Ctx, status int, message string) error {
	return roadmapError(c, status, message)
}
