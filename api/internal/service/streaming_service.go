package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// TraceStream represents an active stream for a specific trace
type TraceStream struct {
	TraceID     uuid.UUID
	ProjectID   uuid.UUID
	Metrics     *domain.LiveMetrics
	Activities  []domain.StreamActivity
	Subscribers map[string]*TraceStreamSubscriber
	mu          sync.RWMutex
	startedAt   time.Time
}

// TraceStreamSubscriber represents a subscriber to a specific trace stream
type TraceStreamSubscriber struct {
	ID      string
	Channel chan *domain.StreamActivity
	Done    chan struct{}
	Filter  domain.StreamSubscription
}

// StreamingService manages per-trace real-time streaming
type StreamingService struct {
	logger        *zap.Logger
	realtime      *RealtimeService
	streams       map[uuid.UUID]*TraceStream
	interventions map[uuid.UUID][]domain.InterventionRequest
	mu            sync.RWMutex
	maxActivities int
}

// NewStreamingService creates a new streaming service
func NewStreamingService(logger *zap.Logger, realtime *RealtimeService) *StreamingService {
	return &StreamingService{
		logger:        logger,
		realtime:      realtime,
		streams:       make(map[uuid.UUID]*TraceStream),
		interventions: make(map[uuid.UUID][]domain.InterventionRequest),
		maxActivities: 1000,
	}
}

// GetOrCreateStream gets or creates a trace stream
func (s *StreamingService) GetOrCreateStream(traceID, projectID uuid.UUID) *TraceStream {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stream, ok := s.streams[traceID]; ok {
		return stream
	}

	stream := &TraceStream{
		TraceID:     traceID,
		ProjectID:   projectID,
		Metrics:     &domain.LiveMetrics{TraceID: traceID, LastUpdated: time.Now()},
		Activities:  make([]domain.StreamActivity, 0),
		Subscribers: make(map[string]*TraceStreamSubscriber),
		startedAt:   time.Now(),
	}
	s.streams[traceID] = stream
	return stream
}

// SubscribeToTrace subscribes to events for a specific trace
func (s *StreamingService) SubscribeToTrace(ctx context.Context, traceID, projectID uuid.UUID, filter domain.StreamSubscription) *TraceStreamSubscriber {
	stream := s.GetOrCreateStream(traceID, projectID)

	stream.mu.Lock()
	defer stream.mu.Unlock()

	sub := &TraceStreamSubscriber{
		ID:      uuid.New().String(),
		Channel: make(chan *domain.StreamActivity, 200),
		Done:    make(chan struct{}),
		Filter:  filter,
	}

	stream.Subscribers[sub.ID] = sub

	go func() {
		<-ctx.Done()
		s.UnsubscribeFromTrace(traceID, sub.ID)
	}()

	return sub
}

// UnsubscribeFromTrace removes a subscriber from a trace stream
func (s *StreamingService) UnsubscribeFromTrace(traceID uuid.UUID, subscriberID string) {
	s.mu.RLock()
	stream, ok := s.streams[traceID]
	s.mu.RUnlock()

	if !ok {
		return
	}

	stream.mu.Lock()
	if sub, exists := stream.Subscribers[subscriberID]; exists {
		delete(stream.Subscribers, subscriberID)
		close(sub.Done)
		close(sub.Channel)
	}
	stream.mu.Unlock()
}

// PublishActivity publishes an activity to a trace stream
func (s *StreamingService) PublishActivity(ctx context.Context, activity domain.StreamActivity) {
	s.mu.RLock()
	stream, ok := s.streams[activity.TraceID]
	s.mu.RUnlock()

	if !ok {
		return
	}

	stream.mu.Lock()
	// Keep bounded activity log
	if len(stream.Activities) >= s.maxActivities {
		stream.Activities = stream.Activities[1:]
	}
	stream.Activities = append(stream.Activities, activity)

	// Update metrics based on activity type
	s.updateMetrics(stream, &activity)
	stream.mu.Unlock()

	// Fan out to subscribers
	stream.mu.RLock()
	for _, sub := range stream.Subscribers {
		if s.matchesFilter(sub.Filter, activity) {
			select {
			case sub.Channel <- &activity:
			default:
				// Skip slow subscribers
			}
		}
	}
	stream.mu.RUnlock()

	// Also publish to project-level realtime
	s.realtime.Publish(ctx, stream.ProjectID, string(activity.Type), activity)
}

