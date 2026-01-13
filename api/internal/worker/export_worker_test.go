package worker

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func TestNewDataExportTask(t *testing.T) {
	payload := &DataExportPayload{
		JobID:     uuid.New(),
		ProjectID: uuid.New(),
		UserID:    uuid.New(),
		Type:      "traces",
		Format:    domain.ExportFormatCSV,
		Filters:   map[string]interface{}{"status": "completed"},
	}

	task, err := NewDataExportTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeDataExport, task.Type())

	// Verify payload
	var decoded DataExportPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.JobID, decoded.JobID)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.Type, decoded.Type)
	assert.Equal(t, payload.Format, decoded.Format)
}

func TestNewDataExportTask_WithDestination(t *testing.T) {
	payload := &DataExportPayload{
		JobID:     uuid.New(),
		ProjectID: uuid.New(),
		UserID:    uuid.New(),
		Type:      "observations",
		Format:    domain.ExportFormatJSON,
		Destination: &ExportDestination{
			Type:   domain.DestinationTypeS3,
			Bucket: "my-bucket",
			Path:   "/exports/data.json",
			Config: map[string]string{"region": "us-east-1"},
		},
	}

	task, err := NewDataExportTask(payload)
	require.NoError(t, err)

	var decoded DataExportPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.NotNil(t, decoded.Destination)
	assert.Equal(t, domain.DestinationTypeS3, decoded.Destination.Type)
	assert.Equal(t, "my-bucket", decoded.Destination.Bucket)
	assert.Equal(t, "/exports/data.json", decoded.Destination.Path)
}

func TestNewDatasetExportTask(t *testing.T) {
	payload := &DatasetExportPayload{
		JobID:       uuid.New(),
		ProjectID:   uuid.New(),
		DatasetID:   uuid.New(),
		UserID:      uuid.New(),
		Format:      domain.ExportFormatCSV,
		IncludeRuns: true,
	}

	task, err := NewDatasetExportTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeDatasetExport, task.Type())

	// Verify payload
	var decoded DatasetExportPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.JobID, decoded.JobID)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.DatasetID, decoded.DatasetID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.Format, decoded.Format)
	assert.True(t, decoded.IncludeRuns)
}

func TestNewDatasetExportTask_WithoutRuns(t *testing.T) {
	payload := &DatasetExportPayload{
		JobID:       uuid.New(),
		ProjectID:   uuid.New(),
		DatasetID:   uuid.New(),
		UserID:      uuid.New(),
		Format:      domain.ExportFormatJSON,
		IncludeRuns: false,
	}

	task, err := NewDatasetExportTask(payload)
	require.NoError(t, err)

	var decoded DatasetExportPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.False(t, decoded.IncludeRuns)
}

func TestDataExportPayload_Serialization(t *testing.T) {
	t.Run("traces export", func(t *testing.T) {
		payload := DataExportPayload{
			JobID:     uuid.New(),
			ProjectID: uuid.New(),
			UserID:    uuid.New(),
			Type:      "traces",
			Format:    domain.ExportFormatCSV,
			Filters: map[string]interface{}{
				"from_date": "2024-01-01",
				"to_date":   "2024-12-31",
			},
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded DataExportPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, payload.Type, decoded.Type)
		assert.Equal(t, payload.Format, decoded.Format)
		assert.NotNil(t, decoded.Filters)
	})

	t.Run("scores export", func(t *testing.T) {
		payload := DataExportPayload{
			JobID:     uuid.New(),
			ProjectID: uuid.New(),
			UserID:    uuid.New(),
			Type:      "scores",
			Format:    domain.ExportFormatJSON,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded DataExportPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "scores", decoded.Type)
		assert.Equal(t, domain.ExportFormatJSON, decoded.Format)
	})
}

func TestDatasetExportPayload_Serialization(t *testing.T) {
	t.Run("with destination", func(t *testing.T) {
		payload := DatasetExportPayload{
			JobID:       uuid.New(),
			ProjectID:   uuid.New(),
			DatasetID:   uuid.New(),
			UserID:      uuid.New(),
			Format:      domain.ExportFormatCSV,
			IncludeRuns: true,
			Destination: &ExportDestination{
				Type:   domain.DestinationTypeS3,
				Bucket: "test-bucket",
				Path:   "/datasets/export.csv",
			},
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded DatasetExportPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.NotNil(t, decoded.Destination)
		assert.Equal(t, "test-bucket", decoded.Destination.Bucket)
	})

	t.Run("without destination", func(t *testing.T) {
		payload := DatasetExportPayload{
			JobID:       uuid.New(),
			ProjectID:   uuid.New(),
			DatasetID:   uuid.New(),
			UserID:      uuid.New(),
			Format:      domain.ExportFormatJSON,
			IncludeRuns: false,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded DatasetExportPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Nil(t, decoded.Destination)
	})
}

func TestExportDestination_Serialization(t *testing.T) {
	destination := ExportDestination{
		Type:   domain.DestinationTypeS3,
		Bucket: "my-exports",
		Path:   "/data/export-2024.csv",
		Config: map[string]string{
			"region":     "us-west-2",
			"access_key": "AKIAIOSFODNN7EXAMPLE",
		},
	}

	data, err := json.Marshal(destination)
	require.NoError(t, err)

	var decoded ExportDestination
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, destination.Type, decoded.Type)
	assert.Equal(t, destination.Bucket, decoded.Bucket)
	assert.Equal(t, destination.Path, decoded.Path)
	assert.Equal(t, "us-west-2", decoded.Config["region"])
}

func TestNewExportWorker(t *testing.T) {
	logger := zap.NewNop()

	worker := NewExportWorker(logger, nil, nil, nil, nil, "test-bucket")

	assert.NotNil(t, worker)
	assert.NotNil(t, worker.logger)
	assert.Equal(t, "test-bucket", worker.bucket)
}

func TestExportTaskTypes(t *testing.T) {
	// Verify task type constants are unique
	types := []string{
		TypeDataExport,
		TypeDatasetExport,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		assert.False(t, seen[typ], "Duplicate task type: %s", typ)
		seen[typ] = true
	}

	// Verify expected values
	assert.Equal(t, "export:data", TypeDataExport)
	assert.Equal(t, "export:dataset", TypeDatasetExport)
}

func TestDataExportPayload_AllTypes(t *testing.T) {
	types := []string{"traces", "observations", "scores"}

	for _, exportType := range types {
		t.Run(exportType, func(t *testing.T) {
			payload := DataExportPayload{
				JobID:     uuid.New(),
				ProjectID: uuid.New(),
				UserID:    uuid.New(),
				Type:      exportType,
				Format:    domain.ExportFormatCSV,
			}

			data, err := json.Marshal(payload)
			require.NoError(t, err)

			var decoded DataExportPayload
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, exportType, decoded.Type)
		})
	}
}

func TestDataExportPayload_AllFormats(t *testing.T) {
	formats := []domain.ExportFormat{domain.ExportFormatCSV, domain.ExportFormatJSON}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			payload := DataExportPayload{
				JobID:     uuid.New(),
				ProjectID: uuid.New(),
				UserID:    uuid.New(),
				Type:      "traces",
				Format:    format,
			}

			data, err := json.Marshal(payload)
			require.NoError(t, err)

			var decoded DataExportPayload
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, format, decoded.Format)
		})
	}
}
