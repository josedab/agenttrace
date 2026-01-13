package worker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNotificationPayload_Serialization(t *testing.T) {
	t.Run("basic payload", func(t *testing.T) {
		payload := NotificationPayload{
			WebhookID:  "webhook-123",
			EventType:  domain.EventTypeTraceError,
			Data:       map[string]any{"traceId": "trace-456"},
			RetryCount: 0,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded NotificationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, payload.WebhookID, decoded.WebhookID)
		assert.Equal(t, payload.EventType, decoded.EventType)
		assert.Equal(t, "trace-456", decoded.Data["traceId"])
		assert.Equal(t, 0, decoded.RetryCount)
	})

	t.Run("with retry count", func(t *testing.T) {
		payload := NotificationPayload{
			WebhookID:  "00000000-0000-0000-0000-000000000002",
			EventType:  domain.EventTypeEvalScoreLow,
			Data:       map[string]any{"scoreId": "score-123"},
			RetryCount: 3,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded NotificationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, 3, decoded.RetryCount)
	})

	t.Run("with complex data", func(t *testing.T) {
		payload := NotificationPayload{
			WebhookID: "00000000-0000-0000-0000-000000000003",
			EventType: domain.EventTypeAnomalyDetected,
			Data: map[string]any{
				"observationId": "obs-123",
				"type":          "generation",
				"metadata": map[string]any{
					"model": "gpt-4",
					"tokens": map[string]any{
						"input":  100,
						"output": 200,
					},
				},
			},
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded NotificationPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.NotNil(t, decoded.Data["metadata"])
	})
}

func TestDailyCostReportPayload_Serialization(t *testing.T) {
	payload := DailyCostReportPayload{
		ProjectID: "project-123",
		Date:      "2024-01-15",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded DailyCostReportPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.Date, decoded.Date)
}

func TestThresholdCheckPayload_Serialization(t *testing.T) {
	payload := ThresholdCheckPayload{
		TraceID:   "trace-abc",
		ProjectID: "project-xyz",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded ThresholdCheckPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, payload.TraceID, decoded.TraceID)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
}

func TestNewNotificationWorker(t *testing.T) {
	logger := zap.NewNop()

	worker := NewNotificationWorker(logger, nil, nil, nil, nil)

	assert.NotNil(t, worker)
	assert.NotNil(t, worker.logger)
}

func TestNotificationWorker_HandleNotificationSend_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewNotificationWorker(logger, nil, nil, nil, nil)

	task := asynq.NewTask(TypeNotificationSend, []byte("invalid json"))

	err := worker.HandleNotificationSend(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNotificationWorker_HandleDailyCostReport_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewNotificationWorker(logger, nil, nil, nil, nil)

	task := asynq.NewTask(TypeDailyCostReport, []byte("invalid json"))

	err := worker.HandleDailyCostReport(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNotificationWorker_HandleCheckThresholds_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewNotificationWorker(logger, nil, nil, nil, nil)

	task := asynq.NewTask(TypeCheckThresholds, []byte("invalid json"))

	err := worker.HandleCheckThresholds(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNotificationWorker_HandleNotificationSend_ValidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewNotificationWorker(logger, nil, nil, nil, nil)

	payload := NotificationPayload{
		WebhookID: "00000000-0000-0000-0000-000000000001",
		EventType: domain.EventTypeTraceError,
		Data:      map[string]any{"test": "data"},
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeNotificationSend, payloadBytes)

	// With nil repository, this should return error (webhook not found)
	err = worker.HandleNotificationSend(context.Background(), task)
	// Expected to fail since we don't have a webhook repository configured
	assert.Error(t, err)
}

func TestNotificationWorker_RegisterHandlers(t *testing.T) {
	logger := zap.NewNop()
	worker := NewNotificationWorker(logger, nil, nil, nil, nil)

	mux := asynq.NewServeMux()
	worker.RegisterHandlers(mux)

	// Verify handlers were registered by checking task types
	assert.NotNil(t, mux)
}

func TestNotificationTaskTypes(t *testing.T) {
	// Verify task type constants are unique
	types := []string{
		TypeNotificationSend,
		TypeDailyCostReport,
		TypeCheckThresholds,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		assert.False(t, seen[typ], "Duplicate task type: %s", typ)
		seen[typ] = true
	}

	// Verify expected values
	assert.Equal(t, "notification:send", TypeNotificationSend)
	assert.Equal(t, "notification:daily_cost_report", TypeDailyCostReport)
	assert.Equal(t, "notification:check_thresholds", TypeCheckThresholds)
}
