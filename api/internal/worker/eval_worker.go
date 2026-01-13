package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/pkg/circuitbreaker"
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
	logger         *zap.Logger
	config         *config.Config
	evalService    *service.EvalService
	scoreService   *service.ScoreService
	queryService   *service.QueryService
	httpClient     *http.Client
	circuitBreaker *circuitbreaker.CircuitBreaker
}

// NewEvalWorker creates a new eval worker
func NewEvalWorker(
	logger *zap.Logger,
	cfg *config.Config,
	evalService *service.EvalService,
	scoreService *service.ScoreService,
	queryService *service.QueryService,
) *EvalWorker {
	// Create circuit breaker for OpenAI API calls
	cb := circuitbreaker.New(circuitbreaker.Config{
		Name:                "openai-eval",
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		MaxHalfOpenRequests: 1,
		OnStateChange: func(name string, from, to circuitbreaker.State) {
			logger.Warn("circuit breaker state changed",
				zap.String("breaker", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})

	return &EvalWorker{
		logger:         logger,
		config:         cfg,
		evalService:    evalService,
		scoreService:   scoreService,
		queryService:   queryService,
		circuitBreaker: cb,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // LLM evaluations can take longer
		},
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

// LLMEvaluationResult represents the expected response from an LLM evaluation
type LLMEvaluationResult struct {
	Score       float64 `json:"score"`
	StringValue string  `json:"string_value,omitempty"`
	Reasoning   string  `json:"reasoning"`
	Passed      *bool   `json:"passed,omitempty"` // For boolean evaluations
}

// runLLMEvaluation runs an LLM-as-Judge evaluation
func (w *EvalWorker) runLLMEvaluation(
	ctx context.Context,
	evaluator *domain.Evaluator,
	trace *domain.Trace,
	observation *domain.Observation,
) (*domain.Score, error) {
	// Check if LLM evaluation is configured
	if w.config.Eval.APIKey == "" {
		return nil, fmt.Errorf("LLM API key not configured (set EVAL_API_KEY)")
	}

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

	// Get model from evaluator config or use default
	model := w.config.Eval.DefaultModel
	if evaluator.Config != "" {
		var evalConfig map[string]interface{}
		if err := json.Unmarshal([]byte(evaluator.Config), &evalConfig); err == nil {
			if m, ok := evalConfig["model"].(string); ok && m != "" {
				model = m
			}
		}
	}

	// Build the system prompt based on score data type
	systemPrompt := w.buildEvaluationSystemPrompt(evaluator.ScoreDataType)

	w.logger.Info("running LLM evaluation",
		zap.String("evaluator_id", evaluator.ID.String()),
		zap.String("model", model),
		zap.Int("prompt_length", len(prompt)),
	)

	// Call the LLM
	response, err := w.callLLM(ctx, model, systemPrompt, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse the response
	result, err := w.parseLLMResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Create score from result
	evalID := evaluator.ID
	score := &domain.Score{
		ID:        uuid.New(),
		TraceID:   trace.ID,
		ProjectID: evaluator.ProjectID,
		Name:      evaluator.ScoreName,
		DataType:  evaluator.ScoreDataType,
		Source:    domain.ScoreSourceEval,
		Comment:   result.Reasoning,
		ConfigID:  &evalID,
		CreatedAt: time.Now(),
	}

	// Set value based on data type
	switch evaluator.ScoreDataType {
	case domain.ScoreDataTypeNumeric:
		score.Value = &result.Score
	case domain.ScoreDataTypeBoolean:
		if result.Passed != nil {
			val := 0.0
			if *result.Passed {
				val = 1.0
			}
			score.Value = &val
		} else {
			// Interpret score as boolean (>= 0.5 is true)
			val := 0.0
			if result.Score >= 0.5 {
				val = 1.0
			}
			score.Value = &val
		}
	case domain.ScoreDataTypeCategorical:
		if result.StringValue != "" {
			score.StringValue = &result.StringValue
		}
	}

	if observation != nil {
		obsID := observation.ID
		score.ObservationID = &obsID
	}

	w.logger.Info("LLM evaluation completed",
		zap.String("evaluator_id", evaluator.ID.String()),
		zap.Float64("score", result.Score),
		zap.String("reasoning", result.Reasoning),
	)

	return score, nil
}

// buildEvaluationSystemPrompt creates the system prompt based on score type
func (w *EvalWorker) buildEvaluationSystemPrompt(dataType domain.ScoreDataType) string {
	basePrompt := `You are an AI evaluation assistant. Your task is to evaluate AI agent outputs based on the criteria provided.

Respond with a JSON object containing your evaluation. The JSON must include:
- "reasoning": A brief explanation of your evaluation (1-3 sentences)
`

	switch dataType {
	case domain.ScoreDataTypeNumeric:
		return basePrompt + `- "score": A numeric score from 0.0 to 1.0 (where 1.0 is best)

Example response:
{"score": 0.85, "reasoning": "The response was accurate and helpful, but could have been more concise."}`

	case domain.ScoreDataTypeBoolean:
		return basePrompt + `- "passed": A boolean (true/false) indicating if the criteria was met
- "score": 1.0 if passed, 0.0 if not

Example response:
{"passed": true, "score": 1.0, "reasoning": "The response correctly followed the instructions."}`

	case domain.ScoreDataTypeCategorical:
		return basePrompt + `- "string_value": The category that best matches the output
- "score": A confidence score from 0.0 to 1.0

Example response:
{"string_value": "positive", "score": 0.9, "reasoning": "The sentiment is clearly positive based on word choice."}`

	default:
		return basePrompt + `- "score": A numeric score from 0.0 to 1.0

Example response:
{"score": 0.75, "reasoning": "The evaluation shows satisfactory results."}`
	}
}

// callLLM makes a call to the configured LLM API (currently OpenAI-compatible)
// Uses circuit breaker to protect against cascading failures
func (w *EvalWorker) callLLM(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	// Use circuit breaker to protect against OpenAI API failures
	return circuitbreaker.ExecuteWithResult(w.circuitBreaker, ctx, func() (string, error) {
		return w.doLLMCall(ctx, model, systemPrompt, userPrompt)
	})
}

// doLLMCall performs the actual LLM API call
func (w *EvalWorker) doLLMCall(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature":     0.1, // Low temperature for consistent evaluations
		"max_tokens":      500,
		"response_format": map[string]string{"type": "json_object"},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.config.Eval.APIKey)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Treat server errors (5xx) and rate limits (429) as failures for circuit breaker
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		// 4xx errors (except 429) are client errors, don't count against circuit breaker
		// We return a special error that wraps the response but doesn't indicate a service failure
		return "", fmt.Errorf("API client error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("LLM API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}

// parseLLMResponse parses the JSON response from the LLM
func (w *EvalWorker) parseLLMResponse(response string) (*LLMEvaluationResult, error) {
	var result LLMEvaluationResult

	// Try to parse the response directly
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Try to extract JSON from response if it contains extra text
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start >= 0 && end > start {
			jsonStr := response[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				return nil, fmt.Errorf("invalid JSON in response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("no JSON found in response")
		}
	}

	// Validate score is in range
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 1 {
		result.Score = 1
	}

	return &result, nil
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

	ruleType, ok := config["rule_type"].(string)
	if !ok || ruleType == "" {
		return nil, fmt.Errorf("rule_type is required and must be a string")
	}

	var passed bool
	var comment string

	switch ruleType {
	case "contains":
		target, ok := config["target"].(string)
		if !ok {
			return nil, fmt.Errorf("'target' must be a string for rule type 'contains'")
		}
		substring, ok := config["substring"].(string)
		if !ok {
			return nil, fmt.Errorf("'substring' must be a string for rule type 'contains'")
		}
		content := w.getTargetContent(target, trace, observation)
		passed = strings.Contains(content, substring)
		comment = fmt.Sprintf("Checked if content contains '%s'", substring)

	case "not_contains":
		target, ok := config["target"].(string)
		if !ok {
			return nil, fmt.Errorf("'target' must be a string for rule type 'not_contains'")
		}
		substring, ok := config["substring"].(string)
		if !ok {
			return nil, fmt.Errorf("'substring' must be a string for rule type 'not_contains'")
		}
		content := w.getTargetContent(target, trace, observation)
		passed = !strings.Contains(content, substring)
		comment = fmt.Sprintf("Checked if content does not contain '%s'", substring)

	case "regex_match":
		// Regex matching would be implemented here
		passed = true
		comment = "Regex match evaluation"

	case "length_check":
		target, ok := config["target"].(string)
		if !ok {
			return nil, fmt.Errorf("'target' must be a string for rule type 'length_check'")
		}
		minLen, _ := config["min_length"].(float64) // optional, defaults to 0
		maxLen, _ := config["max_length"].(float64) // optional, defaults to 0 (no max)
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
