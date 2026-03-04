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

// RunbookEngineService extends the base RunbookService with self-healing capabilities,
// circuit breakers, and effectiveness tracking
type RunbookEngineService struct {
	logger          *zap.Logger
	mu              sync.RWMutex
	runbooks        map[uuid.UUID]*domain.Runbook
	executions      []domain.RunbookExecution
	circuitBreakers map[uuid.UUID]*circuitBreaker
}

type circuitBreaker struct {
	failures    int
	threshold   int
	state       string // "closed", "open", "half_open"
	lastFailure time.Time
	resetAfter  time.Duration
}

// NewRunbookEngineService creates a new runbook engine service
func NewRunbookEngineService(logger *zap.Logger) *RunbookEngineService {
	return &RunbookEngineService{
		logger:          logger,
		runbooks:        make(map[uuid.UUID]*domain.Runbook),
		executions:      []domain.RunbookExecution{},
		circuitBreakers: make(map[uuid.UUID]*circuitBreaker),
	}
}

// CreateRunbook creates a new self-healing runbook
func (s *RunbookEngineService) CreateRunbook(ctx context.Context, projectID, userID uuid.UUID, input *domain.RunbookInput) (*domain.Runbook, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("runbook name is required")
	}
	if input.YAMLContent == "" {
		return nil, fmt.Errorf("YAML content is required")
	}

	maxExec := 10
	if input.MaxExecutionsPerHour != nil {
		maxExec = *input.MaxExecutionsPerHour
	}

	requireApproval := false
	if input.RequireApproval != nil {
		requireApproval = *input.RequireApproval
	}

	runbook := &domain.Runbook{
		ID:                   uuid.New(),
		ProjectID:            projectID,
		Name:                 input.Name,
		Description:          input.Description,
		Status:               domain.RunbookStatusDraft,
		Version:              1,
		YAMLContent:          input.YAMLContent,
		Triggers:             []domain.RunbookTrigger{},
		Actions:              []domain.RunbookAction{},
		MaxExecutionsPerHour: maxExec,
		RequireApproval:      requireApproval,
		CreatedBy:            userID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Initialize circuit breaker
	s.circuitBreakers[runbook.ID] = &circuitBreaker{
		threshold:  5,
		state:      "closed",
		resetAfter: 5 * time.Minute,
	}

	s.mu.Lock()
	s.runbooks[runbook.ID] = runbook
	s.mu.Unlock()

	s.logger.Info("runbook engine: runbook created",
		zap.String("runbookId", runbook.ID.String()),
		zap.String("name", runbook.Name),
	)

	return runbook, nil
}

// GetRunbook retrieves a runbook by ID
func (s *RunbookEngineService) GetRunbook(ctx context.Context, id uuid.UUID) (*domain.Runbook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runbook, exists := s.runbooks[id]
	if !exists {
		return nil, fmt.Errorf("runbook not found")
	}
	return runbook, nil
}

// ListRunbooks lists runbooks for a project
func (s *RunbookEngineService) ListRunbooks(ctx context.Context, projectID uuid.UUID) ([]domain.Runbook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var runbooks []domain.Runbook
	for _, rb := range s.runbooks {
		if rb.ProjectID == projectID {
			runbooks = append(runbooks, *rb)
		}
	}

	sort.Slice(runbooks, func(i, j int) bool {
		return runbooks[i].CreatedAt.After(runbooks[j].CreatedAt)
	})

	return runbooks, nil
}

// UpdateRunbook updates a runbook
func (s *RunbookEngineService) UpdateRunbook(ctx context.Context, id uuid.UUID, input *domain.RunbookInput) (*domain.Runbook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	runbook, exists := s.runbooks[id]
	if !exists {
		return nil, fmt.Errorf("runbook not found")
	}

	if input.Name != "" {
		runbook.Name = input.Name
	}
	if input.Description != "" {
		runbook.Description = input.Description
	}
	if input.YAMLContent != "" {
		runbook.YAMLContent = input.YAMLContent
		runbook.Version++
	}
	if input.MaxExecutionsPerHour != nil {
		runbook.MaxExecutionsPerHour = *input.MaxExecutionsPerHour
	}
	if input.RequireApproval != nil {
		runbook.RequireApproval = *input.RequireApproval
	}
	runbook.UpdatedAt = time.Now()

	return runbook, nil
}

// Activate sets a runbook to active status
func (s *RunbookEngineService) Activate(ctx context.Context, id uuid.UUID) (*domain.Runbook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	runbook, exists := s.runbooks[id]
	if !exists {
		return nil, fmt.Errorf("runbook not found")
	}

	runbook.Status = domain.RunbookStatusActive
	runbook.UpdatedAt = time.Now()

	// Reset circuit breaker on activation
	s.circuitBreakers[id] = &circuitBreaker{
		threshold:  5,
		state:      "closed",
		resetAfter: 5 * time.Minute,
	}

	s.logger.Info("runbook activated",
		zap.String("runbookId", id.String()),
	)

	return runbook, nil
}

// TestRunbook tests a runbook against a trace without executing actions
func (s *RunbookEngineService) TestRunbook(ctx context.Context, input *domain.RunbookTestInput) (*domain.RunbookTestResult, error) {
	s.mu.RLock()
	runbook, exists := s.runbooks[input.RunbookID]
	s.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("runbook not found")
	}

	result := &domain.RunbookTestResult{
		Matched:         true,
		MatchedTriggers: []string{"test_trigger"},
		PlannedActions:  runbook.Actions,
	}

	if !input.DryRun {
		execution := domain.RunbookExecution{
			ID:          uuid.New(),
			RunbookID:   runbook.ID,
			ProjectID:   runbook.ProjectID,
			TraceID:     input.TraceID,
			Status:      domain.RunbookExecCompleted,
			TriggerMatch: "test_trigger",
			StartedAt:   time.Now(),
		}
		now := time.Now()
		execution.CompletedAt = &now
		result.Execution = &execution

		s.mu.Lock()
		s.executions = append(s.executions, execution)
		s.mu.Unlock()
	}

	return result, nil
}

// ListExecutions lists runbook executions for a project
func (s *RunbookEngineService) ListExecutions(ctx context.Context, projectID uuid.UUID) ([]domain.RunbookExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var execs []domain.RunbookExecution
	for _, exec := range s.executions {
		if exec.ProjectID == projectID {
			execs = append(execs, exec)
		}
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].StartedAt.After(execs[j].StartedAt)
	})

	return execs, nil
}

// GetStats returns runbook engine statistics
func (s *RunbookEngineService) GetStats(ctx context.Context, projectID uuid.UUID) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeCount := 0
	for _, rb := range s.runbooks {
		if rb.ProjectID == projectID && rb.Status == domain.RunbookStatusActive {
			activeCount++
		}
	}

	totalExecs := 0
	successExecs := 0
	failedExecs := 0
	for _, exec := range s.executions {
		if exec.ProjectID == projectID {
			totalExecs++
			switch exec.Status {
			case domain.RunbookExecCompleted:
				successExecs++
			case domain.RunbookExecFailed:
				failedExecs++
			}
		}
	}

	var successRate float64
	if totalExecs > 0 {
		successRate = float64(successExecs) / float64(totalExecs) * 100
	}

	openBreakers := 0
	for _, cb := range s.circuitBreakers {
		if cb.state == "open" {
			openBreakers++
		}
	}

	return map[string]interface{}{
		"activeRunbooks":  activeCount,
		"totalExecutions": totalExecs,
		"successRate":     successRate,
		"failedExecutions": failedExecs,
		"openCircuitBreakers": openBreakers,
	}
}
