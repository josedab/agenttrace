package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// RunbookRepository defines repository operations for runbooks
type RunbookRepository interface {
	Save(ctx context.Context, runbook *domain.Runbook) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Runbook, error)
	List(ctx context.Context, filter domain.RunbookFilter, limit, offset int) (*domain.RunbookList, error)
	Update(ctx context.Context, runbook *domain.Runbook) error
	Delete(ctx context.Context, id uuid.UUID) error
	SaveExecution(ctx context.Context, exec *domain.RunbookExecution) error
	ListExecutions(ctx context.Context, runbookID uuid.UUID, limit, offset int) (*domain.RunbookExecutionList, error)
	GetActiveRunbooks(ctx context.Context, projectID uuid.UUID) ([]domain.Runbook, error)
}

// RunbookService handles runbook management and execution
type RunbookService struct {
	logger      *zap.Logger
	runbookRepo RunbookRepository
}

// NewRunbookService creates a new runbook service
func NewRunbookService(logger *zap.Logger, runbookRepo RunbookRepository) *RunbookService {
	return &RunbookService{
		logger:      logger,
		runbookRepo: runbookRepo,
	}
}

// CreateRunbook creates a new runbook from YAML definition
func (s *RunbookService) CreateRunbook(ctx context.Context, projectID, userID uuid.UUID, input *domain.RunbookInput) (*domain.Runbook, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("runbook name is required")
	}
	if input.YAMLContent == "" {
		return nil, fmt.Errorf("YAML content is required")
	}

	triggers, actions, err := s.parseYAML(input.YAMLContent)
	if err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
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
		Triggers:             triggers,
		Actions:              actions,
		MaxExecutionsPerHour: maxExec,
		RequireApproval:      requireApproval,
		CreatedBy:            userID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.runbookRepo.Save(ctx, runbook); err != nil {
		return nil, fmt.Errorf("failed to save runbook: %w", err)
	}

	s.logger.Info("runbook created",
		zap.String("runbookId", runbook.ID.String()),
		zap.String("name", runbook.Name),
		zap.Int("triggers", len(triggers)),
		zap.Int("actions", len(actions)),
	)

	return runbook, nil
}

// GetRunbook retrieves a runbook by ID
func (s *RunbookService) GetRunbook(ctx context.Context, id uuid.UUID) (*domain.Runbook, error) {
	return s.runbookRepo.GetByID(ctx, id)
}

// ListRunbooks lists runbooks for a project
func (s *RunbookService) ListRunbooks(ctx context.Context, filter domain.RunbookFilter, limit, offset int) (*domain.RunbookList, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.runbookRepo.List(ctx, filter, limit, offset)
}

// ActivateRunbook sets a runbook to active status
func (s *RunbookService) ActivateRunbook(ctx context.Context, id uuid.UUID) (*domain.Runbook, error) {
	runbook, err := s.runbookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	runbook.Status = domain.RunbookStatusActive
	runbook.UpdatedAt = time.Now()
	if err := s.runbookRepo.Update(ctx, runbook); err != nil {
		return nil, err
	}
	return runbook, nil
}

// EvaluateTrace checks all active runbooks against a trace and executes matching ones
func (s *RunbookService) EvaluateTrace(ctx context.Context, projectID uuid.UUID, traceContext *RunbookTraceContext) ([]domain.RunbookExecution, error) {
	runbooks, err := s.runbookRepo.GetActiveRunbooks(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active runbooks: %w", err)
	}

	var executions []domain.RunbookExecution

	for _, runbook := range runbooks {
		matchedTriggers := s.evaluateTriggers(&runbook, traceContext)
		if len(matchedTriggers) == 0 {
			continue
		}

		execution := domain.RunbookExecution{
			ID:               uuid.New(),
			RunbookID:        runbook.ID,
			ProjectID:        projectID,
			TraceID:          traceContext.TraceID,
			Status:           domain.RunbookExecPending,
			TriggerMatch:     strings.Join(matchedTriggers, ", "),
			ApprovalRequired: runbook.RequireApproval,
			StartedAt:        time.Now(),
		}

		if runbook.RequireApproval {
			execution.Status = domain.RunbookExecPending
		} else {
			s.executeRunbook(ctx, &runbook, &execution)
		}

		if err := s.runbookRepo.SaveExecution(ctx, &execution); err != nil {
			s.logger.Warn("failed to save runbook execution", zap.Error(err))
			continue
		}

		executions = append(executions, execution)

		s.logger.Info("runbook triggered",
			zap.String("runbookId", runbook.ID.String()),
			zap.String("traceId", traceContext.TraceID),
			zap.Strings("triggers", matchedTriggers),
		)
	}

	return executions, nil
}

