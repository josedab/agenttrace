package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// AnomalyHandler handles anomaly detection HTTP requests
type AnomalyHandler struct {
	logger         *zap.Logger
	anomalyService *service.AnomalyService
}

// NewAnomalyHandler creates a new anomaly handler
func NewAnomalyHandler(
	logger *zap.Logger,
	anomalyService *service.AnomalyService,
) *AnomalyHandler {
	return &AnomalyHandler{
		logger:         logger,
		anomalyService: anomalyService,
	}
}

// ListRules returns all anomaly rules for a project
// @Summary List anomaly rules
// @Description Get all anomaly detection rules for a project
// @Tags anomaly
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Success 200 {array} domain.AnomalyRule
// @Failure 400 {object} ErrorResponse
// @Router /api/public/anomaly/rules [get]
func (h *AnomalyHandler) ListRules(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	h.logger.Debug("List anomaly rules", zap.String("projectId", projectID.String()))

	// Return empty list for now
	rules := []domain.AnomalyRule{}

	return c.JSON(rules)
}

// GetRule returns a specific anomaly rule
// @Summary Get anomaly rule
// @Description Get a specific anomaly detection rule by ID
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} domain.AnomalyRule
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/rules/{id} [get]
func (h *AnomalyHandler) GetRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	h.logger.Debug("Get anomaly rule", zap.String("ruleId", ruleID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Rule not found", nil)
}

// CreateRule creates a new anomaly detection rule
// @Summary Create anomaly rule
// @Description Create a new anomaly detection rule
// @Tags anomaly
// @Accept json
// @Produce json
// @Param rule body domain.AnomalyRuleInput true "Rule configuration"
// @Success 201 {object} domain.AnomalyRule
// @Failure 400 {object} ErrorResponse
// @Router /api/public/anomaly/rules [post]
func (h *AnomalyHandler) CreateRule(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	userID := uuid.New() // In real implementation, get from auth context

	var input domain.AnomalyRuleInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate
	if input.Name == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Name is required", nil)
	}
	if input.Type == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Type is required", nil)
	}
	if input.Method == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Method is required", nil)
	}

	// Get default config and merge with provided
	defaultConfig := h.anomalyService.DefaultRuleConfig(input.Method)
	config := mergeRuleConfig(defaultConfig, input.Config)

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	cooldown := 60 // Default 60 minutes
	if input.Cooldown != nil {
		cooldown = *input.Cooldown
	}

	rule := &domain.AnomalyRule{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		Enabled:         enabled,
		Type:            input.Type,
		Method:          input.Method,
		Config:          config,
		TraceNameFilter: input.TraceNameFilter,
		MetadataFilters: input.MetadataFilters,
		AlertChannels:   input.AlertChannels,
		Severity:        input.Severity,
		Cooldown:        cooldown,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CreatedBy:       userID,
	}

	h.logger.Info("Created anomaly rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("name", rule.Name),
		zap.String("type", string(rule.Type)),
		zap.String("method", string(rule.Method)),
	)

	return c.Status(fiber.StatusCreated).JSON(rule)
}

// UpdateRule updates an anomaly detection rule
// @Summary Update anomaly rule
// @Description Update an anomaly detection rule
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param rule body domain.AnomalyRuleInput true "Updated rule configuration"
// @Success 200 {object} domain.AnomalyRule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/rules/{id} [patch]
func (h *AnomalyHandler) UpdateRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	var input domain.AnomalyRuleInput
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Debug("Update anomaly rule", zap.String("ruleId", ruleID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Rule not found", nil)
}

// DeleteRule deletes an anomaly detection rule
// @Summary Delete anomaly rule
// @Description Delete an anomaly detection rule
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/rules/{id} [delete]
func (h *AnomalyHandler) DeleteRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	h.logger.Info("Delete anomaly rule", zap.String("ruleId", ruleID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Rule not found", nil)
}

// ToggleRule enables or disables an anomaly detection rule
// @Summary Toggle anomaly rule
// @Description Enable or disable an anomaly detection rule
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param body body ToggleRuleRequest true "Enable/disable"
// @Success 200 {object} domain.AnomalyRule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/rules/{id}/toggle [post]
func (h *AnomalyHandler) ToggleRule(c *fiber.Ctx) error {
	ruleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	var req ToggleRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	h.logger.Info("Toggle anomaly rule",
		zap.String("ruleId", ruleID.String()),
		zap.Bool("enabled", req.Enabled),
	)

	return errorResponse(c, fiber.StatusNotFound, "Rule not found", nil)
}

// ToggleRuleRequest represents the request to toggle a rule
type ToggleRuleRequest struct {
	Enabled bool `json:"enabled"`
}

// ListAnomalies returns anomalies for a project
// @Summary List anomalies
// @Description Get all detected anomalies for a project
// @Tags anomaly
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param ruleId query string false "Filter by rule ID"
// @Param type query string false "Filter by anomaly type"
// @Param severity query string false "Filter by severity"
// @Param startTime query string false "Start time (RFC3339)"
// @Param endTime query string false "End time (RFC3339)"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.AnomalyList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/anomaly/anomalies [get]
func (h *AnomalyHandler) ListAnomalies(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	filter := domain.AnomalyFilter{
		ProjectID: projectID,
	}

	if ruleIDStr := c.Query("ruleId"); ruleIDStr != "" {
		ruleID, err := uuid.Parse(ruleIDStr)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
		}
		filter.RuleID = &ruleID
	}

	if typeStr := c.Query("type"); typeStr != "" {
		t := domain.AnomalyType(typeStr)
		filter.Type = &t
	}

	if severityStr := c.Query("severity"); severityStr != "" {
		s := domain.AnomalySeverity(severityStr)
		filter.Severity = &s
	}

	if startTimeStr := c.Query("startTime"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid start time format", err)
		}
		filter.StartTime = &startTime
	}

	if endTimeStr := c.Query("endTime"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid end time format", err)
		}
		filter.EndTime = &endTime
	}

	h.logger.Debug("List anomalies", zap.String("projectId", projectID.String()))

	result := domain.AnomalyList{
		Anomalies:  []domain.Anomaly{},
		TotalCount: 0,
		HasMore:    false,
	}

	return c.JSON(result)
}

