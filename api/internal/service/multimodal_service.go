package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MultiModalService manages multi-modal trace attachments
type MultiModalService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	attachments map[uuid.UUID]*domain.TraceAttachment
}

// NewMultiModalService creates a new multi-modal service
func NewMultiModalService(logger *zap.Logger) *MultiModalService {
	return &MultiModalService{
		logger:      logger,
		attachments: make(map[uuid.UUID]*domain.TraceAttachment),
	}
}

// RegisterAttachment registers a new multi-modal attachment
func (s *MultiModalService) RegisterAttachment(ctx context.Context, projectID string, input *domain.AttachmentInput) (*domain.TraceAttachment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attachment := &domain.TraceAttachment{
		ID:            uuid.New(),
		TraceID:       input.TraceID,
		ObservationID: input.ObservationID,
		Type:          input.Type,
		Filename:      input.Filename,
		MimeType:      input.MimeType,
		SizeBytes:     input.SizeBytes,
		URL:           fmt.Sprintf("https://storage.agenttrace.io/attachments/%s/%s", projectID, input.Filename),
		Metadata:      map[string]string{"projectId": projectID},
		UploadedAt:    time.Now(),
	}

	s.attachments[attachment.ID] = attachment
	s.logger.Info("registered attachment", zap.String("id", attachment.ID.String()), zap.String("type", attachment.Type))
	return attachment, nil
}

// GetTraceAttachments returns all attachments for a trace
func (s *MultiModalService) GetTraceAttachments(ctx context.Context, traceID string) (*domain.MultiModalTrace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tid, _ := uuid.Parse(traceID)
	var attachments []domain.TraceAttachment
	for _, a := range s.attachments {
		if a.TraceID == tid {
			attachments = append(attachments, *a)
		}
	}
	if attachments == nil {
		attachments = []domain.TraceAttachment{}
	}

	summary := s.buildSummary(attachments)
	return &domain.MultiModalTrace{
		TraceID:     tid,
		Attachments: attachments,
		Summary:     summary,
	}, nil
}

// GetAttachment returns a specific attachment by ID
func (s *MultiModalService) GetAttachment(ctx context.Context, attachmentID string) (*domain.TraceAttachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aid, _ := uuid.Parse(attachmentID)
	if a, ok := s.attachments[aid]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("attachment not found: %s", attachmentID)
}

// GetSummary returns a summary of attachments for a trace
func (s *MultiModalService) GetSummary(ctx context.Context, traceID string) (*domain.MultiModalSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tid, _ := uuid.Parse(traceID)
	var attachments []domain.TraceAttachment
	for _, a := range s.attachments {
		if a.TraceID == tid {
			attachments = append(attachments, *a)
		}
	}

	summary := s.buildSummary(attachments)
	return &summary, nil
}

// ListAttachments returns attachments matching the filter
func (s *MultiModalService) ListAttachments(ctx context.Context, projectID string, filter *domain.AttachmentFilter) ([]domain.TraceAttachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.TraceAttachment
	for _, a := range s.attachments {
		if filter != nil {
			if filter.TraceID != "" && a.TraceID.String() != filter.TraceID {
				continue
			}
			if filter.Type != "" && a.Type != filter.Type {
				continue
			}
			if filter.MinSize > 0 && a.SizeBytes < filter.MinSize {
				continue
			}
			if filter.MaxSize > 0 && a.SizeBytes > filter.MaxSize {
				continue
			}
		}
		result = append(result, *a)
	}
	if result == nil {
		result = []domain.TraceAttachment{}
	}
	return result, nil
}

func (s *MultiModalService) buildSummary(attachments []domain.TraceAttachment) domain.MultiModalSummary {
	byType := make(map[string]int)
	var totalSize int64
	for _, a := range attachments {
		byType[a.Type]++
		totalSize += a.SizeBytes
	}
	return domain.MultiModalSummary{
		TotalAttachments: len(attachments),
		ByType:           byType,
		TotalSizeBytes:   totalSize,
	}
}
