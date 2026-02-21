package domain

import (
	"time"

	"github.com/google/uuid"
)

// MigrationStatus represents the status of a migration job
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "PENDING"
	MigrationStatusRunning   MigrationStatus = "RUNNING"
	MigrationStatusCompleted MigrationStatus = "COMPLETED"
	MigrationStatusFailed    MigrationStatus = "FAILED"
)

// IsValid checks if the migration status is valid
func (s MigrationStatus) IsValid() bool {
	switch s {
	case MigrationStatusPending, MigrationStatusRunning, MigrationStatusCompleted, MigrationStatusFailed:
		return true
	}
	return false
}

// IsTerminal checks if the migration status is terminal
func (s MigrationStatus) IsTerminal() bool {
	return s == MigrationStatusCompleted || s == MigrationStatusFailed
}

// MigrationJob represents a data migration job from an external platform
type MigrationJob struct {
	ID          uuid.UUID         `json:"id"`
	ProjectID   uuid.UUID         `json:"projectId"`
	Source      string            `json:"source"` // e.g. "langfuse"
	Status      MigrationStatus   `json:"status"`
	Config      MigrationConfig   `json:"config"`
	Progress    MigrationProgress `json:"progress"`
	Errors      []string          `json:"errors,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	CompletedAt *time.Time        `json:"completedAt,omitempty"`
}

// MigrationConfig holds configuration for a migration job
type MigrationConfig struct {
	SourceDSN       string `json:"sourceDsn"`
	IncrementalMode bool   `json:"incrementalMode"`
	DryRun          bool   `json:"dryRun"`
	IncludeTraces   bool   `json:"includeTraces"`
	IncludePrompts  bool   `json:"includePrompts"`
	IncludeDatasets bool   `json:"includeDatasets"`
	IncludeScores   bool   `json:"includeScores"`
}

// MigrationProgress tracks the progress of a migration job
type MigrationProgress struct {
	TotalItems       int64 `json:"totalItems"`
	ProcessedItems   int64 `json:"processedItems"`
	SkippedItems     int64 `json:"skippedItems"`
	TracesMigrated   int64 `json:"tracesMigrated"`
	PromptsMigrated  int64 `json:"promptsMigrated"`
	DatasetsMigrated int64 `json:"datasetsMigrated"`
	ScoresMigrated   int64 `json:"scoresMigrated"`
}

// MigrationInput represents input for starting a migration
type MigrationInput struct {
	Source string          `json:"source" validate:"required"`
	Config MigrationConfig `json:"config"`
}