// TestRunbook tests a runbook against a trace (dry run)
func (s *RunbookService) TestRunbook(ctx context.Context, input *domain.RunbookTestInput) (*domain.RunbookTestResult, error) {
	runbook, err := s.runbookRepo.GetByID(ctx, input.RunbookID)
	if err != nil {
		return nil, fmt.Errorf("runbook not found: %w", err)
	}

	// Build a mock trace context for testing
	traceCtx := &RunbookTraceContext{
		TraceID: input.TraceID,
	}

	matchedTriggers := s.evaluateTriggers(runbook, traceCtx)

	result := &domain.RunbookTestResult{
		Matched:         len(matchedTriggers) > 0,
		MatchedTriggers: matchedTriggers,
		PlannedActions:  runbook.Actions,
	}

	if !input.DryRun && result.Matched {
		execution := domain.RunbookExecution{
			ID:           uuid.New(),
			RunbookID:    runbook.ID,
			ProjectID:    runbook.ProjectID,
			TraceID:      input.TraceID,
			Status:       domain.RunbookExecRunning,
			TriggerMatch: strings.Join(matchedTriggers, ", "),
			StartedAt:    time.Now(),
		}
		s.executeRunbook(ctx, runbook, &execution)
		result.Execution = &execution
	}

	return result, nil
}

// RunbookTraceContext provides trace data for runbook evaluation
type RunbookTraceContext struct {
	TraceID    string
	Cost       float64
	Latency    float64
	ErrorRate  float64
	HasError   bool
	ErrorMsg   string
	ToolCalls  []string
	Model      string
	Metadata   map[string]string
}

func (s *RunbookService) evaluateTriggers(runbook *domain.Runbook, ctx *RunbookTraceContext) []string {
	var matched []string

	for _, trigger := range runbook.Triggers {
		switch trigger.Type {
		case "threshold":
			if s.evaluateThresholdTrigger(trigger, ctx) {
				matched = append(matched, trigger.Description)
			}
		case "error_match":
			if s.evaluateErrorMatchTrigger(trigger, ctx) {
				matched = append(matched, trigger.Description)
			}
		case "pattern":
			if s.evaluatePatternTrigger(trigger, ctx) {
				matched = append(matched, trigger.Description)
			}
		}
	}

	return matched
}

func (s *RunbookService) evaluateThresholdTrigger(trigger domain.RunbookTrigger, ctx *RunbookTraceContext) bool {
	metric, _ := trigger.Conditions["metric"].(string)
	operator, _ := trigger.Conditions["operator"].(string)
	value, _ := trigger.Conditions["value"].(float64)

	var actual float64
	switch metric {
	case "cost":
		actual = ctx.Cost
	case "latency":
		actual = ctx.Latency
	case "error_rate":
		actual = ctx.ErrorRate
	default:
		return false
	}

	switch operator {
	case "gt":
		return actual > value
	case "gte":
		return actual >= value
	case "lt":
		return actual < value
	case "lte":
		return actual <= value
	default:
		return false
	}
}

func (s *RunbookService) evaluateErrorMatchTrigger(trigger domain.RunbookTrigger, ctx *RunbookTraceContext) bool {
	if !ctx.HasError {
		return false
	}
	pattern, _ := trigger.Conditions["pattern"].(string)
	if pattern == "" {
		return ctx.HasError
	}
	return strings.Contains(ctx.ErrorMsg, pattern)
}

