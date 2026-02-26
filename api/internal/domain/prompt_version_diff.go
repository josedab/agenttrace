package domain

import (
	"time"

	"github.com/google/uuid"
)

// PromptDiffType represents the type of change between prompt versions
type PromptDiffType string

const (
	PromptDiffTypeAdded       PromptDiffType = "added"
	PromptDiffTypeRemoved     PromptDiffType = "removed"
	PromptDiffTypeModified    PromptDiffType = "modified"
	PromptDiffTypeUnchanged   PromptDiffType = "unchanged"
)

// PromptChangeCategory represents the category of a prompt change
type PromptChangeCategory string

const (
	PromptChangeCategoryVariable    PromptChangeCategory = "variable_change"
	PromptChangeCategoryInstruction PromptChangeCategory = "instruction_modification"
	PromptChangeCategorySystemPrompt PromptChangeCategory = "system_prompt_alteration"
	PromptChangeCategoryFormatting  PromptChangeCategory = "formatting"
	PromptChangeCategoryContent     PromptChangeCategory = "content"
	PromptChangeCategoryConfig      PromptChangeCategory = "config"
)

// PromptReviewStatus represents the status of a version review
type PromptReviewStatus string

const (
	PromptReviewPending  PromptReviewStatus = "pending"
	PromptReviewApproved PromptReviewStatus = "approved"
	PromptReviewRejected PromptReviewStatus = "rejected"
)

// PromptVersionDiff represents the diff between two prompt versions
type PromptVersionDiff struct {
	PromptID    uuid.UUID       `json:"promptId"`
	VersionA    int             `json:"versionA"`
	VersionB    int             `json:"versionB"`
	ContentDiff SemanticDiff    `json:"contentDiff"`
	ConfigDiff  *ConfigDiff     `json:"configDiff,omitempty"`
	VariableDiff VariableDiff   `json:"variableDiff"`
	Changes     []PromptChange  `json:"changes"`
	Summary     DiffSummary     `json:"summary"`
}

// SemanticDiff represents a semantic-aware diff of prompt content
type SemanticDiff struct {
	Hunks      []DiffHunk `json:"hunks"`
	AddedLines   int      `json:"addedLines"`
	RemovedLines int      `json:"removedLines"`
	TotalLines   int      `json:"totalLines"`
}

// DiffHunk represents a contiguous block of changes
type DiffHunk struct {
	StartLineA int        `json:"startLineA"`
	StartLineB int        `json:"startLineB"`
	Lines      []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    PromptDiffType `json:"type"`
	Content string         `json:"content"`
	LineA   *int           `json:"lineA,omitempty"`
	LineB   *int           `json:"lineB,omitempty"`
}

// VariableDiff tracks changes in prompt template variables
type VariableDiff struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
	Common  []string `json:"common"`
}

// ConfigDiff tracks changes in prompt configuration
type ConfigDiff struct {
	Added   map[string]interface{} `json:"added,omitempty"`
	Removed map[string]interface{} `json:"removed,omitempty"`
	Changed map[string]ConfigFieldChange `json:"changed,omitempty"`
}

// ConfigFieldChange represents a change in a config field
type ConfigFieldChange struct {
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

// PromptChange represents a categorized change in the prompt
type PromptChange struct {
	Category    PromptChangeCategory `json:"category"`
	Description string               `json:"description"`
	Severity    string               `json:"severity"` // low, medium, high
	LineRange   *[2]int              `json:"lineRange,omitempty"`
}

// DiffSummary provides a summary of all changes
type DiffSummary struct {
	TotalChanges     int                            `json:"totalChanges"`
	ByCategory       map[PromptChangeCategory]int   `json:"byCategory"`
	RiskLevel        string                         `json:"riskLevel"` // low, medium, high
	RiskExplanation  string                         `json:"riskExplanation"`
}

// PromptImpactAnalysis represents impact analysis for a prompt version change
type PromptImpactAnalysis struct {
	PromptID      uuid.UUID                `json:"promptId"`
	VersionBefore int                      `json:"versionBefore"`
	VersionAfter  int                      `json:"versionAfter"`
	Metrics       PromptImpactMetrics      `json:"metrics"`
	TraceComparison PromptTraceComparison  `json:"traceComparison"`
	Recommendation string                  `json:"recommendation"`
	AnalyzedAt    time.Time                `json:"analyzedAt"`
}

// PromptImpactMetrics represents before/after metrics for a prompt change
type PromptImpactMetrics struct {
	Before PromptMetricsSnapshot `json:"before"`
	After  PromptMetricsSnapshot `json:"after"`
	Deltas PromptMetricsDeltas   `json:"deltas"`
}

// PromptMetricsSnapshot represents metrics at a point in time
type PromptMetricsSnapshot struct {
	TraceCount     int     `json:"traceCount"`
	AvgLatency     float64 `json:"avgLatencyMs"`
	AvgCost        float64 `json:"avgCost"`
	AvgTokens      float64 `json:"avgTokens"`
	ErrorRate      float64 `json:"errorRate"`
	AvgQualityScore float64 `json:"avgQualityScore"`
}

// PromptMetricsDeltas represents the change in metrics
type PromptMetricsDeltas struct {
	LatencyDelta      float64 `json:"latencyDelta"`
	LatencyDeltaPct   float64 `json:"latencyDeltaPct"`
	CostDelta         float64 `json:"costDelta"`
	CostDeltaPct      float64 `json:"costDeltaPct"`
	TokenDelta        float64 `json:"tokenDelta"`
	TokenDeltaPct     float64 `json:"tokenDeltaPct"`
	ErrorRateDelta    float64 `json:"errorRateDelta"`
	QualityScoreDelta float64 `json:"qualityScoreDelta"`
}

// PromptTraceComparison compares trace patterns before and after a version change
type PromptTraceComparison struct {
	BeforeTraceCount int                      `json:"beforeTraceCount"`
	AfterTraceCount  int                      `json:"afterTraceCount"`
	TopChanges       []string                 `json:"topChanges"`
}

// PromptVersionReview represents a review of a prompt version
type PromptVersionReview struct {
	ID        uuid.UUID          `json:"id"`
	PromptID  uuid.UUID          `json:"promptId"`
	VersionID uuid.UUID          `json:"versionId"`
	Version   int                `json:"version"`
	Status    PromptReviewStatus `json:"status"`
	ReviewerID uuid.UUID         `json:"reviewerId"`
	Comment   string             `json:"comment,omitempty"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// PromptVersionReviewInput represents input for creating a version review
type PromptVersionReviewInput struct {
	Status  PromptReviewStatus `json:"status" validate:"required"`
	Comment string             `json:"comment,omitempty"`
}

// PromptRollbackInput represents input for rolling back to a previous version
type PromptRollbackInput struct {
	TargetVersion int    `json:"targetVersion" validate:"required,min=1"`
	Reason        string `json:"reason,omitempty"`
}
