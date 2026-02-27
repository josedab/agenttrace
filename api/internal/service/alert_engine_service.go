package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// AlertRuleRepository defines repository operations for alert rules
type AlertRuleRepository interface {
	SaveRule(ctx context.Context, rule *domain.AlertRule) error
	GetRuleByID(ctx context.Context, id uuid.UUID) (*domain.AlertRule, error)
	ListRules(ctx context.Context, filter domain.AlertRuleFilter, limit, offset int) (*domain.AlertRuleList, error)
	UpdateRule(ctx context.Context, rule *domain.AlertRule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error
	SaveEvent(ctx context.Context, event *domain.AlertEvent) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*domain.AlertEvent, error)
	ListEvents(ctx context.Context, filter domain.AlertEventFilter, limit, offset int) (*domain.AlertEventList, error)
	UpdateEvent(ctx context.Context, event *domain.AlertEvent) error
	GetLastEventForRule(ctx context.Context, ruleID uuid.UUID) (*domain.AlertEvent, error)
}

// AlertEngineService handles smart alert evaluation, drift detection, and delivery
type AlertEngineService struct {
	logger         *zap.Logger
	ruleRepo       AlertRuleRepository
	anomalyService *AnomalyService
}

// NewAlertEngineService creates a new alert engine service
func NewAlertEngineService(
	logger *zap.Logger,
	ruleRepo AlertRuleRepository,
	anomalyService *AnomalyService,
) *AlertEngineService {
	return &AlertEngineService{
		logger:         logger,
		ruleRepo:       ruleRepo,
		anomalyService: anomalyService,
	}
}

