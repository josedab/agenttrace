package domain

import (
	"time"

	"github.com/google/uuid"
)

// AlertRuleType represents the type of alert rule
type AlertRuleType string

const (
	AlertRuleTypeThreshold      AlertRuleType = "threshold"
	AlertRuleTypeBehavioralDrift AlertRuleType = "behavioral_drift"
	AlertRuleTypePatternMatch   AlertRuleType = "pattern_match"
	AlertRuleTypeTraceAware     AlertRuleType = "trace_aware"
)

// DeliveryChannel represents the delivery method for alerts
type DeliveryChannel string

const (
	DeliveryChannelSlack     DeliveryChannel = "slack"
	DeliveryChannelPagerDuty DeliveryChannel = "pagerduty"
	DeliveryChannelWebhook   DeliveryChannel = "webhook"
	DeliveryChannelEmail     DeliveryChannel = "email"
	DeliveryChannelTeams     DeliveryChannel = "teams"
)

// BehavioralDriftType represents the kind of behavioral drift detected
type BehavioralDriftType string

const (
	BehavioralDriftToolUsage     BehavioralDriftType = "tool_usage_change"
	BehavioralDriftOutputQuality BehavioralDriftType = "output_quality_degraded"
	BehavioralDriftCostSpike     BehavioralDriftType = "cost_spike"
	BehavioralDriftLatencyShift  BehavioralDriftType = "latency_shift"
	BehavioralDriftErrorPattern  BehavioralDriftType = "error_pattern_change"
	BehavioralDriftDecisionPath  BehavioralDriftType = "decision_path_divergence"
)

