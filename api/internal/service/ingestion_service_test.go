package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MockTraceRepository is a mock implementation of TraceRepository
type MockTraceRepository struct {
	mock.Mock
}

func (m *MockTraceRepository) Create(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepository) CreateBatch(ctx context.Context, traces []*domain.Trace) error {
	args := m.Called(ctx, traces)
	return args.Error(0)
}

func (m *MockTraceRepository) GetByID(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.Trace, error) {
	args := m.Called(ctx, projectID, traceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trace), args.Error(1)
}

func (m *MockTraceRepository) Update(ctx context.Context, trace *domain.Trace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockTraceRepository) UpdateCosts(ctx context.Context, projectID uuid.UUID, traceID string, inputCost, outputCost, totalCost float64) error {
	args := m.Called(ctx, projectID, traceID, inputCost, outputCost, totalCost)
	return args.Error(0)
}

func (m *MockTraceRepository) List(ctx context.Context, filter *domain.TraceFilter, limit, offset int) (*domain.TraceList, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).(*domain.TraceList), args.Error(1)
}

func (m *MockTraceRepository) SetBookmark(ctx context.Context, projectID uuid.UUID, traceID string, bookmarked bool) error {
	args := m.Called(ctx, projectID, traceID, bookmarked)
	return args.Error(0)
}

func (m *MockTraceRepository) GetBySessionID(ctx context.Context, projectID uuid.UUID, sessionID string) ([]domain.Trace, error) {
	args := m.Called(ctx, projectID, sessionID)
	return args.Get(0).([]domain.Trace), args.Error(1)
}

func (m *MockTraceRepository) Delete(ctx context.Context, projectID uuid.UUID, traceID string) error {
	args := m.Called(ctx, projectID, traceID)
	return args.Error(0)
}

// MockObservationRepository is a mock implementation of ObservationRepository
type MockObservationRepository struct {
	mock.Mock
}

func (m *MockObservationRepository) Create(ctx context.Context, obs *domain.Observation) error {
	args := m.Called(ctx, obs)
	return args.Error(0)
}

func (m *MockObservationRepository) CreateBatch(ctx context.Context, observations []*domain.Observation) error {
	args := m.Called(ctx, observations)
	return args.Error(0)
}

func (m *MockObservationRepository) GetByID(ctx context.Context, projectID uuid.UUID, observationID string) (*domain.Observation, error) {
	args := m.Called(ctx, projectID, observationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Observation, error) {
	args := m.Called(ctx, projectID, traceID)
	return args.Get(0).([]domain.Observation), args.Error(1)
}

func (m *MockObservationRepository) Update(ctx context.Context, obs *domain.Observation) error {
	args := m.Called(ctx, obs)
	return args.Error(0)
}

func (m *MockObservationRepository) UpdateCosts(ctx context.Context, projectID uuid.UUID, observationID string, inputCost, outputCost, totalCost float64) error {
	args := m.Called(ctx, projectID, observationID, inputCost, outputCost, totalCost)
	return args.Error(0)
}

func (m *MockObservationRepository) List(ctx context.Context, filter *domain.ObservationFilter, limit, offset int) ([]domain.Observation, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).([]domain.Observation), args.Get(1).(int64), args.Error(2)
}

func (m *MockObservationRepository) GetGenerationsWithoutCost(ctx context.Context, projectID uuid.UUID, limit int) ([]domain.Observation, error) {
	args := m.Called(ctx, projectID, limit)
	return args.Get(0).([]domain.Observation), args.Error(1)
}

func (m *MockObservationRepository) GetTree(ctx context.Context, projectID uuid.UUID, traceID string) (*domain.ObservationTree, error) {
	args := m.Called(ctx, projectID, traceID)
	return args.Get(0).(*domain.ObservationTree), args.Error(1)
}

func TestIngestionService_IngestTrace(t *testing.T) {
	t.Run("creates trace with generated ID", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trace")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		input := &domain.TraceInput{
			Name: "test-trace",
		}

		trace, err := svc.IngestTrace(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.NotEmpty(t, trace.ID)
		assert.Equal(t, "test-trace", trace.Name)
		assert.Equal(t, projectID, trace.ProjectID)
		traceRepo.AssertExpectations(t)
	})

	t.Run("creates trace with custom ID", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trace")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		input := &domain.TraceInput{
			ID:   "custom-trace-id",
			Name: "test-trace",
		}

		trace, err := svc.IngestTrace(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, "custom-trace-id", trace.ID)
		traceRepo.AssertExpectations(t)
	})

	t.Run("creates trace with metadata", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		var capturedTrace *domain.Trace
		traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trace")).
			Run(func(args mock.Arguments) {
				capturedTrace = args.Get(1).(*domain.Trace)
			}).
			Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		input := &domain.TraceInput{
			Name:      "test-trace",
			UserID:    "user-123",
			SessionID: "session-456",
			Metadata: map[string]interface{}{
				"key": "value",
			},
			Tags: []string{"tag1", "tag2"},
		}

		trace, err := svc.IngestTrace(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, "user-123", trace.UserID)
		assert.Equal(t, "session-456", trace.SessionID)
		assert.Equal(t, []string{"tag1", "tag2"}, trace.Tags)

		// Verify metadata was serialized
		var metadata map[string]interface{}
		err = json.Unmarshal([]byte(capturedTrace.Metadata), &metadata)
		require.NoError(t, err)
		assert.Equal(t, "value", metadata["key"])

		traceRepo.AssertExpectations(t)
	})

	t.Run("creates trace with custom timestamp", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		traceRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trace")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		customTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		input := &domain.TraceInput{
			Name:      "test-trace",
			Timestamp: &customTime,
		}

		trace, err := svc.IngestTrace(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, customTime, trace.StartTime)
		traceRepo.AssertExpectations(t)
	})
}

