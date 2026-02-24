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

// CollabPatternService manages agent collaboration patterns
type CollabPatternService struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	patterns    map[uuid.UUID]*domain.CollabPattern
	deployments map[uuid.UUID]*domain.PatternDeployment
}

// NewCollabPatternService creates a new collaboration pattern service with pre-seeded patterns
func NewCollabPatternService(logger *zap.Logger) *CollabPatternService {
	svc := &CollabPatternService{
		logger:      logger,
		patterns:    make(map[uuid.UUID]*domain.CollabPattern),
		deployments: make(map[uuid.UUID]*domain.PatternDeployment),
	}
	svc.seedPatterns()
	return svc
}

// ListPatterns returns all available collaboration patterns
func (s *CollabPatternService) ListPatterns(ctx context.Context) ([]domain.CollabPattern, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var patterns []domain.CollabPattern
	for _, p := range s.patterns {
		patterns = append(patterns, *p)
	}
	if patterns == nil {
		patterns = []domain.CollabPattern{}
	}
	return patterns, nil
}

// GetPattern returns a specific pattern by ID
func (s *CollabPatternService) GetPattern(ctx context.Context, patternID string) (*domain.CollabPattern, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pid, _ := uuid.Parse(patternID)
	if p, ok := s.patterns[pid]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("pattern not found: %s", patternID)
}

// DeployPattern deploys a collaboration pattern to a project
func (s *CollabPatternService) DeployPattern(ctx context.Context, projectID string, input *domain.DeployPatternInput) (*domain.PatternDeployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.patterns[input.PatternID]; !ok {
		return nil, fmt.Errorf("pattern not found: %s", input.PatternID)
	}

	pid, _ := uuid.Parse(projectID)
	deployment := &domain.PatternDeployment{
		ID:         uuid.New(),
		PatternID:  input.PatternID,
		ProjectID:  pid,
		Config:     input.Config,
		Status:     "deployed",
		DeployedAt: time.Now(),
	}

	s.deployments[deployment.ID] = deployment

	// Increment deploy count
	if p, ok := s.patterns[input.PatternID]; ok {
		p.DeployCount++
	}

	s.logger.Info("deployed pattern", zap.String("patternId", input.PatternID.String()), zap.String("projectId", projectID))
	return deployment, nil
}

// GetDeployments returns all pattern deployments for a project
func (s *CollabPatternService) GetDeployments(ctx context.Context, projectID string) ([]domain.PatternDeployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pid, _ := uuid.Parse(projectID)
	var deployments []domain.PatternDeployment
	for _, d := range s.deployments {
		if d.ProjectID == pid {
			deployments = append(deployments, *d)
		}
	}
	if deployments == nil {
		deployments = []domain.PatternDeployment{}
	}
	return deployments, nil
}

// GetPatternAnalytics returns analytics for a specific pattern
func (s *CollabPatternService) GetPatternAnalytics(ctx context.Context, patternID string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pid, _ := uuid.Parse(patternID)
	p, ok := s.patterns[pid]
	if !ok {
		return nil, fmt.Errorf("pattern not found: %s", patternID)
	}

	return map[string]any{
		"patternId":      p.ID,
		"name":           p.Name,
		"deployCount":    p.DeployCount,
		"avgPerformance": p.AvgPerformance,
		"complexity":     p.Complexity,
		"avgLatencyMs":   245.0,
		"successRate":    0.94,
		"errorRate":      0.06,
		"topUseCases":    p.UseCases,
	}, nil
}

