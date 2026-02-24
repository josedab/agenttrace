package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ComplianceReportHandler handles compliance report HTTP requests
type ComplianceReportHandler struct {
	service *service.ComplianceReportService
	logger  *zap.Logger
}

// NewComplianceReportHandler creates a new compliance report handler
func NewComplianceReportHandler(svc *service.ComplianceReportService, logger *zap.Logger) *ComplianceReportHandler {
	return &ComplianceReportHandler{
		service: svc,
		logger:  logger,
	}
}

// Generate handles POST /api/public/compliance-reports
func (h *ComplianceReportHandler) Generate(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.GenerateReportInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Template == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Template is required"})
	}

	report, err := h.service.GenerateReport(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to generate compliance report", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate report"})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// List handles GET /api/public/compliance-reports
func (h *ComplianceReportHandler) List(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	reports := h.service.ListReports(c.Context(), projectID)
	if reports == nil {
		reports = []domain.ComplianceReport{}
	}

	return c.JSON(fiber.Map{"reports": reports, "count": len(reports)})
}

// Get handles GET /api/public/compliance-reports/:reportId
func (h *ComplianceReportHandler) Get(c *fiber.Ctx) error {
	reportID, err := uuid.Parse(c.Params("reportId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid report ID"})
	}

	report, err := h.service.GetReport(c.Context(), reportID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
	}

	return c.JSON(report)
}

// GetTemplates handles GET /api/public/compliance-reports/templates
func (h *ComplianceReportHandler) GetTemplates(c *fiber.Ctx) error {
	templates := h.service.GetTemplates(c.Context())
	return c.JSON(fiber.Map{"templates": templates, "count": len(templates)})
}
