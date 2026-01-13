package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewEvaluationTask(t *testing.T) {
	payload := &EvaluationPayload{
		ProjectID:     uuid.New(),
		EvaluatorID:   uuid.New(),
		TraceID:       "trace-123",
		ObservationID: "obs-456",
	}

	task, err := NewEvaluationTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeEvaluation, task.Type())

	// Verify payload
	var decoded EvaluationPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.EvaluatorID, decoded.EvaluatorID)
	assert.Equal(t, payload.TraceID, decoded.TraceID)
	assert.Equal(t, payload.ObservationID, decoded.ObservationID)
}

func TestNewBatchEvaluationTask(t *testing.T) {
	payload := &BatchEvaluationPayload{
		ProjectID:   uuid.New(),
		EvaluatorID: uuid.New(),
		TraceIDs:    []string{"trace-1", "trace-2", "trace-3"},
	}

	task, err := NewBatchEvaluationTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeBatchEvaluation, task.Type())

	// Verify payload
	var decoded BatchEvaluationPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.EvaluatorID, decoded.EvaluatorID)
	assert.Equal(t, payload.TraceIDs, decoded.TraceIDs)
}

func TestEvalWorker_ProcessTask_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	// Create task with invalid payload
	task := asynq.NewTask(TypeEvaluation, []byte("invalid json"))

	err := worker.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestEvalWorker_ProcessBatchTask_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	// Create task with invalid payload
	task := asynq.NewTask(TypeBatchEvaluation, []byte("invalid json"))

	err := worker.ProcessBatchTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestEvalWorker_BuildEvaluationSystemPrompt(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	tests := []struct {
		name     string
		dataType domain.ScoreDataType
		contains []string
	}{
		{
			name:     "numeric",
			dataType: domain.ScoreDataTypeNumeric,
			contains: []string{"score", "0.0 to 1.0", "reasoning"},
		},
		{
			name:     "boolean",
			dataType: domain.ScoreDataTypeBoolean,
			contains: []string{"passed", "true/false", "reasoning"},
		},
		{
			name:     "categorical",
			dataType: domain.ScoreDataTypeCategorical,
			contains: []string{"string_value", "category", "reasoning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := worker.buildEvaluationSystemPrompt(tt.dataType)
			for _, substr := range tt.contains {
				assert.Contains(t, prompt, substr)
			}
		})
	}
}

func TestEvalWorker_ParseLLMResponse(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	tests := []struct {
		name        string
		response    string
		expectErr   bool
		expectScore float64
	}{
		{
			name:        "valid json",
			response:    `{"score": 0.85, "reasoning": "Good response"}`,
			expectErr:   false,
			expectScore: 0.85,
		},
		{
			name:        "json with extra text",
			response:    `Here is my evaluation: {"score": 0.7, "reasoning": "Adequate"}`,
			expectErr:   false,
			expectScore: 0.7,
		},
		{
			name:        "score above 1 clamped",
			response:    `{"score": 1.5, "reasoning": "Excellent"}`,
			expectErr:   false,
			expectScore: 1.0,
		},
		{
			name:        "score below 0 clamped",
			response:    `{"score": -0.5, "reasoning": "Poor"}`,
			expectErr:   false,
			expectScore: 0.0,
		},
		{
			name:      "invalid json",
			response:  `not json at all`,
			expectErr: true,
		},
		{
			name:      "empty response",
			response:  ``,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := worker.parseLLMResponse(tt.response)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectScore, result.Score)
		})
	}
}

