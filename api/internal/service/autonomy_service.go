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

type AutonomyService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	configs map[string]map[string]*domain.AutonomyConfig // projectID -> agentName -> config
}

func NewAutonomyService(logger *zap.Logger) *AutonomyService {
	return &AutonomyService{
		logger:  logger,
		configs: make(map[string]map[string]*domain.AutonomyConfig),
	}
}

func (s *AutonomyService) SetAutonomy(ctx context.Context, projectID uuid.UUID, input domain.AutonomyConfigInput) (*domain.AutonomyConfig, error) {
	if !input.Level.IsValid() {
		return nil, fmt.Errorf("invalid autonomy level: %s", input.Level)
	}

	perms := s.defaultPermissions(input.Level)
	if input.Permissions != nil {
		perms = *input.Permissions
	}

	now := time.Now()
	config := &domain.AutonomyConfig{
		ID:          uuid.New(),
		ProjectID:   projectID,
		AgentName:   input.AgentName,
		Level:       input.Level,
		Permissions: perms,
		TrustScore:  s.defaultTrustScore(input.Level),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	if s.configs[projectID.String()] == nil {
		s.configs[projectID.String()] = make(map[string]*domain.AutonomyConfig)
	}
	s.configs[projectID.String()][input.AgentName] = config
	s.mu.Unlock()

	s.logger.Info("set autonomy config", zap.String("projectId", projectID.String()), zap.String("agent", input.AgentName))
	return config, nil
}

func (s *AutonomyService) GetConfig(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.AutonomyConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if agents, ok := s.configs[projectID.String()]; ok {
		if config, ok := agents[agentName]; ok {
			return config, nil
		}
	}
	return nil, fmt.Errorf("autonomy config not found for agent %s", agentName)
}

func (s *AutonomyService) GetDashboard(ctx context.Context, projectID uuid.UUID) (*domain.AutonomyDashboard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dashboard := &domain.AutonomyDashboard{
		ProjectID:    projectID,
		Agents:       []domain.AutonomyConfig{},
		Distribution: make(map[domain.AutonomyLevel]int),
	}

	agents := s.configs[projectID.String()]
	var totalTrust float64
	for _, config := range agents {
		dashboard.Agents = append(dashboard.Agents, *config)
		dashboard.Distribution[config.Level]++
		totalTrust += config.TrustScore
	}
	if len(agents) > 0 {
		dashboard.AvgTrust = totalTrust / float64(len(agents))
	}

	return dashboard, nil
}

func (s *AutonomyService) GetTrustEvolution(ctx context.Context, projectID uuid.UUID, agentName string) (*domain.TrustEvolution, error) {
	s.mu.RLock()
	config, exists := s.configs[projectID.String()][agentName]
	s.mu.RUnlock()

	currentScore := 50.0
	if exists {
		currentScore = config.TrustScore
	}

	now := time.Now()
	history := []domain.TrustDataPoint{
		{Timestamp: now.Add(-7 * 24 * time.Hour), TrustScore: currentScore - 15, Level: domain.AutonomyManual, Reason: "initial deployment"},
		{Timestamp: now.Add(-5 * 24 * time.Hour), TrustScore: currentScore - 10, Level: domain.AutonomySupervised, Reason: "successful task completions"},
		{Timestamp: now.Add(-3 * 24 * time.Hour), TrustScore: currentScore - 5, Level: domain.AutonomyHumanGuided, Reason: "consistent performance"},
		{Timestamp: now, TrustScore: currentScore, Level: domain.AutonomyFullAuto, Reason: "trust threshold met"},
	}

	return &domain.TrustEvolution{
		AgentName: agentName,
		History:   history,
		Current:   currentScore,
		Trend:     "improving",
	}, nil
}

func (s *AutonomyService) defaultPermissions(level domain.AutonomyLevel) domain.AutonomyPermissions {
	switch level {
	case domain.AutonomyFullAuto:
		return domain.AutonomyPermissions{
			CanWriteFiles: true, CanDeleteFiles: true, CanExecuteCommands: true,
			CanAccessNetwork: true, CanModifyConfig: true, RequiresApproval: false, MaxCostPerRun: 100.0,
		}
	case domain.AutonomyHumanGuided:
		return domain.AutonomyPermissions{
			CanWriteFiles: true, CanDeleteFiles: false, CanExecuteCommands: true,
			CanAccessNetwork: true, CanModifyConfig: false, RequiresApproval: false, MaxCostPerRun: 50.0,
		}
	case domain.AutonomySupervised:
		return domain.AutonomyPermissions{
			CanWriteFiles: true, CanDeleteFiles: false, CanExecuteCommands: false,
			CanAccessNetwork: false, CanModifyConfig: false, RequiresApproval: true, MaxCostPerRun: 20.0,
		}
	default: // manual
		return domain.AutonomyPermissions{
			CanWriteFiles: false, CanDeleteFiles: false, CanExecuteCommands: false,
			CanAccessNetwork: false, CanModifyConfig: false, RequiresApproval: true, MaxCostPerRun: 5.0,
		}
	}
}

func (s *AutonomyService) defaultTrustScore(level domain.AutonomyLevel) float64 {
	switch level {
	case domain.AutonomyFullAuto:
		return 90.0
	case domain.AutonomyHumanGuided:
		return 70.0
	case domain.AutonomySupervised:
		return 50.0
	default:
		return 25.0
	}
}