// AlertRule defines a smart alerting rule with trace awareness
type AlertRule struct {
	ID        uuid.UUID     `json:"id"`
	ProjectID uuid.UUID     `json:"projectId"`
	Name      string        `json:"name"`
	Type      AlertRuleType `json:"type"`
	Enabled   bool          `json:"enabled"`

	// Conditions for triggering
	Conditions []AlertCondition `json:"conditions"`

	// Behavioral drift settings
	DriftConfig *DriftDetectionConfig `json:"driftConfig,omitempty"`

	// Pattern matching
	PatternConfig *PatternMatchConfig `json:"patternConfig,omitempty"`

	// Delivery configuration
	Deliveries []DeliveryConfig `json:"deliveries"`

	// Suppression and dedup
	CooldownMinutes int               `json:"cooldownMinutes"`
	GroupByFields   []string          `json:"groupByFields,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`

	// Audit
	CreatedBy uuid.UUID `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AlertCondition defines a single condition within an alert rule
type AlertCondition struct {
	Field    string  `json:"field"`    // e.g., "latency", "cost", "error_rate", "tool_call_count"
	Operator string  `json:"operator"` // gt, gte, lt, lte, eq, neq, contains, not_contains
	Value    float64 `json:"value"`
	Window   string  `json:"window,omitempty"` // e.g., "5m", "1h", "24h"
}

// DriftDetectionConfig configures behavioral drift detection
type DriftDetectionConfig struct {
	DriftType       BehavioralDriftType `json:"driftType"`
	BaselineWindow  string              `json:"baselineWindow"`  // e.g., "7d"
	CompareWindow   string              `json:"compareWindow"`   // e.g., "1h"
	Sensitivity     float64             `json:"sensitivity"`     // 0.0-1.0
	MinTraceCount   int                 `json:"minTraceCount"`   // Minimum traces before detection fires
	TraceNameFilter string              `json:"traceNameFilter,omitempty"`
	// Specific drift detection fields
	MonitoredTools  []string `json:"monitoredTools,omitempty"`
	QualityMetric   string   `json:"qualityMetric,omitempty"` // Score field to monitor
}

// PatternMatchConfig configures pattern-based alerting
type PatternMatchConfig struct {
	TracePatterns []TracePattern `json:"tracePatterns"`
	MatchMode     string         `json:"matchMode"` // "all", "any"
}

// TracePattern defines a pattern to match against traces
type TracePattern struct {
	SpanNameRegex   string            `json:"spanNameRegex,omitempty"`
	MetadataMatch   map[string]string `json:"metadataMatch,omitempty"`
	MinDuration     *int64            `json:"minDurationMs,omitempty"`
	MaxDuration     *int64            `json:"maxDurationMs,omitempty"`
	HasError        *bool             `json:"hasError,omitempty"`
	OutputContains  string            `json:"outputContains,omitempty"`
}

// DeliveryConfig defines how to deliver an alert
type DeliveryConfig struct {
	Channel     DeliveryChannel `json:"channel"`
	Target      string          `json:"target"`      // URL, email, channel name
	APIKey      string          `json:"apiKey,omitempty"`
	Template    string          `json:"template,omitempty"` // Custom message template
	MinSeverity AnomalySeverity `json:"minSeverity,omitempty"`
}

// AlertEvent represents a triggered alert event
type AlertEvent struct {
	ID        uuid.UUID       `json:"id"`
	RuleID    uuid.UUID       `json:"ruleId"`
	ProjectID uuid.UUID       `json:"projectId"`
	Severity  AnomalySeverity `json:"severity"`
	Status    AlertStatus     `json:"status"`

	Title       string `json:"title"`
	Description string `json:"description"`

	// Trace context
	TraceID   *string           `json:"traceId,omitempty"`
	SpanID    *string           `json:"spanId,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`

	// Drift details
	DriftDetails *DriftDetails `json:"driftDetails,omitempty"`

	// Delivery status
	DeliveryResults []DeliveryResult `json:"deliveryResults,omitempty"`

	TriggeredAt    time.Time  `json:"triggeredAt"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
}

// DriftDetails contains details about a detected behavioral drift
type DriftDetails struct {
	DriftType      BehavioralDriftType `json:"driftType"`
	BaselineValue  float64             `json:"baselineValue"`
	CurrentValue   float64             `json:"currentValue"`
	DriftScore     float64             `json:"driftScore"`     // 0-1 magnitude
	ConfidenceScore float64            `json:"confidenceScore"` // 0-1
	AffectedTraces int                 `json:"affectedTraces"`
	Summary        string              `json:"summary"`
	Examples       []DriftExample      `json:"examples,omitempty"`
}

// DriftExample provides a concrete example of the detected drift
type DriftExample struct {
	TraceID     string    `json:"traceId"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// DeliveryResult tracks the result of delivering an alert
type DeliveryResult struct {
	Channel    DeliveryChannel `json:"channel"`
	Target     string          `json:"target"`
	Success    bool            `json:"success"`
	SentAt     time.Time       `json:"sentAt"`
	Error      string          `json:"error,omitempty"`
	ResponseID string          `json:"responseId,omitempty"` // e.g., Slack message TS
}

// AlertRuleInput represents input for creating/updating an alert rule
type AlertRuleInput struct {
	Name            string               `json:"name" validate:"required,min=1,max=200"`
	Type            AlertRuleType        `json:"type" validate:"required"`
	Enabled         *bool                `json:"enabled,omitempty"`
	Conditions      []AlertCondition     `json:"conditions,omitempty"`
	DriftConfig     *DriftDetectionConfig `json:"driftConfig,omitempty"`
	PatternConfig   *PatternMatchConfig  `json:"patternConfig,omitempty"`
	Deliveries      []DeliveryConfig     `json:"deliveries" validate:"required,min=1"`
	CooldownMinutes *int                 `json:"cooldownMinutes,omitempty"`
	GroupByFields   []string             `json:"groupByFields,omitempty"`
	Tags            map[string]string    `json:"tags,omitempty"`
}

// AlertRuleFilter represents filter options for querying alert rules
type AlertRuleFilter struct {
	ProjectID uuid.UUID
	Type      *AlertRuleType
	Enabled   *bool
}

// AlertEventFilter represents filter options for querying alert events
type AlertEventFilter struct {
	ProjectID uuid.UUID
	RuleID    *uuid.UUID
	Status    *AlertStatus
	Severity  *AnomalySeverity
	StartTime *time.Time
	EndTime   *time.Time
}

// AlertRuleList represents a paginated list of alert rules
type AlertRuleList struct {
	Rules      []AlertRule `json:"rules"`
	TotalCount int64       `json:"totalCount"`
	HasMore    bool        `json:"hasMore"`
}

// AlertEventList represents a paginated list of alert events
type AlertEventList struct {
	Events     []AlertEvent `json:"events"`
	TotalCount int64        `json:"totalCount"`
	HasMore    bool         `json:"hasMore"`
}

// AlertRuleStats provides summary statistics for alert rules
type AlertRuleStats struct {
	ProjectID       uuid.UUID                 `json:"projectId"`
	TotalRules      int                       `json:"totalRules"`
	EnabledRules    int                       `json:"enabledRules"`
	TotalEvents     int                       `json:"totalEvents"`
	ActiveEvents    int                       `json:"activeEvents"`
	ByType          map[AlertRuleType]int     `json:"byType"`
	BySeverity      map[AnomalySeverity]int   `json:"bySeverity"`
	TopFiringRules  []RuleFiringCount         `json:"topFiringRules"`
	EventsOverTime  []TimeSeriesPoint         `json:"eventsOverTime"`
}

// RuleFiringCount tracks how often a rule fires
type RuleFiringCount struct {
	RuleID   uuid.UUID `json:"ruleId"`
	RuleName string    `json:"ruleName"`
	Count    int       `json:"count"`
}
