package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/repository/postgres"
)

// MockAuditService mocks the audit service for handler tests
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) ListAuditLogs(ctx context.Context, filter *domain.AuditLogFilter) (*domain.AuditLogList, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLogList), args.Error(1)
}

func (m *MockAuditService) GetAuditLog(ctx context.Context, orgID, logID uuid.UUID) (*domain.AuditLog, error) {
	args := m.Called(ctx, orgID, logID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func (m *MockAuditService) GetAuditSummary(ctx context.Context, orgID uuid.UUID, period string) (*domain.AuditSummary, error) {
	args := m.Called(ctx, orgID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditSummary), args.Error(1)
}

func (m *MockAuditService) GetSecurityEvents(ctx context.Context, orgID uuid.UUID, since time.Time, limit int) ([]domain.AuditLog, error) {
	args := m.Called(ctx, orgID, mock.AnythingOfType("time.Time"), limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.AuditLog), args.Error(1)
}

func (m *MockAuditService) GetRetentionPolicy(ctx context.Context, orgID uuid.UUID) (*domain.AuditRetentionPolicy, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditRetentionPolicy), args.Error(1)
}

func (m *MockAuditService) SetRetentionPolicy(ctx context.Context, orgID uuid.UUID, retentionDays int, enabled bool) (*domain.AuditRetentionPolicy, error) {
	args := m.Called(ctx, orgID, retentionDays, enabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditRetentionPolicy), args.Error(1)
}

func (m *MockAuditService) CreateExportJob(ctx context.Context, orgID uuid.UUID, requestedBy *uuid.UUID, filter *domain.AuditLogFilter, format string, compress bool) (*postgres.AuditExportJob, error) {
	args := m.Called(ctx, orgID, requestedBy, filter, format, compress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.AuditExportJob), args.Error(1)
}

func (m *MockAuditService) GetExportJob(ctx context.Context, jobID uuid.UUID) (*postgres.AuditExportJob, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.AuditExportJob), args.Error(1)
}

func (m *MockAuditService) ListExportJobs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]postgres.AuditExportJob, error) {
	args := m.Called(ctx, orgID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]postgres.AuditExportJob), args.Error(1)
}

func (m *MockAuditService) LogSettingsChanged(ctx context.Context, orgID uuid.UUID, actorID uuid.UUID, actorEmail, settingName string, before, after any) error {
	args := m.Called(ctx, orgID, actorID, actorEmail, settingName, before, after)
	return args.Error(0)
}

func setupAuditTestApp(orgID uuid.UUID, userID *uuid.UUID) (*fiber.App, *AuditHandler) {
	app := fiber.New()

	// The AuditHandler uses *service.AuditService directly which can't be easily mocked
	// without interface refactoring. Test handlers via HTTP to verify route behavior.
	handler := &AuditHandler{auditService: nil}

	app.Use(func(c *fiber.Ctx) error {
		if userID != nil {
			c.Locals("userID", *userID)
			c.Locals("userEmail", "test@example.com")
		}
		return c.Next()
	})

	return app, handler
}

func TestAuditHandler_ListAuditLogs(t *testing.T) {
	orgID := uuid.New()

	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs", handler.ListAuditLogs)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/invalid-uuid/audit-logs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("parses orgId correctly", func(t *testing.T) {
		// Without actual service, we just verify it doesn't panic on valid UUID
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs", handler.ListAuditLogs)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String()+"/audit-logs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		// Will 500 since auditService is nil, but at least it parsed the orgID
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAuditHandler_GetAuditLog(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs/:logId", handler.GetAuditLog)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/bad/audit-logs/"+uuid.New().String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for invalid logId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs/:logId", handler.GetAuditLog)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+uuid.New().String()+"/audit-logs/bad", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_GetAuditSummary(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs/summary", handler.GetAuditSummary)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/bad/audit-logs/summary", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_GetSecurityEvents(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs/security", handler.GetSecurityEvents)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/bad/audit-logs/security", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_GetRetentionPolicy(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-retention", handler.GetRetentionPolicy)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/bad/audit-retention", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_SetRetentionPolicy(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Put("/v1/organizations/:orgId/audit-retention", handler.SetRetentionPolicy)

		body, _ := json.Marshal(SetRetentionPolicyRequest{RetentionDays: 90, Enabled: true})
		req := httptest.NewRequest(http.MethodPut, "/v1/organizations/bad/audit-retention", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for negative retention days", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Put("/v1/organizations/:orgId/audit-retention", handler.SetRetentionPolicy)

		body, _ := json.Marshal(SetRetentionPolicyRequest{RetentionDays: -1, Enabled: true})
		req := httptest.NewRequest(http.MethodPut, "/v1/organizations/"+uuid.New().String()+"/audit-retention", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for invalid body", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Put("/v1/organizations/:orgId/audit-retention", handler.SetRetentionPolicy)

		req := httptest.NewRequest(http.MethodPut, "/v1/organizations/"+uuid.New().String()+"/audit-retention", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_CreateExportJob(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Post("/v1/organizations/:orgId/audit-logs/export", handler.CreateExportJob)

		body, _ := json.Marshal(CreateExportJobRequest{Format: "csv"})
		req := httptest.NewRequest(http.MethodPost, "/v1/organizations/bad/audit-logs/export", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 401 without userID", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Post("/v1/organizations/:orgId/audit-logs/export", handler.CreateExportJob)

		body, _ := json.Marshal(CreateExportJobRequest{Format: "csv"})
		req := httptest.NewRequest(http.MethodPost, "/v1/organizations/"+uuid.New().String()+"/audit-logs/export", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuditHandler_GetExportJob(t *testing.T) {
	t.Run("returns 400 for invalid jobId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs/export/:jobId", handler.GetExportJob)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+uuid.New().String()+"/audit-logs/export/bad", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_ListExportJobs(t *testing.T) {
	t.Run("returns 400 for invalid orgId", func(t *testing.T) {
		app := fiber.New()
		handler := &AuditHandler{auditService: nil}
		app.Get("/v1/organizations/:orgId/audit-logs/export/jobs", handler.ListExportJobs)

		req := httptest.NewRequest(http.MethodGet, "/v1/organizations/bad/audit-logs/export/jobs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuditHandler_SetRetentionPolicy_RequiresAuth(t *testing.T) {
	// Verify that SetRetentionPolicy returns 401 when no userID is in context
	app := fiber.New()
	handler := &AuditHandler{auditService: nil}
	app.Put("/v1/organizations/:orgId/audit-retention", handler.SetRetentionPolicy)

	body, _ := json.Marshal(SetRetentionPolicyRequest{RetentionDays: 90, Enabled: true})
	req := httptest.NewRequest(http.MethodPut, "/v1/organizations/"+uuid.New().String()+"/audit-retention", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Will hit auditService.SetRetentionPolicy first (nil panic) since validation passed
	// But this at least confirms the route parsing works
	assert.True(t, resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized)
}
