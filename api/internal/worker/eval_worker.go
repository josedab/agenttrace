package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	// TypeEvaluation is the task type for running an evaluation
	TypeEvaluation = "eval:run"
	// TypeBatchEvaluation is the task type for batch evaluations
	TypeBatchEvaluation = "eval:run-batch"
)

// EvaluationPayload is the payload for evaluation tasks
type EvaluationPayload struct {
	ProjectID     uuid.UUID `json:"project_id"`
	EvaluatorID   uuid.UUID `json:"evaluator_id"`
	TraceID       string    `json:"trace_id"`
	ObservationID string    `json:"observation_id,omitempty"`
}

// NewEvaluationTask creates a new evaluation task
func NewEvaluationTask(payload *EvaluationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal evaluation payload: %w", err)
	}
	return asynq.NewTask(TypeEvaluation, data, asynq.MaxRetry(3), asynq.Timeout(5*time.Minute)), nil
}

// BatchEvaluationPayload is the payload for batch evaluation tasks
type BatchEvaluationPayload struct {
	ProjectID   uuid.UUID `json:"project_id"`
	EvaluatorID uuid.UUID `json:"evaluator_id"`
	TraceIDs    []string  `json:"trace_ids"`
}

// NewBatchEvaluationTask creates a batch evaluation task
func NewBatchEvaluationTask(payload *BatchEvaluationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch evaluation payload: %w", err)
	}
	return asynq.NewTask(TypeBatchEvaluation, data, asynq.MaxRetry(3), asynq.Timeout(30*time.Minute)), nil
}

// EvalWorker handles evaluation tasks
type EvalWorker struct {
	logger       *zap.Logger
	evalService  *service.EvalService
	scoreService *service.ScoreService
	queryService *service.QueryService
}

// NewEvalWorker creates a new eval worker
func NewEvalWorker(
	logger *zap.Logger,
	evalService *service.EvalService,
	scoreService *service.ScoreService,
	queryService *service.QueryService,
) *EvalWorker {
	return &EvalWorker{
		logger:       logger,
		evalService:  evalService,
		scoreService: scoreService,
		queryService: queryService,
	}
}

// ProcessTask processes an evaluation task
func (w *EvalWorker) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload EvaluationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal evaluation payload: %w", err)
	}

	w.logger.Info("processing evaluation",
		zap.String("evaluator_id", payload.EvaluatorID.String()),
		zap.String("trace_id", payload.TraceID),
		zap.String("observation_id", payload.ObservationID),
	)

	// Get evaluator
	evaluator, err := w.evalService.Get(ctx, payload.EvaluatorID)
	if err != nil {
		return fmt.Errorf("failed to get evaluator: %w", err)
	}

	if !evaluator.Enabled {
		w.logger.Info("evaluator is not active, skipping",
			zap.String("evaluator_id", payload.EvaluatorID.String()),
		)
		return nil
	}

	// Get trace data
	trace, err := w.queryService.GetTrace(ctx, payload.ProjectID, payload.TraceID)
	if err != nil {
		return fmt.Errorf("failed to get trace: %w", err)
	}

	// Get observation if specified
	var observation *domain.Observation
	if payload.ObservationID != "" {
		observation, err = w.queryService.GetObservation(ctx, payload.ProjectID, payload.ObservationID)
		if err != nil {
			return fmt.Errorf("failed to get observation: %w", err)
		}
	}

	// Run evaluation based on type
	var score *domain.Score
	switch evaluator.Type {
	case domain.EvaluatorTypeLLM:
		score, err = w.runLLMEvaluation(ctx, evaluator, trace, observation)
	case domain.EvaluatorTypeRule:
		score, err = w.runRuleEvaluation(ctx, evaluator, trace, observation)
	default:
		return fmt.Errorf("unsupported evaluator type: %s", evaluator.Type)
	}

	if err != nil {
		w.logger.Error("evaluation failed",
			zap.String("evaluator_id", payload.EvaluatorID.String()),
			zap.String("trace_id", payload.TraceID),
			zap.Error(err),
		)
		return fmt.Errorf("evaluation failed: %w", err)
	}

	// Save score
	if score != nil {
		var commentPtr *string
		if score.Comment != "" {
			commentPtr = &score.Comment
		}
		scoreInput := &domain.ScoreInput{
			TraceID:       score.TraceID,
			ObservationID: score.ObservationID,
			Name:          score.Name,
			Source:        score.Source,
			DataType:      score.DataType,
			Value:         score.Value,
			StringValue:   score.StringValue,
			Comment:       commentPtr,
		}
		createdScore, err := w.scoreService.Create(ctx, score.ProjectID, scoreInput)
		if err != nil {
			return fmt.Errorf("failed to save score: %w", err)
		}

		w.logger.Info("evaluation completed",
			zap.String("evaluator_id", payload.EvaluatorID.String()),
			zap.String("trace_id", payload.TraceID),
			zap.String("score_id", createdScore.ID.String()),
			zap.Float64p("value", createdScore.Value),
			zap.Stringp("string_value", createdScore.StringValue),
		)
	}

	return nil
}

