package agenttrace

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileOperationType represents the type of file operation.
type FileOperationType string

const (
	FileOpCreate FileOperationType = "create"
	FileOpRead   FileOperationType = "read"
	FileOpUpdate FileOperationType = "update"
	FileOpDelete FileOperationType = "delete"
	FileOpRename FileOperationType = "rename"
	FileOpCopy   FileOperationType = "copy"
	FileOpMove   FileOperationType = "move"
	FileOpChmod  FileOperationType = "chmod"
)

// FileOperationOptions holds options for tracking a file operation.
type FileOperationOptions struct {
	Operation     FileOperationType
	FilePath      string
	ObservationID string
	NewPath       string
	ContentBefore string
	ContentAfter  string
	LinesAdded    *int
	LinesRemoved  *int
	DiffPreview   string
	ToolName      string
	Reason        string
	StartedAt     *time.Time
	CompletedAt   *time.Time
	Success       *bool
	ErrorMessage  string
}

// FileOperationInfo contains information about a tracked file operation.
type FileOperationInfo struct {
	ID            string
	TraceID       string
	ObservationID string
	Operation     FileOperationType
	FilePath      string
	NewPath       string
	FileSize      int64
	ContentHash   string
	LinesAdded    int
	LinesRemoved  int
	Success       bool
	DurationMs    int64
	StartedAt     time.Time
	CompletedAt   time.Time
}

// FileOp tracks a file operation for this trace.
func (t *Trace) FileOp(opts FileOperationOptions) *FileOperationInfo {
	opID := uuid.New().String()
	now := time.Now().UTC()

	startedAt := now
	if opts.StartedAt != nil {
		startedAt = *opts.StartedAt
	}

	completedAt := now
	if opts.CompletedAt != nil {
		completedAt = *opts.CompletedAt
	}

	durationMs := completedAt.Sub(startedAt).Milliseconds()

	// Calculate file info
	var fileSize int64
	var fileMode string
	var contentHash string
	var contentBeforeHash string
	var contentAfterHash string

	if info, err := os.Stat(opts.FilePath); err == nil {
		fileSize = info.Size()
		fileMode = info.Mode().String()
	}

	if opts.ContentBefore != "" {
		h := sha256.Sum256([]byte(opts.ContentBefore))
		contentBeforeHash = hex.EncodeToString(h[:])
	}

	if opts.ContentAfter != "" {
		h := sha256.Sum256([]byte(opts.ContentAfter))
		contentAfterHash = hex.EncodeToString(h[:])
		contentHash = contentAfterHash
	}

	// Auto-calculate lines changed
	linesAdded := 0
	linesRemoved := 0

	if opts.LinesAdded != nil {
		linesAdded = *opts.LinesAdded
	} else if opts.ContentBefore != "" && opts.ContentAfter != "" {
		beforeLines := make(map[string]bool)
		for _, line := range strings.Split(opts.ContentBefore, "\n") {
			beforeLines[line] = true
		}
		afterLines := strings.Split(opts.ContentAfter, "\n")
		for _, line := range afterLines {
			if !beforeLines[line] {
				linesAdded++
			}
		}
	}

	if opts.LinesRemoved != nil {
		linesRemoved = *opts.LinesRemoved
	} else if opts.ContentBefore != "" && opts.ContentAfter != "" {
		afterLines := make(map[string]bool)
		for _, line := range strings.Split(opts.ContentAfter, "\n") {
			afterLines[line] = true
		}
		for _, line := range strings.Split(opts.ContentBefore, "\n") {
			if !afterLines[line] {
				linesRemoved++
			}
		}
	}

	success := true
	if opts.Success != nil {
		success = *opts.Success
	}

	// Send to API
	if t.client.Enabled() {
		t.client.addEvent(map[string]any{
			"type": "file-operation-create",
			"body": map[string]any{
				"id":                opID,
				"traceId":           t.id,
				"observationId":     opts.ObservationID,
				"operation":         string(opts.Operation),
				"filePath":          opts.FilePath,
				"newPath":           opts.NewPath,
				"fileSize":          fileSize,
				"fileMode":          fileMode,
				"contentHash":       contentHash,
				"linesAdded":        linesAdded,
				"linesRemoved":      linesRemoved,
				"diffPreview":       opts.DiffPreview,
				"contentBeforeHash": contentBeforeHash,
				"contentAfterHash":  contentAfterHash,
				"toolName":          opts.ToolName,
				"reason":            opts.Reason,
				"startedAt":         startedAt.Format(time.RFC3339Nano),
				"completedAt":       completedAt.Format(time.RFC3339Nano),
				"durationMs":        durationMs,
				"success":           success,
				"errorMessage":      opts.ErrorMessage,
			},
		})
	}

	return &FileOperationInfo{
		ID:            opID,
		TraceID:       t.id,
		ObservationID: opts.ObservationID,
		Operation:     opts.Operation,
		FilePath:      opts.FilePath,
		NewPath:       opts.NewPath,
		FileSize:      fileSize,
		ContentHash:   contentHash,
		LinesAdded:    linesAdded,
		LinesRemoved:  linesRemoved,
		Success:       success,
		DurationMs:    durationMs,
		StartedAt:     startedAt,
		CompletedAt:   completedAt,
	}
}
