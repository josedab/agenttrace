package handler

import (
	"bytes"
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

// MockScoreService mocks the score service
type MockScoreService struct {
	mock.Mock
}

func (m *MockScoreService) Create(ctx context.Context, projectID uuid.UUID, input *domain.ScoreInput) (*domain.Score, error) {
	args := m.Called(ctx, projectID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Score), args.Error(1)
}

func (m *MockScoreService) CreateBatch(ctx context.Context, projectID uuid.UUID, inputs []*domain.ScoreInput) ([]*domain.Score, error) {
	args := m.Called(ctx, projectID, inputs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Score), args.Error(1)
}

func (m *MockScoreService) Get(ctx context.Context, projectID uuid.UUID, scoreID string) (*domain.Score, error) {
	args := m.Called(ctx, projectID, scoreID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Score), args.Error(1)
}

func (m *MockScoreService) List(ctx context.Context, filter *domain.ScoreFilter, limit, offset int) (*domain.ScoreList, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScoreList), args.Error(1)
}

func (m *MockScoreService) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Score, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Score), args.Error(1)
}

func (m *MockScoreService) Update(ctx context.Context, projectID uuid.UUID, scoreID string, input *domain.ScoreInput) (*domain.Score, error) {
	args := m.Called(ctx, projectID, scoreID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Score), args.Error(1)
}

func (m *MockScoreService) GetStats(ctx context.Context, projectID uuid.UUID, name string) (*domain.ScoreStats, error) {
	args := m.Called(ctx, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScoreStats), args.Error(1)
}

func (m *MockScoreService) SubmitFeedback(ctx context.Context, projectID uuid.UUID, traceID string, feedback *service.FeedbackInput) (*domain.Score, error) {
	args := m.Called(ctx, projectID, traceID, feedback)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Score), args.Error(1)
}

