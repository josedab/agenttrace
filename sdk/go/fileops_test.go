package agenttrace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileOperationType(t *testing.T) {
	t.Run("all file operation types exist", func(t *testing.T) {
		types := []FileOperationType{
			FileOpCreate,
			FileOpRead,
			FileOpUpdate,
			FileOpDelete,
			FileOpRename,
			FileOpCopy,
			FileOpMove,
			FileOpChmod,
		}

		expected := []string{"create", "read", "update", "delete", "rename", "copy", "move", "chmod"}

		for i, op := range types {
			if string(op) != expected[i] {
				t.Errorf("expected %s, got %s", expected[i], op)
			}
		}
	})
}

func TestTraceFileOp(t *testing.T) {
	client := New(Config{
		APIKey: "test-api-key",
	})
	defer client.Shutdown()

	ctx := context.Background()
	trace := client.Trace(ctx, TraceOptions{
		Name: "test-trace",
	})

	t.Run("tracks read operation", func(t *testing.T) {
		op := trace.FileOp(FileOperationOptions{
			Operation: FileOpRead,
			FilePath:  "/path/to/file.go",
		})

		if op.Operation != FileOpRead {
			t.Errorf("expected operation to be 'read', got '%s'", op.Operation)
		}
		if op.FilePath != "/path/to/file.go" {
			t.Errorf("expected filePath to be '/path/to/file.go', got '%s'", op.FilePath)
		}
		if op.TraceID != trace.id {
			t.Error("expected file op to have trace's ID")
		}
		if op.ID == "" {
			t.Error("expected file op ID to be generated")
		}
		if op.Success != true {
			t.Error("expected success to be true by default")
		}
	})

	t.Run("tracks write operation with line changes", func(t *testing.T) {
		linesAdded := 10
		linesRemoved := 5

		op := trace.FileOp(FileOperationOptions{
			Operation:    FileOpUpdate,
			FilePath:     "/path/to/file.go",
			LinesAdded:   &linesAdded,
			LinesRemoved: &linesRemoved,
		})

		if op.Operation != FileOpUpdate {
			t.Errorf("expected operation to be 'update', got '%s'", op.Operation)
		}
		if op.LinesAdded != 10 {
			t.Errorf("expected linesAdded to be 10, got %d", op.LinesAdded)
		}
		if op.LinesRemoved != 5 {
			t.Errorf("expected linesRemoved to be 5, got %d", op.LinesRemoved)
		}
	})

	t.Run("tracks create operation", func(t *testing.T) {
		op := trace.FileOp(FileOperationOptions{
			Operation:    FileOpCreate,
			FilePath:     "/path/to/newfile.go",
			ContentAfter: "package main\n\nfunc main() {}",
		})

		if op.Operation != FileOpCreate {
			t.Errorf("expected operation to be 'create', got '%s'", op.Operation)
		}
	})

	t.Run("tracks delete operation", func(t *testing.T) {
		op := trace.FileOp(FileOperationOptions{
			Operation: FileOpDelete,
			FilePath:  "/path/to/deleted.go",
		})

		if op.Operation != FileOpDelete {
			t.Errorf("expected operation to be 'delete', got '%s'", op.Operation)
		}
	})

	t.Run("tracks rename operation with new path", func(t *testing.T) {
		op := trace.FileOp(FileOperationOptions{
			Operation: FileOpRename,
			FilePath:  "/old/path/file.go",
			NewPath:   "/new/path/file.go",
		})

		if op.Operation != FileOpRename {
			t.Errorf("expected operation to be 'rename', got '%s'", op.Operation)
		}
		if op.FilePath != "/old/path/file.go" {
			t.Errorf("expected filePath to be old path")
		}
		if op.NewPath != "/new/path/file.go" {
			t.Errorf("expected newPath to be '/new/path/file.go', got '%s'", op.NewPath)
		}
	})

	t.Run("tracks with observation ID", func(t *testing.T) {
		op := trace.FileOp(FileOperationOptions{
			Operation:     FileOpRead,
			FilePath:      "/path/file.go",
			ObservationID: "obs-456",
		})

		if op.ObservationID != "obs-456" {
			t.Errorf("expected observationID to be 'obs-456', got '%s'", op.ObservationID)
		}
	})

	t.Run("auto-calculates lines changed from content", func(t *testing.T) {
		op := trace.FileOp(FileOperationOptions{
			Operation:     FileOpUpdate,
			FilePath:      "/path/file.go",
			ContentBefore: "line1\nline2\nline3",
			ContentAfter:  "line1\nline2\nline4\nline5",
		})

		// line4 and line5 are new, line3 was removed
		if op.LinesAdded != 2 {
			t.Errorf("expected linesAdded to be 2, got %d", op.LinesAdded)
		}
		if op.LinesRemoved != 1 {
			t.Errorf("expected linesRemoved to be 1, got %d", op.LinesRemoved)
		}
	})

	t.Run("calculates duration", func(t *testing.T) {
		startedAt := time.Now().Add(-5 * time.Second)
		completedAt := time.Now()

		op := trace.FileOp(FileOperationOptions{
			Operation:   FileOpRead,
			FilePath:    "/path/file.go",
			StartedAt:   &startedAt,
			CompletedAt: &completedAt,
		})

		if op.DurationMs < 4900 || op.DurationMs > 5100 {
			t.Errorf("expected duration to be ~5000ms, got %d", op.DurationMs)
		}
	})

	t.Run("tracks failed operation", func(t *testing.T) {
		success := false
		op := trace.FileOp(FileOperationOptions{
			Operation:    FileOpUpdate,
			FilePath:     "/path/file.go",
			Success:      &success,
			ErrorMessage: "Permission denied",
		})

		if op.Success != false {
			t.Error("expected success to be false")
		}
	})

	t.Run("gets file size for existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.go")
		content := []byte("package main")
		if err := os.WriteFile(tmpFile, content, 0644); err != nil {
			t.Fatal(err)
		}

		op := trace.FileOp(FileOperationOptions{
			Operation: FileOpRead,
			FilePath:  tmpFile,
		})

		if op.FileSize != int64(len(content)) {
			t.Errorf("expected fileSize to be %d, got %d", len(content), op.FileSize)
		}
	})
}