func TestEvalWorker_ExtractVariables(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	t.Run("with trace only", func(t *testing.T) {
		trace := &domain.Trace{
			ID:     "trace-123",
			Name:   "Test Trace",
			Input:  "Hello",
			Output: "World",
		}

		vars := worker.extractVariables(trace, nil)

		assert.Equal(t, "trace-123", vars["trace_id"])
		assert.Equal(t, "Test Trace", vars["trace_name"])
		assert.Equal(t, "Hello", vars["trace_input"])
		assert.Equal(t, "World", vars["trace_output"])
	})

	t.Run("with trace and observation", func(t *testing.T) {
		trace := &domain.Trace{
			ID:     "trace-123",
			Name:   "Test Trace",
			Input:  "Trace Input",
			Output: "Trace Output",
		}

		observation := &domain.Observation{
			ID:     "obs-456",
			Name:   "Test Observation",
			Type:   domain.ObservationTypeGeneration,
			Input:  "Obs Input",
			Output: "Obs Output",
			Model:  "gpt-4",
		}

		vars := worker.extractVariables(trace, observation)

		// Trace variables
		assert.Equal(t, "trace-123", vars["trace_id"])
		assert.Equal(t, "Test Trace", vars["trace_name"])

		// Observation variables
		assert.Equal(t, "obs-456", vars["observation_id"])
		assert.Equal(t, "Test Observation", vars["observation_name"])
		assert.Equal(t, "GENERATION", vars["observation_type"])
		assert.Equal(t, "Obs Input", vars["input"])
		assert.Equal(t, "Obs Output", vars["output"])
		assert.Equal(t, "gpt-4", vars["model"])
	})

	t.Run("nil inputs", func(t *testing.T) {
		vars := worker.extractVariables(nil, nil)
		assert.NotNil(t, vars)
		assert.Empty(t, vars)
	})
}

func TestEvalWorker_GetTargetContent(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	trace := &domain.Trace{
		ID:     "trace-123",
		Input:  "Trace Input Content",
		Output: "Trace Output Content",
	}

	observation := &domain.Observation{
		ID:     "obs-456",
		Input:  "Observation Input Content",
		Output: "Observation Output Content",
	}

	tests := []struct {
		target   string
		expected string
	}{
		{"trace_input", "Trace Input Content"},
		{"trace_output", "Trace Output Content"},
		{"observation_input", "Observation Input Content"},
		{"observation_output", "Observation Output Content"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			result := worker.getTargetContent(tt.target, trace, observation)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("nil trace", func(t *testing.T) {
		result := worker.getTargetContent("trace_input", nil, observation)
		assert.Empty(t, result)
	})

	t.Run("nil observation", func(t *testing.T) {
		result := worker.getTargetContent("observation_output", trace, nil)
		assert.Empty(t, result)
	})
}

func TestEvaluationPayload_Serialization(t *testing.T) {
	t.Run("with observation ID", func(t *testing.T) {
		payload := EvaluationPayload{
			ProjectID:     uuid.New(),
			EvaluatorID:   uuid.New(),
			TraceID:       "trace-abc",
			ObservationID: "obs-xyz",
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded EvaluationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, payload.ProjectID, decoded.ProjectID)
		assert.Equal(t, payload.EvaluatorID, decoded.EvaluatorID)
		assert.Equal(t, payload.TraceID, decoded.TraceID)
		assert.Equal(t, payload.ObservationID, decoded.ObservationID)
	})

	t.Run("without observation ID", func(t *testing.T) {
		payload := EvaluationPayload{
			ProjectID:   uuid.New(),
			EvaluatorID: uuid.New(),
			TraceID:     "trace-abc",
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded EvaluationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Empty(t, decoded.ObservationID)
	})
}

func TestLLMEvaluationResult_Serialization(t *testing.T) {
	t.Run("numeric result", func(t *testing.T) {
		result := LLMEvaluationResult{
			Score:     0.85,
			Reasoning: "Good quality response",
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var decoded LLMEvaluationResult
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, 0.85, decoded.Score)
		assert.Equal(t, "Good quality response", decoded.Reasoning)
	})

	t.Run("boolean result", func(t *testing.T) {
		passed := true
		result := LLMEvaluationResult{
			Score:     1.0,
			Passed:    &passed,
			Reasoning: "Criteria met",
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var decoded LLMEvaluationResult
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, 1.0, decoded.Score)
		assert.NotNil(t, decoded.Passed)
		assert.True(t, *decoded.Passed)
	})

	t.Run("categorical result", func(t *testing.T) {
		result := LLMEvaluationResult{
			Score:       0.9,
			StringValue: "positive",
			Reasoning:   "Clearly positive sentiment",
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var decoded LLMEvaluationResult
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "positive", decoded.StringValue)
	})
}

func TestBatchEvaluationPayload_Serialization(t *testing.T) {
	payload := BatchEvaluationPayload{
		ProjectID:   uuid.New(),
		EvaluatorID: uuid.New(),
		TraceIDs:    []string{"trace-1", "trace-2", "trace-3"},
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded BatchEvaluationPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.EvaluatorID, decoded.EvaluatorID)
	assert.Equal(t, 3, len(decoded.TraceIDs))
	assert.Equal(t, payload.TraceIDs, decoded.TraceIDs)
}

func TestEvalWorker_RunRuleEvaluation_ContainsRule(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	trace := &domain.Trace{
		ID:        "trace-123",
		ProjectID: uuid.New(),
		Output:    "This is a test response with important content",
	}

	t.Run("contains match", func(t *testing.T) {
		configJSON := `{"rule_type": "contains", "target": "trace_output", "substring": "important"}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Contains Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "contains_check",
		}

		score, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 1.0, *score.Value)
	})

	t.Run("contains no match", func(t *testing.T) {
		configJSON := `{"rule_type": "contains", "target": "trace_output", "substring": "missing"}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Contains Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "contains_check",
		}

		score, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 0.0, *score.Value)
	})
}

func TestEvalWorker_RunRuleEvaluation_NotContainsRule(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	trace := &domain.Trace{
		ID:        "trace-123",
		ProjectID: uuid.New(),
		Output:    "This is a safe response",
	}

	t.Run("not_contains pass", func(t *testing.T) {
		configJSON := `{"rule_type": "not_contains", "target": "trace_output", "substring": "dangerous"}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Not Contains Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "safety_check",
		}

		score, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 1.0, *score.Value)
	})

	t.Run("not_contains fail", func(t *testing.T) {
		configJSON := `{"rule_type": "not_contains", "target": "trace_output", "substring": "safe"}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Not Contains Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "safety_check",
		}

		score, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 0.0, *score.Value)
	})
}