func TestIngestionService_IngestObservation(t *testing.T) {
	t.Run("creates span observation", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		traceID := "trace-123"
		obsType := domain.ObservationTypeSpan
		name := "test-span"
		input := &domain.ObservationInput{
			TraceID: &traceID,
			Type:    &obsType,
			Name:    &name,
		}

		obs, err := svc.IngestObservation(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.NotEmpty(t, obs.ID)
		assert.Equal(t, "trace-123", obs.TraceID)
		assert.Equal(t, domain.ObservationTypeSpan, obs.Type)
		assert.Equal(t, "test-span", obs.Name)
		obsRepo.AssertExpectations(t)
	})

	t.Run("creates observation with parent", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		parentID := "parent-obs-id"
		traceID := "trace-123"
		obsType := domain.ObservationTypeSpan
		name := "child-span"
		input := &domain.ObservationInput{
			TraceID:             &traceID,
			Type:                &obsType,
			Name:                &name,
			ParentObservationID: &parentID,
		}

		obs, err := svc.IngestObservation(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, &parentID, obs.ParentObservationID)
		obsRepo.AssertExpectations(t)
	})

	t.Run("creates observation with input/output", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		var capturedObs *domain.Observation
		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).
			Run(func(args mock.Arguments) {
				capturedObs = args.Get(1).(*domain.Observation)
			}).
			Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		traceID := "trace-123"
		obsType := domain.ObservationTypeSpan
		name := "test-span"
		input := &domain.ObservationInput{
			TraceID: &traceID,
			Type:    &obsType,
			Name:    &name,
			Input:   map[string]interface{}{"query": "hello"},
			Output:  map[string]interface{}{"response": "world"},
		}

		_, err := svc.IngestObservation(context.Background(), projectID, input)

		require.NoError(t, err)

		// Verify input/output were serialized
		var inputData map[string]interface{}
		err = json.Unmarshal([]byte(capturedObs.Input), &inputData)
		require.NoError(t, err)
		assert.Equal(t, "hello", inputData["query"])

		var outputData map[string]interface{}
		err = json.Unmarshal([]byte(capturedObs.Output), &outputData)
		require.NoError(t, err)
		assert.Equal(t, "world", outputData["response"])

		obsRepo.AssertExpectations(t)
	})
}

func TestIngestionService_IngestGeneration(t *testing.T) {
	t.Run("creates generation with model info", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)
		// Mock for background cost update goroutine
		obsRepo.On("GetByTraceID", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Observation{}, nil).Maybe()
		traceRepo.On("UpdateCosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		traceID := "trace-123"
		name := "llm-call"
		input := &domain.GenerationInput{
			ObservationInput: domain.ObservationInput{
				TraceID: &traceID,
				Name:    &name,
			},
			Model: "gpt-4",
		}

		obs, err := svc.IngestGeneration(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, domain.ObservationTypeGeneration, obs.Type)
		assert.Equal(t, "gpt-4", obs.Model)
		obsRepo.AssertExpectations(t)
	})

	t.Run("creates generation with usage", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)
		// Mock for background cost update goroutine
		obsRepo.On("GetByTraceID", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Observation{}, nil).Maybe()
		traceRepo.On("UpdateCosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		traceID := "trace-123"
		name := "llm-call"
		inputTokens := int64(100)
		outputTokens := int64(50)
		input := &domain.GenerationInput{
			ObservationInput: domain.ObservationInput{
				TraceID: &traceID,
				Name:    &name,
			},
			Model: "gpt-4",
			Usage: &domain.UsageDetailsInput{
				InputTokens:  &inputTokens,
				OutputTokens: &outputTokens,
			},
		}

		obs, err := svc.IngestGeneration(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, uint64(100), obs.UsageDetails.InputTokens)
		assert.Equal(t, uint64(50), obs.UsageDetails.OutputTokens)
		assert.Equal(t, uint64(150), obs.UsageDetails.TotalTokens) // Auto-calculated
		obsRepo.AssertExpectations(t)
	})

	t.Run("calculates latency from start/end time", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)
		// Mock for background cost update goroutine
		obsRepo.On("GetByTraceID", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Observation{}, nil).Maybe()
		traceRepo.On("UpdateCosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		traceID := "trace-123"
		name := "llm-call"
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 15, 10, 0, 2, 500000000, time.UTC) // 2.5 seconds later
		input := &domain.GenerationInput{
			ObservationInput: domain.ObservationInput{
				TraceID:   &traceID,
				Name:      &name,
				StartTime: &startTime,
				EndTime:   &endTime,
			},
			Model: "gpt-4",
		}

		obs, err := svc.IngestGeneration(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Equal(t, float64(2500), obs.DurationMs)
		obsRepo.AssertExpectations(t)
	})

	t.Run("calculates cost with cost service", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)
		costService := NewCostService(zap.NewNop())

		obsRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)
		// Mock for background cost update goroutine
		obsRepo.On("GetByTraceID", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Observation{}, nil).Maybe()
		traceRepo.On("UpdateCosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, costService, nil)

		projectID := uuid.New()
		traceID := "trace-123"
		name := "llm-call"
		inputTokens := int64(1000)
		outputTokens := int64(500)
		input := &domain.GenerationInput{
			ObservationInput: domain.ObservationInput{
				TraceID: &traceID,
				Name:    &name,
			},
			Model: "gpt-4",
			Usage: &domain.UsageDetailsInput{
				InputTokens:  &inputTokens,
				OutputTokens: &outputTokens,
			},
		}

		obs, err := svc.IngestGeneration(context.Background(), projectID, input)

		require.NoError(t, err)
		assert.Greater(t, obs.CostDetails.TotalCost, 0.0)
		obsRepo.AssertExpectations(t)
	})
}

