package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	AnomalyTypeLatency   AnomalyType = "latency"
	AnomalyTypeCost      AnomalyType = "cost"
	AnomalyTypeErrorRate AnomalyType = "error_rate"
	AnomalyTypeTokens    AnomalyType = "tokens"
	AnomalyTypeCustom    AnomalyType = "custom"
)

// AnomalySeverity represents the severity level of an anomaly
type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "low"
	AnomalySeverityMedium   AnomalySeverity = "medium"
	AnomalySeverityHigh     AnomalySeverity = "high"
	AnomalySeverityCritical AnomalySeverity = "critical"
)

// DetectionMethod represents the algorithm used for detection
type DetectionMethod string

const (
	DetectionMethodZScore         DetectionMethod = "z_score"
	DetectionMethodIQR            DetectionMethod = "iqr"
	DetectionMethodMAD            DetectionMethod = "mad" // Median Absolute Deviation
	DetectionMethodMovingAverage  DetectionMethod = "moving_average"
	DetectionMethodExponentialEMA DetectionMethod = "exponential_ema"
	DetectionMethodThreshold      DetectionMethod = "threshold"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
)

// AnomalyRule defines detection rules for a project
type AnomalyRule struct {
	ID        uuid.UUID       `json:"id"`
	ProjectID uuid.UUID       `json:"projectId"`
	Name      string          `json:"name"`
	Enabled   bool            `json:"enabled"`
	Type      AnomalyType     `json:"type"`
	Method    DetectionMethod `json:"method"`

	// Configuration
	Config AnomalyRuleConfig `json:"config"`

	// Filters
	TraceNameFilter string            `json:"traceNameFilter,omitempty"`
	MetadataFilters map[string]string `json:"metadataFilters,omitempty"`

	// Alert settings
	AlertChannels []uuid.UUID     `json:"alertChannels"` // Webhook IDs to notify
	Severity      AnomalySeverity `json:"severity"`
	Cooldown      int             `json:"cooldownMinutes"` // Minutes between repeated alerts

	// Audit
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy uuid.UUID `json:"createdBy"`
}

// AnomalyRuleConfig contains method-specific configuration
type AnomalyRuleConfig struct {
	// Z-Score configuration
	ZScoreThreshold float64 `json:"zScoreThreshold,omitempty"` // Default: 3.0

	// IQR configuration
	IQRMultiplier float64 `json:"iqrMultiplier,omitempty"` // Default: 1.5

	// MAD configuration
	MADThreshold float64 `json:"madThreshold,omitempty"` // Default: 3.0

	// Moving average configuration
	WindowSize int     `json:"windowSize,omitempty"` // Number of data points
	Deviation  float64 `json:"deviation,omitempty"`  // Percentage deviation

	// Exponential moving average
	Alpha float64 `json:"alpha,omitempty"` // Smoothing factor (0-1)

	// Static threshold
	MinThreshold *float64 `json:"minThreshold,omitempty"`
	MaxThreshold *float64 `json:"maxThreshold,omitempty"`

	// Minimum samples required before detection
	MinSamples int `json:"minSamples,omitempty"` // Default: 30

	// Lookback period for baseline calculation
	LookbackHours int `json:"lookbackHours,omitempty"` // Default: 24
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID        uuid.UUID       `json:"id"`
	ProjectID uuid.UUID       `json:"projectId"`
	RuleID    uuid.UUID       `json:"ruleId"`
	Type      AnomalyType     `json:"type"`
	Severity  AnomalySeverity `json:"severity"`

	// Detection details
	DetectedAt  time.Time       `json:"detectedAt"`
	Method      DetectionMethod `json:"method"`
	Score       float64         `json:"score"`       // Z-score or deviation value
	Value       float64         `json:"value"`       // Actual observed value
	Expected    float64         `json:"expected"`    // Expected/baseline value
	Threshold   float64         `json:"threshold"`   // Threshold that was exceeded
	Description string          `json:"description"` // Human-readable description

	// Context
	TraceID      *uuid.UUID        `json:"traceId,omitempty"`
	TraceName    string            `json:"traceName,omitempty"`
	SpanID       *uuid.UUID        `json:"spanId,omitempty"`
	SpanName     string            `json:"spanName,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	TimeWindow   TimeWindow        `json:"timeWindow"`
	SampleCount  int               `json:"sampleCount"`
	BaselineStats BaselineStats    `json:"baselineStats"`

	// Alert tracking
	AlertsSent []AlertRecord `json:"alertsSent,omitempty"`
}

// TimeWindow represents the time window for anomaly detection
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// BaselineStats contains statistical baseline information
type BaselineStats struct {
	Mean     float64 `json:"mean"`
	StdDev   float64 `json:"stdDev"`
	Median   float64 `json:"median"`
	P95      float64 `json:"p95"`
	P99      float64 `json:"p99"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Q1       float64 `json:"q1"`
	Q3       float64 `json:"q3"`
	IQR      float64 `json:"iqr"`
	MAD      float64 `json:"mad"` // Median Absolute Deviation
}

