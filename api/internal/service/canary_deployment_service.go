package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CanaryDeploymentService manages canary deployments
type CanaryDeploymentService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	deployments map[uuid.UUID]*domain.CanaryDeployment
}

// NewCanaryDeploymentService creates a new canary deployment service
func NewCanaryDeploymentService(logger *zap.Logger) *CanaryDeploymentService {
	return &CanaryDeploymentService{
		logger:      logger,
		deployments: make(map[uuid.UUID]*domain.CanaryDeployment),
	}
}

var defaultStages = []domain.CanaryStage{
	{Percentage: 5, MinDuration: "15m", AutoPromote: false},
	{Percentage: 25, MinDuration: "1h", AutoPromote: false},
	{Percentage: 50, MinDuration: "2h", AutoPromote: false},
	{Percentage: 100, MinDuration: "0", AutoPromote: false},
}

var defaultCriteria = domain.PromotionCriteria{
	MinSampleSize: 100,
}

// CreateDeployment creates a new canary deployment
func (s *CanaryDeploymentService) CreateDeployment(ctx context.Context, projectID uuid.UUID, input *domain.CanaryDeploymentInput) (*domain.CanaryDeployment, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("deployment name is required")
	}
	if input.BaselineVersion == "" || input.CanaryVersion == "" {
		return nil, fmt.Errorf("both baseline and canary versions are required")
	}

	stages := input.Stages
	if len(stages) == 0 {
		stages = defaultStages
	}

	criteria := defaultCriteria
	if input.Criteria != nil {
		criteria = *input.Criteria
	}

	deployment := &domain.CanaryDeployment{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		Description:     input.Description,
		Status:          domain.CanaryStatusPending,
		BaselineVersion: input.BaselineVersion,
		CanaryVersion:   input.CanaryVersion,
		Stages:          stages,
		CurrentStage:    0,
		Criteria:        criteria,
		Metrics: domain.CanaryMetrics{
			Baseline: domain.CanaryVersionMetrics{},
			Canary:   domain.CanaryVersionMetrics{},
		},
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.mu.Lock()
	s.deployments[deployment.ID] = deployment
	s.mu.Unlock()

	s.logger.Info("canary deployment created",
		zap.String("deploymentId", deployment.ID.String()),
		zap.String("baseline", deployment.BaselineVersion),
		zap.String("canary", deployment.CanaryVersion),
	)

	return deployment, nil
}

// GetDeployment retrieves a deployment by ID
func (s *CanaryDeploymentService) GetDeployment(ctx context.Context, id uuid.UUID) (*domain.CanaryDeployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dep, exists := s.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment not found")
	}
	return dep, nil
}

// ListDeployments lists canary deployments for a project
func (s *CanaryDeploymentService) ListDeployments(ctx context.Context, projectID uuid.UUID) (*domain.CanaryDeploymentList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var deps []domain.CanaryDeployment
	for _, dep := range s.deployments {
		if dep.ProjectID == projectID {
			deps = append(deps, *dep)
		}
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].CreatedAt.After(deps[j].CreatedAt)
	})

	return &domain.CanaryDeploymentList{
		Deployments: deps,
		TotalCount:  len(deps),
	}, nil
}

// Promote advances the canary to the next traffic stage
func (s *CanaryDeploymentService) Promote(ctx context.Context, id uuid.UUID) (*domain.CanaryDeployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep, exists := s.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment not found")
	}

	if dep.Status == domain.CanaryStatusCompleted || dep.Status == domain.CanaryStatusRolledBack {
		return nil, fmt.Errorf("deployment already finished")
	}

	// Check promotion criteria
	if !s.meetsCriteria(dep) {
		return nil, fmt.Errorf("promotion criteria not met: check eval scores, error rates, and latency")
	}

	if dep.CurrentStage == 0 {
		dep.Status = domain.CanaryStatusRunning
	}

	dep.CurrentStage++
	if dep.CurrentStage >= len(dep.Stages) {
		dep.Status = domain.CanaryStatusCompleted
		now := time.Now()
		dep.CompletedAt = &now
		s.logger.Info("canary deployment completed — full rollout",
			zap.String("deploymentId", dep.ID.String()),
		)
	} else {
		dep.Status = domain.CanaryStatusPromoting
		s.logger.Info("canary promoted to next stage",
			zap.String("deploymentId", dep.ID.String()),
			zap.Int("stage", dep.CurrentStage),
			zap.Int("percentage", dep.Stages[dep.CurrentStage].Percentage),
		)
	}

	dep.UpdatedAt = time.Now()
	return dep, nil
}

// Rollback rolls back a canary deployment
func (s *CanaryDeploymentService) Rollback(ctx context.Context, id uuid.UUID) (*domain.CanaryDeployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep, exists := s.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment not found")
	}

	dep.Status = domain.CanaryStatusRolledBack
	now := time.Now()
	dep.CompletedAt = &now
	dep.UpdatedAt = now

	s.logger.Warn("canary deployment rolled back",
		zap.String("deploymentId", dep.ID.String()),
		zap.String("canaryVersion", dep.CanaryVersion),
	)

	return dep, nil
}

// GetMetrics returns comparison metrics for a deployment
func (s *CanaryDeploymentService) GetMetrics(ctx context.Context, id uuid.UUID) (*domain.CanaryMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dep, exists := s.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment not found")
	}

	return &dep.Metrics, nil
}

// GetActiveVersion returns the currently active agent version
func (s *CanaryDeploymentService) GetActiveVersion(ctx context.Context, projectID uuid.UUID) (*domain.ActiveVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, dep := range s.deployments {
		if dep.ProjectID == projectID && (dep.Status == domain.CanaryStatusRunning || dep.Status == domain.CanaryStatusPromoting) {
			percentage := 0
			if dep.CurrentStage < len(dep.Stages) {
				percentage = dep.Stages[dep.CurrentStage].Percentage
			}
			return &domain.ActiveVersion{
				Version:      dep.CanaryVersion,
				IsCanary:     true,
				Percentage:   percentage,
				DeploymentID: &dep.ID,
			}, nil
		}
	}

	return &domain.ActiveVersion{
		Version:    "stable",
		IsCanary:   false,
		Percentage: 100,
	}, nil
}

func (s *CanaryDeploymentService) meetsCriteria(dep *domain.CanaryDeployment) bool {
	c := dep.Criteria
	m := dep.Metrics.Canary

	if c.MinEvalScore != nil && m.AvgEvalScore < *c.MinEvalScore {
		return false
	}
	if c.MaxErrorRate != nil && m.ErrorRate > *c.MaxErrorRate {
		return false
	}
	if c.MaxLatencyMs != nil && m.AvgLatencyMs > float64(*c.MaxLatencyMs) {
		return false
	}

	return true
}