func TestIngestionService_IngestBatch(t *testing.T) {
	t.Run("ingests batch of traces and observations", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		traceRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*domain.Trace")).Return(nil)
		obsRepo.On("CreateBatch", mock.Anything, mock.AnythingOfType("[]*domain.Observation")).Return(nil)
		// Mock for background cost update goroutine
		obsRepo.On("GetByTraceID", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Observation{}, nil).Maybe()
		traceRepo.On("UpdateCosts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		traceID1 := "trace-1"
		obsType := domain.ObservationTypeSpan
		spanName := "span-1"
		genName := "gen-1"
		batch := &domain.IngestionBatch{
			Traces: []*domain.TraceInput{
				{Name: "trace-1"},
				{Name: "trace-2"},
			},
			Observations: []*domain.ObservationInput{
				{TraceID: &traceID1, Type: &obsType, Name: &spanName},
			},
			Generations: []*domain.GenerationInput{
				{ObservationInput: domain.ObservationInput{TraceID: &traceID1, Name: &genName}, Model: "gpt-4"},
			},
		}

		err := svc.IngestBatch(context.Background(), projectID, batch)

		require.NoError(t, err)
		traceRepo.AssertExpectations(t)
		obsRepo.AssertExpectations(t)
	})

	t.Run("handles empty batch", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		projectID := uuid.New()
		batch := &domain.IngestionBatch{}

		err := svc.IngestBatch(context.Background(), projectID, batch)

		require.NoError(t, err)
		// No calls should be made for empty batch
		traceRepo.AssertNotCalled(t, "CreateBatch")
		obsRepo.AssertNotCalled(t, "CreateBatch")
	})
}

func TestIngestionService_UpdateTrace(t *testing.T) {
	t.Run("updates trace fields", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		projectID := uuid.New()
		existingTrace := &domain.Trace{
			ID:        "trace-123",
			ProjectID: projectID,
			Name:      "original-name",
		}

		traceRepo.On("GetByID", mock.Anything, projectID, "trace-123").Return(existingTrace, nil)
		traceRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Trace")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		input := &domain.TraceInput{
			Name:   "updated-name",
			UserID: "new-user",
		}

		trace, err := svc.UpdateTrace(context.Background(), projectID, "trace-123", input)

		require.NoError(t, err)
		assert.Equal(t, "updated-name", trace.Name)
		assert.Equal(t, "new-user", trace.UserID)
		traceRepo.AssertExpectations(t)
	})
}

func TestIngestionService_UpdateObservation(t *testing.T) {
	t.Run("updates observation with end time and output", func(t *testing.T) {
		traceRepo := new(MockTraceRepository)
		obsRepo := new(MockObservationRepository)

		projectID := uuid.New()
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		existingObs := &domain.Observation{
			ID:        "obs-123",
			ProjectID: projectID,
			Name:      "test-span",
			StartTime: startTime,
		}

		obsRepo.On("GetByID", mock.Anything, projectID, "obs-123").Return(existingObs, nil)
		obsRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Observation")).Return(nil)

		svc := NewIngestionService(zap.NewNop(), traceRepo, obsRepo, nil, nil)

		endTime := time.Date(2024, 1, 15, 10, 0, 5, 0, time.UTC) // 5 seconds later
		input := &domain.ObservationInput{
			EndTime: &endTime,
			Output:  map[string]interface{}{"result": "success"},
		}

		obs, err := svc.UpdateObservation(context.Background(), projectID, "obs-123", input)

		require.NoError(t, err)
		assert.Equal(t, &endTime, obs.EndTime)
		assert.Equal(t, float64(5000), obs.DurationMs)
		obsRepo.AssertExpectations(t)
	})
}
