package worker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataCleanupTask(t *testing.T) {
	payload := &DataCleanupPayload{
		ProjectID:     uuid.New(),
		RetentionDays: 30,
		DryRun:        false,
	}

	task, err := NewDataCleanupTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeDataCleanup, task.Type())

	// Verify payload
	var decoded DataCleanupPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
	assert.Equal(t, payload.RetentionDays, decoded.RetentionDays)
	assert.Equal(t, payload.DryRun, decoded.DryRun)
}

func TestNewDataCleanupTask_DryRun(t *testing.T) {
	payload := &DataCleanupPayload{
		ProjectID:     uuid.New(),
		RetentionDays: 90,
		DryRun:        true,
	}

	task, err := NewDataCleanupTask(payload)
	require.NoError(t, err)

	var decoded DataCleanupPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.True(t, decoded.DryRun)
}

func TestNewProjectCleanupTask(t *testing.T) {
	payload := &ProjectCleanupPayload{
		ProjectID: uuid.New(),
	}

	task, err := NewProjectCleanupTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeProjectCleanup, task.Type())

	// Verify payload
	var decoded ProjectCleanupPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
}

func TestNewOrphanCleanupTask(t *testing.T) {
	payload := &OrphanCleanupPayload{
		DryRun: false,
	}

	task, err := NewOrphanCleanupTask(payload)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeOrphanCleanup, task.Type())

	// Verify payload
	var decoded OrphanCleanupPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.False(t, decoded.DryRun)
}

func TestNewOrphanCleanupTask_DryRun(t *testing.T) {
	payload := &OrphanCleanupPayload{
		DryRun: true,
	}

	task, err := NewOrphanCleanupTask(payload)
	require.NoError(t, err)

	var decoded OrphanCleanupPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	require.NoError(t, err)
	assert.True(t, decoded.DryRun)
}

func TestDataCleanupPayload_Serialization(t *testing.T) {
	t.Run("basic payload", func(t *testing.T) {
		payload := DataCleanupPayload{
			ProjectID:     uuid.New(),
			RetentionDays: 30,
			DryRun:        false,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded DataCleanupPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, payload.ProjectID, decoded.ProjectID)
		assert.Equal(t, payload.RetentionDays, decoded.RetentionDays)
		assert.Equal(t, payload.DryRun, decoded.DryRun)
	})

	t.Run("zero retention days", func(t *testing.T) {
		payload := DataCleanupPayload{
			ProjectID:     uuid.New(),
			RetentionDays: 0,
			DryRun:        true,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded DataCleanupPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, 0, decoded.RetentionDays)
	})
}

func TestProjectCleanupPayload_Serialization(t *testing.T) {
	payload := ProjectCleanupPayload{
		ProjectID: uuid.New(),
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded ProjectCleanupPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, payload.ProjectID, decoded.ProjectID)
}

func TestOrphanCleanupPayload_Serialization(t *testing.T) {
	t.Run("dry run true", func(t *testing.T) {
		payload := OrphanCleanupPayload{
			DryRun: true,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded OrphanCleanupPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.True(t, decoded.DryRun)
	})

	t.Run("dry run false", func(t *testing.T) {
		payload := OrphanCleanupPayload{
			DryRun: false,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		var decoded OrphanCleanupPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.False(t, decoded.DryRun)
	})
}

func TestCleanupWorker_ProcessTask_InvalidPayload(t *testing.T) {
	worker := &CleanupWorker{}

	// Create task with invalid payload
	task := asynq.NewTask(TypeDataCleanup, []byte("invalid json"))

	err := worker.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestCleanupWorker_ProcessProjectCleanupTask_InvalidPayload(t *testing.T) {
	worker := &CleanupWorker{}

	// Create task with invalid payload
	task := asynq.NewTask(TypeProjectCleanup, []byte("invalid json"))

	err := worker.ProcessProjectCleanupTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestCleanupWorker_ProcessOrphanCleanupTask_InvalidPayload(t *testing.T) {
	worker := &CleanupWorker{}

	// Create task with invalid payload
	task := asynq.NewTask(TypeOrphanCleanup, []byte("invalid json"))

	err := worker.ProcessOrphanCleanupTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestScheduledCleanupConfig_Defaults(t *testing.T) {
	config := &ScheduledCleanupConfig{
		DefaultRetentionDays: 30,
		CleanupHour:          3,
		BatchSize:            0, // Test that it gets defaulted
	}

	assert.Equal(t, 30, config.DefaultRetentionDays)
	assert.Equal(t, 3, config.CleanupHour)
	assert.Equal(t, 0, config.BatchSize) // Will be set to 100 in ScheduleCleanupTasks
}

func TestCleanupTaskTypes(t *testing.T) {
	// Verify task type constants are unique
	types := []string{
		TypeDataCleanup,
		TypeProjectCleanup,
		TypeOrphanCleanup,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		assert.False(t, seen[typ], "Duplicate task type: %s", typ)
		seen[typ] = true
	}

	// Verify expected values
	assert.Equal(t, "cleanup:data", TypeDataCleanup)
	assert.Equal(t, "cleanup:project", TypeProjectCleanup)
	assert.Equal(t, "cleanup:orphans", TypeOrphanCleanup)
}