// GetAnomaly returns a specific anomaly
// @Summary Get anomaly
// @Description Get a specific anomaly by ID
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Anomaly ID"
// @Success 200 {object} domain.Anomaly
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/anomalies/{id} [get]
func (h *AnomalyHandler) GetAnomaly(c *fiber.Ctx) error {
	anomalyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid anomaly ID", err)
	}

	h.logger.Debug("Get anomaly", zap.String("anomalyId", anomalyID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Anomaly not found", nil)
}

// ListAlerts returns alerts for a project
// @Summary List alerts
// @Description Get all alerts for a project
// @Tags anomaly
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param status query string false "Filter by status"
// @Param severity query string false "Filter by severity"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} domain.AlertList
// @Failure 400 {object} ErrorResponse
// @Router /api/public/anomaly/alerts [get]
func (h *AnomalyHandler) ListAlerts(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	filter := domain.AlertFilter{
		ProjectID: projectID,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		s := domain.AlertStatus(statusStr)
		filter.Status = &s
	}

	if severityStr := c.Query("severity"); severityStr != "" {
		s := domain.AnomalySeverity(severityStr)
		filter.Severity = &s
	}

	h.logger.Debug("List alerts", zap.String("projectId", projectID.String()))

	result := domain.AlertList{
		Alerts:     []domain.Alert{},
		TotalCount: 0,
		HasMore:    false,
	}

	return c.JSON(result)
}

// GetAlert returns a specific alert
// @Summary Get alert
// @Description Get a specific alert by ID
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} domain.Alert
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/alerts/{id} [get]
func (h *AnomalyHandler) GetAlert(c *fiber.Ctx) error {
	alertID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid alert ID", err)
	}

	h.logger.Debug("Get alert", zap.String("alertId", alertID.String()))

	return errorResponse(c, fiber.StatusNotFound, "Alert not found", nil)
}

// AcknowledgeAlert acknowledges an alert
// @Summary Acknowledge alert
// @Description Acknowledge an alert
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} domain.Alert
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/alerts/{id}/acknowledge [post]
func (h *AnomalyHandler) AcknowledgeAlert(c *fiber.Ctx) error {
	alertID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid alert ID", err)
	}

	userID := uuid.New() // In real implementation, get from auth context

	h.logger.Info("Acknowledge alert",
		zap.String("alertId", alertID.String()),
		zap.String("userId", userID.String()),
	)

	return errorResponse(c, fiber.StatusNotFound, "Alert not found", nil)
}

// ResolveAlert resolves an alert
// @Summary Resolve alert
// @Description Resolve an alert
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param body body ResolveAlertRequest true "Resolution details"
// @Success 200 {object} domain.Alert
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/alerts/{id}/resolve [post]
func (h *AnomalyHandler) ResolveAlert(c *fiber.Ctx) error {
	alertID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid alert ID", err)
	}

	var req ResolveAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	userID := uuid.New() // In real implementation, get from auth context

	h.logger.Info("Resolve alert",
		zap.String("alertId", alertID.String()),
		zap.String("userId", userID.String()),
	)

	return errorResponse(c, fiber.StatusNotFound, "Alert not found", nil)
}

// ResolveAlertRequest represents the request to resolve an alert
type ResolveAlertRequest struct {
	Note string `json:"note,omitempty"`
}

// AddAlertNote adds a note to an alert
// @Summary Add alert note
// @Description Add a note to an alert
// @Tags anomaly
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param body body AddAlertNoteRequest true "Note content"
// @Success 201 {object} domain.AlertNote
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/public/anomaly/alerts/{id}/notes [post]
func (h *AnomalyHandler) AddAlertNote(c *fiber.Ctx) error {
	alertID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid alert ID", err)
	}

	var req AddAlertNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Content == "" {
		return errorResponse(c, fiber.StatusBadRequest, "Note content is required", nil)
	}

	userID := uuid.New() // In real implementation, get from auth context

	note := &domain.AlertNote{
		ID:        uuid.New(),
		AlertID:   alertID,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	h.logger.Info("Add alert note",
		zap.String("alertId", alertID.String()),
		zap.String("noteId", note.ID.String()),
	)

	return c.Status(fiber.StatusCreated).JSON(note)
}

