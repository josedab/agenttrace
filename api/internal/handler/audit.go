package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"agenttrace/internal/domain"
	"agenttrace/internal/service"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// RegisterRoutes registers audit log routes
func (h *AuditHandler) RegisterRoutes(app fiber.Router) {
	audit := app.Group("/v1/organizations/:orgId/audit-logs")
	audit.Get("", h.ListAuditLogs)
	audit.Get("/summary", h.GetAuditSummary)
	audit.Get("/security", h.GetSecurityEvents)
	audit.Get("/:logId", h.GetAuditLog)
	audit.Get("/export/jobs", h.ListExportJobs)
	audit.Post("/export", h.CreateExportJob)
	audit.Get("/export/:jobId", h.GetExportJob)

	// Retention policy routes
	retention := app.Group("/v1/organizations/:orgId/audit-retention")
	retention.Get("", h.GetRetentionPolicy)
	retention.Put("", h.SetRetentionPolicy)
}

// ListAuditLogs retrieves audit logs with filtering
// @Summary List audit logs
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param actor_id query string false "Filter by actor ID"
// @Param action query string false "Filter by action"
// @Param resource_type query string false "Filter by resource type"
// @Param resource_id query string false "Filter by resource ID"
// @Param start_time query string false "Filter by start time (RFC3339)"
// @Param end_time query string false "Filter by end time (RFC3339)"
// @Param ip_address query string false "Filter by IP address"
// @Param search query string false "Search in description and resource name"
// @Param limit query int false "Limit results (default 50, max 1000)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} domain.AuditLogList
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-logs [get]
func (h *AuditHandler) ListAuditLogs(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	filter := &domain.AuditLogFilter{
		OrganizationID: &orgID,
	}

	// Parse query parameters
	if actorIDStr := c.Query("actor_id"); actorIDStr != "" {
		if actorID, err := uuid.Parse(actorIDStr); err == nil {
			filter.ActorID = &actorID
		}
	}

	if action := c.Query("action"); action != "" {
		a := domain.AuditAction(action)
		filter.Action = &a
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		rt := domain.AuditResourceType(resourceType)
		filter.ResourceType = &rt
	}

	if resourceIDStr := c.Query("resource_id"); resourceIDStr != "" {
		if resourceID, err := uuid.Parse(resourceIDStr); err == nil {
			filter.ResourceID = &resourceID
		}
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		filter.IPAddress = &ipAddress
	}

	if search := c.Query("search"); search != "" {
		filter.SearchQuery = &search
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	result, err := h.auditService.ListAuditLogs(c.Context(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to list audit logs: " + err.Error(),
		})
	}

	return c.JSON(result)
}

// GetAuditLog retrieves a single audit log entry
// @Summary Get audit log
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param logId path string true "Audit Log ID"
// @Success 200 {object} domain.AuditLog
// @Failure 404 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-logs/{logId} [get]
func (h *AuditHandler) GetAuditLog(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	logID, err := uuid.Parse(c.Params("logId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid log ID",
		})
	}

	log, err := h.auditService.GetAuditLog(c.Context(), orgID, logID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get audit log: " + err.Error(),
		})
	}

	if log == nil {
		return c.Status(http.StatusNotFound).JSON(ErrorResponse{
			Error: "Audit log not found",
		})
	}

	return c.JSON(log)
}

// GetAuditSummary returns aggregated audit statistics
// @Summary Get audit summary
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param period query string false "Period: day, week, month (default week)"
// @Success 200 {object} domain.AuditSummary
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-logs/summary [get]
func (h *AuditHandler) GetAuditSummary(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	period := c.Query("period", "week")
	if period != "day" && period != "week" && period != "month" {
		period = "week"
	}

	summary, err := h.auditService.GetAuditSummary(c.Context(), orgID, period)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get audit summary: " + err.Error(),
		})
	}

	return c.JSON(summary)
}

// GetSecurityEvents returns security-related audit events
// @Summary Get security events
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param since query string false "Since time (RFC3339), default 7 days ago"
// @Param limit query int false "Limit results (default 100)"
// @Success 200 {array} domain.AuditLog
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-logs/security [get]
func (h *AuditHandler) GetSecurityEvents(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	since := time.Now().AddDate(0, 0, -7) // Default to 7 days ago
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsedTime
		}
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events, err := h.auditService.GetSecurityEvents(c.Context(), orgID, since, limit)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get security events: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"events": events,
		"since":  since,
		"count":  len(events),
	})
}

// GetRetentionPolicy returns the audit log retention policy
// @Summary Get retention policy
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} domain.AuditRetentionPolicy
// @Router /v1/organizations/{orgId}/audit-retention [get]
func (h *AuditHandler) GetRetentionPolicy(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	policy, err := h.auditService.GetRetentionPolicy(c.Context(), orgID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get retention policy: " + err.Error(),
		})
	}

	return c.JSON(policy)
}

