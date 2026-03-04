package domain

import (
	"time"

	"github.com/google/uuid"
)

// TraceDiffType categorizes structural differences between trace trees
type TraceDiffType string

const (
	DiffTypeAdded     TraceDiffType = "added"
	DiffTypeRemoved   TraceDiffType = "removed"
	DiffTypeModified  TraceDiffType = "modified"
	DiffTypeUnchanged TraceDiffType = "unchanged"
	DiffTypeReordered TraceDiffType = "reordered"
)

// TraceDiffNode represents a single node in the diff tree
type TraceDiffNode struct {
	DiffType     TraceDiffType  `json:"diffType"`
	SpanName     string         `json:"spanName"`
	LeftSpanID   string         `json:"leftSpanId,omitempty"`
	RightSpanID  string         `json:"rightSpanId,omitempty"`
	LeftValue    *SpanSnapshot  `json:"leftValue,omitempty"`
	RightValue   *SpanSnapshot  `json:"rightValue,omitempty"`
	PropertyDiffs []PropertyDiff `json:"propertyDiffs,omitempty"`
	Children     []*TraceDiffNode `json:"children,omitempty"`
}

// SpanSnapshot captures relevant properties of a span for diff comparison
type SpanSnapshot struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Model       string         `json:"model,omitempty"`
	DurationMs  float64        `json:"durationMs"`
	TotalTokens uint64         `json:"totalTokens"`
	TotalCost   float64        `json:"totalCost"`
	Input       string         `json:"input,omitempty"`
	Output      string         `json:"output,omitempty"`
	Level       string         `json:"level"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// PropertyDiff describes a changed property between two matching spans
type PropertyDiff struct {
	Property  string `json:"property"`
	LeftValue any    `json:"leftValue"`
	RightValue any   `json:"rightValue"`
	ChangeType string `json:"changeType"` // increased, decreased, changed
}

// TraceDiffResult contains the full diff output for two traces
type TraceDiffResult struct {
	ID           uuid.UUID        `json:"id"`
	ProjectID    uuid.UUID        `json:"projectId"`
	LeftTraceID  string           `json:"leftTraceId"`
	RightTraceID string           `json:"rightTraceId"`
	RootDiffs    []*TraceDiffNode `json:"rootDiffs"`
	Summary      TraceDiffSummary `json:"summary"`
	CreatedAt    time.Time        `json:"createdAt"`
}

// TraceDiffSummary provides high-level diff statistics
type TraceDiffSummary struct {
	TotalNodes     int     `json:"totalNodes"`
	AddedCount     int     `json:"addedCount"`
	RemovedCount   int     `json:"removedCount"`
	ModifiedCount  int     `json:"modifiedCount"`
	UnchangedCount int     `json:"unchangedCount"`
	CostDelta      float64 `json:"costDelta"`
	LatencyDelta   float64 `json:"latencyDeltaMs"`
	TokenDelta     int64   `json:"tokenDelta"`
}

// TraceDiffInput represents input for creating a trace diff
type TraceDiffInput struct {
	LeftTraceID  string `json:"leftTraceId" validate:"required"`
	RightTraceID string `json:"rightTraceId" validate:"required"`
}

// BisectStatus tracks bisect session state
type BisectStatus string

const (
	BisectStatusActive    BisectStatus = "active"
	BisectStatusCompleted BisectStatus = "completed"
	BisectStatusFailed    BisectStatus = "failed"
	BisectStatusCancelled BisectStatus = "cancelled"
)

// BisectSession represents an active regression bisect session
type BisectSession struct {
	ID             uuid.UUID       `json:"id"`
	ProjectID      uuid.UUID       `json:"projectId"`
	Status         BisectStatus    `json:"status"`
	GoodTraceID    string          `json:"goodTraceId"`
	BadTraceID     string          `json:"badTraceId"`
	TraceHistory   []string        `json:"traceHistory"`
	CurrentIndex   int             `json:"currentIndex"`
	LowIndex       int             `json:"lowIndex"`
	HighIndex      int             `json:"highIndex"`
	Steps          []BisectStep    `json:"steps"`
	RegressionSpan string          `json:"regressionSpan,omitempty"`
	MetricName     string          `json:"metricName"`
	Threshold      float64         `json:"threshold"`
	CreatedAt      time.Time       `json:"createdAt"`
	CompletedAt    *time.Time      `json:"completedAt,omitempty"`
	CreatedBy      uuid.UUID       `json:"createdBy"`
}

// BisectStep records one step in the bisect process
type BisectStep struct {
	StepNumber int       `json:"stepNumber"`
	TraceID    string    `json:"traceId"`
	TraceIndex int       `json:"traceIndex"`
	Verdict    string    `json:"verdict"` // good, bad, skip
	MetricValue float64  `json:"metricValue"`
	Timestamp  time.Time `json:"timestamp"`
}

// BisectStartInput represents input for starting a bisect session
type BisectStartInput struct {
	GoodTraceID  string   `json:"goodTraceId" validate:"required"`
	BadTraceID   string   `json:"badTraceId" validate:"required"`
	TraceHistory []string `json:"traceHistory" validate:"required,min=2"`
	MetricName   string   `json:"metricName" validate:"required"`
	Threshold    float64  `json:"threshold"`
}

// BisectVerdictInput represents a verdict for the current bisect step
type BisectVerdictInput struct {
	Verdict string `json:"verdict" validate:"required,oneof=good bad skip"`
}

// BisectResult is returned when a bisect session completes
type BisectResult struct {
	SessionID       uuid.UUID `json:"sessionId"`
	RegressionTrace string    `json:"regressionTrace"`
	PreviousTrace   string    `json:"previousTrace"`
	StepsTaken      int       `json:"stepsTaken"`
	RegressionSpan  string    `json:"regressionSpan,omitempty"`
	Diff            *TraceDiffResult `json:"diff,omitempty"`
}
