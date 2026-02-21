package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ComplianceExportHandler handles compliance export HTTP requests
type ComplianceExportHandler struct {
	complianceExportService *service.ComplianceExportService
	logger                  *zap.Logger
}

// NewComplianceExportHandler creates a new compliance export handler
func NewComplianceExportHandler(complianceExportService *service.ComplianceExportService, logger *zap.Logger) *ComplianceExportHandler {
	return &ComplianceExportHandler{
		complianceExportService: complianceExportService,
		logger:                  logger,
	}
}

// StartExport handles POST /compliance/exports
func (h *ComplianceExportHandler) StartExport(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ComplianceExportInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	job, err := h.complianceExportService.StartExport(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to start compliance export",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to start compliance export",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(job)
}

// GetExport handles GET /compliance/exports/:id
func (h *ComplianceExportHandler) GetExport(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid export job ID",
		})
	}

	job, err := h.complianceExportService.GetExportJob(c.Context(), jobID)
	if err != nil {
		h.logger.Error("failed to get compliance export",
			zap.String("jobId", jobID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get compliance export",
		})
	}

	return c.JSON(job)
}

// ListExports handles GET /compliance/exports
func (h *ComplianceExportHandler) ListExports(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	jobs, err := h.complianceExportService.ListExports(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list compliance exports",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to list compliance exports",
		})
	}

	return c.JSON(jobs)
}

// GetTemplates handles GET /compliance/templates
func (h *ComplianceExportHandler) GetTemplates(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	templates := h.complianceExportService.GetTemplates()
	return c.JSON(templates)
}