// AddAlertNoteRequest represents the request to add an alert note
type AddAlertNoteRequest struct {
	Content string `json:"content"`
}

// GetStats returns anomaly statistics for a project
// @Summary Get anomaly stats
// @Description Get anomaly statistics for a project
// @Tags anomaly
// @Accept json
// @Produce json
// @Param projectId query string true "Project ID"
// @Param startTime query string false "Start time (RFC3339)"
// @Param endTime query string false "End time (RFC3339)"
// @Success 200 {object} domain.AnomalyStats
// @Failure 400 {object} ErrorResponse
// @Router /api/public/anomaly/stats [get]
func (h *AnomalyHandler) GetStats(c *fiber.Ctx) error {
	projectID, err := getProjectIDFromContext(c)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid project ID", err)
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if startTimeStr := c.Query("startTime"); startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid start time format", err)
		}
		startTime = t
	}

	if endTimeStr := c.Query("endTime"); endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid end time format", err)
		}
		endTime = t
	}

	period := domain.TimeWindow{
		Start: startTime,
		End:   endTime,
	}

	// In real implementation, fetch from database
	stats := h.anomalyService.GetAnomalyStats(c.Context(), projectID, []domain.Anomaly{}, 0, period)

	return c.JSON(stats)
}

// TestRule tests an anomaly rule against historical data
// @Summary Test anomaly rule
// @Description Test an anomaly rule against historical data
// @Tags anomaly
// @Accept json
// @Produce json
// @Param body body TestRuleRequest true "Test configuration"
// @Success 200 {object} TestRuleResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/public/anomaly/rules/test [post]
func (h *AnomalyHandler) TestRule(c *fiber.Ctx) error {
	var req TestRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if len(req.HistoricalData) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "Historical data is required", nil)
	}

	// Create a temporary rule for testing
	rule := &domain.AnomalyRule{
		ID:      uuid.New(),
		Type:    req.Type,
		Method:  req.Method,
		Config:  req.Config,
		Enabled: true,
	}

	// Run detection on the test value
	result, err := h.anomalyService.DetectAnomaly(
		c.Context(),
		rule,
		req.TestValue,
		req.HistoricalData,
	)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	// Calculate baseline stats
	stats := h.anomalyService.CalculateBaselineStats(req.HistoricalData)

	response := TestRuleResponse{
		IsAnomaly:     result.IsAnomaly,
		Score:         result.Score,
		Threshold:     result.Threshold,
		Expected:      result.Expected,
		TestValue:     req.TestValue,
		Description:   result.Description,
		Severity:      result.Severity,
		BaselineStats: stats,
	}

	return c.JSON(response)
}

// TestRuleRequest represents the request to test an anomaly rule
type TestRuleRequest struct {
	Type           domain.AnomalyType       `json:"type"`
	Method         domain.DetectionMethod   `json:"method"`
	Config         domain.AnomalyRuleConfig `json:"config"`
	TestValue      float64                  `json:"testValue"`
	HistoricalData []float64                `json:"historicalData"`
}

// TestRuleResponse represents the response from testing an anomaly rule
type TestRuleResponse struct {
	IsAnomaly     bool                  `json:"isAnomaly"`
	Score         float64               `json:"score"`
	Threshold     float64               `json:"threshold"`
	Expected      float64               `json:"expected"`
	TestValue     float64               `json:"testValue"`
	Description   string                `json:"description"`
	Severity      domain.AnomalySeverity `json:"severity,omitempty"`
	BaselineStats domain.BaselineStats  `json:"baselineStats"`
}

// mergeRuleConfig merges default config with provided config
func mergeRuleConfig(defaultConfig, provided domain.AnomalyRuleConfig) domain.AnomalyRuleConfig {
	result := defaultConfig

	if provided.ZScoreThreshold != 0 {
		result.ZScoreThreshold = provided.ZScoreThreshold
	}
	if provided.IQRMultiplier != 0 {
		result.IQRMultiplier = provided.IQRMultiplier
	}
	if provided.MADThreshold != 0 {
		result.MADThreshold = provided.MADThreshold
	}
	if provided.WindowSize != 0 {
		result.WindowSize = provided.WindowSize
	}
	if provided.Deviation != 0 {
		result.Deviation = provided.Deviation
	}
	if provided.Alpha != 0 {
		result.Alpha = provided.Alpha
	}
	if provided.MinThreshold != nil {
		result.MinThreshold = provided.MinThreshold
	}
	if provided.MaxThreshold != nil {
		result.MaxThreshold = provided.MaxThreshold
	}
	if provided.MinSamples != 0 {
		result.MinSamples = provided.MinSamples
	}
	if provided.LookbackHours != 0 {
		result.LookbackHours = provided.LookbackHours
	}

	return result
}
