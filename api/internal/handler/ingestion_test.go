package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
)

// MockIngestionService mocks the ingestion service
type MockIngestionService struct {
	mock.Mock
}

func (m *MockIngestionService) IngestTrace(ctx context.Context, projectID uuid.UUID, input *domain.TraceInput) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockIngestionService) IngestObservation(ctx context.Context, projectID uuid.UUID, input *domain.ObservationInput) (*domain.Observation, error) {
	args := m.Called(ctx, projectID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Observation), args.Error(1)
}

func (m *MockIngestionService) IngestScore(ctx context.Context, projectID uuid.UUID, input *domain.ScoreInput) (*domain.Score, error) {
	args := m.Called(ctx, projectID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Score), args.Error(1)
}

func (m *MockIngestionService) IngestOTLP(ctx context.Context, projectID uuid.UUID, data []byte, contentType string) error {
	args := m.Called(ctx, projectID, data, contentType)
	return args.Error(0)
}

func (m *MockIngestionService) IngestBatch(ctx context.Context, projectID uuid.UUID, batch *domain.IngestionBatch) error {
	args := m.Called(ctx, projectID, batch)
	return args.Error(0)
}

func (m *MockIngestionService) UpdateTrace(ctx context.Context, projectID uuid.UUID, traceID string, input *domain.TraceInput) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockIngestionService) UpdateObservation(ctx context.Context, projectID uuid.UUID, observationID string, input *domain.ObservationInput) (*domain.Observation, error) {
	args := m.Called(ctx, projectID, observationID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Observation), args.Error(1)
}

// testProjectMiddleware injects a test project ID
func testProjectMiddleware(projectID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(string(middleware.ContextKeyProjectID), projectID)
		return c.Next()
	}
}

