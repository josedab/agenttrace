package agenttrace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckpointType(t *testing.T) {
	t.Run("all checkpoint types exist", func(t *testing.T) {
		types := []CheckpointType{
			CheckpointTypeManual,
			CheckpointTypeAuto,
			CheckpointTypeToolCall,
			CheckpointTypeError,
			CheckpointTypeMilestone,
			CheckpointTypeRestore,
		}

		expected := []string{"manual", "auto", "tool_call", "error", "milestone", "restore"}

		for i, cp := range types {
			if string(cp) != expected[i] {
				t.Errorf("expected %s, got %s", expected[i], cp)
			}
		}
	})
}

func TestTraceCheckpoint(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("creates checkpoint with minimal options", func(t *testing.T) {
		cp := trace.Checkpoint(CheckpointOptions{
			Name: "test-checkpoint",
		})

		if cp.Name != "test-checkpoint" {
			t.Errorf("expected name to be 'test-checkpoint', got '%s'", cp.Name)
		}
		if cp.Type != CheckpointTypeManual {
			t.Errorf("expected type to be 'manual', got '%s'", cp.Type)
		}
		if cp.TraceID != trace.id {
			t.Error("expected checkpoint to have trace's ID")
		}
		if cp.ID == "" {
			t.Error("expected checkpoint ID to be generated")
		}
		if cp.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("creates checkpoint with custom type", func(t *testing.T) {
		cp := trace.Checkpoint(CheckpointOptions{
			Name: "milestone-checkpoint",
			Type: CheckpointTypeMilestone,
		})

		if cp.Type != CheckpointTypeMilestone {
			t.Errorf("expected type to be 'milestone', got '%s'", cp.Type)
		}
	})

	t.Run("creates checkpoint with observation ID", func(t *testing.T) {
		cp := trace.Checkpoint(CheckpointOptions{
			Name:          "obs-checkpoint",
			ObservationID: "obs-456",
		})

		if cp.ObservationID != "obs-456" {
			t.Errorf("expected observationID to be 'obs-456', got '%s'", cp.ObservationID)
		}
	})

	t.Run("creates checkpoint with files", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.go")
		if err := os.WriteFile(tmpFile, []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}

		cp := trace.Checkpoint(CheckpointOptions{
			Name:           "file-checkpoint",
			Files:          []string{tmpFile},
			IncludeGitInfo: false,
		})

		if cp.TotalFiles != 1 {
			t.Errorf("expected TotalFiles to be 1, got %d", cp.TotalFiles)
		}
		if cp.TotalSizeBytes == 0 {
			t.Error("expected TotalSizeBytes to be > 0")
		}
		if len(cp.FilesChanged) != 1 {
			t.Errorf("expected 1 file in FilesChanged, got %d", len(cp.FilesChanged))
		}
	})

	t.Run("handles nonexistent files gracefully", func(t *testing.T) {
		cp := trace.Checkpoint(CheckpointOptions{
			Name:           "no-file-checkpoint",
			Files:          []string{"/nonexistent/file.go"},
			IncludeGitInfo: false,
		})

		if cp.TotalFiles != 1 {
			t.Errorf("expected TotalFiles to be 1, got %d", cp.TotalFiles)
		}
		if cp.TotalSizeBytes != 0 {
			t.Errorf("expected TotalSizeBytes to be 0 for nonexistent file, got %d", cp.TotalSizeBytes)
		}
	})

	t.Run("creates checkpoint with empty files list", func(t *testing.T) {
		cp := trace.Checkpoint(CheckpointOptions{
			Name:           "empty-files-checkpoint",
			Files:          []string{},
			IncludeGitInfo: false,
		})

		if cp.TotalFiles != 0 {
			t.Errorf("expected TotalFiles to be 0, got %d", cp.TotalFiles)
		}
		if len(cp.FilesChanged) != 0 {
			t.Errorf("expected empty FilesChanged, got %d files", len(cp.FilesChanged))
		}
	})
}

func TestCheckpointDisabledClient(t *testing.T) {
	enabled := false
	client := New(Config{
		APIKey:  "test-api-key",
		Enabled: &enabled,
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	cp := trace.Checkpoint(CheckpointOptions{
		Name: "disabled-checkpoint",
	})

	// Should still return checkpoint info
	if cp.Name != "disabled-checkpoint" {
		t.Errorf("expected checkpoint name even when disabled")
	}
	if cp.ID == "" {
		t.Error("expected checkpoint ID even when disabled")
	}
}

func TestCheckpointInfo(t *testing.T) {
	t.Run("has all expected fields", func(t *testing.T) {
		info := CheckpointInfo{
			ID:             "cp-123",
			Name:           "test",
			Type:           CheckpointTypeManual,
			TraceID:        "trace-456",
			ObservationID:  "obs-789",
			GitCommitSha:   "abc123",
			GitBranch:      "main",
			FilesChanged:   []string{"file1.go", "file2.go"},
			TotalFiles:     2,
			TotalSizeBytes: 1024,
			CreatedAt:      time.Now(),
		}

		if info.ID != "cp-123" {
			t.Error("ID not set correctly")
		}
		if info.Name != "test" {
			t.Error("Name not set correctly")
		}
		if info.Type != CheckpointTypeManual {
			t.Error("Type not set correctly")
		}
		if info.TraceID != "trace-456" {
			t.Error("TraceID not set correctly")
		}
		if info.ObservationID != "obs-789" {
			t.Error("ObservationID not set correctly")
		}
		if info.GitCommitSha != "abc123" {
			t.Error("GitCommitSha not set correctly")
		}
		if info.GitBranch != "main" {
			t.Error("GitBranch not set correctly")
		}
		if len(info.FilesChanged) != 2 {
			t.Error("FilesChanged not set correctly")
		}
		if info.TotalFiles != 2 {
			t.Error("TotalFiles not set correctly")
		}
		if info.TotalSizeBytes != 1024 {
			t.Error("TotalSizeBytes not set correctly")
		}
		if info.CreatedAt.IsZero() {
			t.Error("CreatedAt not set correctly")
		}
	})
}

func TestGetGitInfo(t *testing.T) {
	// This test will behave differently depending on whether
	// we're running in a git repo or not
	info := getGitInfo()

	// Just verify it doesn't panic and returns a GitInfo struct
	_ = info.CommitSha
	_ = info.Branch
	_ = info.RepoURL
}

func TestGitInfo(t *testing.T) {
	t.Run("struct holds git information", func(t *testing.T) {
		info := GitInfo{
			CommitSha: "abc123def456",
			Branch:    "main",
			RepoURL:   "https://github.com/test/repo.git",
		}

		if info.CommitSha != "abc123def456" {
			t.Error("CommitSha not set correctly")
		}
		if info.Branch != "main" {
			t.Error("Branch not set correctly")
		}
		if info.RepoURL != "https://github.com/test/repo.git" {
			t.Error("RepoURL not set correctly")
		}
	})
}
