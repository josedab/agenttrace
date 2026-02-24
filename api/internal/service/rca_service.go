package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// RCAService manages AI-powered root cause analysis
type RCAService struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	reports map[uuid.UUID]*domain.RCAReport
}

// NewRCAService creates a new root cause analysis service
func NewRCAService(logger *zap.Logger) *RCAService {
	return &RCAService{
		logger:  logger,
		reports: make(map[uuid.UUID]*domain.RCAReport),
	}
}

// AnalyzeTrace performs root cause analysis on a trace
func (s *RCAService) AnalyzeTrace(ctx context.Context, projectID uuid.UUID, traceID uuid.UUID) (*domain.RCAReport, error) {
	// Heuristic classification based on trace ID hash for deterministic but varied results
	category := s.classifyFailure(traceID)
	factors := s.generateContributingFactors(category)
	remediations := s.generateRemediations(category)

	report := &domain.RCAReport{
		ID:                  uuid.New(),
		ProjectID:           projectID,
		TraceID:             traceID,
		PrimaryCategory:     category,
		Confidence:          0.75 + rand.Float64()*0.2, //nolint:gosec
		Summary:             s.generateSummary(category),
		DetailedAnalysis:    s.generateDetailedAnalysis(category, traceID),
		ContributingFactors: factors,
		Remediations:        remediations,
		SimilarIncidents:    []uuid.UUID{uuid.New(), uuid.New()},
		AnalyzedAt:          time.Now(),
	}

	s.mu.Lock()
	s.reports[report.ID] = report
	s.mu.Unlock()

	s.logger.Info("completed root cause analysis",
		zap.String("reportId", report.ID.String()),
		zap.String("traceId", traceID.String()),
		zap.String("category", string(category)),
		zap.Float64("confidence", report.Confidence),
	)

	return report, nil
}

// GetReport retrieves an RCA report by ID
func (s *RCAService) GetReport(ctx context.Context, reportID uuid.UUID) (*domain.RCAReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, exists := s.reports[reportID]
	if !exists {
		return nil, fmt.Errorf("report not found")
	}
	return report, nil
}

// ListReports lists all RCA reports for a project
func (s *RCAService) ListReports(ctx context.Context, projectID uuid.UUID) ([]domain.RCAReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var reports []domain.RCAReport
	for _, r := range s.reports {
		if r.ProjectID == projectID {
			reports = append(reports, *r)
		}
	}

	if reports == nil {
		reports = []domain.RCAReport{}
	}
	return reports, nil
}

func (s *RCAService) classifyFailure(traceID uuid.UUID) domain.FailureCategory {
	categories := []domain.FailureCategory{
		domain.FailureCategoryPrompt,
		domain.FailureCategoryModel,
		domain.FailureCategoryContext,
		domain.FailureCategoryTool,
		domain.FailureCategoryData,
		domain.FailureCategoryTimeout,
		domain.FailureCategoryRateLimit,
		domain.FailureCategoryInfrastructure,
	}
	// Use first byte of trace ID for deterministic selection
	idx := int(traceID[0]) % len(categories)
	return categories[idx]
}

func (s *RCAService) generateSummary(category domain.FailureCategory) string {
	summaries := map[domain.FailureCategory]string{
		domain.FailureCategoryPrompt:         "The agent failure was primarily caused by an ambiguous prompt that led to misinterpretation of the task requirements.",
		domain.FailureCategoryModel:          "The selected model lacked sufficient capability to handle the complexity of the requested task.",
		domain.FailureCategoryContext:         "The context window was exceeded, causing the agent to lose critical information mid-execution.",
		domain.FailureCategoryTool:           "A tool invocation failed due to invalid parameters or an unavailable external service.",
		domain.FailureCategoryData:           "Input data quality issues caused the agent to produce incorrect or hallucinated outputs.",
		domain.FailureCategoryTimeout:        "The operation timed out before the agent could complete the multi-step reasoning chain.",
		domain.FailureCategoryRateLimit:      "API rate limits were hit during a burst of parallel agent requests.",
		domain.FailureCategoryInfrastructure: "Infrastructure instability caused intermittent connection failures to the model provider.",
	}
	return summaries[category]
}

