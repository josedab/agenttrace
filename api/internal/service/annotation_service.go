package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AnnotationService manages collaborative annotations
type AnnotationService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	annotations map[string]*domain.CollabAnnotation
	presence    map[string]map[string]*domain.PresenceUser // traceID -> userID -> user
}

// NewAnnotationService creates a new annotation service
func NewAnnotationService(logger *zap.Logger) *AnnotationService {
	return &AnnotationService{
		logger:      logger,
		annotations: make(map[string]*domain.CollabAnnotation),
		presence:    make(map[string]map[string]*domain.PresenceUser),
	}
}

// CreateAnnotation creates a new annotation on a trace
func (s *AnnotationService) CreateAnnotation(ctx context.Context, projectID, userID, userName string, input *domain.AnnotationInput) (*domain.CollabAnnotation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	annotation := &domain.CollabAnnotation{
		ID:        fmt.Sprintf("ann_%d", time.Now().UnixNano()),
		TraceID:   input.TraceID,
		ProjectID: projectID,
		SpanID:    input.SpanID,
		UserID:    userID,
		UserName:  userName,
		Content:   input.Content,
		Position:  input.Position,
		Thread:    []domain.AnnotationReply{},
		Resolved:  false,
		CreatedAt: time.Now(),
	}

	s.annotations[annotation.ID] = annotation
	s.logger.Info("created annotation",
		zap.String("id", annotation.ID),
		zap.String("traceId", input.TraceID),
	)
	return annotation, nil
}

// ListAnnotations returns all annotations for a trace
func (s *AnnotationService) ListAnnotations(ctx context.Context, traceID string) ([]domain.CollabAnnotation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.CollabAnnotation
	for _, a := range s.annotations {
		if a.TraceID == traceID {
			result = append(result, *a)
		}
	}
	if result == nil {
		result = []domain.CollabAnnotation{}
	}
	return result, nil
}

// AddReply adds a reply to an annotation thread
func (s *AnnotationService) AddReply(ctx context.Context, annotationID, userID, userName string, input *domain.ReplyInput) (*domain.CollabAnnotation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	annotation, ok := s.annotations[annotationID]
	if !ok {
		return nil, fmt.Errorf("annotation not found: %s", annotationID)
	}

	reply := domain.AnnotationReply{
		ID:        fmt.Sprintf("reply_%d", time.Now().UnixNano()),
		UserID:    userID,
		UserName:  userName,
		Content:   input.Content,
		CreatedAt: time.Now(),
	}

	annotation.Thread = append(annotation.Thread, reply)
	s.logger.Info("added reply to annotation",
		zap.String("annotationId", annotationID),
		zap.String("replyId", reply.ID),
	)
	return annotation, nil
}

// ResolveAnnotation marks an annotation as resolved
func (s *AnnotationService) ResolveAnnotation(ctx context.Context, annotationID string) (*domain.CollabAnnotation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	annotation, ok := s.annotations[annotationID]
	if !ok {
		return nil, fmt.Errorf("annotation not found: %s", annotationID)
	}

	annotation.Resolved = true
	s.logger.Info("resolved annotation", zap.String("id", annotationID))
	return annotation, nil
}

// GetPresence returns presence information for a trace
func (s *AnnotationService) GetPresence(ctx context.Context, traceID string) (*domain.AnnotationPresence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	presence := &domain.AnnotationPresence{
		TraceID:     traceID,
		ActiveUsers: []domain.PresenceUser{},
	}

	if users, ok := s.presence[traceID]; ok {
		for _, u := range users {
			// Only include users active in the last 5 minutes
			if time.Since(u.LastActive) < 5*time.Minute {
				presence.ActiveUsers = append(presence.ActiveUsers, *u)
			}
		}
	}

	return presence, nil
}

// JoinTrace marks a user as present on a trace
func (s *AnnotationService) JoinTrace(ctx context.Context, traceID, userID, userName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.presence[traceID] == nil {
		s.presence[traceID] = make(map[string]*domain.PresenceUser)
	}

	s.presence[traceID][userID] = &domain.PresenceUser{
		UserID:     userID,
		UserName:   userName,
		LastActive: time.Now(),
	}

	s.logger.Info("user joined trace",
		zap.String("traceId", traceID),
		zap.String("userId", userID),
	)
	return nil
}

// LeaveTrace removes a user from a trace's presence
func (s *AnnotationService) LeaveTrace(ctx context.Context, traceID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if users, ok := s.presence[traceID]; ok {
		delete(users, userID)
	}

	s.logger.Info("user left trace",
		zap.String("traceId", traceID),
		zap.String("userId", userID),
	)
	return nil
}
