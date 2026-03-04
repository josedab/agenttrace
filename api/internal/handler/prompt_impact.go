package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// PromptImpactHandler handles prompt impact analysis HTTP requests
type PromptImpactHandler struct {
	service *service.PromptImpactService
	logger  *zap.Logger
}

// NewPromptImpactHandler creates a new prompt impact handler
func NewPromptImpactHandler(svc *service.PromptImpactService, logger *zap.Logger) *PromptImpactHandler {
	return &PromptImpactHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateAnalysis handles POST /api/public/prompt-impact/analyses
func (h *PromptImpactHandler) CreateAnalysis(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.PromptImpactInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.PromptName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Prompt name is required"})
	}
	if input.VersionBefore == "" || input.VersionAfter == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Both versions are required"})
	}

	userID := uuid.New()
	analysis, err := h.service.CreateAnalysis(c.Context(), projectID, userID, &input)
	if err != nil {
		h.logger.Error("failed to create impact analysis", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(analysis)
}

// GetAnalysis handles GET /api/public/prompt-impact/analyses/:analysisId
func (h *PromptImpactHandler) GetAnalysis(c *fiber.Ctx) error {
	analysisID, err := uuid.Parse(c.Params("analysisId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid analysis ID"})
	}

	analysis, err := h.service.GetAnalysis(c.Context(), analysisID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Analysis not found"})
	}

	return c.JSON(analysis)
}

// ListAnalyses handles GET /api/public/prompt-impact/analyses
func (h *PromptImpactHandler) ListAnalyses(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	analyses, err := h.service.ListAnalyses(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if analyses == nil {
		analyses = []domain.PromptVersionImpactAnalysis{}
	}

	return c.JSON(fiber.Map{"analyses": analyses, "count": len(analyses)})
}

// GetReport handles GET /api/public/prompt-impact/analyses/:analysisId/report
func (h *PromptImpactHandler) GetReport(c *fiber.Ctx) error {
	analysisID, err := uuid.Parse(c.Params("analysisId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid analysis ID"})
	}

	report, err := h.service.GetReport(c.Context(), analysisID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// CompareVersions handles POST /api/public/prompt-impact/compare
func (h *PromptImpactHandler) CompareVersions(c *fiber.Ctx) error {
	projectID, err := RequireProjectID(c)
	if err != nil {
		return err
	}

	var input domain.PromptCompareInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.PromptName == "" || input.VersionA == "" || input.VersionB == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Prompt name and both versions are required"})
	}

	analysis, err := h.service.CompareVersions(c.Context(), projectID, &input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(analysis)
}
