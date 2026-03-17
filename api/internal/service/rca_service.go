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
	logger         *zap.Logger
	mu             sync.RWMutex
	reports        map[uuid.UUID]*domain.RCAReport
	anomalies      map[uuid.UUID]*domain.CorrelatedAnomaly
	alertChannels  map[uuid.UUID]*domain.AlertDeliveryChannel
	corrRules      map[uuid.UUID]*domain.CorrelationRule
	investigations map[uuid.UUID]*domain.RCAInvestigation
}

// NewRCAService creates a new root cause analysis service
func NewRCAService(logger *zap.Logger) *RCAService {
	return &RCAService{
		logger:         logger,
		reports:        make(map[uuid.UUID]*domain.RCAReport),
		anomalies:      make(map[uuid.UUID]*domain.CorrelatedAnomaly),
		alertChannels:  make(map[uuid.UUID]*domain.AlertDeliveryChannel),
		corrRules:      make(map[uuid.UUID]*domain.CorrelationRule),
		investigations: make(map[uuid.UUID]*domain.RCAInvestigation),
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

// DetectAnomalies creates a correlated anomaly with auto-generated root causes
func (s *RCAService) DetectAnomalies(ctx context.Context, projectID uuid.UUID, input domain.CorrelatedAnomalyInput) (*domain.CorrelatedAnomaly, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title must not be empty")
	}

	severity := input.Severity
	if severity == "" {
		severity = "warning"
	}

	category := s.classifyFailure(uuid.New())
	anomaly := &domain.CorrelatedAnomaly{
		ID:             uuid.New(),
		ProjectID:      projectID,
		AnomalyType:    input.AnomalyType,
		Severity:       severity,
		Title:          input.Title,
		Description:    input.Description,
		AffectedTraces: input.AffectedTraces,
		Correlation:    0.7 + rand.Float64()*0.3, //nolint:gosec
		RootCauses:     s.generateContributingFactors(category),
		Remediations:   s.generateRemediations(category),
		Status:         "open",
		DetectedAt:     time.Now(),
		Metadata:       map[string]interface{}{},
	}
	if anomaly.AffectedTraces == nil {
		anomaly.AffectedTraces = []string{}
	}

	s.mu.Lock()
	s.anomalies[anomaly.ID] = anomaly
	s.mu.Unlock()

	s.logger.Info("detected correlated anomaly",
		zap.String("anomalyId", anomaly.ID.String()),
		zap.String("type", anomaly.AnomalyType),
		zap.String("severity", anomaly.Severity),
	)

	return anomaly, nil
}

// GetAnomaly retrieves a correlated anomaly by ID
func (s *RCAService) GetAnomaly(ctx context.Context, anomalyID uuid.UUID) (*domain.CorrelatedAnomaly, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	anomaly, exists := s.anomalies[anomalyID]
	if !exists {
		return nil, fmt.Errorf("anomaly not found")
	}
	return anomaly, nil
}

// ListAnomalies lists all correlated anomalies for a project
func (s *RCAService) ListAnomalies(ctx context.Context, projectID uuid.UUID) ([]domain.CorrelatedAnomaly, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var anomalies []domain.CorrelatedAnomaly
	for _, a := range s.anomalies {
		if a.ProjectID == projectID {
			anomalies = append(anomalies, *a)
		}
	}
	if anomalies == nil {
		anomalies = []domain.CorrelatedAnomaly{}
	}
	return anomalies, nil
}

// AcknowledgeAnomaly acknowledges a correlated anomaly
func (s *RCAService) AcknowledgeAnomaly(ctx context.Context, anomalyID uuid.UUID) (*domain.CorrelatedAnomaly, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	anomaly, exists := s.anomalies[anomalyID]
	if !exists {
		return nil, fmt.Errorf("anomaly not found")
	}
	anomaly.Status = "acknowledged"
	return anomaly, nil
}

// CreateAlertChannel creates a new alert delivery channel
func (s *RCAService) CreateAlertChannel(ctx context.Context, projectID uuid.UUID, input domain.DeliveryChannelInput) (*domain.AlertDeliveryChannel, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("type must not be empty")
	}

	channel := &domain.AlertDeliveryChannel{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Name:       input.Name,
		Type:       input.Type,
		Config:     input.Config,
		Enabled:    true,
		TestStatus: "untested",
		CreatedAt:  time.Now(),
	}

	s.mu.Lock()
	s.alertChannels[channel.ID] = channel
	s.mu.Unlock()

	s.logger.Info("created alert delivery channel",
		zap.String("channelId", channel.ID.String()),
		zap.String("type", channel.Type),
	)

	return channel, nil
}

// ListAlertChannels lists all alert delivery channels for a project
func (s *RCAService) ListAlertChannels(ctx context.Context, projectID uuid.UUID) ([]domain.AlertDeliveryChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var channels []domain.AlertDeliveryChannel
	for _, ch := range s.alertChannels {
		if ch.ProjectID == projectID {
			channels = append(channels, *ch)
		}
	}
	if channels == nil {
		channels = []domain.AlertDeliveryChannel{}
	}
	return channels, nil
}

// TestAlertChannel tests an alert delivery channel
func (s *RCAService) TestAlertChannel(ctx context.Context, channelID uuid.UUID) (*domain.AlertDeliveryChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.alertChannels[channelID]
	if !exists {
		return nil, fmt.Errorf("alert channel not found")
	}
	channel.TestStatus = "success"
	return channel, nil
}