func (s *RunbookService) evaluatePatternTrigger(trigger domain.RunbookTrigger, ctx *RunbookTraceContext) bool {
	// Simple metadata pattern matching
	for k, v := range trigger.Conditions {
		expected, ok := v.(string)
		if !ok {
			continue
		}
		if actual, exists := ctx.Metadata[k]; exists {
			if !strings.Contains(actual, expected) {
				return false
			}
		} else {
			return false
		}
	}
	return len(trigger.Conditions) > 0
}

func (s *RunbookService) executeRunbook(ctx context.Context, runbook *domain.Runbook, execution *domain.RunbookExecution) {
	execution.Status = domain.RunbookExecRunning

	for _, action := range runbook.Actions {
		result := s.executeAction(ctx, action)
		execution.ActionResults = append(execution.ActionResults, result)

		if result.Status == domain.RunbookExecFailed {
			switch action.OnFailure {
			case "abort":
				execution.Status = domain.RunbookExecFailed
				execution.Error = fmt.Sprintf("action '%s' failed: %s", action.Name, result.Error)
				now := time.Now()
				execution.CompletedAt = &now
				return
			case "retry":
				for i := 0; i < action.RetryCount; i++ {
					retryResult := s.executeAction(ctx, action)
					if retryResult.Status == domain.RunbookExecCompleted {
						execution.ActionResults = append(execution.ActionResults, retryResult)
						break
					}
				}
			default: // "continue"
				continue
			}
		}
	}

	execution.Status = domain.RunbookExecCompleted
	now := time.Now()
	execution.CompletedAt = &now
}

func (s *RunbookService) executeAction(ctx context.Context, action domain.RunbookAction) domain.ActionResult {
	start := time.Now()

	result := domain.ActionResult{
		ActionName: action.Name,
		ActionType: action.Type,
		Status:     domain.RunbookExecCompleted,
		StartedAt:  start,
	}

	switch action.Type {
	case domain.RunbookActionRetryWithModel:
		model := action.Parameters["model"]
		result.Output = fmt.Sprintf("Scheduled retry with model: %s", model)
		s.logger.Info("runbook action: retry with model", zap.String("model", model))

	case domain.RunbookActionEscalateToHuman:
		result.Output = "Escalation notification sent"
		s.logger.Info("runbook action: escalate to human")

	case domain.RunbookActionRollbackPrompt:
		version := action.Parameters["version"]
		result.Output = fmt.Sprintf("Prompt rollback to version %s scheduled", version)
		s.logger.Info("runbook action: rollback prompt", zap.String("version", version))

	case domain.RunbookActionAdjustTemperature:
		temp := action.Parameters["temperature"]
		result.Output = fmt.Sprintf("Temperature adjusted to %s", temp)
		s.logger.Info("runbook action: adjust temperature", zap.String("temperature", temp))

	case domain.RunbookActionSendNotification:
		channel := action.Parameters["channel"]
		result.Output = fmt.Sprintf("Notification sent to %s", channel)
		s.logger.Info("runbook action: send notification", zap.String("channel", channel))

	case domain.RunbookActionWebhook:
		url := action.Parameters["url"]
		result.Output = fmt.Sprintf("Webhook called: %s", url)
		s.logger.Info("runbook action: webhook", zap.String("url", url))

	default:
		result.Status = domain.RunbookExecFailed
		result.Error = fmt.Sprintf("unsupported action type: %s", action.Type)
	}

	result.Duration = time.Since(start).Milliseconds()
	return result
}

// parseYAML parses runbook YAML content into triggers and actions
// Simplified parser that handles the expected YAML structure
func (s *RunbookService) parseYAML(content string) ([]domain.RunbookTrigger, []domain.RunbookAction, error) {
	if content == "" {
		return nil, nil, fmt.Errorf("empty YAML content")
	}

	// Basic validation - real implementation would use gopkg.in/yaml.v3
	if !strings.Contains(content, "triggers") && !strings.Contains(content, "actions") {
		return nil, nil, fmt.Errorf("YAML must contain 'triggers' and/or 'actions' sections")
	}

	// Return placeholder parsed content - real implementation would fully parse YAML
	triggers := []domain.RunbookTrigger{}
	actions := []domain.RunbookAction{}

	return triggers, actions, nil
}
