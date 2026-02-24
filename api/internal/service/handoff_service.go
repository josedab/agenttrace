package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// HandoffService manages agent handoff operations
type HandoffService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	handoffs map[string]*domain.Handoff
}

// NewHandoffService creates a new handoff service
func NewHandoffService(logger *zap.Logger) *HandoffService {
	return &HandoffService{
		logger:   logger,
		handoffs: make(map[string]*domain.Handoff),
	}
}

// InitiateHandoff creates a new agent handoff
func (s *HandoffService) InitiateHandoff(ctx context.Context, projectID string, input *domain.HandoffInput) (*domain.Handoff, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	priority := input.Priority
	if priority == "" {
		priority = domain.HandoffPriorityNormal
	}

	handoff := &domain.Handoff{
		ID:              fmt.Sprintf("ho_%d", time.Now().UnixNano()),
		ProjectID:       projectID,
		FromAgent:       input.FromAgent,
		ToAgent:         input.ToAgent,
		TaskDescription: input.TaskDescription,
		Context:         input.Context,
		Priority:        priority,
		Status:          domain.HandoffStatusInitiated,
		CreatedAt:       time.Now(),
	}

	s.handoffs[handoff.ID] = handoff
	s.logger.Info("initiated handoff",
		zap.String("id", handoff.ID),
		zap.String("from", handoff.FromAgent),
		zap.String("to", handoff.ToAgent),
	)
	return handoff, nil
}

// AcceptHandoff marks a handoff as accepted
func (s *HandoffService) AcceptHandoff(ctx context.Context, handoffID string) (*domain.Handoff, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	handoff, ok := s.handoffs[handoffID]
	if !ok {
		return nil, fmt.Errorf("handoff not found: %s", handoffID)
	}

	if handoff.Status != domain.HandoffStatusInitiated {
		return nil, fmt.Errorf("handoff cannot be accepted in status: %s", handoff.Status)
	}

	handoff.Status = domain.HandoffStatusAccepted
	s.logger.Info("accepted handoff", zap.String("id", handoffID))
	return handoff, nil
}

// CompleteHandoff marks a handoff as completed with simulated metrics
func (s *HandoffService) CompleteHandoff(ctx context.Context, handoffID string) (*domain.Handoff, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	handoff, ok := s.handoffs[handoffID]
	if !ok {
		return nil, fmt.Errorf("handoff not found: %s", handoffID)
	}

	if handoff.Status != domain.HandoffStatusAccepted {
		return nil, fmt.Errorf("handoff cannot be completed in status: %s", handoff.Status)
	}

	now := time.Now()
	handoff.Status = domain.HandoffStatusCompleted
	handoff.CompletedAt = &now
	handoff.ResolutionTimeMs = now.Sub(handoff.CreatedAt).Milliseconds()

	// Simulate context preservation based on priority
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch handoff.Priority {
	case domain.HandoffPriorityLow, domain.HandoffPriorityNormal:
		handoff.ContextPreservationPct = 90.0 + r.Float64()*10.0 // 90-100%
	case domain.HandoffPriorityHigh, domain.HandoffPriorityCritical:
		handoff.ContextPreservationPct = 60.0 + r.Float64()*20.0 // 60-80%
	}

	s.logger.Info("completed handoff",
		zap.String("id", handoffID),
		zap.Float64("preservationPct", handoff.ContextPreservationPct),
	)
	return handoff, nil
}

// GetChain returns the handoff chain for a trace
func (s *HandoffService) GetChain(ctx context.Context, traceID string) (*domain.HandoffChain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var handoffs []domain.Handoff
	agents := make(map[string]bool)
	var totalPreservation float64
	var totalResolution int64
	failures := 0
	preserved := 0

	for _, h := range s.handoffs {
		handoffs = append(handoffs, *h)
		agents[h.FromAgent] = true
		agents[h.ToAgent] = true
		if h.Status == domain.HandoffStatusFailed {
			failures++
		}
		if h.ContextPreservationPct > 0 {
			totalPreservation += h.ContextPreservationPct
			preserved++
		}
		if h.ResolutionTimeMs > 0 {
			totalResolution += h.ResolutionTimeMs
		}
	}

	var avgPreservation float64
	if preserved > 0 {
		avgPreservation = totalPreservation / float64(preserved)
	}
	var avgResolution int64
	if len(handoffs) > 0 {
		avgResolution = totalResolution / int64(len(handoffs))
	}

	chain := &domain.HandoffChain{
		ID:              fmt.Sprintf("chain_%s", traceID),
		Handoffs:        handoffs,
		TotalAgents:     len(agents),
		AvgPreservation: avgPreservation,
		AvgResolutionMs: avgResolution,
		Failures:        failures,
	}

	return chain, nil
}

// GetStats returns handoff statistics for a project
func (s *HandoffService) GetStats(ctx context.Context, projectID string) (*domain.HandoffStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &domain.HandoffStats{
		ByPriority: make(map[string]int),
	}

	var totalPreservation float64
	var totalResolution int64
	completed := 0
	preserved := 0

	for _, h := range s.handoffs {
		if h.ProjectID != projectID && projectID != "" {
			continue
		}
		stats.TotalHandoffs++
		stats.ByPriority[string(h.Priority)]++
		if h.Status == domain.HandoffStatusCompleted {
			completed++
		}
		if h.ContextPreservationPct > 0 {
			totalPreservation += h.ContextPreservationPct
			preserved++
		}
		totalResolution += h.ResolutionTimeMs
	}

	if stats.TotalHandoffs > 0 {
		stats.SuccessRate = float64(completed) / float64(stats.TotalHandoffs) * 100.0
		stats.AvgResolutionMs = totalResolution / int64(stats.TotalHandoffs)
	}
	if preserved > 0 {
		stats.AvgPreservation = totalPreservation / float64(preserved)
	}

	return stats, nil
}
