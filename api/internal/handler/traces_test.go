package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// MockQueryService mocks the query service
type MockQueryService struct {
	mock.Mock
}

func (m *MockQueryService) GetTrace(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockQueryService) ListTraces(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TraceList), args.Error(1)
}

func (m *MockQueryService) GetObservationTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ObservationTree), args.Error(1)
}

func (m *MockQueryService) GetTraceStats(ctx context.Context, filter *domain.TraceFilter) (*service.TraceStats, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TraceStats), args.Error(1)
}

func (m *MockQueryService) GetSessionTraces(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	args := m.Called(ctx, projectID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trace), args.Error(1)
}

func setupTracesTestApp(mockSvc *MockQueryService, projectID uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	app.Use(testProjectMiddleware(projectID))

	// ListTraces
	app.Get("/v1/traces", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		filter := &domain.TraceFilter{
			ProjectID: pid,
		}

		limit := 50
		if l := c.Query("limit"); l != "" {
			limit = parseIntParam(c, "limit", 50)
		}
		offset := parseIntParam(c, "offset", 0)

		list, err := mockSvc.ListTraces(c.Context(), filter, limit, offset)
		if err != nil {
			logger.Error("failed to list traces")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(list)
	})

	// GetTrace
	app.Get("/v1/traces/:traceId", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		traceID := c.Params("traceId")
		if traceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Trace ID required",
			})
		}

		trace, err := mockSvc.GetTrace(c.Context(), pid, traceID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}

		return c.JSON(trace)
	})

	// GetTraceObservations
	app.Get("/v1/traces/:traceId/observations", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		traceID := c.Params("traceId")
		if traceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Trace ID required",
			})
		}

		tree, err := mockSvc.GetObservationTree(c.Context(), pid, traceID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Trace not found",
			})
		}

		return c.JSON(tree)
	})

	// GetTraceStats
	app.Get("/v1/traces/:traceId/stats", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		traceID := c.Params("traceId")
		if traceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Trace ID required",
			})
		}

		filter := &domain.TraceFilter{
			ProjectID: pid,
			IDs:       []string{traceID},
		}

		stats, err := mockSvc.GetTraceStats(c.Context(), filter)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(stats)
	})

	// GetSession
	app.Get("/v1/sessions/:sessionId", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		sessionID := c.Params("sessionId")
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Session ID required",
			})
		}

		traces, err := mockSvc.GetSessionTraces(c.Context(), pid, sessionID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		if len(traces) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Session not found",
			})
		}

		return c.JSON(fiber.Map{
			"id":     sessionID,
			"traces": traces,
		})
	})

	return app
}

func TestTracesHandler_ListTraces(t *testing.T) {
	t.Run("successfully lists traces", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		now := time.Now()
		expectedList := &domain.TraceList{
			Traces: []domain.Trace{
				{
					ID:        "trace-1",
					ProjectID: projectID,
					Name:      "trace-1",
					StartTime: now,
				},
				{
					ID:        "trace-2",
					ProjectID: projectID,
					Name:      "trace-2",
					StartTime: now,
				},
			},
			TotalCount: 2,
			HasMore:    false,
		}

		mockSvc.On("ListTraces", mock.Anything, mock.AnythingOfType("*domain.TraceFilter"), 50, 0).
			Return(expectedList, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.TraceList
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Len(t, result.Traces, 2)
		assert.Equal(t, int64(2), result.TotalCount)

		mockSvc.AssertExpectations(t)
	})

	t.Run("uses limit parameter", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		expectedList := &domain.TraceList{
			Traces:     []domain.Trace{},
			TotalCount: 0,
			HasMore:    false,
		}

		mockSvc.On("ListTraces", mock.Anything, mock.AnythingOfType("*domain.TraceFilter"), 10, 0).
			Return(expectedList, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces?limit=10", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

func TestTracesHandler_GetTrace(t *testing.T) {
	t.Run("successfully gets trace", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		expectedTrace := &domain.Trace{
			ID:        "trace-123",
			ProjectID: projectID,
			Name:      "test-trace",
			StartTime: time.Now(),
		}

		mockSvc.On("GetTrace", mock.Anything, projectID, "trace-123").
			Return(expectedTrace, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces/trace-123", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Trace
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "trace-123", result.ID)
		assert.Equal(t, "test-trace", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent trace", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		mockSvc.On("GetTrace", mock.Anything, projectID, "non-existent").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces/non-existent", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

func TestTracesHandler_GetTraceObservations(t *testing.T) {
	t.Run("successfully gets trace observations", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		rootObs := domain.Observation{
			ID:      "obs-1",
			TraceID: "trace-123",
			Name:    "root-span",
			Type:    domain.ObservationTypeSpan,
		}
		childObs := domain.Observation{
			ID:                  "obs-2",
			TraceID:             "trace-123",
			ParentObservationID: stringPtr("obs-1"),
			Name:                "child-span",
			Type:                domain.ObservationTypeSpan,
		}
		expectedTree := &domain.ObservationTree{
			Observation: &rootObs,
			Children: []*domain.ObservationTree{
				{
					Observation: &childObs,
					Children:    nil,
				},
			},
		}

		mockSvc.On("GetObservationTree", mock.Anything, projectID, "trace-123").
			Return(expectedTree, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces/trace-123/observations", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.ObservationTree
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.NotNil(t, result.Observation)
		assert.Equal(t, "obs-1", result.Observation.ID)

		mockSvc.AssertExpectations(t)
	})
}

func TestTracesHandler_GetTraceStats(t *testing.T) {
	t.Run("successfully gets trace stats", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		expectedStats := &service.TraceStats{
			TotalCount:  1,
			AvgDuration: 150.5,
			TotalCost:   0.05,
			TotalTokens: 1500,
			ErrorCount:  0,
			ErrorRate:   0,
		}

		mockSvc.On("GetTraceStats", mock.Anything, mock.AnythingOfType("*domain.TraceFilter")).
			Return(expectedStats, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces/trace-123/stats", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result service.TraceStats
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, 150.5, result.AvgDuration)

		mockSvc.AssertExpectations(t)
	})
}

func TestTracesHandler_GetSession(t *testing.T) {
	t.Run("successfully gets session", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		expectedTraces := []domain.Trace{
			{
				ID:        "trace-1",
				ProjectID: projectID,
				SessionID: "session-123",
				Name:      "trace-1",
			},
			{
				ID:        "trace-2",
				ProjectID: projectID,
				SessionID: "session-123",
				Name:      "trace-2",
			},
		}

		mockSvc.On("GetSessionTraces", mock.Anything, projectID, "session-123").
			Return(expectedTraces, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/sessions/session-123", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "session-123", result["id"])
		traces := result["traces"].([]interface{})
		assert.Len(t, traces, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent session", func(t *testing.T) {
		mockSvc := new(MockQueryService)
		projectID := uuid.New()
		app := setupTracesTestApp(mockSvc, projectID)

		mockSvc.On("GetSessionTraces", mock.Anything, projectID, "non-existent").
			Return([]domain.Trace{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/sessions/non-existent", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

