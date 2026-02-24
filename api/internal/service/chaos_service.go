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

// ChaosService manages chaos engineering experiments for agents
type ChaosService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	experiments map[uuid.UUID]*domain.ChaosExperiment
}

// NewChaosService creates a new chaos service
func NewChaosService(logger *zap.Logger) *ChaosService {
	return &ChaosService{
		logger:      logger,
		experiments: make(map[uuid.UUID]*domain.ChaosExperiment),
	}
}

// CreateExperiment creates a new chaos experiment
func (s *ChaosService) CreateExperiment(ctx context.Context, projectID uuid.UUID, input *domain.ChaosExperimentInput) (*domain.ChaosExperiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	experiment := &domain.ChaosExperiment{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      input.Name,
		Status:    domain.ChaosStatusDraft,
		Faults:    input.Faults,
		CreatedAt: time.Now(),
	}

	for i := range experiment.Faults {
		if experiment.Faults[i].ID == uuid.Nil {
			experiment.Faults[i].ID = uuid.New()
		}
	}

	s.experiments[experiment.ID] = experiment
	s.logger.Info("created chaos experiment", zap.String("id", experiment.ID.String()), zap.String("name", input.Name))
	return experiment, nil
}

// RunExperiment runs a chaos experiment and generates mock results
func (s *ChaosService) RunExperiment(ctx context.Context, projectID uuid.UUID, experimentID uuid.UUID) (*domain.ChaosExperiment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp, ok := s.experiments[experimentID]
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}
	if exp.ProjectID != projectID {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	now := time.Now()
	exp.Status = domain.ChaosStatusCompleted
	exp.CompletedAt = &now
	exp.Results = &domain.ChaosResults{
		Resilience:          78.5,
		GracefulDegradation: true,
		ErrorRecovery:       true,
		FallbackBehavior:    "Agent switched to fallback model and reduced context window",
		Metrics: map[string]float64{
			"response_time_increase_pct": 45.2,
			"error_rate_under_fault":     0.12,
			"recovery_time_seconds":      3.5,
			"quality_degradation_pct":    15.0,
		},
	}

	s.logger.Info("completed chaos experiment", zap.String("id", experimentID.String()))
	return exp, nil
}

// GetExperiment returns a specific chaos experiment
func (s *ChaosService) GetExperiment(ctx context.Context, projectID uuid.UUID, experimentID uuid.UUID) (*domain.ChaosExperiment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, ok := s.experiments[experimentID]
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}
	if exp.ProjectID != projectID {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	return exp, nil
}

// ListExperiments returns all chaos experiments for a project
func (s *ChaosService) ListExperiments(ctx context.Context, projectID uuid.UUID) ([]*domain.ChaosExperiment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.ChaosExperiment
	for _, exp := range s.experiments {
		if exp.ProjectID == projectID {
			result = append(result, exp)
		}
	}

	if result == nil {
		result = []*domain.ChaosExperiment{}
	}
	return result, nil
}

// GetResilienceScorecard returns a resilience scorecard for an agent
func (s *ChaosService) GetResilienceScorecard(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.ResilienceScorecard, error) {
	s.logger.Debug("getting resilience scorecard", zap.String("agent", agentName))

	return &domain.ResilienceScorecard{
		AgentName:    agentName,
		OverallScore: 82.5,
		Scores: map[string]float64{
			"latency_resilience":    85.0,
			"error_handling":        90.0,
			"rate_limit_handling":   75.0,
			"model_fallback":        80.0,
			"context_management":    82.0,
		},
		TestedScenarios: 12,
		PassedScenarios: 10,
		Recommendations: []string{
			"Improve rate limit handling with exponential backoff",
			"Add circuit breaker for external API calls",
			"Implement graceful degradation for context truncation",
		},
	}, nil
}