func setupScoresTestApp(mockSvc *MockScoreService, projectID uuid.UUID) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()

	app.Use(testProjectMiddleware(projectID))

	// ListScores
	app.Get("/v1/scores", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		filter := &domain.ScoreFilter{
			ProjectID: pid,
		}

		limit := parseIntParam(c, "limit", 50)
		offset := parseIntParam(c, "offset", 0)

		list, err := mockSvc.List(c.Context(), filter, limit, offset)
		if err != nil {
			logger.Error("failed to list scores")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(fiber.Map{
			"data":       list.Scores,
			"totalCount": list.TotalCount,
		})
	})

	// GetScoreStats - must be registered before :scoreId to avoid matching "stats" as a scoreId
	app.Get("/v1/scores/stats", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		scoreName := c.Query("name")
		if scoreName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Score name required",
			})
		}

		stats, err := mockSvc.GetStats(c.Context(), pid, scoreName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(stats)
	})

	// BatchCreateScores - must be registered before :scoreId
	app.Post("/v1/scores/batch", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var request struct {
			Scores []*domain.ScoreInput `json:"scores"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		if len(request.Scores) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "At least one score required",
			})
		}

		results, err := mockSvc.CreateBatch(c.Context(), pid, request.Scores)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"data": results,
		})
	})

	// GetScore
	app.Get("/v1/scores/:scoreId", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		scoreID := c.Params("scoreId")
		if scoreID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Score ID required",
			})
		}

		score, err := mockSvc.Get(c.Context(), pid, scoreID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Score not found",
			})
		}

		return c.JSON(score)
	})

	// CreateScore
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
				"message": "Invalid request body",
			})
		}

		if input.TraceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "traceId is required",
			})
		}

		if input.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "name is required",
			})
		}

		score, err := mockSvc.Create(c.Context(), pid, &input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(score)
	})

	// UpdateScore
	app.Patch("/v1/scores/:scoreId", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		scoreID := c.Params("scoreId")
		if scoreID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Score ID required",
			})
		}

		var input domain.ScoreInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		score, err := mockSvc.Update(c.Context(), pid, scoreID, &input)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "Score not found",
			})
		}

		return c.JSON(score)
	})

	// GetTraceScores
	app.Get("/v1/traces/:traceId/scores", func(c *fiber.Ctx) error {
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

		scores, err := mockSvc.GetByTraceID(c.Context(), pid, traceID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.JSON(fiber.Map{
			"data": scores,
		})
	})

	// SubmitFeedback
	app.Post("/v1/feedback", func(c *fiber.Ctx) error {
		pid, ok := middleware.GetProjectID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var request struct {
			TraceID  string              `json:"traceId"`
			Name     string              `json:"name"`
			Value    *float64            `json:"value"`
			DataType domain.ScoreDataType `json:"dataType"`
			Comment  *string             `json:"comment"`
		}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		if request.TraceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "traceId is required",
			})
		}

		feedback := &service.FeedbackInput{
			Name:     request.Name,
			Value:    request.Value,
			DataType: request.DataType,
			Comment:  request.Comment,
		}

		score, err := mockSvc.SubmitFeedback(c.Context(), pid, request.TraceID, feedback)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(score)
	})

	return app
}

func TestScoresHandler_ListScores(t *testing.T) {
	t.Parallel()
	t.Run("successfully lists scores", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		value := 0.95
		expectedList := &domain.ScoreList{
			Scores: []domain.Score{
				{
					ID:        uuid.New(),
					ProjectID: projectID,
					TraceID:   "trace-1",
					Name:      "accuracy",
					Value:     &value,
					Source:    domain.ScoreSourceAPI,
					DataType:  domain.ScoreDataTypeNumeric,
					CreatedAt: time.Now(),
				},
			},
			TotalCount: 1,
			HasMore:    false,
		}

		mockSvc.On("List", mock.Anything, mock.AnythingOfType("*domain.ScoreFilter"), 50, 0).
			Return(expectedList, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/scores", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, float64(1), result["totalCount"])

		mockSvc.AssertExpectations(t)
	})
}

func TestScoresHandler_GetScore(t *testing.T) {
	t.Parallel()
	t.Run("successfully gets score", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		scoreID := uuid.New()
		value := 0.85
		expectedScore := &domain.Score{
			ID:        scoreID,
			ProjectID: projectID,
			TraceID:   "trace-123",
			Name:      "quality",
			Value:     &value,
			Source:    domain.ScoreSourceAPI,
			DataType:  domain.ScoreDataTypeNumeric,
			CreatedAt: time.Now(),
		}

		mockSvc.On("Get", mock.Anything, projectID, scoreID.String()).
			Return(expectedScore, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/scores/"+scoreID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Score
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "quality", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 404 for non-existent score", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		scoreID := uuid.New()
		mockSvc.On("Get", mock.Anything, projectID, scoreID.String()).
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/v1/scores/"+scoreID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})
}

func TestScoresHandler_CreateScore(t *testing.T) {
	t.Parallel()
	t.Run("successfully creates score", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		value := 0.9
		scoreID := uuid.New()
		expectedScore := &domain.Score{
			ID:        scoreID,
			ProjectID: projectID,
			TraceID:   "trace-123",
			Name:      "relevance",
			Value:     &value,
			Source:    domain.ScoreSourceAPI,
			DataType:  domain.ScoreDataTypeNumeric,
			CreatedAt: time.Now(),
		}

		mockSvc.On("Create", mock.Anything, projectID, mock.AnythingOfType("*domain.ScoreInput")).
			Return(expectedScore, nil)

		body := map[string]interface{}{
			"traceId": "trace-123",
			"name":    "relevance",
			"value":   0.9,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/scores", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Score
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "relevance", result.Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing traceId", func(t *testing.T) {
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"name":  "relevance",
			"value": 0.9,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/scores", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"traceId": "trace-123",
			"value":   0.9,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/scores", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestScoresHandler_UpdateScore(t *testing.T) {
	t.Parallel()
	t.Run("successfully updates score", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		scoreID := uuid.New()
		newValue := 0.95
		expectedScore := &domain.Score{
			ID:        scoreID,
			ProjectID: projectID,
			TraceID:   "trace-123",
			Name:      "relevance",
			Value:     &newValue,
			Source:    domain.ScoreSourceAPI,
			DataType:  domain.ScoreDataTypeNumeric,
			UpdatedAt: time.Now(),
		}

		mockSvc.On("Update", mock.Anything, projectID, scoreID.String(), mock.AnythingOfType("*domain.ScoreInput")).
			Return(expectedScore, nil)

		body := map[string]interface{}{
			"value": 0.95,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/v1/scores/"+scoreID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.Score
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, newValue, *result.Value)

		mockSvc.AssertExpectations(t)
	})
}

func TestScoresHandler_GetTraceScores(t *testing.T) {
	t.Parallel()
	t.Run("successfully gets trace scores", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		value1 := 0.9
		value2 := 0.8
		expectedScores := []domain.Score{
			{
				ID:        uuid.New(),
				ProjectID: projectID,
				TraceID:   "trace-123",
				Name:      "accuracy",
				Value:     &value1,
			},
			{
				ID:        uuid.New(),
				ProjectID: projectID,
				TraceID:   "trace-123",
				Name:      "relevance",
				Value:     &value2,
			},
		}

		mockSvc.On("GetByTraceID", mock.Anything, projectID, "trace-123").
			Return(expectedScores, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/traces/trace-123/scores", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		data := result["data"].([]interface{})
		assert.Len(t, data, 2)

		mockSvc.AssertExpectations(t)
	})
}

func TestScoresHandler_BatchCreateScores(t *testing.T) {
	t.Parallel()
	t.Run("successfully batch creates scores", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		value1 := 0.9
		value2 := 0.8
		expectedScores := []*domain.Score{
			{
				ID:        uuid.New(),
				ProjectID: projectID,
				TraceID:   "trace-1",
				Name:      "accuracy",
				Value:     &value1,
			},
			{
				ID:        uuid.New(),
				ProjectID: projectID,
				TraceID:   "trace-2",
				Name:      "relevance",
				Value:     &value2,
			},
		}

		mockSvc.On("CreateBatch", mock.Anything, projectID, mock.AnythingOfType("[]*domain.ScoreInput")).
			Return(expectedScores, nil)

		body := map[string]interface{}{
			"scores": []map[string]interface{}{
				{"traceId": "trace-1", "name": "accuracy", "value": 0.9},
				{"traceId": "trace-2", "name": "relevance", "value": 0.8},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/scores/batch", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		data := result["data"].([]interface{})
		assert.Len(t, data, 2)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for empty scores array", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"scores": []map[string]interface{}{},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/scores/batch", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestScoresHandler_GetScoreStats(t *testing.T) {
	t.Parallel()
	t.Run("successfully gets score stats", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		avgValue := 0.85
		minValue := 0.5
		maxValue := 1.0
		expectedStats := &domain.ScoreStats{
			Name:     "accuracy",
			Count:    100,
			AvgValue: &avgValue,
			MinValue: &minValue,
			MaxValue: &maxValue,
		}

		mockSvc.On("GetStats", mock.Anything, projectID, "accuracy").
			Return(expectedStats, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/scores/stats?name=accuracy", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result domain.ScoreStats
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, "accuracy", result.Name)
		assert.Equal(t, int64(100), result.Count)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing name", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		req := httptest.NewRequest(http.MethodGet, "/v1/scores/stats", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestScoresHandler_SubmitFeedback(t *testing.T) {
	t.Parallel()
	t.Run("successfully submits feedback", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		value := 1.0
		expectedScore := &domain.Score{
			ID:        uuid.New(),
			ProjectID: projectID,
			TraceID:   "trace-123",
			Name:      "user-feedback",
			Value:     &value,
			Source:    domain.ScoreSourceAnnotation,
			DataType:  domain.ScoreDataTypeNumeric,
			CreatedAt: time.Now(),
		}

		mockSvc.On("SubmitFeedback", mock.Anything, projectID, "trace-123", mock.AnythingOfType("*service.FeedbackInput")).
			Return(expectedScore, nil)

		body := map[string]interface{}{
			"traceId": "trace-123",
			"value":   1.0,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result domain.Score
		respBody, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)
		assert.Equal(t, value, *result.Value)

		mockSvc.AssertExpectations(t)
	})

	t.Run("returns 400 for missing traceId", func(t *testing.T) {
		t.Parallel()
		mockSvc := new(MockScoreService)
		projectID := uuid.New()
		app := setupScoresTestApp(mockSvc, projectID)

		body := map[string]interface{}{
			"value": 1.0,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
