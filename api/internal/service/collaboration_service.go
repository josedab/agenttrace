package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CollaborationRepository defines repository operations for collaboration data
type CollaborationRepository interface {
	SaveAnnotation(ctx context.Context, annotation *domain.TraceAnnotation) error
	GetAnnotationByID(ctx context.Context, id uuid.UUID) (*domain.TraceAnnotation, error)
	ListAnnotations(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TraceAnnotation, error)
	UpdateAnnotation(ctx context.Context, annotation *domain.TraceAnnotation) error
	SaveSharedSession(ctx context.Context, session *domain.SharedSession) error
	GetSharedSessionByID(ctx context.Context, id uuid.UUID) (*domain.SharedSession, error)
}

// CollaborationService manages real-time collaboration on traces
type CollaborationService struct {
	logger      *zap.Logger
	collabRepo  CollaborationRepository
	realtimeSvc *RealtimeService
}

// NewCollaborationService creates a new collaboration service
func NewCollaborationService(
	logger *zap.Logger,
	collabRepo CollaborationRepository,
	realtimeSvc *RealtimeService,
) *CollaborationService {
	return &CollaborationService{
		logger:      logger,
		collabRepo:  collabRepo,
		realtimeSvc: realtimeSvc,
	}
}

// UpdatePresence tracks which user is viewing which trace
func (s *CollaborationService) UpdatePresence(ctx context.Context, presence *domain.UserPresence) error {
	presence.LastSeen = time.Now()

	// Broadcast presence update to other viewers
	s.realtimeSvc.Publish(ctx, presence.ProjectID, "presence_update", presence)

	return nil
}

// GetPresence retrieves all users currently viewing a trace
func (s *CollaborationService) GetPresence(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.UserPresence, error) {
	// In a real implementation, this would query an in-memory store or cache
	// for active presence records filtered by projectID and traceID
	return []domain.UserPresence{}, nil
}

// AddAnnotation creates a new annotation on a trace event
func (s *CollaborationService) AddAnnotation(ctx context.Context, projectID uuid.UUID, annotation *domain.TraceAnnotation) (*domain.TraceAnnotation, error) {
	annotation.ID = uuid.New()
	annotation.ProjectID = projectID
	annotation.CreatedAt = time.Now()

	if err := s.collabRepo.SaveAnnotation(ctx, annotation); err != nil {
		return nil, fmt.Errorf("failed to save annotation: %w", err)
	}

	// Broadcast annotation to collaborators
	s.realtimeSvc.Publish(ctx, projectID, "annotation_added", annotation)

	s.logger.Info("annotation added",
		zap.String("annotationId", annotation.ID.String()),
		zap.String("traceId", annotation.TraceID),
		zap.String("userId", annotation.UserID.String()),
	)

	return annotation, nil
}

// ListAnnotations retrieves all annotations for a trace
func (s *CollaborationService) ListAnnotations(ctx context.Context, projectID uuid.UUID, traceID string) ([]domain.TraceAnnotation, error) {
	annotations, err := s.collabRepo.ListAnnotations(ctx, projectID, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list annotations: %w", err)
	}
	return annotations, nil
}

// ResolveAnnotation marks an annotation as resolved
func (s *CollaborationService) ResolveAnnotation(ctx context.Context, annotationID uuid.UUID, userID uuid.UUID) error {
	annotation, err := s.collabRepo.GetAnnotationByID(ctx, annotationID)
	if err != nil {
		return fmt.Errorf("failed to get annotation: %w", err)
	}

	now := time.Now()
	annotation.ResolvedAt = &now

	if err := s.collabRepo.UpdateAnnotation(ctx, annotation); err != nil {
		return fmt.Errorf("failed to resolve annotation: %w", err)
	}

	// Broadcast resolution to collaborators
	s.realtimeSvc.Publish(ctx, annotation.ProjectID, "annotation_resolved", annotation)

	return nil
}

// CreateSharedSession creates a new shared collaboration session on a trace
func (s *CollaborationService) CreateSharedSession(ctx context.Context, projectID uuid.UUID, input *domain.SharedSession) (*domain.SharedSession, error) {
	input.ID = uuid.New()
	input.ProjectID = projectID
	input.CreatedAt = time.Now()

	if err := s.collabRepo.SaveSharedSession(ctx, input); err != nil {
		return nil, fmt.Errorf("failed to create shared session: %w", err)
	}

	s.logger.Info("shared session created",
		zap.String("sessionId", input.ID.String()),
		zap.String("traceId", input.TraceID),
	)

	return input, nil
}

// GetSharedSession retrieves a shared session by ID
func (s *CollaborationService) GetSharedSession(ctx context.Context, sessionID uuid.UUID) (*domain.SharedSession, error) {
	session, err := s.collabRepo.GetSharedSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared session: %w", err)
	}
	return session, nil
}
