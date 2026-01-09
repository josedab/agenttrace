package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CheckpointRepository defines checkpoint repository operations
type CheckpointRepository interface {
	Create(ctx context.Context, checkpoint *domain.Checkpoint) error
	GetByID(ctx context.Context, projectID, checkpointID uuid.UUID) (*domain.Checkpoint, error)
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Checkpoint, error)
	List(ctx context.Context, filter *domain.CheckpointFilter, limit, offset int) (*domain.CheckpointList, error)
}

// CheckpointService handles checkpoint operations
type CheckpointService struct {
	checkpointRepo CheckpointRepository
	traceRepo      TraceRepository
}

// NewCheckpointService creates a new checkpoint service
func NewCheckpointService(
	checkpointRepo CheckpointRepository,
	traceRepo TraceRepository,
) *CheckpointService {
	return &CheckpointService{
		checkpointRepo: checkpointRepo,
		traceRepo:      traceRepo,
	}
}

// Create creates a new checkpoint
func (s *CheckpointService) Create(ctx context.Context, projectID uuid.UUID, input *domain.CheckpointInput) (*domain.Checkpoint, error) {
	// Verify trace exists
	_, err := s.traceRepo.GetByID(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	checkpointType := input.Type
	if checkpointType == "" {
		checkpointType = domain.CheckpointTypeManual
	}

	var description string
	if input.Description != nil {
		description = *input.Description
	}

	checkpoint := &domain.Checkpoint{
		ID:             uuid.New(),
		ProjectID:      projectID,
		TraceID:        input.TraceID,
		ObservationID:  input.ObservationID,
		Name:           input.Name,
		Description:    description,
		Type:           checkpointType,
		GitCommitSha:   input.GitCommitSha,
		GitBranch:      input.GitBranch,
		GitRepoURL:     input.GitRepoURL,
		FilesSnapshot:  input.FilesSnapshot,
		FilesChanged:   input.FilesChanged,
		TotalFiles:     input.TotalFiles,
		TotalSizeBytes: input.TotalSizeBytes,
		CreatedAt:      time.Now(),
	}

	if err := s.checkpointRepo.Create(ctx, checkpoint); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	return checkpoint, nil
}

// Get retrieves a checkpoint by ID
func (s *CheckpointService) Get(ctx context.Context, projectID, checkpointID uuid.UUID) (*domain.Checkpoint, error) {
	return s.checkpointRepo.GetByID(ctx, projectID, checkpointID)
}

// GetByTraceID retrieves checkpoints for a trace
func (s *CheckpointService) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.Checkpoint, error) {
	return s.checkpointRepo.GetByTraceID(ctx, projectID, traceID)
}

// List retrieves checkpoints with filtering
func (s *CheckpointService) List(ctx context.Context, filter *domain.CheckpointFilter, limit, offset int) (*domain.CheckpointList, error) {
	return s.checkpointRepo.List(ctx, filter, limit, offset)
}

// Restore restores from a checkpoint (creates a new rollback checkpoint)
func (s *CheckpointService) Restore(ctx context.Context, projectID uuid.UUID, input *domain.RestoreCheckpointInput) (*domain.Checkpoint, error) {
	// Get the checkpoint to restore from
	checkpoint, err := s.checkpointRepo.GetByID(ctx, projectID, input.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}

	// Create a new rollback checkpoint
	now := time.Now()
	rollbackCheckpoint := &domain.Checkpoint{
		ID:             uuid.New(),
		ProjectID:      projectID,
		TraceID:        input.TraceID,
		Name:           fmt.Sprintf("Rollback from %s", checkpoint.Name),
		Type:           domain.CheckpointTypeRollback,
		GitCommitSha:   checkpoint.GitCommitSha,
		GitBranch:      checkpoint.GitBranch,
		GitRepoURL:     checkpoint.GitRepoURL,
		FilesSnapshot:  checkpoint.FilesSnapshot,
		FilesChanged:   checkpoint.FilesChanged,
		TotalFiles:     checkpoint.TotalFiles,
		TotalSizeBytes: checkpoint.TotalSizeBytes,
		RestoredFrom:   &checkpoint.ID,
		RestoredAt:     &now,
		CreatedAt:      now,
	}

	if err := s.checkpointRepo.Create(ctx, rollbackCheckpoint); err != nil {
		return nil, fmt.Errorf("failed to create rollback checkpoint: %w", err)
	}

	return rollbackCheckpoint, nil
}