func TestEvalWorker_RunRuleEvaluation_LengthCheckRule(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	trace := &domain.Trace{
		ID:        "trace-123",
		ProjectID: uuid.New(),
		Output:    "Short", // 5 characters
	}

	t.Run("length within range", func(t *testing.T) {
		configJSON := `{"rule_type": "length_check", "target": "trace_output", "min_length": 1, "max_length": 10}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Length Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "length_check",
		}

		score, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 1.0, *score.Value)
	})

	t.Run("length below minimum", func(t *testing.T) {
		configJSON := `{"rule_type": "length_check", "target": "trace_output", "min_length": 10, "max_length": 100}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Length Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "length_check",
		}

		score, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		require.NoError(t, err)
		assert.NotNil(t, score)
		assert.Equal(t, 0.0, *score.Value)
	})
}

func TestEvalWorker_RunRuleEvaluation_InvalidConfig(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	trace := &domain.Trace{
		ID:        "trace-123",
		ProjectID: uuid.New(),
		Output:    "Test output",
	}

	t.Run("empty config", func(t *testing.T) {
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Invalid Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    "",
			ScoreName: "test",
		}

		_, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		assert.Error(t, err)
	})

	t.Run("missing rule_type", func(t *testing.T) {
		configJSON := `{"target": "trace_output"}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Invalid Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "test",
		}

		_, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rule_type")
	})

	t.Run("unsupported rule type", func(t *testing.T) {
		configJSON := `{"rule_type": "unknown_rule"}`
		evaluator := &domain.Evaluator{
			ID:        uuid.New(),
			ProjectID: trace.ProjectID,
			Name:      "Invalid Test",
			Type:      domain.EvaluatorTypeRule,
			Config:    configJSON,
			ScoreName: "test",
		}

		_, err := worker.runRuleEvaluation(context.Background(), evaluator, trace, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported rule type")
	})
}

func TestTaskTypes(t *testing.T) {
	// Verify task type constants are unique
	types := []string{
		TypeEvaluation,
		TypeBatchEvaluation,
		TypeCostCalculation,
		TypeBatchCostCalculation,
		TypeDailyAggregation,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		assert.False(t, seen[typ], "Duplicate task type: %s", typ)
		seen[typ] = true
	}
}

func TestNewEvalWorker(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}

	worker := NewEvalWorker(logger, cfg, nil, nil, nil)

	assert.NotNil(t, worker)
	assert.NotNil(t, worker.logger)
	assert.NotNil(t, worker.config)
	assert.NotNil(t, worker.httpClient)
	assert.Equal(t, 60*time.Second, worker.httpClient.Timeout)
}
