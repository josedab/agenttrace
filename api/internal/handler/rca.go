package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// RCAHandler handles root cause analysis HTTP requests
type RCAHandler struct {
	service *service.RCAService
	logger  *zap.Logger
}

// NewRCAHandler creates a new RCA handler
func NewRCAHandler(svc *service.RCAService, logger *zap.Logger) *RCAHandler {
	return &RCAHandler{
		service: svc,
		logger:  logger,
	}
}

// Analyze handles POST /api/public/rca/analyze
func (h *RCAHandler) Analyze(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.RCAInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.TraceID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "traceId is required"})
	}

	report, err := h.service.AnalyzeTrace(c.Context(), projectID, input.TraceID)
	if err != nil {
		h.logger.Error("failed to analyze trace", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to analyze trace"})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetReport handles GET /api/public/rca/reports/:reportId
func (h *RCAHandler) GetReport(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

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

// ListReports handles GET /api/public/rca/reports
func (h *RCAHandler) ListReports(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	reports, err := h.service.ListReports(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list RCA reports", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list reports"})
	}

	return c.JSON(fiber.Map{"reports": reports, "count": len(reports)})
}

// DetectAnomalies handles POST /api/public/rca/anomalies
func (h *RCAHandler) DetectAnomalies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CorrelatedAnomalyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.AnomalyType == "" || input.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "anomalyType and title are required"})
	}

	anomaly, err := h.service.DetectAnomalies(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to detect anomalies", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to detect anomalies"})
	}

	return c.Status(fiber.StatusCreated).JSON(anomaly)
}

// GetAnomaly handles GET /api/public/rca/anomalies/:anomalyId
func (h *RCAHandler) GetAnomaly(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	anomalyID, err := uuid.Parse(c.Params("anomalyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid anomaly ID"})
	}

	anomaly, err := h.service.GetAnomaly(c.Context(), anomalyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Anomaly not found"})
	}

	return c.JSON(anomaly)
}

// ListAnomalies handles GET /api/public/rca/anomalies
func (h *RCAHandler) ListAnomalies(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	anomalies, err := h.service.ListAnomalies(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list anomalies", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list anomalies"})
	}

	return c.JSON(fiber.Map{"anomalies": anomalies, "count": len(anomalies)})
}

// AcknowledgeAnomaly handles POST /api/public/rca/anomalies/:anomalyId/acknowledge
func (h *RCAHandler) AcknowledgeAnomaly(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	anomalyID, err := uuid.Parse(c.Params("anomalyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid anomaly ID"})
	}

	anomaly, err := h.service.AcknowledgeAnomaly(c.Context(), anomalyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Anomaly not found"})
	}

	return c.JSON(anomaly)
}

// CreateAlertChannel handles POST /api/public/rca/alert-channels
func (h *RCAHandler) CreateAlertChannel(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.DeliveryChannelInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" || input.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and type are required"})
	}

	channel, err := h.service.CreateAlertChannel(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create alert channel", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create alert channel"})
	}

	return c.Status(fiber.StatusCreated).JSON(channel)
}

// ListAlertChannels handles GET /api/public/rca/alert-channels
func (h *RCAHandler) ListAlertChannels(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	channels, err := h.service.ListAlertChannels(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list alert channels", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list alert channels"})
	}

	return c.JSON(fiber.Map{"channels": channels, "count": len(channels)})
}

// TestAlertChannel handles POST /api/public/rca/alert-channels/:channelId/test
func (h *RCAHandler) TestAlertChannel(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	channelID, err := uuid.Parse(c.Params("channelId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	channel, err := h.service.TestAlertChannel(c.Context(), channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Alert channel not found"})
	}

	return c.JSON(fiber.Map{"channel": channel, "testResult": "success", "message": "Test alert delivered successfully"})
}

// CreateCorrelationRule handles POST /api/public/rca/correlation-rules
func (h *RCAHandler) CreateCorrelationRule(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input domain.CorrelationRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Name == "" || len(input.AnomalyTypes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and anomalyTypes are required"})
	}

	rule, err := h.service.CreateCorrelationRule(c.Context(), projectID, input)
	if err != nil {
		h.logger.Error("failed to create correlation rule", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create correlation rule"})
	}

	return c.Status(fiber.StatusCreated).JSON(rule)
}

// ListCorrelationRules handles GET /api/public/rca/correlation-rules
func (h *RCAHandler) ListCorrelationRules(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	rules, err := h.service.ListCorrelationRules(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list correlation rules", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list correlation rules"})
	}

	return c.JSON(fiber.Map{"rules": rules, "count": len(rules)})
}

// GetAlertDashboardStats handles GET /api/public/rca/dashboard
func (h *RCAHandler) GetAlertDashboardStats(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	stats, err := h.service.GetAlertDashboardStats(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get alert dashboard stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get dashboard stats"})
	}

	return c.JSON(stats)
}

// CreateInvestigation handles POST /api/public/rca/investigations
func (h *RCAHandler) CreateInvestigation(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	var input struct {
		AnomalyID      uuid.UUID `json:"anomalyId"`
		Title          string    `json:"title"`
		InvestigatorID uuid.UUID `json:"investigatorId"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.AnomalyID == uuid.Nil || input.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "anomalyId and title are required"})
	}

	if input.InvestigatorID == uuid.Nil {
		input.InvestigatorID = uuid.New()
	}

	investigation, err := h.service.CreateInvestigation(c.Context(), projectID, input.AnomalyID, input.Title, input.InvestigatorID)
	if err != nil {
		h.logger.Error("failed to create investigation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"}) // Error logged via middleware
	}

	return c.Status(fiber.StatusCreated).JSON(investigation)
}

// UpdateInvestigation handles PUT /api/public/rca/investigations/:investigationId
func (h *RCAHandler) UpdateInvestigation(c *fiber.Ctx) error {
	_, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	investigationID, err := uuid.Parse(c.Params("investigationId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid investigation ID"})
	}

	var input struct {
		Status     string `json:"status"`
		RootCause  string `json:"rootCause"`
		Resolution string `json:"resolution"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	investigation, err := h.service.UpdateInvestigation(c.Context(), investigationID, input.Status, input.RootCause, input.Resolution)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Investigation not found"})
	}

	return c.JSON(investigation)
}

// ListInvestigations handles GET /api/public/rca/investigations
func (h *RCAHandler) ListInvestigations(c *fiber.Ctx) error {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Project ID not found"})
	}

	investigations, err := h.service.ListInvestigations(c.Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list investigations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list investigations"})
	}

	return c.JSON(fiber.Map{"investigations": investigations, "count": len(investigations)})
}