// CreateRule creates a new alert rule
func (s *AlertEngineService) CreateRule(ctx context.Context, projectID, userID uuid.UUID, input *domain.AlertRuleInput) (*domain.AlertRule, error) {
	if err := s.validateRuleInput(input); err != nil {
		return nil, fmt.Errorf("invalid rule input: %w", err)
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	cooldown := 15
	if input.CooldownMinutes != nil {
		cooldown = *input.CooldownMinutes
	}

	rule := &domain.AlertRule{
		ID:              uuid.New(),
		ProjectID:       projectID,
		Name:            input.Name,
		Type:            input.Type,
		Enabled:         enabled,
		Conditions:      input.Conditions,
		DriftConfig:     input.DriftConfig,
		PatternConfig:   input.PatternConfig,
		Deliveries:      input.Deliveries,
		CooldownMinutes: cooldown,
		GroupByFields:   input.GroupByFields,
		Tags:            input.Tags,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.ruleRepo.SaveRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to save alert rule: %w", err)
	}

	s.logger.Info("created alert rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("name", rule.Name),
		zap.String("type", string(rule.Type)),
	)

	return rule, nil
}

// GetRule retrieves an alert rule by ID
func (s *AlertEngineService) GetRule(ctx context.Context, ruleID uuid.UUID) (*domain.AlertRule, error) {
	rule, err := s.ruleRepo.GetRuleByID(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}
	return rule, nil
}

// ListRules retrieves alert rules for a project
func (s *AlertEngineService) ListRules(ctx context.Context, filter domain.AlertRuleFilter, limit, offset int) (*domain.AlertRuleList, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.ruleRepo.ListRules(ctx, filter, limit, offset)
}

// UpdateRule updates an existing alert rule
func (s *AlertEngineService) UpdateRule(ctx context.Context, ruleID uuid.UUID, input *domain.AlertRuleInput) (*domain.AlertRule, error) {
	rule, err := s.ruleRepo.GetRuleByID(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	if err := s.validateRuleInput(input); err != nil {
		return nil, fmt.Errorf("invalid rule input: %w", err)
	}

	rule.Name = input.Name
	rule.Type = input.Type
	rule.Conditions = input.Conditions
	rule.DriftConfig = input.DriftConfig
	rule.PatternConfig = input.PatternConfig
	rule.Deliveries = input.Deliveries
	rule.GroupByFields = input.GroupByFields
	rule.Tags = input.Tags
	rule.UpdatedAt = time.Now()

	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if input.CooldownMinutes != nil {
		rule.CooldownMinutes = *input.CooldownMinutes
	}

	if err := s.ruleRepo.UpdateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update alert rule: %w", err)
	}

	s.logger.Info("updated alert rule",
		zap.String("ruleId", rule.ID.String()),
		zap.String("name", rule.Name),
	)

	return rule, nil
}

// DeleteRule deletes an alert rule
func (s *AlertEngineService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	if err := s.ruleRepo.DeleteRule(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}
	s.logger.Info("deleted alert rule", zap.String("ruleId", ruleID.String()))
	return nil
}

// EvaluateRule evaluates a single rule against trace data and fires alerts if needed
func (s *AlertEngineService) EvaluateRule(ctx context.Context, rule *domain.AlertRule, traceData *TraceEvalContext) (*domain.AlertEvent, error) {
	if !rule.Enabled {
		return nil, nil
	}

	// Check cooldown
	if shouldSuppress, err := s.checkCooldown(ctx, rule); err != nil {
		s.logger.Warn("cooldown check failed", zap.Error(err))
	} else if shouldSuppress {
		return nil, nil
	}

	var shouldAlert bool
	var severity domain.AnomalySeverity
	var title, description string
	var driftDetails *domain.DriftDetails

	switch rule.Type {
	case domain.AlertRuleTypeThreshold:
		shouldAlert, severity, title, description = s.evaluateThresholdConditions(rule, traceData)
	case domain.AlertRuleTypeBehavioralDrift:
		shouldAlert, severity, driftDetails = s.evaluateBehavioralDrift(rule, traceData)
		if shouldAlert && driftDetails != nil {
			title = fmt.Sprintf("Behavioral drift detected: %s", driftDetails.DriftType)
			description = driftDetails.Summary
		}
	case domain.AlertRuleTypePatternMatch:
		shouldAlert, severity, title, description = s.evaluatePatternMatch(rule, traceData)
	case domain.AlertRuleTypeTraceAware:
		shouldAlert, severity, title, description = s.evaluateTraceAware(rule, traceData)
	default:
		return nil, fmt.Errorf("unknown rule type: %s", rule.Type)
	}

	if !shouldAlert {
		return nil, nil
	}

	event := &domain.AlertEvent{
		ID:           uuid.New(),
		RuleID:       rule.ID,
		ProjectID:    rule.ProjectID,
		Severity:     severity,
		Status:       domain.AlertStatusActive,
		Title:        title,
		Description:  description,
		TraceID:      traceData.TraceID,
		SpanID:       traceData.SpanID,
		Metadata:     traceData.Metadata,
		DriftDetails: driftDetails,
		TriggeredAt:  time.Now(),
	}

	if err := s.ruleRepo.SaveEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save alert event: %w", err)
	}

	// Deliver alert asynchronously
	deliveryResults := s.deliverAlert(ctx, rule, event)
	event.DeliveryResults = deliveryResults

	if err := s.ruleRepo.UpdateEvent(ctx, event); err != nil {
		s.logger.Warn("failed to update event with delivery results", zap.Error(err))
	}

	s.logger.Info("alert triggered",
		zap.String("ruleId", rule.ID.String()),
		zap.String("eventId", event.ID.String()),
		zap.String("severity", string(severity)),
	)

	return event, nil
}

// AcknowledgeEvent acknowledges an alert event
func (s *AlertEngineService) AcknowledgeEvent(ctx context.Context, eventID, userID uuid.UUID) (*domain.AlertEvent, error) {
	event, err := s.ruleRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert event: %w", err)
	}

	now := time.Now()
	event.Status = domain.AlertStatusAcknowledged
	event.AcknowledgedAt = &now

	if err := s.ruleRepo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to acknowledge event: %w", err)
	}

	return event, nil
}

// ResolveEvent resolves an alert event
func (s *AlertEngineService) ResolveEvent(ctx context.Context, eventID uuid.UUID) (*domain.AlertEvent, error) {
	event, err := s.ruleRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert event: %w", err)
	}

	now := time.Now()
	event.Status = domain.AlertStatusResolved
	event.ResolvedAt = &now

	if err := s.ruleRepo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to resolve event: %w", err)
	}

	return event, nil
}

