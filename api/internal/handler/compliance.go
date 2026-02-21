package handler

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// ComplianceHandler handles compliance HTTP requests
type ComplianceHandler struct {
	complianceService *service.ComplianceService
	logger            *zap.Logger
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(complianceService *service.ComplianceService, logger *zap.Logger) *ComplianceHandler {
	return &ComplianceHandler{
		complianceService: complianceService,
		logger:            logger,
	}
}

// AssessProject handles GET /compliance/assess
func (h *ComplianceHandler) AssessProject(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	record, err := h.complianceService.AssessProject(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to assess project compliance",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to assess project compliance",
		})
	}

	return c.JSON(record)
}

// GetStatus handles GET /compliance/status
func (h *ComplianceHandler) GetStatus(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	record, err := h.complianceService.GetComplianceStatus(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get compliance status",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get compliance status",
		})
	}

	return c.JSON(record)
}

// GetAuditTrail handles GET /compliance/audit-trail?start=&end=
func (h *ComplianceHandler) GetAuditTrail(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if s := c.Query("start"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid start time format",
			})
		}
		startTime = t
	}

	if s := c.Query("end"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid end time format",
			})
		}
		endTime = t
	}

	entries, err := h.complianceService.GetAuditTrail(c.Context(), projectID, startTime, endTime)
	if err != nil {
		h.logger.Error("failed to get audit trail",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get audit trail",
		})
	}

	return c.JSON(entries)
}

// CreateAssessment handles POST /compliance/assessments
func (h *ComplianceHandler) CreateAssessment(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ConformityAssessment
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}

	assessment, err := h.complianceService.CreateConformityAssessment(c.Context(), projectID, &input)
	if err != nil {
		h.logger.Error("failed to create assessment",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create assessment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(assessment)
}

// GetAssessment handles GET /compliance/assessments/:id
func (h *ComplianceHandler) GetAssessment(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	assessmentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid assessment ID",
		})
	}

	assessment, err := h.complianceService.GetConformityAssessment(c.Context(), assessmentID)
	if err != nil {
		h.logger.Error("failed to get assessment",
			zap.String("assessmentId", assessmentID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to get assessment",
		})
	}

	return c.JSON(assessment)
}

// GenerateReport handles POST /compliance/reports
func (h *ComplianceHandler) GenerateReport(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Unauthorized",
			"message": "Project ID not found",
		})
	}

	var input domain.ComplianceReportInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid request body",
		})
	}
	input.ProjectID = projectID

	data, err := h.complianceService.GenerateComplianceReport(c.Context(), input)
	if err != nil {
		h.logger.Error("failed to generate compliance report",
			zap.String("projectId", projectID.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to generate compliance report",
		})
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to parse compliance report",
		})
	}

	return c.JSON(result)
}
