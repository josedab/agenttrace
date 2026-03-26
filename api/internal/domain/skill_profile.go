package domain

import (
	"time"

	"github.com/google/uuid"
)

// SkillDimension represents a measurable capability area for AI agents.
type SkillDimension string

const (
	SkillCodeGeneration SkillDimension = "code_generation"
	SkillRefactoring    SkillDimension = "refactoring"
	SkillBugFixing      SkillDimension = "bug_fixing"
	SkillTesting        SkillDimension = "testing"
	SkillDebugging      SkillDimension = "debugging"
	SkillDocumentation  SkillDimension = "documentation"
	SkillCodeReview     SkillDimension = "code_review"
)

// AgentSkillProfile aggregates an agent's performance scores across skill
// dimensions, languages, and models for a given project.
type AgentSkillProfile struct {
	AgentName      string                        `json:"agentName"`
	ProjectID      uuid.UUID                     `json:"projectId"`
	Skills         map[SkillDimension]SkillScore `json:"skills"`
	LanguageStats  map[string]LanguageStat       `json:"languageStats"`
	ModelStats     map[string]ModelStat          `json:"modelStats"`
	TotalTraces    int                           `json:"totalTraces"`
	SuccessRate    float64                       `json:"successRate"`
	AvgCostPerTask float64                       `json:"avgCostPerTask"`
	AvgLatencyMs   float64                       `json:"avgLatencyMs"`
	LastActive     time.Time                     `json:"lastActive"`
	UpdatedAt      time.Time                     `json:"updatedAt"`
}

// SkillScore holds the performance metrics for a single skill dimension.
type SkillScore struct {
	Score       float64 `json:"score"`       // 0-100
	Confidence  float64 `json:"confidence"`  // 0-1 based on sample size
	TraceCount  int     `json:"traceCount"`
	SuccessRate float64 `json:"successRate"` // 0-1
	AvgLatency  float64 `json:"avgLatencyMs"`
	AvgCost     float64 `json:"avgCost"`
}

// LanguageStat contains per-language performance statistics for an agent.
type LanguageStat struct {
	Language    string  `json:"language"`
	TraceCount  int     `json:"traceCount"`
	SuccessRate float64 `json:"successRate"`
	AvgQuality  float64 `json:"avgQuality"`
}

// ModelStat contains per-model performance statistics for an agent.
type ModelStat struct {
	Model       string  `json:"model"`
	TraceCount  int     `json:"traceCount"`
	AvgCost     float64 `json:"avgCost"`
	AvgLatency  float64 `json:"avgLatencyMs"`
	SuccessRate float64 `json:"successRate"`
}

// AgentComparison presents a side-by-side comparison of multiple agent skill
// profiles, identifying the best-performing agent for each skill dimension.
type AgentComparison struct {
	Agents    []AgentSkillProfile       `json:"agents"`
	BestAgent map[SkillDimension]string `json:"bestAgent"`
}

// SkillProfileFilter specifies criteria for querying agent skill profiles.
type SkillProfileFilter struct {
	ProjectID uuid.UUID
	AgentName string
	MinTraces int
}