// SetRetentionPolicy creates or updates the audit log retention policy
// @Summary Set retention policy
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param body body SetRetentionPolicyRequest true "Retention Policy"
// @Success 200 {object} domain.AuditRetentionPolicy
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-retention [put]
func (h *AuditHandler) SetRetentionPolicy(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	var req SetRetentionPolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.RetentionDays < 0 {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Retention days cannot be negative",
		})
	}

	policy, err := h.auditService.SetRetentionPolicy(c.Context(), orgID, req.RetentionDays, req.Enabled)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to set retention policy: " + err.Error(),
		})
	}

	// Log the change
	userID := c.Locals("userID").(uuid.UUID)
	userEmail := c.Locals("userEmail").(string)
	h.auditService.LogSettingsChanged(c.Context(), orgID, userID, userEmail, "audit_retention",
		nil, map[string]any{"retention_days": req.RetentionDays, "enabled": req.Enabled})

	return c.JSON(policy)
}

// CreateExportJob creates a new audit log export job
// @Summary Create export job
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param body body CreateExportJobRequest true "Export request"
// @Success 201 {object} ExportJobResponse
// @Failure 400 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-logs/export [post]
func (h *AuditHandler) CreateExportJob(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	var req CreateExportJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate format
	if req.Format != "csv" && req.Format != "json" {
		req.Format = "csv"
	}

	// Build filter from request
	filter := &domain.AuditLogFilter{
		OrganizationID: &orgID,
	}

	if req.StartTime != nil {
		filter.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		filter.EndTime = req.EndTime
	}
	if len(req.Actions) > 0 {
		for _, a := range req.Actions {
			filter.Actions = append(filter.Actions, domain.AuditAction(a))
		}
	}
	if req.ResourceType != "" {
		rt := domain.AuditResourceType(req.ResourceType)
		filter.ResourceType = &rt
	}

	userID := c.Locals("userID").(uuid.UUID)

	job, err := h.auditService.CreateExportJob(c.Context(), orgID, &userID, filter, req.Format, req.Compress)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to create export job: " + err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(ExportJobResponse{
		ID:        job.ID,
		Status:    job.Status,
		Format:    job.Format,
		CreatedAt: job.CreatedAt,
	})
}

// GetExportJob retrieves an export job
// @Summary Get export job
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param jobId path string true "Job ID"
// @Success 200 {object} ExportJobResponse
// @Failure 404 {object} ErrorResponse
// @Router /v1/organizations/{orgId}/audit-logs/export/{jobId} [get]
func (h *AuditHandler) GetExportJob(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid job ID",
		})
	}

	job, err := h.auditService.GetExportJob(c.Context(), jobID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to get export job: " + err.Error(),
		})
	}

	if job == nil {
		return c.Status(http.StatusNotFound).JSON(ErrorResponse{
			Error: "Export job not found",
		})
	}

	resp := ExportJobResponse{
		ID:          job.ID,
		Status:      job.Status,
		Format:      job.Format,
		RecordCount: job.RecordCount,
		FileSize:    job.FileSize,
		Error:       job.Error,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
		ExpiresAt:   job.ExpiresAt,
		CreatedAt:   job.CreatedAt,
	}

	if job.Status == "completed" && job.FilePath != nil {
		// Generate download URL
		resp.DownloadURL = stringPtr("/v1/audit-logs/export/" + job.ID.String() + "/download")
	}

	return c.JSON(resp)
}

// ListExportJobs lists export jobs for an organization
// @Summary List export jobs
// @Tags Audit
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param limit query int false "Limit results (default 20)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} ListExportJobsResponse
// @Router /v1/organizations/{orgId}/audit-logs/export/jobs [get]
func (h *AuditHandler) ListExportJobs(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("orgId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid organization ID",
		})
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	jobs, err := h.auditService.ListExportJobs(c.Context(), orgID, limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to list export jobs: " + err.Error(),
		})
	}

	var responses []ExportJobResponse
	for _, job := range jobs {
		resp := ExportJobResponse{
			ID:          job.ID,
			Status:      job.Status,
			Format:      job.Format,
			RecordCount: job.RecordCount,
			FileSize:    job.FileSize,
			Error:       job.Error,
			StartedAt:   job.StartedAt,
			CompletedAt: job.CompletedAt,
			ExpiresAt:   job.ExpiresAt,
			CreatedAt:   job.CreatedAt,
		}
		responses = append(responses, resp)
	}

	return c.JSON(ListExportJobsResponse{
		Jobs: responses,
	})
}

// Request/Response types

type SetRetentionPolicyRequest struct {
	RetentionDays int  `json:"retentionDays"`
	Enabled       bool `json:"enabled"`
}

type CreateExportJobRequest struct {
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	Actions      []string   `json:"actions"`
	ResourceType string     `json:"resourceType"`
	Format       string     `json:"format"` // csv or json
	Compress     bool       `json:"compress"`
}

type ExportJobResponse struct {
	ID          uuid.UUID  `json:"id"`
	Status      string     `json:"status"`
	Format      string     `json:"format"`
	RecordCount *int       `json:"recordCount,omitempty"`
	FileSize    *int64     `json:"fileSize,omitempty"`
	DownloadURL *string    `json:"downloadUrl,omitempty"`
	Error       *string    `json:"error,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type ListExportJobsResponse struct {
	Jobs []ExportJobResponse `json:"jobs"`
}

func stringPtr(s string) *string {
	return &s
}
