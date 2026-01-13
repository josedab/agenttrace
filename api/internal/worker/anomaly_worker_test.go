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

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewAnomalyDetectionTask(t *testing.T) {
	traceID := uuid.New()
	payload := AnomalyDetectionPayload{
		ProjectID:    uuid.New(),
		RuleID:       uuid.New(),
		CurrentValue: 150.0,
		TraceID:      &traceID,
		TraceName:    "test-trace",
		MetricType:   domain.AnomalyTypeLatency,
		Metadata:     map[string]string{"environment": "production"},
	}

	task, err := NewAnomalyDetectionTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeAnomalyDetection, task.Type())

	// Verify payload
	var decoded AnomalyDetectionPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.RuleID, decoded.RuleID)
	assert.Equal(t, payload.CurrentValue, decoded.CurrentValue)
	assert.Equal(t, payload.TraceName, decoded.TraceName)
	assert.Equal(t, payload.MetricType, decoded.MetricType)
}

func TestNewAnomalyAlertTask(t *testing.T) {
	payload := AnomalyAlertPayload{
		AnomalyID:  uuid.New(),
		ProjectID:  uuid.New(),
		RuleID:     uuid.New(),
		WebhookIDs: []uuid.UUID{uuid.New(), uuid.New()},
		AlertTitle: "High Latency Detected",
		AlertBody:  "Latency exceeded threshold",
		Severity:   domain.AnomalySeverityHigh,
	}

	task, err := NewAnomalyAlertTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeAnomalyAlert, task.Type())

	// Verify payload
	var decoded AnomalyAlertPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.AnomalyID, decoded.AnomalyID)
	assert.Equal(t, payload.AlertTitle, decoded.AlertTitle)
	assert.Equal(t, payload.Severity, decoded.Severity)
	assert.Len(t, decoded.WebhookIDs, 2)
}

func TestNewAnomalyCleanupTask(t *testing.T) {
	payload := AnomalyCleanupPayload{
		ProjectID:     uuid.New(),
		RetentionDays: 30,
	}

	task, err := NewAnomalyCleanupTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeAnomalyCleanup, task.Type())

	// Verify payload
	var decoded AnomalyCleanupPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.RetentionDays, decoded.RetentionDays)
}

func TestNewAnomalyScheduledScanTask(t *testing.T) {
	payload := AnomalyScheduledScanPayload{
		ProjectID: uuid.New(),
		RuleID:    uuid.New(),
	}

	task, err := NewAnomalyScheduledScanTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeAnomalyScheduledScan, task.Type())

	// Verify payload
	var decoded AnomalyScheduledScanPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.RuleID, decoded.RuleID)
}

func TestAnomalyDetectionPayload_Serialization(t *testing.T) {
	t.Run("with trace info", func(t *testing.T) {
		traceID := uuid.New()
		payload := AnomalyDetectionPayload{
			ProjectID:    uuid.New(),
			RuleID:       uuid.New(),
			CurrentValue: 100.5,
			TraceID:      &traceID,
			TraceName:    "api-call",
			MetricType:   domain.AnomalyTypeErrorRate,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded AnomalyDetectionPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.NotNil(t, decoded.TraceID)
		assert.Equal(t, traceID, *decoded.TraceID)
	})

	t.Run("without trace info", func(t *testing.T) {
		payload := AnomalyDetectionPayload{
			ProjectID:    uuid.New(),
			RuleID:       uuid.New(),
			CurrentValue: 50.0,
			MetricType:   domain.AnomalyTypeCost,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded AnomalyDetectionPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Nil(t, decoded.TraceID)
	})

	t.Run("with span info", func(t *testing.T) {
		spanID := uuid.New()
		payload := AnomalyDetectionPayload{
			ProjectID:    uuid.New(),
			RuleID:       uuid.New(),
			CurrentValue: 200.0,
			SpanID:       &spanID,
			SpanName:     "database-query",
			MetricType:   domain.AnomalyTypeLatency,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded AnomalyDetectionPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.NotNil(t, decoded.SpanID)
		assert.Equal(t, spanID, *decoded.SpanID)
		assert.Equal(t, "database-query", decoded.SpanName)
	})
}

func TestAnomalyAlertPayload_Serialization(t *testing.T) {
	t.Run("with multiple webhooks", func(t *testing.T) {
		webhooks := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		payload := AnomalyAlertPayload{
			AnomalyID:  uuid.New(),
			ProjectID:  uuid.New(),
			RuleID:     uuid.New(),
			WebhookIDs: webhooks,
			AlertTitle: "Critical Alert",
			AlertBody:  "Service is experiencing high error rates",
			Severity:   domain.AnomalySeverityCritical,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded AnomalyAlertPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Len(t, decoded.WebhookIDs, 3)
		assert.Equal(t, domain.AnomalySeverityCritical, decoded.Severity)
	})

	t.Run("with empty webhooks", func(t *testing.T) {
		payload := AnomalyAlertPayload{
			AnomalyID:  uuid.New(),
			ProjectID:  uuid.New(),
			RuleID:     uuid.New(),
			WebhookIDs: []uuid.UUID{},
			AlertTitle: "Test Alert",
			Severity:   domain.AnomalySeverityLow,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded AnomalyAlertPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Empty(t, decoded.WebhookIDs)
	})
}

func TestAnomalyWorker_HandleAnomalyDetection_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewAnomalyWorker(logger, nil, nil)

	task := asynq.NewTask(TypeAnomalyDetection, []byte("invalid json"))

	err := worker.HandleAnomalyDetection(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestAnomalyWorker_HandleAnomalyAlert_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewAnomalyWorker(logger, nil, nil)

	task := asynq.NewTask(TypeAnomalyAlert, []byte("invalid json"))

	err := worker.HandleAnomalyAlert(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestAnomalyWorker_HandleAnomalyCleanup_InvalidPayload(t *testing.T) {
	logger := zap.NewNop()
	worker := NewAnomalyWorker(logger, nil, nil)

	task := asynq.NewTask(TypeAnomalyCleanup, []byte("invalid json"))

	err := worker.HandleAnomalyCleanup(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNewAnomalyWorker(t *testing.T) {
	logger := zap.NewNop()

	worker := NewAnomalyWorker(logger, nil, nil)

	assert.NotNil(t, worker)
	assert.NotNil(t, worker.logger)
}

func TestAnomalyTaskTypes(t *testing.T) {
	// Verify task type constants are unique
	types := []string{
		TypeAnomalyDetection,
		TypeAnomalyAlert,
		TypeAnomalyCleanup,
		TypeAnomalyScheduledScan,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		assert.False(t, seen[typ], "Duplicate task type: %s", typ)
		seen[typ] = true
	}

	// Verify expected values
	assert.Equal(t, "anomaly:detect", TypeAnomalyDetection)
	assert.Equal(t, "anomaly:alert", TypeAnomalyAlert)
	assert.Equal(t, "anomaly:cleanup", TypeAnomalyCleanup)
	assert.Equal(t, "anomaly:scheduled_scan", TypeAnomalyScheduledScan)
}

func TestAnomalyCleanupPayload_Serialization(t *testing.T) {
	payload := AnomalyCleanupPayload{
		ProjectID:     uuid.New(),
		RetentionDays: 90,
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded AnomalyCleanupPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, 90, decoded.RetentionDays)
}

func TestAnomalyScheduledScanPayload_Serialization(t *testing.T) {
	payload := AnomalyScheduledScanPayload{
		ProjectID: uuid.New(),
		RuleID:    uuid.New(),
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded AnomalyScheduledScanPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.RuleID, decoded.RuleID)
}
