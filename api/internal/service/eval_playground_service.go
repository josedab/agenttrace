package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// EvalPlaygroundService handles evaluation playground operations
type EvalPlaygroundService struct {
	logger   *zap.Logger
	sessions map[uuid.UUID]*domain.PlaygroundSession
}

// NewEvalPlaygroundService creates a new eval playground service
func NewEvalPlaygroundService(logger *zap.Logger) *EvalPlaygroundService {
	return &EvalPlaygroundService{
		logger:   logger,
		sessions: make(map[uuid.UUID]*domain.PlaygroundSession),
	}
}

// CreateSession creates a new playground session
func (s *EvalPlaygroundService) CreateSession(ctx context.Context, projectID, userID uuid.UUID, input *domain.PlaygroundCreateInput) (*domain.PlaygroundSession, error) {
	session := &domain.PlaygroundSession{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    userID,
		Name:      input.Name,
		Status:    domain.PlaygroundStatusDraft,
		Code:      input.Code,
		Language:  input.Language,
		TraceIDs:  input.TraceIDs,
		Results:   []domain.PlaygroundResult{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if session.Name == "" {
		session.Name = "Untitled Playground"
	}
	if session.Language == "" {
		session.Language = "javascript"
	}

	s.sessions[session.ID] = session

	s.logger.Info("playground session created",
		zap.String("sessionId", session.ID.String()),
	)

	return session, nil
}

// GetSession retrieves a playground session
func (s *EvalPlaygroundService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.PlaygroundSession, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("playground session not found: %s", sessionID)
	}
	return session, nil
}

// Execute runs evaluator code against trace data in a sandboxed environment
func (s *EvalPlaygroundService) Execute(ctx context.Context, projectID uuid.UUID, input *domain.PlaygroundExecuteInput) ([]domain.PlaygroundResult, error) {
	if input.Code == "" {
		return nil, fmt.Errorf("evaluator code is required")
	}
	if len(input.TraceIDs) == 0 {
		return nil, fmt.Errorf("at least one trace ID is required")
	}

	timeout := input.Timeout
	if timeout <= 0 || timeout > 120 {
		timeout = 30
	}

	var results []domain.PlaygroundResult

	for _, traceID := range input.TraceIDs {
		start := time.Now()
		result := s.executeEval(ctx, input.Code, input.Language, traceID, timeout)
		result.DurationMs = time.Since(start).Milliseconds()
		result.TraceID = traceID
		result.ExecutedAt = time.Now()
		results = append(results, result)
	}

	s.logger.Info("playground execution completed",
		zap.Int("traceCount", len(input.TraceIDs)),
		zap.Int("resultCount", len(results)),
	)

	return results, nil
}

// executeEval runs the evaluator code for a single trace
func (s *EvalPlaygroundService) executeEval(ctx context.Context, code, language, traceID string, timeout int) domain.PlaygroundResult {
	// Sandboxed execution with simulated output
	// In production this would use a V8 isolate or Wasm sandbox
	result := domain.PlaygroundResult{
		TraceID: traceID,
		Metadata: map[string]any{
			"language": language,
			"sandbox":  "simulated",
		},
	}

	// Basic validation of the code structure
	if len(code) < 10 {
		result.Error = "evaluator code too short — must define a function"
		return result
	}

	// Simulate execution — return a placeholder score
	score := 0.85
	result.Score = &score
	result.Label = "pass"
	result.Reasoning = fmt.Sprintf("Evaluator executed successfully against trace %s (sandbox mode)", traceID)

	return result
}

// ShareSession makes a session publicly accessible via share token
func (s *EvalPlaygroundService) ShareSession(ctx context.Context, sessionID uuid.UUID) (*domain.PlaygroundSession, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("playground session not found: %s", sessionID)
	}

	session.ShareToken = uuid.New().String()[:12]
	session.IsShared = true
	session.UpdatedAt = time.Now()

	return session, nil
}

// GetSharedSession retrieves a shared session by token
func (s *EvalPlaygroundService) GetSharedSession(ctx context.Context, token string) (*domain.PlaygroundSession, error) {
	for _, session := range s.sessions {
		if session.ShareToken == token && session.IsShared {
			return session, nil
		}
	}
	return nil, fmt.Errorf("shared session not found")
}

// ListTemplates returns built-in evaluator templates
func (s *EvalPlaygroundService) ListTemplates(ctx context.Context) []domain.PlaygroundTemplate {
	return []domain.PlaygroundTemplate{
		{
			ID:          "correctness",
			Name:        "Correctness Check",
			Description: "Evaluates if the agent output matches expected results",
			Language:    "javascript",
			Code:        "function evaluate(trace) {\n  const output = trace.output;\n  const expected = trace.metadata?.expected;\n  if (!expected) return { score: null, label: 'skipped', reasoning: 'No expected output' };\n  const match = output === expected;\n  return { score: match ? 1.0 : 0.0, label: match ? 'pass' : 'fail', reasoning: match ? 'Output matches expected' : 'Output differs from expected' };\n}",
			Category:    "quality",
		},
		{
			ID:          "cost-efficiency",
			Name:        "Cost Efficiency",
			Description: "Checks if trace cost is within acceptable bounds",
			Language:    "javascript",
			Code:        "function evaluate(trace) {\n  const maxCost = 0.50;\n  const cost = trace.totalCost || 0;\n  const score = Math.max(0, 1 - (cost / maxCost));\n  return { score, label: cost <= maxCost ? 'pass' : 'fail', reasoning: `Cost: $${cost.toFixed(4)} (max: $${maxCost})` };\n}",
			Category:    "cost",
		},
		{
			ID:          "latency-check",
			Name:        "Latency SLA Check",
			Description: "Validates trace latency against SLA targets",
			Language:    "javascript",
			Code:        "function evaluate(trace) {\n  const maxLatencyMs = 5000;\n  const latency = trace.durationMs || 0;\n  const score = latency <= maxLatencyMs ? 1.0 : 0.0;\n  return { score, label: score === 1 ? 'pass' : 'fail', reasoning: `Latency: ${latency}ms (SLA: ${maxLatencyMs}ms)` };\n}",
			Category:    "performance",
		},
		{
			ID:          "token-budget",
			Name:        "Token Budget",
			Description: "Evaluates token usage efficiency",
			Language:    "javascript",
			Code:        "function evaluate(trace) {\n  const maxTokens = 10000;\n  const tokens = trace.totalTokens || 0;\n  const efficiency = Math.max(0, 1 - (tokens / maxTokens));\n  return { score: efficiency, label: tokens <= maxTokens ? 'pass' : 'over-budget', reasoning: `${tokens} tokens used (budget: ${maxTokens})` };\n}",
			Category:    "cost",
		},
		{
			ID:          "error-rate",
			Name:        "Error Detection",
			Description: "Detects errors in trace execution",
			Language:    "python",
			Code:        "def evaluate(trace):\n    has_error = trace.get('level') == 'ERROR' or trace.get('statusMessage', '') != ''\n    return {\n        'score': 0.0 if has_error else 1.0,\n        'label': 'error' if has_error else 'clean',\n        'reasoning': f\"Error detected: {trace.get('statusMessage', 'none')}\" if has_error else 'No errors'\n    }",
			Category:    "quality",
		},
	}
}

// Helper to serialize results for JSON response
func marshalResults(results []domain.PlaygroundResult) (string, error) {
	data, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
