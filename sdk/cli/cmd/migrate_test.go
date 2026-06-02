package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateCommand(t *testing.T) {
	t.Run("has expected flags", func(t *testing.T) {
		f := migrateCmd.PersistentFlags()
		assert.NotNil(t, f.Lookup("source"))
		assert.NotNil(t, f.Lookup("source-dsn"))
		assert.NotNil(t, f.Lookup("source-file"))
		assert.NotNil(t, f.Lookup("target-host"))
		assert.NotNil(t, f.Lookup("dry-run"))
		assert.NotNil(t, f.Lookup("incremental"))
		assert.NotNil(t, f.Lookup("batch-size"))
	})

	t.Run("has validate subcommand", func(t *testing.T) {
		found := false
		for _, c := range migrateCmd.Commands() {
			if c.Name() == "validate" {
				found = true
				break
			}
		}
		assert.True(t, found, "migrate should have validate subcommand")
	})

	t.Run("source defaults to langfuse", func(t *testing.T) {
		f := migrateCmd.PersistentFlags().Lookup("source")
		assert.Equal(t, "langfuse", f.DefValue)
	})
}

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "postgres://user:password@host:5432/db",
			expected: "postgres://user:****@host:5432/db",
		},
		{
			input:    "postgres://admin:s3cr3t@localhost:5432/langfuse",
			expected: "postgres://admin:****@localhost:5432/langfuse",
		},
		{
			input:    "no-scheme-dsn",
			expected: "no-scheme-dsn",
		},
		{
			input:    "postgres://user@host/db",
			expected: "postgres://user@host/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, maskDSN(tt.input))
		})
	}
}

func TestGetMigrateHost(t *testing.T) {
	t.Run("uses flag when set", func(t *testing.T) {
		original := migrateHost
		defer func() { migrateHost = original }()

		migrateHost = "http://custom:8080"
		assert.Equal(t, "http://custom:8080", getMigrateHost())
	})

	t.Run("falls back to env var", func(t *testing.T) {
		original := migrateHost
		defer func() { migrateHost = original }()

		migrateHost = ""
		t.Setenv("AGENTTRACE_API_URL", "http://env-host:8080")
		assert.Equal(t, "http://env-host:8080", getMigrateHost())
	})
}

func TestRunMigrate_RequiresAPIKey(t *testing.T) {
	original := apiKey
	defer func() { apiKey = original }()

	apiKey = ""
	t.Setenv("AGENTTRACE_API_KEY", "")

	err := runMigrate(migrateCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key required")
}

func TestRunMigrate_RequiresSourceFile(t *testing.T) {
	original := apiKey
	origDSN := migrateSourceDSN
	origFile := migrateSourceFile
	defer func() {
		apiKey = original
		migrateSourceDSN = origDSN
		migrateSourceFile = origFile
	}()

	apiKey = "test-key"
	migrateSourceDSN = ""
	migrateSourceFile = ""

	err := runMigrate(migrateCmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--source-file is required")
}

func TestReadLangfuseExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "langfuse.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		"traces":[{"id":"trace-1","timestamp":"2026-07-25T10:00:00Z"}],
		"scores":[{"id":"score-1","traceId":"trace-1","name":"quality","value":0.9}]
	}`), 0o600))

	export, fingerprint, err := readLangfuseExport(path)

	require.NoError(t, err)
	assert.Len(t, export.Traces, 1)
	assert.Len(t, export.Scores, 1)
	assert.Len(t, fingerprint, 64)
}

func TestReadLangfuseExportRejectsEmptyOrInvalid(t *testing.T) {
	directory := t.TempDir()
	invalid := filepath.Join(directory, "invalid.json")
	require.NoError(t, os.WriteFile(invalid, []byte(`{`), 0o600))
	_, _, err := readLangfuseExport(invalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")

	empty := filepath.Join(directory, "empty.json")
	require.NoError(t, os.WriteFile(empty, []byte(`{"traces":[]}`), 0o600))
	_, _, err = readLangfuseExport(empty)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no supported records")
}

func TestRunLangfuseJSONMigrationBatchesAndResumes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "langfuse.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		"traces":[
			{"id":"trace-1","timestamp":"2026-07-25T10:00:00Z"},
			{"id":"trace-2","timestamp":"2026-07-25T10:01:00Z"},
			{"id":"trace-3","timestamp":"2026-07-25T10:02:00Z"}
		]
	}`), 0o600))

	requestCount := 0
	var jobID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, "Bearer local-key", r.Header.Get("Authorization"))
		var payload struct {
			JobID      string `json:"jobId"`
			FinalBatch bool   `json:"finalBatch"`
			Records    struct {
				Traces []json.RawMessage `json:"traces"`
			} `json:"records"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		if jobID == "" {
			jobID = payload.JobID
		} else {
			assert.Equal(t, jobID, payload.JobID)
		}
		if requestCount == 1 {
			assert.False(t, payload.FinalBatch)
			assert.Len(t, payload.Records.Traces, 2)
		} else {
			assert.True(t, payload.FinalBatch)
			assert.Len(t, payload.Records.Traces, 1)
		}
		w.Header().Set("Content-Type", "application/json")
		status := "RUNNING"
		if payload.FinalBatch {
			status = "COMPLETED"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": status,
			"progress": map[string]any{
				"totalItems":     3,
				"processedItems": requestCount * 2,
				"tracesMigrated": requestCount * 2,
			},
			"errors": []string{},
		})
	}))
	defer server.Close()

	originalBatchSize := migrateBatchSize
	originalDryRun := migrateDryRun
	defer func() {
		migrateBatchSize = originalBatchSize
		migrateDryRun = originalDryRun
	}()
	migrateBatchSize = 2
	migrateDryRun = false

	err := runLangfuseJSONMigration(path, server.URL, "local-key")

	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
	assert.NotEmpty(t, jobID)
}