// GetLiveMetrics returns current live metrics for a trace
func (s *StreamingService) GetLiveMetrics(traceID uuid.UUID) *domain.LiveMetrics {
	s.mu.RLock()
	stream, ok := s.streams[traceID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	stream.mu.RLock()
	defer stream.mu.RUnlock()

	metrics := *stream.Metrics
	metrics.ElapsedMs = time.Since(stream.startedAt).Milliseconds()
	if metrics.ElapsedMs > 0 {
		seconds := float64(metrics.ElapsedMs) / 1000.0
		metrics.TokensPerSecond = float64(metrics.TotalTokens) / seconds
		minutes := seconds / 60.0
		if minutes > 0 {
			metrics.CostPerMinute = metrics.TotalCost / minutes
		}
	}
	return &metrics
}

// GetRecentActivities returns recent activities for a trace
func (s *StreamingService) GetRecentActivities(traceID uuid.UUID, limit int) []domain.StreamActivity {
	s.mu.RLock()
	stream, ok := s.streams[traceID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	stream.mu.RLock()
	defer stream.mu.RUnlock()

	activities := stream.Activities
	if limit > 0 && len(activities) > limit {
		activities = activities[len(activities)-limit:]
	}

	result := make([]domain.StreamActivity, len(activities))
	copy(result, activities)
	return result
}

// RequestIntervention sends an intervention request to a running agent
func (s *StreamingService) RequestIntervention(ctx context.Context, req domain.InterventionRequest) error {
	s.mu.Lock()
	s.interventions[req.TraceID] = append(s.interventions[req.TraceID], req)
	s.mu.Unlock()

	// Publish intervention as an activity
	activity := domain.StreamActivity{
		ID:          req.ID.String(),
		TraceID:     req.TraceID,
		Type:        domain.StreamEventIntervention,
		Title:       "Intervention: " + string(req.Action),
		Description: req.Message,
		Timestamp:   req.CreatedAt,
		Status:      "pending",
		Metadata: map[string]any{
			"action": req.Action,
			"userId": req.UserID.String(),
		},
	}
	s.PublishActivity(ctx, activity)

	// Also publish via realtime for SDK consumption
	s.realtime.Publish(ctx, req.ProjectID, "intervention.requested", req)

	return nil
}

// GetPendingInterventions returns pending interventions for a trace
func (s *StreamingService) GetPendingInterventions(traceID uuid.UUID) []domain.InterventionRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	interventions := s.interventions[traceID]
	var pending []domain.InterventionRequest
	for _, i := range interventions {
		if i.Status == "pending" {
			pending = append(pending, i)
		}
	}
	return pending
}

// AcknowledgeIntervention marks an intervention as acknowledged
func (s *StreamingService) AcknowledgeIntervention(traceID, interventionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	interventions := s.interventions[traceID]
	for i := range interventions {
		if interventions[i].ID == interventionID {
			interventions[i].Status = "acknowledged"
			return nil
		}
	}
	return nil
}

// GetActiveStreams returns info about active streams for a project
func (s *StreamingService) GetActiveStreams(projectID uuid.UUID) []domain.LiveMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.LiveMetrics
	for _, stream := range s.streams {
		if stream.ProjectID == projectID {
			stream.mu.RLock()
			m := *stream.Metrics
			m.ElapsedMs = time.Since(stream.startedAt).Milliseconds()
			stream.mu.RUnlock()
			result = append(result, m)
		}
	}
	return result
}

// CleanupStream removes a completed trace stream
func (s *StreamingService) CleanupStream(traceID uuid.UUID) {
	s.mu.Lock()
	stream, ok := s.streams[traceID]
	if ok {
		delete(s.streams, traceID)
	}
	s.mu.Unlock()

	if ok {
		stream.mu.Lock()
		for _, sub := range stream.Subscribers {
			close(sub.Done)
			close(sub.Channel)
		}
		stream.mu.Unlock()
	}
}

func (s *StreamingService) updateMetrics(stream *TraceStream, activity *domain.StreamActivity) {
	m := stream.Metrics
	m.LastUpdated = time.Now()

	switch activity.Type {
	case domain.StreamEventObservationStart:
		m.ActiveSpans++
	case domain.StreamEventObservationEnd:
		if m.ActiveSpans > 0 {
			m.ActiveSpans--
		}
		m.CompletedSpans++
		if tokens, ok := activity.Metadata["tokens"].(float64); ok {
			m.TotalTokens += int(tokens)
		}
		if cost, ok := activity.Metadata["cost"].(float64); ok {
			m.TotalCost += cost
		}
	case domain.StreamEventFileChange:
		m.FilesModified++
	case domain.StreamEventTerminalOutput:
		m.TerminalCommands++
	case domain.StreamEventErrorOccurred:
		m.ErrorCount++
	case domain.StreamEventCostUpdate:
		if cost, ok := activity.Metadata["totalCost"].(float64); ok {
			m.TotalCost = cost
		}
		if tokens, ok := activity.Metadata["totalTokens"].(float64); ok {
			m.TotalTokens = int(tokens)
		}
	}
}

func (s *StreamingService) matchesFilter(filter domain.StreamSubscription, activity domain.StreamActivity) bool {
	if len(filter.EventTypes) == 0 {
		return true
	}
	for _, t := range filter.EventTypes {
		if t == activity.Type {
			return true
		}
	}
	return false
}
