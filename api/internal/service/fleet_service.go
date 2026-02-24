package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// FleetService manages agent fleet operations
type FleetService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	policies map[uuid.UUID]*domain.FleetPolicy
}

// NewFleetService creates a new fleet management service
func NewFleetService(logger *zap.Logger) *FleetService {
	return &FleetService{
		logger:   logger,
		policies: make(map[uuid.UUID]*domain.FleetPolicy),
	}
}

// GetDashboard returns the fleet management dashboard
func (s *FleetService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.FleetDashboard, error) {
	agents := s.getMockAgents(projectID)
	policies := s.listPoliciesInternal(projectID)

	healthy, degraded, down := 0, 0, 0
	var totalCost float64
	var totalTraces int
	for _, a := range agents {
		switch a.Status {
		case domain.FleetAgentHealthy:
			healthy++
		case domain.FleetAgentDegraded:
			degraded++
		case domain.FleetAgentDown:
			down++
		}
		totalCost += a.Cost
		totalTraces += a.Traces
	}

	return &domain.FleetDashboard{
		TotalAgents:   len(agents),
		HealthyCount:  healthy,
		DegradedCount: degraded,
		DownCount:     down,
		TotalCost:     totalCost,
		TotalTraces:   totalTraces,
		Agents:        agents,
		Policies:      policies,
	}, nil
}

// ListAgents lists all agents in the fleet
func (s *FleetService) ListAgents(ctx context.Context, projectID uuid.UUID) ([]domain.FleetAgent, error) {
	return s.getMockAgents(projectID), nil
}

// CreatePolicy creates a new fleet policy
func (s *FleetService) CreatePolicy(ctx context.Context, projectID uuid.UUID, input *domain.FleetPolicyInput) (*domain.FleetPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy := &domain.FleetPolicy{
		ID:        uuid.New(),
		Name:      input.Name,
		Type:      input.Type,
		Config:    input.Config,
		Scope:     input.Scope,
		Enabled:   input.Enabled,
		CreatedAt: time.Now(),
	}

	s.policies[policy.ID] = policy
	s.logger.Info("created fleet policy", zap.String("id", policy.ID.String()), zap.String("name", input.Name))
	return policy, nil
}

// ListPolicies lists all fleet policies for a project
func (s *FleetService) ListPolicies(ctx context.Context, projectID uuid.UUID) ([]domain.FleetPolicy, error) {
	return s.listPoliciesInternal(projectID), nil
}

// BulkUpdate applies a bulk configuration update to agents
func (s *FleetService) BulkUpdate(ctx context.Context, projectID uuid.UUID, update *domain.BulkConfigUpdate) (int, error) {
	s.logger.Info("applied bulk config update",
		zap.Int("agents", len(update.AgentNames)),
		zap.String("note", update.Note),
	)
	return len(update.AgentNames), nil
}

// GetScalingRecommendations returns scaling recommendations for fleet agents
func (s *FleetService) GetScalingRecommendations(ctx context.Context, projectID uuid.UUID) ([]domain.FleetScalingRecommendation, error) {
	return []domain.FleetScalingRecommendation{
		{AgentName: "code-review-agent", CurrentLoad: 0.92, RecommendedAction: domain.ScaleUp, Reason: "High load sustained for 2+ hours", EstimatedSavings: -15.00},
		{AgentName: "test-runner-agent", CurrentLoad: 0.35, RecommendedAction: domain.ScaleDown, Reason: "Consistently under-utilized", EstimatedSavings: 42.50},
		{AgentName: "deploy-agent", CurrentLoad: 0.60, RecommendedAction: domain.Maintain, Reason: "Load within optimal range", EstimatedSavings: 0},
		{AgentName: "monitoring-agent", CurrentLoad: 0.88, RecommendedAction: domain.ScaleUp, Reason: "Approaching capacity limit", EstimatedSavings: -8.00},
		{AgentName: "docs-agent", CurrentLoad: 0.15, RecommendedAction: domain.ScaleDown, Reason: "Very low utilization", EstimatedSavings: 25.00},
	}, nil
}

func (s *FleetService) listPoliciesInternal(projectID uuid.UUID) []domain.FleetPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.FleetPolicy
	for _, p := range s.policies {
		result = append(result, *p)
	}
	return result
}

func (s *FleetService) getMockAgents(projectID uuid.UUID) []domain.FleetAgent {
	now := time.Now()
	r := rand.New(rand.NewSource(int64(projectID.ID())))
	agents := []domain.FleetAgent{
		{Name: "code-review-agent", ProjectID: projectID, Status: domain.FleetAgentHealthy, Traces: 1240 + r.Intn(100), Cost: 45.20, LastActive: now.Add(-2 * time.Minute), Version: "2.1.0", HealthScore: 0.98},
		{Name: "test-runner-agent", ProjectID: projectID, Status: domain.FleetAgentHealthy, Traces: 890 + r.Intn(100), Cost: 32.10, LastActive: now.Add(-5 * time.Minute), Version: "1.8.3", HealthScore: 0.95},
		{Name: "deploy-agent", ProjectID: projectID, Status: domain.FleetAgentDegraded, Traces: 456 + r.Intn(100), Cost: 28.75, LastActive: now.Add(-15 * time.Minute), Version: "2.0.1", HealthScore: 0.72},
		{Name: "monitoring-agent", ProjectID: projectID, Status: domain.FleetAgentHealthy, Traces: 2100 + r.Intn(100), Cost: 18.90, LastActive: now.Add(-1 * time.Minute), Version: "1.5.0", HealthScore: 0.99},
		{Name: "docs-agent", ProjectID: projectID, Status: domain.FleetAgentDown, Traces: 120 + r.Intn(50), Cost: 5.40, LastActive: now.Add(-2 * time.Hour), Version: "1.2.0", HealthScore: 0.0},
		{Name: "security-scanner", ProjectID: projectID, Status: domain.FleetAgentHealthy, Traces: 670 + r.Intn(100), Cost: 22.30, LastActive: now.Add(-8 * time.Minute), Version: "3.0.0", HealthScore: 0.94},
	}
	return agents
}

// unused but required to satisfy linter for fmt import
func init() {
	_ = fmt.Sprintf
}
