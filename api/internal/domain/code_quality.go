package domain

import (
	"time"

	"github.com/google/uuid"
)

// CodeQualityAnalyzer represents the type of static analysis tool
type CodeQualityAnalyzer string

const (
	CodeQualityAnalyzerESLint   CodeQualityAnalyzer = "eslint"
	CodeQualityAnalyzerSemgrep  CodeQualityAnalyzer = "semgrep"
	CodeQualityAnalyzerSonarQube CodeQualityAnalyzer = "sonarqube"
	CodeQualityAnalyzerCustom   CodeQualityAnalyzer = "custom"
)

// IsValid checks if the code quality analyzer is valid
func (a CodeQualityAnalyzer) IsValid() bool {
	switch a {
	case CodeQualityAnalyzerESLint, CodeQualityAnalyzerSemgrep, CodeQualityAnalyzerSonarQube, CodeQualityAnalyzerCustom:
		return true
	}
	return false
}

// CodeQualitySeverity represents the severity level of a finding
type CodeQualitySeverity string

const (
	CodeQualitySeverityBlocker  CodeQualitySeverity = "blocker"
	CodeQualitySeverityCritical CodeQualitySeverity = "critical"
	CodeQualitySeverityMajor    CodeQualitySeverity = "major"
	CodeQualitySeverityMinor    CodeQualitySeverity = "minor"
	CodeQualitySeverityInfo     CodeQualitySeverity = "info"
)

// IsValid checks if the code quality severity is valid
func (s CodeQualitySeverity) IsValid() bool {
	switch s {
	case CodeQualitySeverityBlocker, CodeQualitySeverityCritical, CodeQualitySeverityMajor, CodeQualitySeverityMinor, CodeQualitySeverityInfo:
		return true
	}
	return false
}

// CodeQualityConfig defines the configuration for code quality analysis
type CodeQualityConfig struct {
	ID                uuid.UUID        `json:"id"`
	ProjectID         uuid.UUID        `json:"projectId"`
	Name              string           `json:"name"`
	Analyzers         []AnalyzerConfig `json:"analyzers"`
	AutoRunOnTrace    bool             `json:"autoRunOnTrace"`
	MinScoreThreshold float64          `json:"minScoreThreshold"`
	FailOnBlocker     bool             `json:"failOnBlocker"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

// AnalyzerConfig defines the configuration for a specific analyzer
type AnalyzerConfig struct {
	Type    CodeQualityAnalyzer `json:"type"`
	Enabled bool                `json:"enabled"`
	Rules   map[string]any      `json:"rules,omitempty"`
	Weight  float64             `json:"weight"`
}

// CodeQualityReport represents the results of a code quality analysis
type CodeQualityReport struct {
	ID                uuid.UUID              `json:"id"`
	ProjectID         uuid.UUID              `json:"projectId"`
	TraceID           uuid.UUID              `json:"traceId"`
	ConfigID          *uuid.UUID             `json:"configId,omitempty"`
	OverallScore      float64                `json:"overallScore"`
	Grade             string                 `json:"grade"`
	Findings          []CodeQualityFinding   `json:"findings"`
	AnalyzerResults   []AnalyzerResult       `json:"analyzerResults"`
	LinesAnalyzed     int                    `json:"linesAnalyzed"`
	FilesAnalyzed     int                    `json:"filesAnalyzed"`
	TotalFindings     int                    `json:"totalFindings"`
	FindingsBySeverity map[string]int        `json:"findingsBySeverity"`
	Passed            bool                   `json:"passed"`
	DurationMs        int64                  `json:"durationMs"`
	Duration          time.Duration          `json:"-"`
	CreatedAt         time.Time              `json:"createdAt"`
}

// CodeQualityFinding represents a single finding from an analyzer
type CodeQualityFinding struct {
	ID         string              `json:"id"`
	Analyzer   CodeQualityAnalyzer `json:"analyzer"`
	Severity   CodeQualitySeverity `json:"severity"`
	Rule       string              `json:"rule"`
	Message    string              `json:"message"`
	FilePath   string              `json:"filePath"`
	Line       int                 `json:"line"`
	Column     int                 `json:"column"`
	EndLine    *int                `json:"endLine,omitempty"`
	EndColumn  *int                `json:"endColumn,omitempty"`
	Snippet    string              `json:"snippet,omitempty"`
	Suggestion *string             `json:"suggestion,omitempty"`
	Effort     string              `json:"effort,omitempty"`
}

// AnalyzerResult represents the result from a single analyzer
type AnalyzerResult struct {
	Analyzer      CodeQualityAnalyzer `json:"analyzer"`
	Score         float64             `json:"score"`
	FindingsCount int                 `json:"findingsCount"`
	Duration      int64               `json:"duration"`
	Passed        bool                `json:"passed"`
	Details       map[string]any      `json:"details,omitempty"`
}

// CodeQualityInput represents the input for running a code quality analysis
type CodeQualityInput struct {
	TraceID   uuid.UUID             `json:"traceId" validate:"required"`
	ConfigID  *uuid.UUID            `json:"configId,omitempty"`
	Analyzers []CodeQualityAnalyzer `json:"analyzers,omitempty"`
}

// CodeQualityConfigInput represents the input for creating or updating a config
type CodeQualityConfigInput struct {
	Name              string           `json:"name" validate:"required,min=1,max=200"`
	Analyzers         []AnalyzerConfig `json:"analyzers"`
	AutoRunOnTrace    bool             `json:"autoRunOnTrace"`
	MinScoreThreshold float64          `json:"minScoreThreshold"`
	FailOnBlocker     bool             `json:"failOnBlocker"`
}

// CodeQualityDashboard represents aggregated code quality metrics
type CodeQualityDashboard struct {
	TotalScans        int                    `json:"totalScans"`
	TotalFindings     int                    `json:"totalFindings"`
	AvgScore          float64                `json:"avgScore"`
	GradeDistribution map[string]int         `json:"gradeDistribution"`
	TopIssues         []CodeQualityFinding   `json:"topIssues"`
	ScoreTrend        []CodeQualityTrendPoint `json:"scoreTrend"`
	AnalyzerBreakdown map[string]int         `json:"analyzerBreakdown"`
}

// CodeQualityTrendPoint represents a single point in the score trend
type CodeQualityTrendPoint struct {
	Date          time.Time `json:"date"`
	AvgScore      float64   `json:"avgScore"`
	TotalFindings int       `json:"totalFindings"`
}

// CodeQualityReportList represents a paginated list of code quality reports
type CodeQualityReportList struct {
	Reports    []CodeQualityReport `json:"reports"`
	TotalCount int64               `json:"totalCount"`
	HasMore    bool                `json:"hasMore"`
}
