package domain

import (
	"time"

	"github.com/google/uuid"
)

// FileOperation represents a file operation performed during agent execution
type FileOperation struct {
	ID            uuid.UUID         `json:"id" ch:"id"`
	ProjectID     uuid.UUID         `json:"projectId" ch:"project_id"`
	TraceID       string            `json:"traceId" ch:"trace_id"`
	ObservationID *string           `json:"observationId,omitempty" ch:"observation_id"`

	// Operation details
	Operation FileOperationType `json:"operation" ch:"operation"`
	FilePath  string            `json:"filePath" ch:"file_path"`
	NewPath   string            `json:"newPath,omitempty" ch:"new_path"`

	// File info
	FileSize    uint64 `json:"fileSize" ch:"file_size"`
	FileMode    string `json:"fileMode,omitempty" ch:"file_mode"`
	ContentHash string `json:"contentHash,omitempty" ch:"content_hash"`
	MimeType    string `json:"mimeType,omitempty" ch:"mime_type"`

	// Change details
	LinesAdded   uint32 `json:"linesAdded" ch:"lines_added"`
	LinesRemoved uint32 `json:"linesRemoved" ch:"lines_removed"`
	DiffPreview  string `json:"diffPreview,omitempty" ch:"diff_preview"`

	// Content snapshots
	ContentBeforeHash string `json:"contentBeforeHash,omitempty" ch:"content_before_hash"`
	ContentAfterHash  string `json:"contentAfterHash,omitempty" ch:"content_after_hash"`

	// Context
	ToolName string `json:"toolName,omitempty" ch:"tool_name"`
	Reason   string `json:"reason,omitempty" ch:"reason"`

	// Timing
	StartedAt   time.Time  `json:"startedAt" ch:"started_at"`
	CompletedAt *time.Time `json:"completedAt,omitempty" ch:"completed_at"`
	DurationMs  uint32     `json:"durationMs" ch:"duration_ms"`

	// Status
	Success      bool   `json:"success" ch:"success"`
	ErrorMessage string `json:"errorMessage,omitempty" ch:"error_message"`
}

// FileOperationInput represents input for creating a file operation
type FileOperationInput struct {
	TraceID       string            `json:"traceId" validate:"required"`
	ObservationID *string           `json:"observationId,omitempty"`
	Operation     FileOperationType `json:"operation" validate:"required"`
	FilePath      string            `json:"filePath" validate:"required"`
	NewPath       *string           `json:"newPath,omitempty"`

	FileSize *uint64 `json:"fileSize,omitempty"`
	FileMode *string `json:"fileMode,omitempty"`
	ContentHash *string `json:"contentHash,omitempty"`
	MimeType    *string `json:"mimeType,omitempty"`

	LinesAdded   *uint32 `json:"linesAdded,omitempty"`
	LinesRemoved *uint32 `json:"linesRemoved,omitempty"`
	DiffPreview  *string `json:"diffPreview,omitempty"`

	ContentBeforeHash *string `json:"contentBeforeHash,omitempty"`
	ContentAfterHash  *string `json:"contentAfterHash,omitempty"`

	ToolName *string `json:"toolName,omitempty"`
	Reason   *string `json:"reason,omitempty"`

	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	DurationMs  *uint32    `json:"durationMs,omitempty"`

	Success      *bool   `json:"success,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// FileOperationFilter represents filter options for querying file operations
type FileOperationFilter struct {
	ProjectID     uuid.UUID
	TraceID       *string
	ObservationID *string
	Operation     *FileOperationType
	FilePath      *string
	Success       *bool
	FromTime      *time.Time
	ToTime        *time.Time
}

// FileOperationList represents a paginated list of file operations
type FileOperationList struct {
	FileOperations []FileOperation `json:"fileOperations"`
	TotalCount     int64           `json:"totalCount"`
	HasMore        bool            `json:"hasMore"`
}

// FileOperationStats represents statistics for file operations
type FileOperationStats struct {
	TotalOperations uint64 `json:"totalOperations"`
	CreateCount     uint64 `json:"createCount"`
	ReadCount       uint64 `json:"readCount"`
	UpdateCount     uint64 `json:"updateCount"`
	DeleteCount     uint64 `json:"deleteCount"`
	SuccessCount    uint64 `json:"successCount"`
	FailureCount    uint64 `json:"failureCount"`
	TotalLinesAdded uint64 `json:"totalLinesAdded"`
	TotalLinesRemoved uint64 `json:"totalLinesRemoved"`
}