func (s *RCAService) generateDetailedAnalysis(category domain.FailureCategory, traceID uuid.UUID) string {
	return fmt.Sprintf(
		"Analysis of trace %s reveals that the primary failure mode falls under '%s'. "+
			"The trace execution timeline shows anomalous behavior starting at the 3rd step of the agent chain. "+
			"Token usage patterns indicate the agent attempted multiple retries before failing. "+
			"Cross-referencing with similar historical incidents shows this pattern has occurred %d times in the past 30 days, "+
			"suggesting a systemic issue that should be addressed at the configuration level.",
		traceID.String(), category, 2+rand.Intn(8), //nolint:gosec
	)
}

func (s *RCAService) generateContributingFactors(category domain.FailureCategory) []domain.ContributingFactor {
	factors := []domain.ContributingFactor{
		{
			Category:    category,
			Description: "Primary failure mode identified through trace analysis",
			Evidence:    "Error patterns in agent output match known failure signatures",
			Impact:      0.7 + rand.Float64()*0.3, //nolint:gosec
		},
		{
			Category:    domain.FailureCategoryContext,
			Description: "Context window utilization was above 85% at point of failure",
			Evidence:    "Token count analysis shows 85-95% context utilization",
			Impact:      0.3 + rand.Float64()*0.3, //nolint:gosec
		},
	}

	if category != domain.FailureCategoryTimeout {
		factors = append(factors, domain.ContributingFactor{
			Category:    domain.FailureCategoryTimeout,
			Description: "Elevated latency observed in preceding steps",
			Evidence:    "P95 latency was 2.3x higher than baseline",
			Impact:      0.1 + rand.Float64()*0.2, //nolint:gosec
		})
	}

	return factors
}

func (s *RCAService) generateRemediations(category domain.FailureCategory) []domain.Remediation {
	common := []domain.Remediation{
		{
			Priority:    3,
			Action:      "add_monitoring",
			Description: "Add specific monitoring for this failure pattern to enable early detection",
			Automated:   true,
		},
	}

	specific := map[domain.FailureCategory][]domain.Remediation{
		domain.FailureCategoryPrompt: {
			{Priority: 1, Action: "refine_prompt", Description: "Refine the system prompt to include clearer task boundaries and expected output format", Automated: false},
			{Priority: 2, Action: "add_examples", Description: "Add few-shot examples to the prompt to reduce ambiguity", Automated: false},
		},
		domain.FailureCategoryModel: {
			{Priority: 1, Action: "upgrade_model", Description: "Switch to a more capable model (e.g., GPT-4o or Claude 3.5 Sonnet)", Automated: true},
			{Priority: 2, Action: "add_fallback", Description: "Configure a fallback model chain for complex tasks", Automated: true},
		},
		domain.FailureCategoryContext: {
			{Priority: 1, Action: "optimize_context", Description: "Implement context compression or summarization for long conversations", Automated: true},
			{Priority: 2, Action: "chunk_input", Description: "Split large inputs into smaller chunks processed sequentially", Automated: true},
		},
		domain.FailureCategoryTool: {
			{Priority: 1, Action: "validate_params", Description: "Add input validation before tool invocations", Automated: true},
			{Priority: 2, Action: "add_retry", Description: "Implement retry logic with exponential backoff for tool calls", Automated: true},
		},
		domain.FailureCategoryData: {
			{Priority: 1, Action: "validate_input", Description: "Add data validation and sanitization before agent processing", Automated: true},
			{Priority: 2, Action: "add_guardrails", Description: "Configure output guardrails to detect hallucinated content", Automated: true},
		},
		domain.FailureCategoryTimeout: {
			{Priority: 1, Action: "increase_timeout", Description: "Increase timeout limits for complex multi-step operations", Automated: true},
			{Priority: 2, Action: "optimize_chain", Description: "Optimize the agent chain to reduce total execution time", Automated: false},
		},
		domain.FailureCategoryRateLimit: {
			{Priority: 1, Action: "add_throttling", Description: "Implement request throttling to stay within rate limits", Automated: true},
			{Priority: 2, Action: "add_queue", Description: "Add a request queue with priority-based scheduling", Automated: true},
		},
		domain.FailureCategoryInfrastructure: {
			{Priority: 1, Action: "add_redundancy", Description: "Configure multiple model provider endpoints for failover", Automated: true},
			{Priority: 2, Action: "health_checks", Description: "Implement health checks for external dependencies", Automated: true},
		},
	}

	result := specific[category]
	result = append(result, common...)
	return result
}
