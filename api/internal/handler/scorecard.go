package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ScorecardHandler handles scorecard HTTP requests
type ScorecardHandler struct {
	scorecardService *service.ScorecardService
	logger           *zap.Logger
}

// NewScorecardHandler creates a new scorecard handler
func NewScorecardHandler(scorecardService *service.ScorecardService, logger *zap.Logger) *ScorecardHandler {
	return &ScorecardHandler{
		scorecardService: scorecardService,
		logger:           logger,
	}
}

// Generate handles POST /scorecards
func (h *ScorecardHandler) Generate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ScorecardInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	scorecard, err := h.scorecardService.Generate(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to generate scorecard",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to generate scorecard",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(scorecard)
}

// GetScorecard handles GET /scorecards/:id
func (h *ScorecardHandler) GetScorecard(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	scorecardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid scorecard ID",
		})
	}

	scorecard, err := h.scorecardService.GetScorecard(c.Context(), scorecardID)
	if err != nil {
		h.logger.Error("failed to get scorecard",
			zap.String("scorecardId", scorecardID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get scorecard",
		})
	}

	return c.JSON(scorecard)
}

// ListScorecards handles GET /scorecards?agentName=
func (h *ScorecardHandler) ListScorecards(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	agentName := c.Query("agentName")

	scorecards, err := h.scorecardService.ListScorecards(c.Context(), projectID, agentName)
	if err != nil {
		h.logger.Error("failed to list scorecards",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list scorecards",
		})
	}

	return c.JSON(scorecards)
}

// ConfigureAuto handles POST /scorecards/config
func (h *ScorecardHandler) ConfigureAuto(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var config domain.ScorecardConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}
	config.ProjectID = projectID

	if err := h.scorecardService.ConfigureAutoGeneration(c.Context(), config); err != nil {
		h.logger.Error("failed to configure scorecard auto-generation",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to configure scorecard auto-generation",
		})
	}

	return c.JSON(fiber.Map{"status": "configured"})
}

// GetConfig handles GET /scorecards/config
func (h *ScorecardHandler) GetConfig(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	config, err := h.scorecardService.GetConfig(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get scorecard config",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get scorecard config",
		})
	}

	return c.JSON(config)
}
