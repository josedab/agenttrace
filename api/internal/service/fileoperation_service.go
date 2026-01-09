package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// FileOperationRepository defines file operation repository operations
type FileOperationRepository interface {
	Create(ctx context.Context, fileOp *domain.FileOperation) error
	CreateBatch(ctx context.Context, fileOps []*domain.FileOperation) error
	GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.FileOperation, error)
	List(ctx context.Context, filter *domain.FileOperationFilter, limit, offset int) (*domain.FileOperationList, error)
	GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.FileOperationStats, error)
}

// FileOperationService handles file operation tracking
type FileOperationService struct {
	fileOpRepo FileOperationRepository
	traceRepo  TraceRepository
}

// NewFileOperationService creates a new file operation service
func NewFileOperationService(
	fileOpRepo FileOperationRepository,
	traceRepo TraceRepository,
) *FileOperationService {
	return &FileOperationService{
		fileOpRepo: fileOpRepo,
		traceRepo:  traceRepo,
	}
}

// Track records a file operation
func (s *FileOperationService) Track(ctx context.Context, projectID uuid.UUID, input *domain.FileOperationInput) (*domain.FileOperation, error) {
	// Verify trace exists
	_, err := s.traceRepo.GetByID(ctx, projectID, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	now := time.Now()
	startedAt := now
	if input.StartedAt != nil {
		startedAt = *input.StartedAt
	}

	var completedAt *time.Time
	if input.CompletedAt != nil {
		completedAt = input.CompletedAt
	} else {
		completedAt = &now
	}

	durationMs := uint32(completedAt.Sub(startedAt).Milliseconds())

	var newPath, contentHash, mimeType, diffPreview string
	var contentBeforeHash, contentAfterHash, toolName, reason, errorMessage string

	if input.NewPath != nil {
		newPath = *input.NewPath
	}
	if input.ContentHash != nil {
		contentHash = *input.ContentHash
	}
	if input.MimeType != nil {
		mimeType = *input.MimeType
	}
	if input.DiffPreview != nil {
		diffPreview = *input.DiffPreview
	}
	if input.ContentBeforeHash != nil {
		contentBeforeHash = *input.ContentBeforeHash
	}
	if input.ContentAfterHash != nil {
		contentAfterHash = *input.ContentAfterHash
	}
	if input.ToolName != nil {
		toolName = *input.ToolName
	}
	if input.Reason != nil {
		reason = *input.Reason
	}
	if input.ErrorMessage != nil {
		errorMessage = *input.ErrorMessage
	}

	var fileSize uint64
	var fileMode string
	var linesAdded, linesRemoved uint32

	if input.FileSize != nil {
		fileSize = *input.FileSize
	}
	if input.FileMode != nil {
		fileMode = *input.FileMode
	}
	if input.LinesAdded != nil {
		linesAdded = *input.LinesAdded
	}
	if input.LinesRemoved != nil {
		linesRemoved = *input.LinesRemoved
	}

	success := true
	if input.Success != nil {
		success = *input.Success
	}

	fileOp := &domain.FileOperation{
		ID:                uuid.New(),
		ProjectID:         projectID,
		TraceID:           input.TraceID,
		ObservationID:     input.ObservationID,
		Operation:         input.Operation,
		FilePath:          input.FilePath,
		NewPath:           newPath,
		FileSize:          fileSize,
		FileMode:          fileMode,
		ContentHash:       contentHash,
		MimeType:          mimeType,
		LinesAdded:        linesAdded,
		LinesRemoved:      linesRemoved,
		DiffPreview:       diffPreview,
		ContentBeforeHash: contentBeforeHash,
		ContentAfterHash:  contentAfterHash,
		ToolName:          toolName,
		Reason:            reason,
		StartedAt:         startedAt,
		CompletedAt:       completedAt,
		DurationMs:        durationMs,
		Success:           success,
		ErrorMessage:      errorMessage,
	}

	if err := s.fileOpRepo.Create(ctx, fileOp); err != nil {
		return nil, fmt.Errorf("failed to create file operation: %w", err)
	}

	return fileOp, nil
}

// TrackBatch records multiple file operations
func (s *FileOperationService) TrackBatch(ctx context.Context, projectID uuid.UUID, inputs []*domain.FileOperationInput) ([]*domain.FileOperation, error) {
	if len(inputs) == 0 {
		return []*domain.FileOperation{}, nil
	}

	// Verify trace exists for first input (assuming all belong to same trace)
	_, err := s.traceRepo.GetByID(ctx, projectID, inputs[0].TraceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	now := time.Now()
	fileOps := make([]*domain.FileOperation, len(inputs))

	for i, input := range inputs {
		startedAt := now
		if input.StartedAt != nil {
			startedAt = *input.StartedAt
		}

		var completedAt *time.Time
		if input.CompletedAt != nil {
			completedAt = input.CompletedAt
		} else {
			nowCopy := now
			completedAt = &nowCopy
		}

		durationMs := uint32(completedAt.Sub(startedAt).Milliseconds())

		var newPath, contentHash, mimeType, diffPreview string
		var contentBeforeHash, contentAfterHash, toolName, reason, errorMessage string

		if input.NewPath != nil {
			newPath = *input.NewPath
		}
		if input.ContentHash != nil {
			contentHash = *input.ContentHash
		}
		if input.MimeType != nil {
			mimeType = *input.MimeType
		}
		if input.DiffPreview != nil {
			diffPreview = *input.DiffPreview
		}
		if input.ContentBeforeHash != nil {
			contentBeforeHash = *input.ContentBeforeHash
		}
		if input.ContentAfterHash != nil {
			contentAfterHash = *input.ContentAfterHash
		}
		if input.ToolName != nil {
			toolName = *input.ToolName
		}
		if input.Reason != nil {
			reason = *input.Reason
		}
		if input.ErrorMessage != nil {
			errorMessage = *input.ErrorMessage
		}

		var fileSize uint64
		var fileMode string
		var linesAdded, linesRemoved uint32

		if input.FileSize != nil {
			fileSize = *input.FileSize
		}
		if input.FileMode != nil {
			fileMode = *input.FileMode
		}
		if input.LinesAdded != nil {
			linesAdded = *input.LinesAdded
		}
		if input.LinesRemoved != nil {
			linesRemoved = *input.LinesRemoved
		}

		success := true
		if input.Success != nil {
			success = *input.Success
		}

		fileOps[i] = &domain.FileOperation{
			ID:                uuid.New(),
			ProjectID:         projectID,
			TraceID:           input.TraceID,
			ObservationID:     input.ObservationID,
			Operation:         input.Operation,
			FilePath:          input.FilePath,
			NewPath:           newPath,
			FileSize:          fileSize,
			FileMode:          fileMode,
			ContentHash:       contentHash,
			MimeType:          mimeType,
			LinesAdded:        linesAdded,
			LinesRemoved:      linesRemoved,
			DiffPreview:       diffPreview,
			ContentBeforeHash: contentBeforeHash,
			ContentAfterHash:  contentAfterHash,
			ToolName:          toolName,
			Reason:            reason,
			StartedAt:         startedAt,
			CompletedAt:       completedAt,
			DurationMs:        durationMs,
			Success:           success,
			ErrorMessage:      errorMessage,
		}
	}

	if err := s.fileOpRepo.CreateBatch(ctx, fileOps); err != nil {
		return nil, fmt.Errorf("failed to create file operations: %w", err)
	}

	return fileOps, nil
}

// GetByTraceID retrieves file operations for a trace
func (s *FileOperationService) GetByTraceID(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.FileOperation, error) {
	return s.fileOpRepo.GetByTraceID(ctx, projectID, traceID)
}

// List retrieves file operations with filtering
func (s *FileOperationService) List(ctx context.Context, filter *domain.FileOperationFilter, limit, offset int) (*domain.FileOperationList, error) {
	return s.fileOpRepo.List(ctx, filter, limit, offset)
}

// GetStats retrieves file operation statistics
func (s *FileOperationService) GetStats(ctx context.Context, projectID uuid.UUID, traceID *string) (*domain.FileOperationStats, error) {
	return s.fileOpRepo.GetStats(ctx, projectID, traceID)
}
