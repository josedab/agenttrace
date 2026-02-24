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

// FederatedLearningService manages federated learning from traces
type FederatedLearningService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	rings   map[uuid.UUID]*domain.FederationRing
	configs map[string]*domain.FederationConfig // projectID -> config
}

// NewFederatedLearningService creates a new federated learning service with pre-seeded rings
func NewFederatedLearningService(logger *zap.Logger) *FederatedLearningService {
	svc := &FederatedLearningService{
		logger:  logger,
		rings:   make(map[uuid.UUID]*domain.FederationRing),
		configs: make(map[string]*domain.FederationConfig),
	}
	svc.seedRings()
	return svc
}

// JoinRing joins or creates a federation ring
func (s *FederatedLearningService) JoinRing(ctx context.Context, projectID string, input *domain.FederationJoinInput) (*domain.FederationRing, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find existing ring by name
	for _, ring := range s.rings {
		if ring.Name == input.RingName {
			ring.Participants++
			s.logger.Info("joined federation ring", zap.String("ring", ring.Name), zap.Int("participants", ring.Participants))
			return ring, nil
		}
	}

	// Create new ring
	ring := &domain.FederationRing{
		ID:                 uuid.New(),
		Name:               input.RingName,
		Participants:       1,
		Status:             "active",
		PrivacyLevel:       input.PrivacyLevel,
		AggregatedInsights: []domain.FederatedInsight{},
		CreatedAt:          time.Now(),
	}

	s.rings[ring.ID] = ring
	s.logger.Info("created federation ring", zap.String("ring", ring.Name))
	return ring, nil
}

// ListRings returns all federation rings
func (s *FederatedLearningService) ListRings(ctx context.Context) ([]domain.FederationRing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rings []domain.FederationRing
	for _, r := range s.rings {
		rings = append(rings, *r)
	}
	if rings == nil {
		rings = []domain.FederationRing{}
	}
	return rings, nil
}

// GetInsights returns insights for a specific federation ring
func (s *FederatedLearningService) GetInsights(ctx context.Context, ringID string) ([]domain.FederatedInsight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rid, _ := uuid.Parse(ringID)
	if ring, ok := s.rings[rid]; ok {
		if ring.AggregatedInsights == nil {
			return []domain.FederatedInsight{}, nil
		}
		return ring.AggregatedInsights, nil
	}
	return nil, fmt.Errorf("ring not found: %s", ringID)
}

// GetConfig returns the federation configuration for a project
func (s *FederatedLearningService) GetConfig(ctx context.Context, projectID string) (*domain.FederationConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if config, ok := s.configs[projectID]; ok {
		return config, nil
	}

	pid, _ := uuid.Parse(projectID)
	return &domain.FederationConfig{
		ProjectID:                  pid,
		Enabled:                    false,
		ParticipatingRings:         []uuid.UUID{},
		SharingCategories:          []string{},
		DifferentialPrivacyEpsilon: 1.0,
	}, nil
}

// UpdateConfig updates the federation configuration for a project
func (s *FederatedLearningService) UpdateConfig(ctx context.Context, projectID string, config *domain.FederationConfig) (*domain.FederationConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid, _ := uuid.Parse(projectID)
	config.ProjectID = pid
	s.configs[projectID] = config
	s.logger.Info("updated federation config", zap.String("projectId", projectID), zap.Bool("enabled", config.Enabled))
	return config, nil
}

func (s *FederatedLearningService) seedRings() {
	ring1ID := uuid.New()
	ring2ID := uuid.New()

	s.rings[ring1ID] = &domain.FederationRing{
		ID:           ring1ID,
		Name:         "Global Prompt Optimization",
		Participants: 47,
		Status:       "active",
		PrivacyLevel: "strict",
		AggregatedInsights: []domain.FederatedInsight{
			{ID: uuid.New(), RingID: ring1ID, Category: "prompt_optimization", Insight: "Chain-of-thought prompts improve accuracy by 23% across participants", Confidence: 0.87, ContributingOrgs: 34, GeneratedAt: time.Now().Add(-2 * time.Hour)},
			{ID: uuid.New(), RingID: ring1ID, Category: "cost_reduction", Insight: "Prompt caching reduces costs by 31% for repeated query patterns", Confidence: 0.92, ContributingOrgs: 41, GeneratedAt: time.Now().Add(-6 * time.Hour)},
			{ID: uuid.New(), RingID: ring1ID, Category: "model_selection", Insight: "GPT-3.5 handles 78% of queries at equivalent quality to GPT-4", Confidence: 0.79, ContributingOrgs: 28, GeneratedAt: time.Now().Add(-12 * time.Hour)},
		},
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
	}

	s.rings[ring2ID] = &domain.FederationRing{
		ID:           ring2ID,
		Name:         "Error Pattern Detection",
		Participants: 23,
		Status:       "active",
		PrivacyLevel: "moderate",
		AggregatedInsights: []domain.FederatedInsight{
			{ID: uuid.New(), RingID: ring2ID, Category: "error_pattern", Insight: "Token limit errors spike during business hours across 67% of participants", Confidence: 0.91, ContributingOrgs: 19, GeneratedAt: time.Now().Add(-1 * time.Hour)},
			{ID: uuid.New(), RingID: ring2ID, Category: "error_pattern", Insight: "Rate limiting errors correlate with batch processing schedules", Confidence: 0.84, ContributingOrgs: 15, GeneratedAt: time.Now().Add(-4 * time.Hour)},
		},
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
	}
}