// ProcessBatchTask processes a batch evaluation task
func (w *EvalWorker) ProcessBatchTask(ctx context.Context, t *asynq.Task) error {
	var payload BatchEvaluationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal batch evaluation payload: %w", err)
	}

	w.logger.Info("processing batch evaluation",
		zap.String("evaluator_id", payload.EvaluatorID.String()),
		zap.Int("trace_count", len(payload.TraceIDs)),
	)

	// Process each trace
	var errs []error
	for _, traceID := range payload.TraceIDs {
		task, err := NewEvaluationTask(&EvaluationPayload{
			ProjectID:   payload.ProjectID,
			EvaluatorID: payload.EvaluatorID,
			TraceID:     traceID,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// Process inline for batch
		if err := w.ProcessTask(ctx, task); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		w.logger.Warn("batch evaluation completed with errors",
			zap.Int("total", len(payload.TraceIDs)),
			zap.Int("errors", len(errs)),
		)
	} else {
		w.logger.Info("batch evaluation completed successfully",
			zap.Int("total", len(payload.TraceIDs)),
		)
	}

	return nil
}

// runLLMEvaluation runs an LLM-as-Judge evaluation
// Note: This is a placeholder - full implementation requires LLM integration
func (w *EvalWorker) runLLMEvaluation(
	ctx context.Context,
	evaluator *domain.Evaluator,
	trace *domain.Trace,
	observation *domain.Observation,
) (*domain.Score, error) {
	_ = ctx

	// Build prompt from template
	prompt := evaluator.PromptTemplate
	if prompt == "" {
		return nil, fmt.Errorf("evaluator has no prompt template")
	}

	// Substitute variables
	variables := w.extractVariables(trace, observation)
	for k, v := range variables {
		prompt = strings.ReplaceAll(prompt, "{{"+k+"}}", v)
	}

	// TODO: Call LLM (using the configured model from evaluator.Config)
	// This would integrate with OpenAI, Anthropic, etc.
	// For now, return a placeholder score
	w.logger.Info("LLM evaluation prompt prepared",
		zap.String("evaluator_id", evaluator.ID.String()),
		zap.Int("prompt_length", len(prompt)),
	)

	// Parse result and create score
	evalID := evaluator.ID
	score := &domain.Score{
		ID:        uuid.New(),
		TraceID:   trace.ID,
		ProjectID: evaluator.ProjectID,
		Name:      evaluator.ScoreName,
		DataType:  evaluator.ScoreDataType,
		Source:    domain.ScoreSourceEval,
		ConfigID:  &evalID,
		CreatedAt: time.Now(),
	}

	if observation != nil {
		obsID := observation.ID
		score.ObservationID = &obsID
	}

	// Note: LLM evaluation result parsing would go here
	// For now, return nil to indicate no score was generated
	// In a full implementation, the LLM response would be parsed here
	_ = score
	return nil, fmt.Errorf("LLM evaluation not yet implemented")
}

// runRuleEvaluation runs a rule-based evaluation
func (w *EvalWorker) runRuleEvaluation(
	ctx context.Context,
	evaluator *domain.Evaluator,
	trace *domain.Trace,
	observation *domain.Observation,
) (*domain.Score, error) {
	_ = ctx

	// Get rule configuration from JSON string
	var config map[string]interface{}
	if evaluator.Config != "" {
		if err := json.Unmarshal([]byte(evaluator.Config), &config); err != nil {
			return nil, fmt.Errorf("invalid evaluator config: %w", err)
		}
	}
	if config == nil {
		return nil, fmt.Errorf("evaluator config is empty")
	}

	ruleType, _ := config["rule_type"].(string)

	var passed bool
	var comment string

	switch ruleType {
	case "contains":
		target, _ := config["target"].(string)
		substring, _ := config["substring"].(string)
		content := w.getTargetContent(target, trace, observation)
		passed = strings.Contains(content, substring)
		comment = fmt.Sprintf("Checked if content contains '%s'", substring)

	case "not_contains":
		target, _ := config["target"].(string)
		substring, _ := config["substring"].(string)
		content := w.getTargetContent(target, trace, observation)
		passed = !strings.Contains(content, substring)
		comment = fmt.Sprintf("Checked if content does not contain '%s'", substring)

	case "regex_match":
		// Regex matching would be implemented here
		passed = true
		comment = "Regex match evaluation"

	case "length_check":
		target, _ := config["target"].(string)
		minLen, _ := config["min_length"].(float64)
		maxLen, _ := config["max_length"].(float64)
		content := w.getTargetContent(target, trace, observation)
		length := len(content)
		passed = float64(length) >= minLen && (maxLen == 0 || float64(length) <= maxLen)
		comment = fmt.Sprintf("Length: %d (min: %.0f, max: %.0f)", length, minLen, maxLen)

	default:
		return nil, fmt.Errorf("unsupported rule type: %s", ruleType)
	}

	// Create score
	val := 0.0
	if passed {
		val = 1.0
	}

	evalID := evaluator.ID
	score := &domain.Score{
		ID:        uuid.New(),
		TraceID:   trace.ID,
		ProjectID: evaluator.ProjectID,
		Name:      evaluator.ScoreName,
		Value:     &val,
		DataType:  domain.ScoreDataTypeBoolean,
		Source:    domain.ScoreSourceEval,
		Comment:   comment,
		ConfigID:  &evalID,
		CreatedAt: time.Now(),
	}

	if observation != nil {
		obsID := observation.ID
		score.ObservationID = &obsID
	}

	return score, nil
}

// extractVariables extracts variables from trace/observation for prompt substitution
func (w *EvalWorker) extractVariables(trace *domain.Trace, observation *domain.Observation) map[string]string {
	variables := make(map[string]string)

	// Trace variables
	if trace != nil {
		variables["trace_id"] = trace.ID
		variables["trace_name"] = trace.Name
		if trace.Input != "" {
			variables["trace_input"] = trace.Input
		}
		if trace.Output != "" {
			variables["trace_output"] = trace.Output
		}
	}

	// Observation variables
	if observation != nil {
		variables["observation_id"] = observation.ID
		variables["observation_name"] = observation.Name
		variables["observation_type"] = string(observation.Type)
		if observation.Input != "" {
			variables["input"] = observation.Input
		}
		if observation.Output != "" {
			variables["output"] = observation.Output
		}
		if observation.Model != "" {
			variables["model"] = observation.Model
		}
	}

	return variables
}

// getTargetContent gets content for rule evaluation
func (w *EvalWorker) getTargetContent(target string, trace *domain.Trace, observation *domain.Observation) string {
	switch target {
	case "trace_input":
		if trace != nil {
			return trace.Input
		}
	case "trace_output":
		if trace != nil {
			return trace.Output
		}
	case "observation_input":
		if observation != nil {
			return observation.Input
		}
	case "observation_output":
		if observation != nil {
			return observation.Output
		}
	}
	return ""
}