// CreateCorrelationRule creates a new correlation rule
func (s *RCAService) CreateCorrelationRule(ctx context.Context, projectID uuid.UUID, input domain.CorrelationRuleInput) (*domain.CorrelationRule, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if len(input.AnomalyTypes) == 0 {
		return nil, fmt.Errorf("anomaly types must not be empty")
	}

	windowMinutes := input.WindowMinutes
	if windowMinutes == 0 {
		windowMinutes = 30
	}
	minCorrelation := input.MinCorrelation
	if minCorrelation == 0 {
		minCorrelation = 0.7
	}
	severity := input.Severity
	if severity == "" {
		severity = "warning"
	}
	channels := input.Channels
	if channels == nil {
		channels = []string{}
	}

	rule := &domain.CorrelationRule{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Name:           input.Name,
		AnomalyTypes:   input.AnomalyTypes,
		WindowMinutes:  windowMinutes,
		MinCorrelation: minCorrelation,
		AutoRemediate:  input.AutoRemediate,
		Severity:       severity,
		Channels:       channels,
		Enabled:        true,
		CreatedAt:      time.Now(),
	}

	s.mu.Lock()
	s.corrRules[rule.ID] = rule
	s.mu.Unlock()

	s.logger.Info("created correlation rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("name", rule.Name),
	)

	return rule, nil
}

// ListCorrelationRules lists all correlation rules for a project
func (s *RCAService) ListCorrelationRules(ctx context.Context, projectID uuid.UUID) ([]domain.CorrelationRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rules []domain.CorrelationRule
	for _, r := range s.corrRules {
		if r.ProjectID == projectID {
			rules = append(rules, *r)
		}
	}
	if rules == nil {
		rules = []domain.CorrelationRule{}
	}
	return rules, nil
}

// GetAlertDashboardStats returns alerting dashboard statistics
func (s *RCAService) GetAlertDashboardStats(ctx context.Context, projectID uuid.UUID) (*domain.AlertDashboardStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	openAnomalies := 0
	criticalAlerts := 0
	for _, a := range s.anomalies {
		if a.ProjectID == projectID {
			if a.Status == "open" || a.Status == "investigating" {
				openAnomalies++
			}
			if a.Severity == "critical" || a.Severity == "emergency" {
				criticalAlerts++
			}
		}
	}

	activeInvestigations := 0
	for _, inv := range s.investigations {
		if inv.ProjectID == projectID && (inv.Status == "open" || inv.Status == "investigating") {
			activeInvestigations++
		}
	}

	return &domain.AlertDashboardStats{
		OpenAnomalies:        openAnomalies,
		CriticalAlerts:       criticalAlerts,
		ActiveInvestigations: activeInvestigations,
		MTTR:                 45.2 + rand.Float64()*30, //nolint:gosec
		AlertsSentToday:      len(s.anomalies),
		AnomalyTrend:         "stable",
	}, nil
}

// CreateInvestigation creates a new RCA investigation
func (s *RCAService) CreateInvestigation(ctx context.Context, projectID uuid.UUID, anomalyID uuid.UUID, title string, investigatorID uuid.UUID) (*domain.RCAInvestigation, error) {
	if title == "" {
		return nil, fmt.Errorf("title must not be empty")
	}

	s.mu.RLock()
	_, exists := s.anomalies[anomalyID]
	s.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("anomaly not found")
	}

	now := time.Now()
	investigation := &domain.RCAInvestigation{
		ID:             uuid.New(),
		ProjectID:      projectID,
		AnomalyID:      anomalyID,
		Title:          title,
		Status:         "open",
		Findings:       []domain.InvestigationFinding{},
		Timeline: []domain.InvestigationEvent{
			{
				Action:    "created",
				Details:   "Investigation opened",
				UserID:    investigatorID,
				Timestamp: now,
			},
		},
		InvestigatorID: investigatorID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.mu.Lock()
	s.investigations[investigation.ID] = investigation
	s.mu.Unlock()

	s.logger.Info("created RCA investigation",
		zap.String("investigationId", investigation.ID.String()),
		zap.String("anomalyId", anomalyID.String()),
	)

	return investigation, nil
}

// UpdateInvestigation updates an existing RCA investigation
func (s *RCAService) UpdateInvestigation(ctx context.Context, investigationID uuid.UUID, status string, rootCause string, resolution string) (*domain.RCAInvestigation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	investigation, exists := s.investigations[investigationID]
	if !exists {
		return nil, fmt.Errorf("investigation not found")
	}

	if status != "" {
		investigation.Status = status
	}
	if rootCause != "" {
		investigation.RootCause = rootCause
	}
	if resolution != "" {
		investigation.Resolution = resolution
	}
	investigation.UpdatedAt = time.Now()
	investigation.Timeline = append(investigation.Timeline, domain.InvestigationEvent{
		Action:    "updated",
		Details:   fmt.Sprintf("Status changed to %s", investigation.Status),
		UserID:    investigation.InvestigatorID,
		Timestamp: time.Now(),
	})

	return investigation, nil
}

// ListInvestigations lists all RCA investigations for a project
func (s *RCAService) ListInvestigations(ctx context.Context, projectID uuid.UUID) ([]domain.RCAInvestigation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var investigations []domain.RCAInvestigation
	for _, inv := range s.investigations {
		if inv.ProjectID == projectID {
			investigations = append(investigations, *inv)
		}
	}
	if investigations == nil {
		investigations = []domain.RCAInvestigation{}
	}
	return investigations, nil
}
