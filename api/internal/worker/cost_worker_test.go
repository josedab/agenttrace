package worker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCostCalculationTask(t *testing.T) {
	payload := &CostCalculationPayload{
		ProjectID:        uuid.New().String(),
		TraceID:          "trace-123",
		ObservationID:    "obs-456",
		Model:            "gpt-4",
		PromptTokens:     100,
		CompletionTokens: 50,
	}

	task, err := NewCostCalculationTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeCostCalculation, task.Type())

	// Verify payload can be deserialized
	var decoded CostCalculationPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.TraceID, decoded.TraceID)
	assert.Equal(t, payload.Model, decoded.Model)
}

func TestNewBatchCostCalculationTask(t *testing.T) {
	payload := &BatchCostCalculationPayload{
		ProjectID: uuid.New().String(),
		TraceID:   "trace-123",
	}

	task, err := NewBatchCostCalculationTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeBatchCostCalculation, task.Type())

	// Verify payload
	var decoded BatchCostCalculationPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.TraceID, decoded.TraceID)
}

func TestNewDailyAggregationTask(t *testing.T) {
	payload := &DailyAggregationPayload{
		ProjectID: uuid.New().String(),
		Date:      "2024-01-15",
	}

	task, err := NewDailyAggregationTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeDailyAggregation, task.Type())

	// Verify payload
	var decoded DailyAggregationPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.Date, decoded.Date)
}

func TestCostWorker_ProcessTask_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	// Create task with invalid payload
	task := asynq.NewTask(TypeCostCalculation, []byte("invalid json"))

	err := worker.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestCostWorker_ProcessTask_InvalidProjectID(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	payload := &CostCalculationPayload{
		ProjectID:        "invalid-uuid",
		TraceID:          "trace-123",
		ObservationID:    "obs-456",
		Model:            "gpt-4",
		PromptTokens:     100,
		CompletionTokens: 50,
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeCostCalculation, data)

	err = worker.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project ID")
}

func TestCostWorker_ProcessBatchCostTask_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	// Create task with invalid payload
	task := asynq.NewTask(TypeBatchCostCalculation, []byte("invalid json"))

	err := worker.ProcessBatchCostTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestCostWorker_ProcessDailyAggregationTask_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	// Create task with invalid payload
	task := asynq.NewTask(TypeDailyAggregation, []byte("invalid json"))

	err := worker.ProcessDailyAggregationTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestCostWorker_ProcessBatchCostTask_InvalidProjectID(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	payload := &BatchCostCalculationPayload{
		ProjectID: "invalid-uuid",
		TraceID:   "trace-123",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeBatchCostCalculation, data)

	err = worker.ProcessBatchCostTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project ID")
}

func TestCostWorker_ProcessDailyAggregationTask_InvalidProjectID(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	payload := &DailyAggregationPayload{
		ProjectID: "invalid-uuid",
		Date:      "2024-01-15",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeDailyAggregation, data)

	err = worker.ProcessDailyAggregationTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project ID")
}

func TestCostWorker_ProcessDailyAggregationTask_InvalidDate(t *testing.T) {
	logger := zap.NewNop()
	worker := NewCostWorker(logger, nil, nil, nil)

	payload := &DailyAggregationPayload{
		ProjectID: uuid.New().String(),
		Date:      "invalid-date",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeDailyAggregation, data)

	err = worker.ProcessDailyAggregationTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

func TestCostCalculationPayload_Serialization(t *testing.T) {
	t.Run("complete payload", func(t *testing.T) {
		payload := CostCalculationPayload{
			ProjectID:        uuid.New().String(),
			TraceID:          "trace-abc",
			ObservationID:    "obs-xyz",
			Model:            "claude-3-opus",
			PromptTokens:     1500,
			CompletionTokens: 500,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded CostCalculationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, payload.ProjectID, decoded.ProjectID)
		assert.Equal(t, payload.TraceID, decoded.TraceID)
		assert.Equal(t, payload.ObservationID, decoded.ObservationID)
		assert.Equal(t, payload.Model, decoded.Model)
		assert.Equal(t, payload.PromptTokens, decoded.PromptTokens)
		assert.Equal(t, payload.CompletionTokens, decoded.CompletionTokens)
	})

	t.Run("zero tokens", func(t *testing.T) {
		payload := CostCalculationPayload{
			ProjectID:        uuid.New().String(),
			TraceID:          "trace-abc",
			Model:            "gpt-4",
			PromptTokens:     0,
			CompletionTokens: 0,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded CostCalculationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, 0, decoded.PromptTokens)
		assert.Equal(t, 0, decoded.CompletionTokens)
	})
}