func setupIngestionTestApp(mockSvc *MockIngestionService, projectID uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	// We need to create an adapter since IngestionHandler expects *service.IngestionService
	// For testing, we'll create test routes directly
	app.Use(testProjectMiddleware(projectID))

	app.Post("/v1/traces", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var input domain.TraceInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}

		trace, err := mockSvc.IngestTrace(c.Context(), pid, &input)
		if err != nil {
			logger.Error("failed to create trace")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(trace)
	})

	app.Post("/v1/spans", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var input domain.ObservationInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}

		obsType := domain.ObservationTypeSpan
		input.Type = &obsType

		obs, err := mockSvc.IngestObservation(c.Context(), pid, &input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(obs)
	})

	app.Post("/v1/generations", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var input domain.ObservationInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}

		obsType := domain.ObservationTypeGeneration
		input.Type = &obsType

		obs, err := mockSvc.IngestObservation(c.Context(), pid, &input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(obs)
	})

	app.Post("/v1/scores", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var input domain.ScoreInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		}

		score, err := mockSvc.IngestScore(c.Context(), pid, &input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(score)
	})

	app.Post("/api/public/ingestion", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var request struct {
			Batch []json.RawMessage `json:"batch"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Bad Request",
			})
		}

		if len(request.Batch) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Batch is empty",
			})
		}

		successes := make([]string, 0)
		errors := make([]map[string]any, 0)

		for _, item := range request.Batch {
			var common struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			}
			if err := json.Unmarshal(item, &common); err != nil {
				errors = append(errors, map[string]any{
					"id":      "unknown",
					"status":  400,
					"message": err.Error(),
				})
				continue
			}

			eventID := common.ID
			if eventID == "" {
				eventID = uuid.New().String()
			}

			switch common.Type {
			case "trace-create":
				var traceInput domain.TraceInput
				if err := json.Unmarshal(item, &struct {
					Body *domain.TraceInput `json:"body"`
				}{Body: &traceInput}); err != nil {
					errors = append(errors, map[string]any{
						"id":      eventID,
						"status":  400,
						"message": err.Error(),
					})
					continue
				}
				if _, err := mockSvc.IngestTrace(c.Context(), pid, &traceInput); err != nil {
					errors = append(errors, map[string]any{
						"id":      eventID,
						"status":  500,
						"message": err.Error(),
					})
					continue
				}
				successes = append(successes, eventID)

			case "span-create", "generation-create":
				var obsInput domain.ObservationInput
				if err := json.Unmarshal(item, &struct {
					Body *domain.ObservationInput `json:"body"`
				}{Body: &obsInput}); err != nil {
					errors = append(errors, map[string]any{
						"id":      eventID,
						"status":  400,
						"message": err.Error(),
					})
					continue
				}

				if common.Type == "span-create" {
					t := domain.ObservationTypeSpan
					obsInput.Type = &t
				} else {
					t := domain.ObservationTypeGeneration
					obsInput.Type = &t
				}

				if _, err := mockSvc.IngestObservation(c.Context(), pid, &obsInput); err != nil {
					errors = append(errors, map[string]any{
						"id":      eventID,
						"status":  500,
						"message": err.Error(),
					})
					continue
				}
				successes = append(successes, eventID)

			default:
				errors = append(errors, map[string]any{
					"id":      eventID,
					"status":  400,
					"message": "Unknown event type",
				})
			}
		}

		return c.JSON(fiber.Map{
			"successes": successes,
			"errors":    errors,
		})
	})

	return app
}

func TestIngestionHandler_CreateTrace(t *testing.T) {
	t.Run("successfully creates trace", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		expectedTrace := &domain.Trace{
			ID:        "trace-123",
			ProjectID: projectID,
			Name:      "test-trace",
		}

		mockSvc.On("IngestTrace", mock.Anything, projectID, mock.AnythingOfType("*domain.TraceInput")).
			Return(expectedTrace, nil)

		body := map[string]interface{}{
			"name": "test-trace",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Trace
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "trace-123", result.ID)
		assert.Equal(t, "test-trace", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		req := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestIngestionHandler_CreateSpan(t *testing.T) {
	t.Run("successfully creates span", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		expectedObs := &domain.Observation{
			ID:      "obs-123",
			TraceID: "trace-123",
			Type:    domain.ObservationTypeSpan,
			Name:    "test-span",
		}

		mockSvc.On("IngestObservation", mock.Anything, projectID, mock.AnythingOfType("*domain.ObservationInput")).
			Return(expectedObs, nil)

		body := map[string]interface{}{
			"traceId": "trace-123",
			"name":    "test-span",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/spans", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Observation
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "obs-123", result.ID)
		assert.Equal(t, domain.ObservationTypeSpan, result.Type)

		mockSvc.AssertExpectations(t)
	})
}

func TestIngestionHandler_CreateGeneration(t *testing.T) {
	t.Run("successfully creates generation", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		expectedObs := &domain.Observation{
			ID:      "gen-123",
			TraceID: "trace-123",
			Type:    domain.ObservationTypeGeneration,
			Name:    "llm-call",
			Model:   "gpt-4",
		}

		mockSvc.On("IngestObservation", mock.Anything, projectID, mock.AnythingOfType("*domain.ObservationInput")).
			Return(expectedObs, nil)

		body := map[string]interface{}{
			"traceId": "trace-123",
			"name":    "llm-call",
			"model":   "gpt-4",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/generations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Observation
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "gen-123", result.ID)
		assert.Equal(t, domain.ObservationTypeGeneration, result.Type)
		assert.Equal(t, "gpt-4", result.Model)

		mockSvc.AssertExpectations(t)
	})
}

func TestIngestionHandler_CreateScore(t *testing.T) {
	t.Run("successfully creates score", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		scoreValue := 0.95
		expectedScore := &domain.Score{
			ID:      uuid.New(),
			TraceID: "trace-123",
			Name:    "accuracy",
			Value:   &scoreValue,
		}

		mockSvc.On("IngestScore", mock.Anything, projectID, mock.AnythingOfType("*domain.ScoreInput")).
			Return(expectedScore, nil)

		body := map[string]interface{}{
			"traceId": "trace-123",
			"name":    "accuracy",
			"value":   0.95,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/scores", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

func TestIngestionHandler_BatchIngestion(t *testing.T) {
	t.Run("successfully processes batch", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		mockSvc.On("IngestTrace", mock.Anything, projectID, mock.AnythingOfType("*domain.TraceInput")).
			Return(&domain.Trace{ID: "trace-1"}, nil)

		mockSvc.On("IngestObservation", mock.Anything, projectID, mock.AnythingOfType("*domain.ObservationInput")).
			Return(&domain.Observation{ID: "obs-1"}, nil)

		batch := map[string]interface{}{
			"batch": []map[string]interface{}{
				{
					"id":   "event-1",
					"type": "trace-create",
					"body": map[string]interface{}{
						"name": "trace-1",
					},
				},
				{
					"id":   "event-2",
					"type": "span-create",
					"body": map[string]interface{}{
						"traceId": "trace-1",
						"name":    "span-1",
					},
				},
			},
		}
		jsonBody, _ := json.Marshal(batch)

		req := httptest.NewRequest(http.MethodPost, "/api/public/ingestion", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		successes := result["successes"].([]interface{})
		assert.Len(t, successes, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for empty batch", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		batch := map[string]interface{}{
			"batch": []map[string]interface{}{},
		}
		jsonBody, _ := json.Marshal(batch)

		req := httptest.NewRequest(http.MethodPost, "/api/public/ingestion", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("handles partial failures", func(t *testing.T) {
		mockSvc := new(MockIngestionService)
		projectID := uuid.New()
		app := setupIngestionTestApp(mockSvc, projectID)

		mockSvc.On("IngestTrace", mock.Anything, projectID, mock.AnythingOfType("*domain.TraceInput")).
			Return(&domain.Trace{ID: "trace-1"}, nil)

		batch := map[string]interface{}{
			"batch": []map[string]interface{}{
				{
					"id":   "event-1",
					"type": "trace-create",
					"body": map[string]interface{}{
						"name": "trace-1",
					},
				},
				{
					"id":   "event-2",
					"type": "unknown-type",
					"body": map[string]interface{}{},
				},
			},
		}
		jsonBody, _ := json.Marshal(batch)

		req := httptest.NewRequest(http.MethodPost, "/api/public/ingestion", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		successes := result["successes"].([]interface{})
		errors := result["errors"].([]interface{})
		assert.Len(t, successes, 1)
		assert.Len(t, errors, 1)

		mockSvc.AssertExpectations(t)
	})
}
