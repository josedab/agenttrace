package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AgentBuilderService manages natural language agent blueprint generation
type AgentBuilderService struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	blueprints map[uuid.UUID]*domain.AgentBlueprint
}

// NewAgentBuilderService creates a new agent builder service
func NewAgentBuilderService(logger *zap.Logger) *AgentBuilderService {
	return &AgentBuilderService{
		logger:     logger,
		blueprints: make(map[uuid.UUID]*domain.AgentBlueprint),
	}
}

// GenerateBlueprint generates an agent blueprint from a task description
func (s *AgentBuilderService) GenerateBlueprint(ctx context.Context, projectID uuid.UUID, input *domain.BuilderInput) (*domain.AgentBlueprint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.generateConfig(input)

	blueprint := &domain.AgentBlueprint{
		ID:               uuid.New(),
		ProjectID:        projectID,
		TaskDescription:  input.TaskDescription,
		GeneratedConfig:  config,
		EstimatedCost:    s.estimateCost(input.Complexity),
		EstimatedLatency: s.estimateLatency(input.Complexity),
		EstimatedQuality: s.estimateQuality(input.Complexity),
		Status:           domain.BlueprintStatusReady,
		CreatedAt:        time.Now(),
	}

	s.blueprints[blueprint.ID] = blueprint
	s.logger.Info("generated agent blueprint", zap.String("id", blueprint.ID.String()), zap.String("task", input.TaskDescription))
	return blueprint, nil
}

// GetBlueprint retrieves a blueprint by ID
func (s *AgentBuilderService) GetBlueprint(ctx context.Context, id uuid.UUID) (*domain.AgentBlueprint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bp, ok := s.blueprints[id]
	if !ok {
		return nil, fmt.Errorf("blueprint not found: %s", id)
	}
	return bp, nil
}

// ListBlueprints lists all blueprints for a project
func (s *AgentBuilderService) ListBlueprints(ctx context.Context, projectID uuid.UUID) ([]domain.AgentBlueprint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.AgentBlueprint
	for _, bp := range s.blueprints {
		if bp.ProjectID == projectID {
			result = append(result, *bp)
		}
	}
	return result, nil
}

// DeployBlueprint marks a blueprint as deployed
func (s *AgentBuilderService) DeployBlueprint(ctx context.Context, id uuid.UUID) (*domain.AgentBlueprint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bp, ok := s.blueprints[id]
	if !ok {
		return nil, fmt.Errorf("blueprint not found: %s", id)
	}
	bp.Status = domain.BlueprintStatusDeployed
	s.logger.Info("deployed blueprint", zap.String("id", id.String()))
	return bp, nil
}

func (s *AgentBuilderService) generateConfig(input *domain.BuilderInput) domain.BlueprintConfig {
	desc := strings.ToLower(input.TaskDescription)

	config := domain.BlueprintConfig{
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.7,
		Tools:       []string{"code_search", "file_read"},
		Parameters:  map[string]any{},
	}

	if strings.Contains(desc, "test") {
		config.Tools = append(config.Tools, "test_runner", "coverage_analyzer")
		config.SystemPrompt = "You are an expert testing agent. Write comprehensive tests with high coverage."
		config.Temperature = 0.3
	} else if strings.Contains(desc, "refactor") {
		config.Tools = append(config.Tools, "ast_parser", "dependency_analyzer")
		config.SystemPrompt = "You are a refactoring agent. Improve code quality while preserving behavior."
		config.Temperature = 0.2
	} else if strings.Contains(desc, "debug") {
		config.Tools = append(config.Tools, "debugger", "log_analyzer", "stack_trace_parser")
		config.SystemPrompt = "You are a debugging agent. Identify and fix bugs systematically."
		config.Temperature = 0.1
	} else if strings.Contains(desc, "review") {
		config.Tools = append(config.Tools, "diff_viewer", "lint_runner")
		config.SystemPrompt = "You are a code review agent. Provide thorough, constructive feedback."
		config.Temperature = 0.5
	} else if strings.Contains(desc, "document") {
		config.Tools = append(config.Tools, "doc_generator", "markdown_renderer")
		config.SystemPrompt = "You are a documentation agent. Create clear, comprehensive documentation."
		config.Temperature = 0.6
	} else {
		config.SystemPrompt = "You are a general-purpose coding agent. Complete the assigned task efficiently."
	}

	switch input.Complexity {
	case domain.ComplexitySimple:
		config.MaxTokens = 2048
	case domain.ComplexityComplex:
		config.MaxTokens = 8192
		config.Model = "gpt-4-turbo"
	}

	if input.TargetLanguage != "" {
		config.Parameters["language"] = input.TargetLanguage
	}

	return config
}

func (s *AgentBuilderService) estimateCost(complexity domain.BuilderComplexity) float64 {
	switch complexity {
	case domain.ComplexitySimple:
		return 0.05
	case domain.ComplexityComplex:
		return 0.50
	default:
		return 0.15
	}
}

func (s *AgentBuilderService) estimateLatency(complexity domain.BuilderComplexity) float64 {
	switch complexity {
	case domain.ComplexitySimple:
		return 2.0
	case domain.ComplexityComplex:
		return 15.0
	default:
		return 5.0
	}
}

func (s *AgentBuilderService) estimateQuality(complexity domain.BuilderComplexity) float64 {
	switch complexity {
	case domain.ComplexitySimple:
		return 0.95
	case domain.ComplexityComplex:
		return 0.82
	default:
		return 0.90
	}
}
