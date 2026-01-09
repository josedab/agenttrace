package domain

import (
	"time"

	"github.com/google/uuid"
)

// Checkpoint represents a code state snapshot during agent execution
type Checkpoint struct {
	ID            uuid.UUID      `json:"id" ch:"id"`
	ProjectID     uuid.UUID      `json:"projectId" ch:"project_id"`
	TraceID       string         `json:"traceId" ch:"trace_id"`
	ObservationID *string        `json:"observationId,omitempty" ch:"observation_id"`
	Name          string         `json:"name" ch:"name"`
	Description   string         `json:"description,omitempty" ch:"description"`
	Type          CheckpointType `json:"type" ch:"checkpoint_type"`

	// Git context at checkpoint time
	GitCommitSha string `json:"gitCommitSha,omitempty" ch:"git_commit_sha"`
	GitBranch    string `json:"gitBranch,omitempty" ch:"git_branch"`
	GitRepoURL   string `json:"gitRepoUrl,omitempty" ch:"git_repo_url"`

	// File state
	FilesSnapshot  string   `json:"filesSnapshot,omitempty" ch:"files_snapshot"`
	FilesChanged   []string `json:"filesChanged,omitempty" ch:"files_changed"`
	StoragePath    string   `json:"storagePath,omitempty" ch:"storage_path"`
	TotalFiles     uint32   `json:"totalFiles" ch:"total_files"`
	TotalSizeBytes uint64   `json:"totalSizeBytes" ch:"total_size_bytes"`

	// Restoration info
	RestoredFrom *uuid.UUID `json:"restoredFrom,omitempty" ch:"restored_from"`
	RestoredAt   *time.Time `json:"restoredAt,omitempty" ch:"restored_at"`

	CreatedAt time.Time `json:"createdAt" ch:"created_at"`
}

// CheckpointInput represents input for creating a checkpoint
type CheckpointInput struct {
	TraceID       string         `json:"traceId" validate:"required"`
	ObservationID *string        `json:"observationId,omitempty"`
	Name          string         `json:"name,omitempty"`
	Description   *string        `json:"description,omitempty"`
	Type          CheckpointType `json:"type,omitempty"`

	// Git context
	GitCommitSha string `json:"gitCommitSha,omitempty"`
	GitBranch    string `json:"gitBranch,omitempty"`
	GitRepoURL   string `json:"gitRepoUrl,omitempty"`

	// File state
	FilesSnapshot  string   `json:"filesSnapshot,omitempty"`
	FilesChanged   []string `json:"filesChanged,omitempty"`
	TotalFiles     uint32   `json:"totalFiles,omitempty"`
	TotalSizeBytes uint64   `json:"totalSizeBytes,omitempty"`
}

// CheckpointFilter represents filter options for querying checkpoints
type CheckpointFilter struct {
	ProjectID     uuid.UUID
	TraceID       *string
	ObservationID *string
	Type          *CheckpointType
	GitCommitSha  *string
	GitBranch     *string
	FromTime      *time.Time
	ToTime        *time.Time
}

// CheckpointList represents a paginated list of checkpoints
type CheckpointList struct {
	Checkpoints []Checkpoint `json:"checkpoints"`
	TotalCount  int64        `json:"totalCount"`
	HasMore     bool         `json:"hasMore"`
}

// RestoreCheckpointInput represents input for restoring a checkpoint
type RestoreCheckpointInput struct {
	CheckpointID uuid.UUID `json:"checkpointId" validate:"required"`
	TraceID      string    `json:"traceId" validate:"required"`
}
