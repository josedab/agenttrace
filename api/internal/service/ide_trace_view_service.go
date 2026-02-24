package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// IDETraceViewService handles IDE trace view logic
type IDETraceViewService struct {
	logger *zap.Logger
}

// NewIDETraceViewService creates a new IDE trace view service
func NewIDETraceViewService(logger *zap.Logger) *IDETraceViewService {
	return &IDETraceViewService{
		logger: logger,
	}
}

// GetFileMapping returns trace mapping for a specific file
func (s *IDETraceViewService) GetFileMapping(ctx context.Context, projectID uuid.UUID, filePath string) (*domain.FileTraceMapping, error) {
	s.logger.Info("fetching file trace mapping",
		zap.String("projectId", projectID.String()),
		zap.String("filePath", filePath),
	)

	mapping := &domain.FileTraceMapping{
		FilePath:    filePath,
		ProjectID:   projectID,
		Annotations: []domain.LineAnnotation{},
		Summary: domain.FileTraceSummary{
			TopAgents: []string{},
		},
	}

	return mapping, nil
}

// GetTraceContext returns detailed trace context for IDE display
func (s *IDETraceViewService) GetTraceContext(ctx context.Context, traceID uuid.UUID) (*domain.IDETraceContext, error) {
	s.logger.Info("fetching IDE trace context", zap.String("traceId", traceID.String()))

	traceCtx := &domain.IDETraceContext{
		TraceID:     traceID,
		FileChanges: []domain.IDEFileChange{},
	}

	return traceCtx, nil
}

// GetBatchMappings returns trace mappings for multiple files at once
func (s *IDETraceViewService) GetBatchMappings(ctx context.Context, projectID uuid.UUID, filePaths []string) ([]domain.FileTraceMapping, error) {
	s.logger.Info("fetching batch file trace mappings",
		zap.String("projectId", projectID.String()),
		zap.Int("fileCount", len(filePaths)),
	)

	mappings := make([]domain.FileTraceMapping, 0, len(filePaths))
	for _, fp := range filePaths {
		mapping, err := s.GetFileMapping(ctx, projectID, fp)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, *mapping)
	}

	return mappings, nil
}
