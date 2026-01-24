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

func TestNewOTelExportTask(t *testing.T) {
	t.Parallel()

	payload := OTelExportPayload{
		ProjectID:  uuid.New(),
		ExporterID: uuid.New(),
		TraceID:    uuid.New(),
	}

	task, err := NewOTelExportTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeOTelExport, task.Type())

	// Verify payload roundtrips correctly
	var decoded OTelExportPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.ExporterID, decoded.ExporterID)
	assert.Equal(t, payload.TraceID, decoded.TraceID)
}

func TestNewOTelBatchExportTask(t *testing.T) {
	t.Parallel()

	payload := OTelBatchExportPayload{
		ProjectID:  uuid.New(),
		ExporterID: uuid.New(),
		Spans:      nil,
	}

	task, err := NewOTelBatchExportTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeOTelBatchExport, task.Type())

	var decoded OTelBatchExportPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.ExporterID, decoded.ExporterID)
}

func TestNewOTelHealthCheckTask(t *testing.T) {
	t.Parallel()

	payload := OTelHealthCheckPayload{
		ExporterID: uuid.New(),
	}

	task, err := NewOTelHealthCheckTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeOTelExporterHealthCheck, task.Type())

	var decoded OTelHealthCheckPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ExporterID, decoded.ExporterID)
}

func TestOTelWorker_HandleOTelExport_InvalidPayload(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)

	task := asynq.NewTask(TypeOTelExport, []byte("invalid json"))
	err := worker.HandleOTelExport(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestOTelWorker_HandleOTelExport_ValidPayload(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)

	payload := OTelExportPayload{
		ProjectID:  uuid.New(),
		ExporterID: uuid.New(),
		TraceID:    uuid.New(),
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeOTelExport, data)
	err = worker.HandleOTelExport(context.Background(), task)
	assert.NoError(t, err)
}

func TestOTelWorker_HandleOTelBatchExport_InvalidPayload(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)

	task := asynq.NewTask(TypeOTelBatchExport, []byte("not json"))
	err := worker.HandleOTelBatchExport(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestOTelWorker_HandleOTelBatchExport_ValidPayload(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)

	payload := OTelBatchExportPayload{
		ProjectID:  uuid.New(),
		ExporterID: uuid.New(),
		Spans:      nil,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeOTelBatchExport, data)
	err = worker.HandleOTelBatchExport(context.Background(), task)
	assert.NoError(t, err)
}

func TestOTelWorker_HandleOTelHealthCheck_InvalidPayload(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)

	task := asynq.NewTask(TypeOTelExporterHealthCheck, []byte("{bad"))
	err := worker.HandleOTelHealthCheck(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestOTelWorker_HandleOTelHealthCheck_ValidPayload(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)

	payload := OTelHealthCheckPayload{
		ExporterID: uuid.New(),
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeOTelExporterHealthCheck, data)
	err = worker.HandleOTelHealthCheck(context.Background(), task)
	assert.NoError(t, err)
}

func TestOTelWorker_RegisterHandlers(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	worker := NewOTelWorker(logger, nil)
	mux := asynq.NewServeMux()

	// Should not panic
	worker.RegisterHandlers(mux)
}

func TestOTelTaskTypes(t *testing.T) {
	t.Parallel()

	// Verify task type constants are unique
	types := map[string]bool{
		TypeOTelExport:             true,
		TypeOTelBatchExport:        true,
		TypeOTelExporterHealthCheck: true,
	}
	assert.Len(t, types, 3, "Task types should be unique")
}