func (s *CollabPatternService) seedPatterns() {
	patterns := []domain.CollabPattern{
		{
			ID:          uuid.New(),
			Name:        "Coordinator Pattern",
			Description: "Central coordinator dispatches tasks to specialized agents and aggregates results",
			Type:        "coordinator",
			AgentRoles: []domain.PatternRole{
				{Name: "Coordinator", Type: "orchestrator", Responsibilities: []string{"task decomposition", "result aggregation", "error handling"}, Model: "gpt-4"},
				{Name: "Researcher", Type: "worker", Responsibilities: []string{"information gathering", "fact checking"}, Model: "claude-3-sonnet"},
				{Name: "Writer", Type: "worker", Responsibilities: []string{"content generation", "formatting"}, Model: "gpt-4"},
			},
			MessageFlow: []domain.PatternMessage{
				{From: "Coordinator", To: "Researcher", Type: "request", Description: "Research task assignment"},
				{From: "Researcher", To: "Coordinator", Type: "response", Description: "Research results"},
				{From: "Coordinator", To: "Writer", Type: "request", Description: "Writing task with context"},
				{From: "Writer", To: "Coordinator", Type: "response", Description: "Generated content"},
			},
			Complexity:     "moderate",
			UseCases:       []string{"report generation", "content creation", "data analysis"},
			DeployCount:    156,
			AvgPerformance: 0.89,
			CreatedAt:      time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Pipeline Pattern",
			Description: "Sequential processing where each agent transforms and passes data to the next",
			Type:        "pipeline",
			AgentRoles: []domain.PatternRole{
				{Name: "Extractor", Type: "processor", Responsibilities: []string{"data extraction", "parsing"}, Model: "gpt-3.5-turbo"},
				{Name: "Transformer", Type: "processor", Responsibilities: []string{"data transformation", "enrichment"}, Model: "claude-3-haiku"},
				{Name: "Validator", Type: "processor", Responsibilities: []string{"quality validation", "error detection"}, Model: "gpt-4"},
			},
			MessageFlow: []domain.PatternMessage{
				{From: "Extractor", To: "Transformer", Type: "request", Description: "Extracted data"},
				{From: "Transformer", To: "Validator", Type: "request", Description: "Transformed data"},
				{From: "Validator", To: "Extractor", Type: "response", Description: "Validation result"},
			},
			Complexity:     "simple",
			UseCases:       []string{"ETL pipelines", "document processing", "data migration"},
			DeployCount:    234,
			AvgPerformance: 0.92,
			CreatedAt:      time.Now().Add(-45 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Voting Pattern",
			Description: "Multiple agents independently solve a problem and results are aggregated by majority vote",
			Type:        "voting",
			AgentRoles: []domain.PatternRole{
				{Name: "Voter-1", Type: "voter", Responsibilities: []string{"independent analysis", "solution generation"}, Model: "gpt-4"},
				{Name: "Voter-2", Type: "voter", Responsibilities: []string{"independent analysis", "solution generation"}, Model: "claude-3-opus"},
				{Name: "Voter-3", Type: "voter", Responsibilities: []string{"independent analysis", "solution generation"}, Model: "gemini-pro"},
				{Name: "Aggregator", Type: "aggregator", Responsibilities: []string{"vote counting", "consensus building"}, Model: "gpt-4"},
			},
			MessageFlow: []domain.PatternMessage{
				{From: "Aggregator", To: "Voter-1", Type: "broadcast", Description: "Problem statement"},
				{From: "Voter-1", To: "Aggregator", Type: "vote", Description: "Solution vote"},
				{From: "Voter-2", To: "Aggregator", Type: "vote", Description: "Solution vote"},
				{From: "Voter-3", To: "Aggregator", Type: "vote", Description: "Solution vote"},
			},
			Complexity:     "moderate",
			UseCases:       []string{"decision making", "classification", "quality assurance"},
			DeployCount:    89,
			AvgPerformance: 0.95,
			CreatedAt:      time.Now().Add(-20 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Debate Pattern",
			Description: "Agents argue different perspectives and a judge synthesizes the best answer",
			Type:        "debate",
			AgentRoles: []domain.PatternRole{
				{Name: "Proponent", Type: "debater", Responsibilities: []string{"argue for", "provide evidence"}, Model: "gpt-4"},
				{Name: "Opponent", Type: "debater", Responsibilities: []string{"argue against", "find weaknesses"}, Model: "claude-3-opus"},
				{Name: "Judge", Type: "judge", Responsibilities: []string{"evaluate arguments", "synthesize answer"}, Model: "gpt-4"},
			},
			MessageFlow: []domain.PatternMessage{
				{From: "Judge", To: "Proponent", Type: "request", Description: "Initial prompt"},
				{From: "Proponent", To: "Judge", Type: "response", Description: "Argument for"},
				{From: "Judge", To: "Opponent", Type: "request", Description: "Counter-argument request"},
				{From: "Opponent", To: "Judge", Type: "response", Description: "Counter-argument"},
			},
			Complexity:     "complex",
			UseCases:       []string{"critical analysis", "policy evaluation", "risk assessment"},
			DeployCount:    67,
			AvgPerformance: 0.91,
			CreatedAt:      time.Now().Add(-15 * 24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "Swarm Pattern",
			Description: "Emergent behavior from many simple agents working on sub-tasks in parallel",
			Type:        "swarm",
			AgentRoles: []domain.PatternRole{
				{Name: "Queen", Type: "coordinator", Responsibilities: []string{"task decomposition", "swarm management"}, Model: "gpt-4"},
				{Name: "Worker-Pool", Type: "swarm", Responsibilities: []string{"sub-task execution", "local optimization"}, Model: "gpt-3.5-turbo"},
			},
			MessageFlow: []domain.PatternMessage{
				{From: "Queen", To: "Worker-Pool", Type: "broadcast", Description: "Sub-task distribution"},
				{From: "Worker-Pool", To: "Queen", Type: "response", Description: "Sub-task results"},
			},
			Complexity:     "complex",
			UseCases:       []string{"large-scale data processing", "distributed search", "parallel exploration"},
			DeployCount:    45,
			AvgPerformance: 0.87,
			CreatedAt:      time.Now().Add(-10 * 24 * time.Hour),
		},
	}

	for i := range patterns {
		s.patterns[patterns[i].ID] = &patterns[i]
	}
}
