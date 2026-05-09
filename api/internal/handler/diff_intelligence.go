package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// DiffIntelligenceHandler handles diff intelligence endpoints
type DiffIntelligenceHandler struct {
	diffService *service.DiffIntelligenceService
	logger      *zap.Logger
}

// NewDiffIntelligenceHandler creates a new handler
func NewDiffIntelligenceHandler(diffService *service.DiffIntelligenceService, logger *zap.Logger) *DiffIntelligenceHandler {
	return &DiffIntelligenceHandler{
		diffService: diffService,
		logger:      logger,
	}
}

// AnalyzeDiff handles POST /api/public/diff-analysis
func (h *DiffIntelligenceHandler) AnalyzeDiff(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.DiffAnalysisInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	analysis, err := h.diffService.AnalyzeDiff(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to analyze diff", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to analyze diff"})
	}

	return c.Status(fiber.StatusCreated).JSON(analysis)
}

// GetAnalysis handles GET /api/public/diff-analysis/:id
func (h *DiffIntelligenceHandler) GetAnalysis(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid analysis ID"})
	}

	analysis, err := h.diffService.GetAnalysis(c.Context(), projectID, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Analysis not found"})
	}

	return c.JSON(analysis)
}

// ListAnalyses handles GET /api/public/diff-analysis
func (h *DiffIntelligenceHandler) ListAnalyses(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filter := &domain.DiffAnalysisFilter{ProjectID: projectID}

	analyses, total, err := h.diffService.ListAnalyses(c.Context(), filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list analyses", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list analyses"})
	}

	return c.JSON(fiber.Map{
		"analyses":   analyses,
		"totalCount": total,
		"hasMore":    int64(offset+limit) < total,
	})
}

// GetTraceAnalyses handles GET /api/public/traces/:traceId/diff-analysis
func (h *DiffIntelligenceHandler) GetTraceAnalyses(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	traceID, err := uuid.Parse(c.Params("traceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trace ID"})
	}

	analyses, err := h.diffService.GetTraceAnalyses(c.Context(), projectID, traceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get analyses"})
	}

	return c.JSON(fiber.Map{"analyses": analyses})
}

// GetQualityTrend handles GET /api/public/diff-analysis/trend
func (h *DiffIntelligenceHandler) GetQualityTrend(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days <= 0 || days > 365 {
		days = 30
	}

	trend, err := h.diffService.GetQualityTrend(c.Context(), projectID, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get trend"})
	}

	return c.JSON(trend)
}