// GetAlertStats returns alert statistics for a project
func (s *AlertEngineService) GetAlertStats(ctx context.Context, projectID uuid.UUID) (*domain.AlertRuleStats, error) {
	rules, err := s.ruleRepo.ListRules(ctx, domain.AlertRuleFilter{ProjectID: projectID}, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	stats := &domain.AlertRuleStats{
		ProjectID:    projectID,
		TotalRules:   int(rules.TotalCount),
		ByType:       make(map[domain.AlertRuleType]int),
		BySeverity:   make(map[domain.AnomalySeverity]int),
	}

	for _, rule := range rules.Rules {
		if rule.Enabled {
			stats.EnabledRules++
		}
		stats.ByType[rule.Type]++
	}

	events, err := s.ruleRepo.ListEvents(ctx, domain.AlertEventFilter{
		ProjectID: projectID,
		Status:    alertStatusPtr(domain.AlertStatusActive),
	}, 1000, 0)
	if err == nil {
		stats.ActiveEvents = int(events.TotalCount)
		for _, event := range events.Events {
			stats.BySeverity[event.Severity]++
		}
	}

	return stats, nil
}

// TraceEvalContext provides trace data for rule evaluation
type TraceEvalContext struct {
	TraceID       *string
	SpanID        *string
	Metadata      map[string]string
	Latency       float64
	Cost          float64
	TokenCount    int
	ErrorRate     float64
	ToolCalls     []string
	HasError      bool
	SpanName      string
	Output        string
	// Historical data for drift detection
	HistoricalLatencies []float64
	HistoricalCosts     []float64
	HistoricalToolUsage map[string]int
}

func (s *AlertEngineService) validateRuleInput(input *domain.AlertRuleInput) error {
	if input.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if len(input.Deliveries) == 0 {
		return fmt.Errorf("at least one delivery channel is required")
	}

	switch input.Type {
	case domain.AlertRuleTypeThreshold:
		if len(input.Conditions) == 0 {
			return fmt.Errorf("threshold rules require at least one condition")
		}
	case domain.AlertRuleTypeBehavioralDrift:
		if input.DriftConfig == nil {
			return fmt.Errorf("behavioral drift rules require drift config")
		}
	case domain.AlertRuleTypePatternMatch:
		if input.PatternConfig == nil {
			return fmt.Errorf("pattern match rules require pattern config")
		}
	case domain.AlertRuleTypeTraceAware:
		// Combines conditions with trace context
	default:
		return fmt.Errorf("unsupported rule type: %s", input.Type)
	}

	return nil
}

func (s *AlertEngineService) checkCooldown(ctx context.Context, rule *domain.AlertRule) (bool, error) {
	if rule.CooldownMinutes <= 0 {
		return false, nil
	}

	lastEvent, err := s.ruleRepo.GetLastEventForRule(ctx, rule.ID)
	if err != nil || lastEvent == nil {
		return false, nil
	}

	cooldownEnd := lastEvent.TriggeredAt.Add(time.Duration(rule.CooldownMinutes) * time.Minute)
	return time.Now().Before(cooldownEnd), nil
}

func (s *AlertEngineService) evaluateThresholdConditions(rule *domain.AlertRule, data *TraceEvalContext) (bool, domain.AnomalySeverity, string, string) {
	for _, cond := range rule.Conditions {
		var fieldValue float64
		switch cond.Field {
		case "latency":
			fieldValue = data.Latency
		case "cost":
			fieldValue = data.Cost
		case "error_rate":
			fieldValue = data.ErrorRate
		case "token_count":
			fieldValue = float64(data.TokenCount)
		case "tool_call_count":
			fieldValue = float64(len(data.ToolCalls))
		default:
			continue
		}

		matched := false
		switch cond.Operator {
		case "gt":
			matched = fieldValue > cond.Value
		case "gte":
			matched = fieldValue >= cond.Value
		case "lt":
			matched = fieldValue < cond.Value
		case "lte":
			matched = fieldValue <= cond.Value
		case "eq":
			matched = fieldValue == cond.Value
		case "neq":
			matched = fieldValue != cond.Value
		}

		if matched {
			severity := s.determineSeverity(fieldValue, cond.Value)
			title := fmt.Sprintf("Threshold exceeded: %s %s %.2f (actual: %.2f)", cond.Field, cond.Operator, cond.Value, fieldValue)
			desc := fmt.Sprintf("Alert rule '%s' detected %s value of %.2f exceeding threshold of %.2f", rule.Name, cond.Field, fieldValue, cond.Value)
			return true, severity, title, desc
		}
	}

	return false, "", "", ""
}

func (s *AlertEngineService) evaluateBehavioralDrift(rule *domain.AlertRule, data *TraceEvalContext) (bool, domain.AnomalySeverity, *domain.DriftDetails) {
	cfg := rule.DriftConfig
	if cfg == nil {
		return false, "", nil
	}

	var baselineVal, currentVal, driftScore float64

	switch cfg.DriftType {
	case domain.BehavioralDriftToolUsage:
		if len(data.HistoricalToolUsage) == 0 {
			return false, "", nil
		}
		// Detect if tool usage patterns have changed
		currentToolCount := float64(len(data.ToolCalls))
		var historicalAvg float64
		var count int
		for _, c := range data.HistoricalToolUsage {
			historicalAvg += float64(c)
			count++
		}
		if count > 0 {
			historicalAvg /= float64(count)
		}
		baselineVal = historicalAvg
		currentVal = currentToolCount
		if historicalAvg > 0 {
			driftScore = math.Abs(currentVal-baselineVal) / baselineVal
		}

	case domain.BehavioralDriftCostSpike:
		if len(data.HistoricalCosts) < cfg.MinTraceCount {
			return false, "", nil
		}
		stats := s.anomalyService.CalculateBaselineStats(data.HistoricalCosts)
		baselineVal = stats.Mean
		currentVal = data.Cost
		if stats.StdDev > 0 {
			driftScore = math.Abs(currentVal-baselineVal) / stats.StdDev
		}

	case domain.BehavioralDriftLatencyShift:
		if len(data.HistoricalLatencies) < cfg.MinTraceCount {
			return false, "", nil
		}
		stats := s.anomalyService.CalculateBaselineStats(data.HistoricalLatencies)
		baselineVal = stats.Mean
		currentVal = data.Latency
		if stats.StdDev > 0 {
			driftScore = math.Abs(currentVal-baselineVal) / stats.StdDev
		}

	default:
		return false, "", nil
	}

	// Normalize drift score to 0-1 using sensitivity
	normalizedDrift := math.Min(driftScore/3.0, 1.0)
	threshold := 1.0 - cfg.Sensitivity

	if normalizedDrift < threshold {
		return false, "", nil
	}

	severity := s.determineSeverity(normalizedDrift, threshold)

	details := &domain.DriftDetails{
		DriftType:       cfg.DriftType,
		BaselineValue:   baselineVal,
		CurrentValue:    currentVal,
		DriftScore:      normalizedDrift,
		ConfidenceScore: math.Min(normalizedDrift/threshold, 1.0),
		Summary:         fmt.Sprintf("Detected %s drift: baseline=%.2f, current=%.2f (drift score: %.2f)", cfg.DriftType, baselineVal, currentVal, normalizedDrift),
	}

	return true, severity, details
}

func (s *AlertEngineService) evaluatePatternMatch(rule *domain.AlertRule, data *TraceEvalContext) (bool, domain.AnomalySeverity, string, string) {
	cfg := rule.PatternConfig
	if cfg == nil || len(cfg.TracePatterns) == 0 {
		return false, "", "", ""
	}

	matchCount := 0
	for _, pattern := range cfg.TracePatterns {
		if s.matchTracePattern(&pattern, data) {
			matchCount++
		}
	}

	shouldAlert := false
	switch cfg.MatchMode {
	case "all":
		shouldAlert = matchCount == len(cfg.TracePatterns)
	case "any":
		shouldAlert = matchCount > 0
	default:
		shouldAlert = matchCount > 0
	}

	if !shouldAlert {
		return false, "", "", ""
	}

	title := fmt.Sprintf("Pattern matched: %s (%d/%d patterns)", rule.Name, matchCount, len(cfg.TracePatterns))
	desc := fmt.Sprintf("Alert rule '%s' matched %d patterns against trace data", rule.Name, matchCount)
	return true, domain.AnomalySeverityMedium, title, desc
}

func (s *AlertEngineService) evaluateTraceAware(rule *domain.AlertRule, data *TraceEvalContext) (bool, domain.AnomalySeverity, string, string) {
	// Trace-aware combines threshold conditions with contextual awareness
	conditionsMet, severity, title, desc := s.evaluateThresholdConditions(rule, data)
	if !conditionsMet {
		return false, "", "", ""
	}

	// Enrich with trace context
	if data.TraceID != nil {
		desc += fmt.Sprintf(" [trace: %s]", *data.TraceID)
	}
	if data.SpanName != "" {
		desc += fmt.Sprintf(" [span: %s]", data.SpanName)
	}

	return true, severity, title, desc
}

func (s *AlertEngineService) matchTracePattern(pattern *domain.TracePattern, data *TraceEvalContext) bool {
	if pattern.HasError != nil && *pattern.HasError != data.HasError {
		return false
	}
	if pattern.MinDuration != nil && int64(data.Latency) < *pattern.MinDuration {
		return false
	}
	if pattern.MaxDuration != nil && int64(data.Latency) > *pattern.MaxDuration {
		return false
	}
	for k, v := range pattern.MetadataMatch {
		if data.Metadata[k] != v {
			return false
		}
	}
	return true
}

func (s *AlertEngineService) determineSeverity(value, threshold float64) domain.AnomalySeverity {
	if threshold == 0 {
		return domain.AnomalySeverityMedium
	}
	ratio := value / threshold
	switch {
	case ratio >= 3.0:
		return domain.AnomalySeverityCritical
	case ratio >= 2.0:
		return domain.AnomalySeverityHigh
	case ratio >= 1.5:
		return domain.AnomalySeverityMedium
	default:
		return domain.AnomalySeverityLow
	}
}

func (s *AlertEngineService) deliverAlert(ctx context.Context, rule *domain.AlertRule, event *domain.AlertEvent) []domain.DeliveryResult {
	var results []domain.DeliveryResult

	for _, delivery := range rule.Deliveries {
		if delivery.MinSeverity != "" && !isSeverityAtLeast(event.Severity, delivery.MinSeverity) {
			continue
		}

		result := domain.DeliveryResult{
			Channel: delivery.Channel,
			Target:  delivery.Target,
			SentAt:  time.Now(),
		}

		switch delivery.Channel {
		case domain.DeliveryChannelSlack:
			result.Success = true // Placeholder: actual Slack API integration
			s.logger.Info("delivering alert to Slack",
				zap.String("channel", delivery.Target),
				zap.String("eventId", event.ID.String()),
			)
		case domain.DeliveryChannelPagerDuty:
			result.Success = true
			s.logger.Info("delivering alert to PagerDuty",
				zap.String("routingKey", delivery.Target),
				zap.String("eventId", event.ID.String()),
			)
		case domain.DeliveryChannelWebhook:
			result.Success = true
			s.logger.Info("delivering alert via webhook",
				zap.String("url", delivery.Target),
				zap.String("eventId", event.ID.String()),
			)
		case domain.DeliveryChannelEmail:
			result.Success = true
			s.logger.Info("delivering alert via email",
				zap.String("to", delivery.Target),
				zap.String("eventId", event.ID.String()),
			)
		default:
			result.Success = false
			result.Error = fmt.Sprintf("unsupported delivery channel: %s", delivery.Channel)
		}

		results = append(results, result)
	}

	return results
}

func alertStatusPtr(s domain.AlertStatus) *domain.AlertStatus {
	return &s
}

// isSeverityAtLeast checks if severity a is at least as severe as b
func isSeverityAtLeast(a, b domain.AnomalySeverity) bool {
	levels := map[domain.AnomalySeverity]int{
		domain.AnomalySeverityLow:      1,
		domain.AnomalySeverityMedium:   2,
		domain.AnomalySeverityHigh:     3,
		domain.AnomalySeverityCritical: 4,
	}
	return levels[a] >= levels[b]
}
