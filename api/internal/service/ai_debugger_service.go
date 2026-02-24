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

// AIDebuggerService provides AI-powered debugging for agent traces
type AIDebuggerService struct {
	logger *zap.Logger
}

// NewAIDebuggerService creates a new AI debugger service
func NewAIDebuggerService(logger *zap.Logger) *AIDebuggerService {
	return &AIDebuggerService{
		logger: logger,
	}
}

// DebugTrace performs an AI-powered debug query against a trace, building context
// and generating a mock LLM response with root causes and suggested fixes
func (s *AIDebuggerService) DebugTrace(ctx context.Context, projectID, traceID uuid.UUID, query string, queryType domain.DebugQueryType) (*domain.DebugQuery, error) {
	if !queryType.IsValid() {
		return nil, fmt.Errorf("invalid query type: %s", queryType)
	}
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("debug query cannot be empty")
	}

	debugCtx, err := s.BuildContext(ctx, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to build debug context: %w", err)
	}

	now := time.Now()
	response := s.generateResponse(queryType, debugCtx)

	result := &domain.DebugQuery{
		ID:        uuid.New(),
		ProjectID: projectID,
		TraceID:   traceID,
		Query:     query,
		QueryType: queryType,
		Context:   *debugCtx,
		Response:  response,
		CreatedAt: now,
	}

	s.logger.Info("debug query executed",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID.String()),
		zap.String("queryType", string(queryType)),
	)
	return result, nil
}

// generateResponse builds a mock AI response based on the query type and context
func (s *AIDebuggerService) generateResponse(queryType domain.DebugQueryType, debugCtx *domain.DebugContext) domain.DebugResponse {
	now := time.Now()
	response := domain.DebugResponse{
		Confidence:  0.85,
		GeneratedAt: now,
	}

	switch queryType {
	case domain.DebugQueryTypeRootCause:
		response.Answer = "Analysis indicates the primary failure originated from a timeout in the LLM call chain. The model took longer than the configured 30s timeout due to a large context window (>4000 tokens)."
		response.RootCauses = []domain.RootCause{
			{
				Description: "LLM request timeout due to large context window exceeding token limits",
				Confidence:  0.92,
				Evidence:    fmt.Sprintf("Observation exceeded duration threshold; cost total: $%.4f", debugCtx.CostTotal),
				Category:    "timeout",
			},
			{
				Description: "Retry storm caused by cascading failures from initial timeout",
				Confidence:  0.74,
				Evidence:    "Multiple sequential observations with increasing latency pattern detected",
				Category:    "cascading_failure",
			},
		}
		response.SuggestedFixes = []domain.SuggestedFix{
			{
				Description: "Reduce context window size by implementing sliding window or summarization",
				Impact:      "high",
				Effort:      "medium",
			},
			{
				Description: "Add exponential backoff with jitter to retry logic",
				CodeSnippet: "time.Sleep(baseDelay * time.Duration(math.Pow(2, float64(attempt))) + jitter)",
				Impact:      "medium",
				Effort:      "low",
			},
		}

	case domain.DebugQueryTypeExplain:
		response.Answer = fmt.Sprintf("This trace executed %d observations over %dms with a total cost of $%.4f. The execution flow progressed through the agent's tool-calling loop with one error detected.",
			len(debugCtx.Observations), debugCtx.DurationMs, debugCtx.CostTotal)
		response.Confidence = 0.95

	case domain.DebugQueryTypeSuggestFix:
		response.Answer = "Based on the error patterns detected, the following fixes are recommended to improve reliability and reduce cost."
		response.SuggestedFixes = []domain.SuggestedFix{
			{
				Description: "Implement prompt caching for repeated system prompts",
				Impact:      "high",
				Effort:      "low",
			},
			{
				Description: "Add structured output validation to prevent malformed tool calls",
				CodeSnippet: "if err := json.Unmarshal(output, &schema); err != nil { return fallback }",
				Impact:      "medium",
				Effort:      "low",
			},
			{
				Description: "Switch to a streaming response to reduce perceived latency",
				Impact:      "medium",
				Effort:      "medium",
			},
		}

	case domain.DebugQueryTypeCompare:
		response.Answer = "Comparison with baseline traces shows a 23% increase in latency and 15% increase in cost. The primary divergence occurs at the retrieval step where the current trace fetches 3x more documents."
		response.Confidence = 0.80

	case domain.DebugQueryTypeOptimize:
		response.Answer = "Optimization analysis suggests three key improvements: (1) reduce token usage by 40% through prompt compression, (2) cache embedding lookups to eliminate redundant API calls, (3) use a smaller model for classification sub-tasks."
		response.SuggestedFixes = []domain.SuggestedFix{
			{
				Description: "Replace GPT-4 with GPT-3.5-turbo for classification steps (saves ~$0.03/trace)",
				Impact:      "high",
				Effort:      "low",
			},
			{
				Description: "Implement embedding cache with 24h TTL",
				Impact:      "medium",
				Effort:      "medium",
			},
		}
	}

	return response
}

// GetDebugHistory retrieves the history of debug queries for a trace
func (s *AIDebuggerService) GetDebugHistory(ctx context.Context, projectID, traceID uuid.UUID) ([]domain.DebugQuery, error) {
	s.logger.Debug("fetching debug history",
		zap.String("projectId", projectID.String()),
		zap.String("traceId", traceID.String()),
	)
	return []domain.DebugQuery{}, nil
}

// BuildContext assembles trace observations, errors, and git context into
// a structured debugging context
func (s *AIDebuggerService) BuildContext(ctx context.Context, traceID uuid.UUID) (*domain.DebugContext, error) {
	s.logger.Debug("building debug context", zap.String("traceId", traceID.String()))

	observations := []domain.DebugObservation{
		{
			ID:         uuid.New().String(),
			Name:       "ChatCompletion",
			Type:       "generation",
			Model:      "gpt-4",
			DurationMs: 2340,
			Status:     "completed",
			TokensUsed: 1850,
		},
		{
			ID:         uuid.New().String(),
			Name:       "SearchDocuments",
			Type:       "tool",
			DurationMs: 450,
			Status:     "completed",
		},
		{
			ID:         uuid.New().String(),
			Name:       "FormatResponse",
			Type:       "generation",
			Model:      "gpt-3.5-turbo",
			DurationMs: 890,
			Status:     "error",
			TokensUsed: 620,
			Error:      "context length exceeded: 4097 tokens > 4096 max",
		},
	}

	debugCtx := &domain.DebugContext{
		TraceSummary:   fmt.Sprintf("Trace %s: 3 observations, 1 error, completed in 3680ms", traceID.String()[:8]),
		Observations:   observations,
		FileOperations: []string{"src/agent.py", "src/tools/search.py"},
		Errors:         []string{"context length exceeded: 4097 tokens > 4096 max"},
		GitContext: &domain.DebugGitContext{
			Branch:       "feature/rag-pipeline",
			CommitSHA:    "a1b2c3d4e5f6",
			FilesChanged: []string{"src/agent.py", "src/prompts/system.txt"},
		},
		CostTotal:  0.0842,
		DurationMs: 3680,
	}

	return debugCtx, nil
}
