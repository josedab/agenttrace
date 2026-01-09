package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RealtimeEvent represents an event to be sent to clients
type RealtimeEvent struct {
	Type      string    `json:"type"`
	ProjectID uuid.UUID `json:"projectId"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// Subscriber represents a connected client
type Subscriber struct {
	ID        string
	ProjectID uuid.UUID
	Channel   chan *RealtimeEvent
	Done      chan struct{}
}

// RealtimeService handles real-time event streaming
type RealtimeService struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber
}

// NewRealtimeService creates a new realtime service
func NewRealtimeService() *RealtimeService {
	return &RealtimeService{
		subscribers: make(map[string]*Subscriber),
	}
}

// Subscribe creates a new subscription for a project
func (s *RealtimeService) Subscribe(ctx context.Context, projectID uuid.UUID) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub := &Subscriber{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Channel:   make(chan *RealtimeEvent, 100),
		Done:      make(chan struct{}),
	}

	s.subscribers[sub.ID] = sub

	// Clean up when context is done
	go func() {
		select {
		case <-ctx.Done():
			s.Unsubscribe(sub.ID)
		case <-sub.Done:
		}
	}()

	return sub
}

// Unsubscribe removes a subscription
func (s *RealtimeService) Unsubscribe(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sub, ok := s.subscribers[id]; ok {
		close(sub.Done)
		close(sub.Channel)
		delete(s.subscribers, id)
	}
}

// Publish sends an event to all subscribers of a project
func (s *RealtimeService) Publish(ctx context.Context, projectID uuid.UUID, eventType string, data any) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event := &RealtimeEvent{
		Type:      eventType,
		ProjectID: projectID,
		Data:      data,
		Timestamp: time.Now(),
	}

	for _, sub := range s.subscribers {
		if sub.ProjectID == projectID {
			select {
			case sub.Channel <- event:
			default:
				// Channel is full, skip this subscriber
			}
		}
	}
}

// PublishTraceCreated publishes a trace created event
func (s *RealtimeService) PublishTraceCreated(ctx context.Context, projectID uuid.UUID, traceID string) {
	s.Publish(ctx, projectID, "trace.created", map[string]string{
		"traceId": traceID,
	})
}

// PublishTraceUpdated publishes a trace updated event
func (s *RealtimeService) PublishTraceUpdated(ctx context.Context, projectID uuid.UUID, traceID string) {
	s.Publish(ctx, projectID, "trace.updated", map[string]string{
		"traceId": traceID,
	})
}

// PublishObservationCreated publishes an observation created event
func (s *RealtimeService) PublishObservationCreated(ctx context.Context, projectID uuid.UUID, traceID, observationID string) {
	s.Publish(ctx, projectID, "observation.created", map[string]string{
		"traceId":       traceID,
		"observationId": observationID,
	})
}

// PublishScoreCreated publishes a score created event
func (s *RealtimeService) PublishScoreCreated(ctx context.Context, projectID uuid.UUID, traceID, scoreID string) {
	s.Publish(ctx, projectID, "score.created", map[string]string{
		"traceId": traceID,
		"scoreId": scoreID,
	})
}

// PublishEvaluationCompleted publishes an evaluation completed event
func (s *RealtimeService) PublishEvaluationCompleted(ctx context.Context, projectID uuid.UUID, evaluatorID, traceID string, score float64) {
	s.Publish(ctx, projectID, "evaluation.completed", map[string]any{
		"evaluatorId": evaluatorID,
		"traceId":     traceID,
		"score":       score,
	})
}

// GetSubscriberCount returns the number of active subscribers for a project
func (s *RealtimeService) GetSubscriberCount(projectID uuid.UUID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, sub := range s.subscribers {
		if sub.ProjectID == projectID {
			count++
		}
	}
	return count
}

// FormatSSE formats an event for SSE
func FormatSSE(event *RealtimeEvent) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return append([]byte("data: "), append(data, '\n', '\n')...), nil
}

// EventTypes constants for event types
const (
	EventTypeTraceCreated        = "trace.created"
	EventTypeTraceUpdated        = "trace.updated"
	EventTypeObservationCreated  = "observation.created"
	EventTypeScoreCreated        = "score.created"
	EventTypeEvaluationCompleted = "evaluation.completed"
)