func TestFileOpDisabledClient(t *testing.T) {
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

	op := trace.FileOp(FileOperationOptions{
		Operation: FileOpRead,
		FilePath:  "/path/file.go",
	})

	// Should still return operation info
	if op.Operation != FileOpRead {
		t.Error("expected operation info even when disabled")
	}
	if op.ID == "" {
		t.Error("expected operation ID even when disabled")
	}
}

func TestFileOperationInfo(t *testing.T) {
	t.Run("has all expected fields", func(t *testing.T) {
		now := time.Now()
		info := FileOperationInfo{
			ID:            "op-123",
			TraceID:       "trace-456",
			ObservationID: "obs-789",
			Operation:     FileOpUpdate,
			FilePath:      "/path/to/file.go",
			NewPath:       "/new/path.go",
			FileSize:      1024,
			ContentHash:   "abc123hash",
			LinesAdded:    10,
			LinesRemoved:  5,
			Success:       true,
			DurationMs:    100,
			StartedAt:     now,
			CompletedAt:   now.Add(100 * time.Millisecond),
		}

		if info.ID != "op-123" {
			t.Error("ID not set correctly")
		}
		if info.TraceID != "trace-456" {
			t.Error("TraceID not set correctly")
		}
		if info.ObservationID != "obs-789" {
			t.Error("ObservationID not set correctly")
		}
		if info.Operation != FileOpUpdate {
			t.Error("Operation not set correctly")
		}
		if info.FilePath != "/path/to/file.go" {
			t.Error("FilePath not set correctly")
		}
		if info.NewPath != "/new/path.go" {
			t.Error("NewPath not set correctly")
		}
		if info.FileSize != 1024 {
			t.Error("FileSize not set correctly")
		}
		if info.ContentHash != "abc123hash" {
			t.Error("ContentHash not set correctly")
		}
		if info.LinesAdded != 10 {
			t.Error("LinesAdded not set correctly")
		}
		if info.LinesRemoved != 5 {
			t.Error("LinesRemoved not set correctly")
		}
		if info.Success != true {
			t.Error("Success not set correctly")
		}
		if info.DurationMs != 100 {
			t.Error("DurationMs not set correctly")
		}
	})
}

func TestFileOperationOptions(t *testing.T) {
	t.Run("optional fields can be nil", func(t *testing.T) {
		opts := FileOperationOptions{
			Operation: FileOpRead,
			FilePath:  "/path/file.go",
			// All other fields are optional
		}

		if opts.LinesAdded != nil {
			t.Error("LinesAdded should be nil by default")
		}
		if opts.LinesRemoved != nil {
			t.Error("LinesRemoved should be nil by default")
		}
		if opts.StartedAt != nil {
			t.Error("StartedAt should be nil by default")
		}
		if opts.CompletedAt != nil {
			t.Error("CompletedAt should be nil by default")
		}
		if opts.Success != nil {
			t.Error("Success should be nil by default")
		}
	})
}
