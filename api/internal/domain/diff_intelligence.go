package domain

import (
	"time"

	"github.com/google/uuid"
)

// DiffAnalysisStatus represents the status of a diff analysis
type DiffAnalysisStatus string

const (
	DiffAnalysisPending   DiffAnalysisStatus = "pending"
	DiffAnalysisRunning   DiffAnalysisStatus = "running"
	DiffAnalysisCompleted DiffAnalysisStatus = "completed"
	DiffAnalysisFailed    DiffAnalysisStatus = "failed"
)

// IsValid checks if the diff analysis status is valid
func (s DiffAnalysisStatus) IsValid() bool {
	switch s {
	case DiffAnalysisPending, DiffAnalysisRunning, DiffAnalysisCompleted, DiffAnalysisFailed:
		return true
	}
	return false
}

// QualityDimension represents a dimension of code quality
type QualityDimension string

const (
	QualityComplexity      QualityDimension = "complexity"
	QualityReadability     QualityDimension = "readability"
	QualityMaintainability QualityDimension = "maintainability"
	QualityTestCoverage    QualityDimension = "test_coverage"
	QualitySecurity        QualityDimension = "security"
	QualityPerformance     QualityDimension = "performance"
)

// FindingSeverity represents the severity level for findings
type FindingSeverity string

const (
	FindingSeverityInfo     FindingSeverity = "info"
	FindingSeverityWarning  FindingSeverity = "warning"
	FindingSeverityError    FindingSeverity = "error"
	FindingSeverityCritical FindingSeverity = "critical"
)

// DiffAnalysis represents a complete analysis of code changes from an agent
type DiffAnalysis struct {
	ID        uuid.UUID          `json:"id"`
	ProjectID uuid.UUID          `json:"projectId"`
	TraceID   uuid.UUID          `json:"traceId"`
	Status    DiffAnalysisStatus `json:"status"`

	// File changes summary
	FilesAdded    int `json:"filesAdded"`
	FilesModified int `json:"filesModified"`
	FilesDeleted  int `json:"filesDeleted"`
	LinesAdded    int `json:"linesAdded"`
	LinesRemoved  int `json:"linesRemoved"`

	// Quality scores (0-100)
	OverallScore    float64                      `json:"overallScore"`
	DimensionScores map[QualityDimension]float64 `json:"dimensionScores"`

	// Findings
	Findings     []DiffFinding  `json:"findings"`
	FileAnalyses []FileAnalysis `json:"fileAnalyses"`

	// Metadata
	AgentName    string     `json:"agentName,omitempty"`
	GitCommitSha string     `json:"gitCommitSha,omitempty"`
	GitBranch    string     `json:"gitBranch,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}

// DiffFinding represents a single finding from the analysis
type DiffFinding struct {
	ID          string          `json:"id"`
	Severity    FindingSeverity `json:"severity"`
	Category    string          `json:"category"` // security, quality, performance, style
	Title       string          `json:"title"`
	Description string          `json:"description"`
	FilePath    string          `json:"filePath"`
	StartLine   int             `json:"startLine,omitempty"`
	EndLine     int             `json:"endLine,omitempty"`
	Suggestion  string          `json:"suggestion,omitempty"`
	RuleID      string          `json:"ruleId,omitempty"`
	Confidence  float64         `json:"confidence"` // 0-1
}

// FileAnalysis represents analysis for a single file
type FileAnalysis struct {
	FilePath        string        `json:"filePath"`
	Language        string        `json:"language"`
	LinesAdded      int           `json:"linesAdded"`
	LinesRemoved    int           `json:"linesRemoved"`
	ComplexityDelta int           `json:"complexityDelta"`
	QualityScore    float64       `json:"qualityScore"`
	Findings        []DiffFinding `json:"findings"`
	Diff            string        `json:"diff,omitempty"`
}

// DiffAnalysisInput represents input for creating a new analysis
type DiffAnalysisInput struct {
	TraceID      uuid.UUID         `json:"traceId" validate:"required"`
	FileChanges  []FileChangeInput `json:"fileChanges"`
	GitCommitSha string            `json:"gitCommitSha,omitempty"`
	GitBranch    string            `json:"gitBranch,omitempty"`
}

// FileChangeInput represents a file change to analyze
type FileChangeInput struct {
	FilePath      string `json:"filePath"`
	Operation     string `json:"operation"` // add, modify, delete
	ContentBefore string `json:"contentBefore,omitempty"`
	ContentAfter  string `json:"contentAfter,omitempty"`
	Diff          string `json:"diff,omitempty"`
	Language      string `json:"language,omitempty"`
}

// DiffAnalysisFilter for querying analyses
type DiffAnalysisFilter struct {
	ProjectID   uuid.UUID
	TraceID     *uuid.UUID
	Status      *DiffAnalysisStatus
	MinScore    *float64
	MaxScore    *float64
	HasFindings *bool
	GitBranch   string
}

// DiffAnalysisSummary for listing
type DiffAnalysisSummary struct {
	ID           uuid.UUID          `json:"id"`
	TraceID      uuid.UUID          `json:"traceId"`
	Status       DiffAnalysisStatus `json:"status"`
	OverallScore float64            `json:"overallScore"`
	FindingCount int                `json:"findingCount"`
	FilesChanged int                `json:"filesChanged"`
	CreatedAt    time.Time          `json:"createdAt"`
}

// QualityTrend represents quality score over time
type QualityTrend struct {
	Points  []QualityTrendPoint `json:"points"`
	Average float64             `json:"average"`
	Trend   string              `json:"trend"` // improving, declining, stable
}

// QualityTrendPoint represents a single point in the quality trend
type QualityTrendPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	OverallScore float64   `json:"overallScore"`
	TraceID      uuid.UUID `json:"traceId"`
}