// AlertRecord tracks alerts sent for an anomaly
type AlertRecord struct {
	WebhookID   uuid.UUID `json:"webhookId"`
	WebhookName string    `json:"webhookName"`
	SentAt      time.Time `json:"sentAt"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
}

// Alert represents an active alert for user action
type Alert struct {
	ID        uuid.UUID       `json:"id"`
	ProjectID uuid.UUID       `json:"projectId"`
	AnomalyID uuid.UUID       `json:"anomalyId"`
	RuleID    uuid.UUID       `json:"ruleId"`
	Status    AlertStatus     `json:"status"`
	Severity  AnomalySeverity `json:"severity"`

	// Alert details
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        AnomalyType `json:"type"`

	// Metrics snapshot
	CurrentValue  float64 `json:"currentValue"`
	ExpectedValue float64 `json:"expectedValue"`
	Deviation     float64 `json:"deviation"` // Percentage deviation

	// Timeline
	TriggeredAt    time.Time  `json:"triggeredAt"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy *uuid.UUID `json:"acknowledgedBy,omitempty"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy     *uuid.UUID `json:"resolvedBy,omitempty"`

	// Notes
	Notes []AlertNote `json:"notes,omitempty"`
}

// AlertNote represents a note added to an alert
type AlertNote struct {
	ID        uuid.UUID `json:"id"`
	AlertID   uuid.UUID `json:"alertId"`
	UserID    uuid.UUID `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// AnomalyRuleInput represents input for creating/updating a rule
type AnomalyRuleInput struct {
	Name            string            `json:"name" validate:"required,min=1,max=100"`
	Enabled         *bool             `json:"enabled,omitempty"`
	Type            AnomalyType       `json:"type" validate:"required"`
	Method          DetectionMethod   `json:"method" validate:"required"`
	Config          AnomalyRuleConfig `json:"config"`
	TraceNameFilter string            `json:"traceNameFilter,omitempty"`
	MetadataFilters map[string]string `json:"metadataFilters,omitempty"`
	AlertChannels   []uuid.UUID       `json:"alertChannels,omitempty"`
	Severity        AnomalySeverity   `json:"severity" validate:"required"`
	Cooldown        *int              `json:"cooldownMinutes,omitempty"`
}

// AnomalyFilter represents filter options for querying anomalies
type AnomalyFilter struct {
	ProjectID  uuid.UUID
	RuleID     *uuid.UUID
	Type       *AnomalyType
	Severity   *AnomalySeverity
	StartTime  *time.Time
	EndTime    *time.Time
	TraceID    *uuid.UUID
	TraceName  string
}

// AlertFilter represents filter options for querying alerts
type AlertFilter struct {
	ProjectID uuid.UUID
	Status    *AlertStatus
	Severity  *AnomalySeverity
	Type      *AnomalyType
	StartTime *time.Time
	EndTime   *time.Time
}

// AnomalyList represents a paginated list of anomalies
type AnomalyList struct {
	Anomalies  []Anomaly `json:"anomalies"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}

// AlertList represents a paginated list of alerts
type AlertList struct {
	Alerts     []Alert `json:"alerts"`
	TotalCount int64   `json:"totalCount"`
	HasMore    bool    `json:"hasMore"`
}

// AnomalyStats provides summary statistics for a project
type AnomalyStats struct {
	ProjectID           uuid.UUID              `json:"projectId"`
	Period              TimeWindow             `json:"period"`
	TotalAnomalies      int                    `json:"totalAnomalies"`
	ActiveAlerts        int                    `json:"activeAlerts"`
	BySeverity          map[AnomalySeverity]int `json:"bySeverity"`
	ByType              map[AnomalyType]int    `json:"byType"`
	TopAffectedTraces   []TraceAnomalyCount    `json:"topAffectedTraces"`
	AnomaliesOverTime   []TimeSeriesPoint      `json:"anomaliesOverTime"`
}

// TraceAnomalyCount tracks anomaly counts per trace
type TraceAnomalyCount struct {
	TraceName string `json:"traceName"`
	Count     int    `json:"count"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// DetectionResult represents the result of running anomaly detection
type DetectionResult struct {
	IsAnomaly   bool            `json:"isAnomaly"`
	Score       float64         `json:"score"`
	Threshold   float64         `json:"threshold"`
	Method      DetectionMethod `json:"method"`
	Value       float64         `json:"value"`
	Expected    float64         `json:"expected"`
	Severity    AnomalySeverity `json:"severity"`
	Description string          `json:"description"`
	Stats       BaselineStats   `json:"stats"`
}
