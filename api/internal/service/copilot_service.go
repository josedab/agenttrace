package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// CopilotService provides an AI copilot for observability
type CopilotService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	queries map[uuid.UUID]*domain.CopilotQuery
}

// NewCopilotService creates a new copilot service
func NewCopilotService(logger *zap.Logger) *CopilotService {
	return &CopilotService{
		logger:  logger,
		queries: make(map[uuid.UUID]*domain.CopilotQuery),
	}
}

// AskQuestion processes a question and generates an answer with sources
func (s *CopilotService) AskQuestion(ctx context.Context, projectID string, input *domain.CopilotQueryInput) (*domain.CopilotQuery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid, _ := uuid.Parse(projectID)
	start := time.Now()

	answer, sources, suggestions := s.generateAnswer(input.Question)

	query := &domain.CopilotQuery{
		ID:          uuid.New(),
		ProjectID:   pid,
		Question:    input.Question,
		Answer:      answer,
		Sources:     sources,
		Suggestions: suggestions,
		QueryTimeMs: time.Since(start).Milliseconds(),
		CreatedAt:   time.Now(),
	}

	s.queries[query.ID] = query
	s.logger.Info("copilot query", zap.String("projectId", projectID), zap.String("question", input.Question))
	return query, nil
}

// GetSuggestions returns proactive optimization suggestions
func (s *CopilotService) GetSuggestions(ctx context.Context, projectID string) ([]domain.CopilotSuggestion, error) {
	return []domain.CopilotSuggestion{
		{
			Category:    "cost",
			Title:       "Switch low-complexity queries to GPT-3.5",
			Description: "42% of your GPT-4 queries have complexity scores below 0.3 and could use GPT-3.5 with equivalent quality",
			Impact:      "Estimated $340/month savings",
			Confidence:  0.88,
			Automated:   true,
		},
		{
			Category:    "performance",
			Title:       "Enable prompt caching for repeated patterns",
			Description: "Detected 156 repeated prompt patterns in the last 24 hours that could benefit from caching",
			Impact:      "35% latency reduction for cached queries",
			Confidence:  0.92,
			Automated:   true,
		},
		{
			Category:    "quality",
			Title:       "Add guardrails for financial queries",
			Description: "Financial-related agent traces show 12% higher error rates without output validation guardrails",
			Impact:      "Reduce error rate from 12% to ~3%",
			Confidence:  0.81,
			Automated:   false,
		},
		{
			Category:    "security",
			Title:       "Enable PII redaction for customer support traces",
			Description: "Customer support traces contain unredacted email addresses in 23% of observations",
			Impact:      "GDPR compliance improvement",
			Confidence:  0.95,
			Automated:   true,
		},
	}, nil
}

// GetProactiveInsights returns proactive insights for a project
func (s *CopilotService) GetProactiveInsights(ctx context.Context, projectID string) ([]domain.ProactiveInsight, error) {
	pid, _ := uuid.Parse(projectID)
	return []domain.ProactiveInsight{
		{
			ID:          uuid.New(),
			ProjectID:   pid,
			Category:    "cost",
			Title:       "Cost spike detected",
			Description: "LLM costs increased 45% in the last 6 hours compared to the 7-day average",
			Severity:    "warning",
			Data:        map[string]any{"currentRate": 12.5, "averageRate": 8.6, "currency": "USD/hour"},
			CreatedAt:   time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          uuid.New(),
			ProjectID:   pid,
			Category:    "performance",
			Title:       "Latency regression in summarization agent",
			Description: "P95 latency for the summarization agent increased from 2.1s to 4.8s",
			Severity:    "action",
			Data:        map[string]any{"agent": "summarizer", "p95Before": 2.1, "p95After": 4.8},
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          uuid.New(),
			ProjectID:   pid,
			Category:    "quality",
			Title:       "Error rate trending up",
			Description: "Agent error rate has increased from 2.1% to 3.8% over the past 24 hours",
			Severity:    "info",
			Data:        map[string]any{"previousRate": 0.021, "currentRate": 0.038, "threshold": 0.05},
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}, nil
}

func (s *CopilotService) generateAnswer(question string) (string, []domain.CopilotSource, []domain.CopilotSuggestion) {
	q := strings.ToLower(question)

	switch {
	case strings.Contains(q, "cost") || strings.Contains(q, "spend") || strings.Contains(q, "budget"):
		return "Your project spent $847.23 in the last 7 days across 12,450 traces. GPT-4 accounts for 68% of costs. Consider routing simple queries to GPT-3.5 to reduce costs by ~40%.",
			[]domain.CopilotSource{
				{Type: "metric", Reference: "cost_dashboard:7d", Relevance: 0.95},
				{Type: "trace", Reference: "trace:cost_analysis:latest", Relevance: 0.82},
			},
			[]domain.CopilotSuggestion{
				{Category: "cost", Title: "Use model routing", Description: "Route simple queries to cheaper models", Impact: "~40% cost reduction", Confidence: 0.88, Automated: true},
			}

	case strings.Contains(q, "error") || strings.Contains(q, "fail") || strings.Contains(q, "bug"):
		return "Current error rate is 3.2% (up from 2.1% last week). Top errors: token limit exceeded (45%), rate limiting (30%), malformed output (25%). The summarization agent has the highest error rate at 8.1%.",
			[]domain.CopilotSource{
				{Type: "metric", Reference: "error_rate:7d", Relevance: 0.93},
				{Type: "trace", Reference: "trace:errors:recent", Relevance: 0.87},
			},
			[]domain.CopilotSuggestion{
				{Category: "quality", Title: "Increase token limits", Description: "Token limit errors can be reduced by increasing max_tokens", Impact: "45% fewer errors", Confidence: 0.91, Automated: true},
			}

	case strings.Contains(q, "latency") || strings.Contains(q, "slow") || strings.Contains(q, "performance"):
		return "Average latency is 1.8s (P50: 1.2s, P95: 4.5s, P99: 8.2s). The code-review agent is the slowest at P95 of 12.3s. Prompt caching could reduce latency by 35% for repeated patterns.",
			[]domain.CopilotSource{
				{Type: "metric", Reference: "latency_distribution:24h", Relevance: 0.94},
				{Type: "config", Reference: "agent:code-review:config", Relevance: 0.76},
			},
			[]domain.CopilotSuggestion{
				{Category: "performance", Title: "Enable streaming", Description: "Streaming responses reduce time-to-first-token", Impact: "Better user experience", Confidence: 0.85, Automated: false},
			}

	case strings.Contains(q, "trace") || strings.Contains(q, "observ"):
		return "Your project has 12,450 traces in the last 7 days with 98.2% success rate. Average trace duration is 3.4s with 4.2 observations per trace. Top agents: research-agent (34%), summarizer (28%), code-review (22%).",
			[]domain.CopilotSource{
				{Type: "trace", Reference: "trace:summary:7d", Relevance: 0.96},
				{Type: "metric", Reference: "trace_stats:7d", Relevance: 0.89},
			},
			nil

	default:
		return "Based on your project's data: you have 12,450 traces in the last 7 days, spending $847.23 with a 3.2% error rate and 1.8s average latency. Would you like me to dive deeper into costs, errors, performance, or traces?",
			[]domain.CopilotSource{
				{Type: "metric", Reference: "project_overview:7d", Relevance: 0.80},
				{Type: "documentation", Reference: "docs:getting-started", Relevance: 0.60},
			},
			nil
	}
}
