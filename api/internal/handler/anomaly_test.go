package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

func setupAnomalyTestApp(projectID uuid.UUID, userID *uuid.UUID) (*fiber.App, *AnomalyHandler) {
	app := fiber.New()
	logger := zap.NewNop()
	anomalySvc := service.NewAnomalyService(logger)
	handler := NewAnomalyHandler(logger, anomalySvc)

	// Middleware to inject projectID and optionally userID into context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("apiKey", &domain.APIKey{ProjectID: projectID})
		c.Locals(string(middleware.ContextKeyProjectID), projectID)
		if userID != nil {
			c.Locals(string(middleware.ContextKeyUserID), *userID)
		}
		return c.Next()
	})

	return app, handler
}

func TestAnomalyHandler_ListRules(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Get("/anomaly/rules", handler.ListRules)

	req := httptest.NewRequest(http.MethodGet, "/anomaly/rules?projectId="+projectID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAnomalyHandler_GetRule(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Get("/anomaly/rules/:id", handler.GetRule)

	t.Run("valid UUID returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/anomaly/rules/"+uuid.New().String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/anomaly/rules/not-a-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAnomalyHandler_CreateRule(t *testing.T) {
	projectID := uuid.New()

	t.Run("creates rule with valid input and auth", func(t *testing.T) {
		userID := uuid.New()
		app, handler := setupAnomalyTestApp(projectID, &userID)
		app.Post("/anomaly/rules", handler.CreateRule)

		body, _ := json.Marshal(domain.AnomalyRuleInput{
			Name:   "test-rule",
			Type:   "latency",
			Method: "z_score",
		})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/rules?projectId="+projectID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var rule domain.AnomalyRule
		err = json.NewDecoder(resp.Body).Decode(&rule)
		require.NoError(t, err)
		assert.Equal(t, "test-rule", rule.Name)
		assert.Equal(t, userID, rule.CreatedBy)
	})

	t.Run("returns 401 without auth", func(t *testing.T) {
		app, handler := setupAnomalyTestApp(projectID, nil) // no userID
		app.Post("/anomaly/rules", handler.CreateRule)

		body, _ := json.Marshal(domain.AnomalyRuleInput{
			Name:   "test-rule",
			Type:   "latency",
			Method: "z_score",
		})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/rules?projectId="+projectID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		userID := uuid.New()
		app, handler := setupAnomalyTestApp(projectID, &userID)
		app.Post("/anomaly/rules", handler.CreateRule)

		body, _ := json.Marshal(domain.AnomalyRuleInput{
			Type:   "latency",
			Method: "z_score",
		})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/rules?projectId="+projectID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAnomalyHandler_DeleteRule(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Delete("/anomaly/rules/:id", handler.DeleteRule)

	t.Run("valid UUID returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/anomaly/rules/"+uuid.New().String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestAnomalyHandler_ListAnomalies(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Get("/anomaly/anomalies", handler.ListAnomalies)

	t.Run("returns empty list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/anomaly/anomalies?projectId="+projectID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("validates bad time format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/anomaly/anomalies?projectId="+projectID.String()+"&startTime=bad", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAnomalyHandler_ListAlerts(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Get("/anomaly/alerts", handler.ListAlerts)

	req := httptest.NewRequest(http.MethodGet, "/anomaly/alerts?projectId="+projectID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAnomalyHandler_AcknowledgeAlert(t *testing.T) {
	projectID := uuid.New()

	t.Run("returns 401 without auth", func(t *testing.T) {
		app, handler := setupAnomalyTestApp(projectID, nil)
		app.Post("/anomaly/alerts/:id/acknowledge", handler.AcknowledgeAlert)

		req := httptest.NewRequest(http.MethodPost, "/anomaly/alerts/"+uuid.New().String()+"/acknowledge", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 404 with valid auth", func(t *testing.T) {
		userID := uuid.New()
		app, handler := setupAnomalyTestApp(projectID, &userID)
		app.Post("/anomaly/alerts/:id/acknowledge", handler.AcknowledgeAlert)

		req := httptest.NewRequest(http.MethodPost, "/anomaly/alerts/"+uuid.New().String()+"/acknowledge", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestAnomalyHandler_ResolveAlert(t *testing.T) {
	projectID := uuid.New()

	t.Run("returns 401 without auth", func(t *testing.T) {
		app, handler := setupAnomalyTestApp(projectID, nil)
		app.Post("/anomaly/alerts/:id/resolve", handler.ResolveAlert)

		body, _ := json.Marshal(ResolveAlertRequest{Note: "resolved"})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/alerts/"+uuid.New().String()+"/resolve", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAnomalyHandler_AddAlertNote(t *testing.T) {
	projectID := uuid.New()

	t.Run("returns 401 without auth", func(t *testing.T) {
		app, handler := setupAnomalyTestApp(projectID, nil)
		app.Post("/anomaly/alerts/:id/notes", handler.AddAlertNote)

		body, _ := json.Marshal(AddAlertNoteRequest{Content: "test note"})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/alerts/"+uuid.New().String()+"/notes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("creates note with valid auth", func(t *testing.T) {
		userID := uuid.New()
		app, handler := setupAnomalyTestApp(projectID, &userID)
		app.Post("/anomaly/alerts/:id/notes", handler.AddAlertNote)

		body, _ := json.Marshal(AddAlertNoteRequest{Content: "test note"})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/alerts/"+uuid.New().String()+"/notes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var note domain.AlertNote
		err = json.NewDecoder(resp.Body).Decode(&note)
		require.NoError(t, err)
		assert.Equal(t, userID, note.UserID)
		assert.Equal(t, "test note", note.Content)
	})

	t.Run("returns 400 for empty content", func(t *testing.T) {
		userID := uuid.New()
		app, handler := setupAnomalyTestApp(projectID, &userID)
		app.Post("/anomaly/alerts/:id/notes", handler.AddAlertNote)

		body, _ := json.Marshal(AddAlertNoteRequest{Content: ""})
		req := httptest.NewRequest(http.MethodPost, "/anomaly/alerts/"+uuid.New().String()+"/notes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAnomalyHandler_GetStats(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Get("/anomaly/stats", handler.GetStats)

	t.Run("returns stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/anomaly/stats?projectId="+projectID.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("validates bad time format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/anomaly/stats?projectId="+projectID.String()+"&startTime=invalid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAnomalyHandler_ToggleRule(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	app, handler := setupAnomalyTestApp(projectID, &userID)
	app.Post("/anomaly/rules/:id/toggle", handler.ToggleRule)

	body, _ := json.Marshal(ToggleRuleRequest{Enabled: true})
	req := httptest.NewRequest(http.MethodPost, "/anomaly/rules/"+uuid.New().String()+"/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
