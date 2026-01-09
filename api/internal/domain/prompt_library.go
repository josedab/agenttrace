package domain

import (
	"time"

	"github.com/google/uuid"
)

// PromptVisibility represents the visibility of a prompt in the library
type PromptVisibility string

const (
	PromptVisibilityPrivate      PromptVisibility = "private"
	PromptVisibilityOrganization PromptVisibility = "organization"
	PromptVisibilityPublic       PromptVisibility = "public"
)

// PromptCategory represents a category for organizing prompts
type PromptCategory string

const (
	PromptCategoryAgent        PromptCategory = "agent"
	PromptCategoryChat         PromptCategory = "chat"
	PromptCategoryCompletion   PromptCategory = "completion"
	PromptCategorySummarization PromptCategory = "summarization"
	PromptCategoryExtraction   PromptCategory = "extraction"
	PromptCategoryClassification PromptCategory = "classification"
	PromptCategoryCodeGen      PromptCategory = "code_generation"
	PromptCategoryTranslation  PromptCategory = "translation"
	PromptCategoryCustom       PromptCategory = "custom"
)

// LibraryPrompt represents a prompt template in the community library
type LibraryPrompt struct {
	ID          uuid.UUID        `json:"id"`
	AuthorID    uuid.UUID        `json:"authorId"`
	AuthorName  string           `json:"authorName"`
	ProjectID   *uuid.UUID       `json:"projectId,omitempty"` // nil for org-level prompts

	// Basic info
	Name        string           `json:"name"`
	Slug        string           `json:"slug"` // URL-friendly identifier
	Description string           `json:"description"`
	Visibility  PromptVisibility `json:"visibility"`
	Category    PromptCategory   `json:"category"`
	Tags        []string         `json:"tags"`

	// Content
	Template    string           `json:"template"`
	Variables   []PromptVariable `json:"variables"`
	Examples    []PromptExample  `json:"examples,omitempty"`

	// Model configuration hints
	RecommendedModels []string          `json:"recommendedModels,omitempty"`
	ModelParams       map[string]any    `json:"modelParams,omitempty"`

	// Versioning
	Version       int              `json:"version"`
	LatestVersion int              `json:"latestVersion"`
	VersionNotes  string           `json:"versionNotes,omitempty"`

	// Forking
	ForkOf        *uuid.UUID       `json:"forkOf,omitempty"`
	ForkCount     int              `json:"forkCount"`

	// Usage stats
	UsageCount    int              `json:"usageCount"`
	StarCount     int              `json:"starCount"`
	ViewCount     int              `json:"viewCount"`

	// Benchmark results
	Benchmarks    []PromptBenchmark `json:"benchmarks,omitempty"`

	// Audit
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

// PromptVariable represents a variable in a prompt template
type PromptVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, number, boolean, array, object
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
	Example     any    `json:"example,omitempty"`
}

// PromptExample represents an example usage of a prompt
type PromptExample struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Variables   map[string]any `json:"variables"`
	Expected    string         `json:"expected,omitempty"`
}

// PromptBenchmark represents benchmark results for a prompt
type PromptBenchmark struct {
	ID            uuid.UUID          `json:"id"`
	PromptID      uuid.UUID          `json:"promptId"`
	PromptVersion int                `json:"promptVersion"`

	// Configuration
	Model         string             `json:"model"`
	DatasetID     *uuid.UUID         `json:"datasetId,omitempty"`
	DatasetName   string             `json:"datasetName,omitempty"`
	SampleCount   int                `json:"sampleCount"`

	// Results
	Metrics       BenchmarkMetrics   `json:"metrics"`

	// Audit
	RunBy         uuid.UUID          `json:"runBy"`
	RunAt         time.Time          `json:"runAt"`
	Duration      int                `json:"durationSeconds"`
}

// BenchmarkMetrics contains performance metrics from a benchmark
type BenchmarkMetrics struct {
	// Quality scores (0-1)
	Accuracy      *float64 `json:"accuracy,omitempty"`
	Relevance     *float64 `json:"relevance,omitempty"`
	Coherence     *float64 `json:"coherence,omitempty"`
	Helpfulness   *float64 `json:"helpfulness,omitempty"`

	// Custom scores
	CustomScores  map[string]float64 `json:"customScores,omitempty"`

	// Performance
	AvgLatency    float64  `json:"avgLatencyMs"`
	P95Latency    float64  `json:"p95LatencyMs"`
	AvgTokens     float64  `json:"avgTokens"`
	TotalCost     float64  `json:"totalCost"`
	AvgCostPerCall float64 `json:"avgCostPerCall"`

	// Error rate
	SuccessRate   float64  `json:"successRate"`
	ErrorCount    int      `json:"errorCount"`
}

// PromptStar represents a user starring a prompt
type PromptStar struct {
	UserID    uuid.UUID `json:"userId"`
	PromptID  uuid.UUID `json:"promptId"`
	StarredAt time.Time `json:"starredAt"`
}

// PromptFork represents a fork of a library prompt
type PromptFork struct {
	ID              uuid.UUID `json:"id"`
	SourcePromptID  uuid.UUID `json:"sourcePromptId"`
	SourceVersion   int       `json:"sourceVersion"`
	ForkedPromptID  uuid.UUID `json:"forkedPromptId"`
	ForkedBy        uuid.UUID `json:"forkedBy"`
	ForkedAt        time.Time `json:"forkedAt"`
}

// PromptVersion represents a historical version of a prompt
type PromptVersion struct {
	PromptID    uuid.UUID        `json:"promptId"`
	Version     int              `json:"version"`
	Template    string           `json:"template"`
	Variables   []PromptVariable `json:"variables"`
	VersionNotes string          `json:"versionNotes,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	CreatedBy   uuid.UUID        `json:"createdBy"`
}

// LibraryPromptInput represents input for creating a library prompt
type LibraryPromptInput struct {
	Name              string             `json:"name" validate:"required,min=1,max=100"`
	Description       string             `json:"description" validate:"max=1000"`
	Visibility        PromptVisibility   `json:"visibility"`
	Category          PromptCategory     `json:"category" validate:"required"`
	Tags              []string           `json:"tags,omitempty"`
	Template          string             `json:"template" validate:"required"`
	Variables         []PromptVariable   `json:"variables,omitempty"`
	Examples          []PromptExample    `json:"examples,omitempty"`
	RecommendedModels []string           `json:"recommendedModels,omitempty"`
	ModelParams       map[string]any     `json:"modelParams,omitempty"`
	VersionNotes      string             `json:"versionNotes,omitempty"`
}

// LibraryPromptUpdateInput represents input for updating a library prompt
type LibraryPromptUpdateInput struct {
	Name              *string            `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description       *string            `json:"description,omitempty" validate:"omitempty,max=1000"`
	Visibility        *PromptVisibility  `json:"visibility,omitempty"`
	Category          *PromptCategory    `json:"category,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	Template          *string            `json:"template,omitempty"`
	Variables         []PromptVariable   `json:"variables,omitempty"`
	Examples          []PromptExample    `json:"examples,omitempty"`
	RecommendedModels []string           `json:"recommendedModels,omitempty"`
	ModelParams       map[string]any     `json:"modelParams,omitempty"`
	VersionNotes      string             `json:"versionNotes,omitempty"`
	BumpVersion       bool               `json:"bumpVersion,omitempty"`
}

// LibraryPromptFilter represents filter options for querying library prompts
type LibraryPromptFilter struct {
	AuthorID     *uuid.UUID
	ProjectID    *uuid.UUID
	Category     *PromptCategory
	Visibility   *PromptVisibility
	Tags         []string
	Search       string
	ForkOf       *uuid.UUID
	OnlyStarred  bool
	StarredBy    *uuid.UUID
	SortBy       string // "popular", "recent", "stars", "usage"
	SortOrder    string // "asc", "desc"
}

// LibraryPromptList represents a paginated list of library prompts
type LibraryPromptList struct {
	Prompts    []LibraryPrompt `json:"prompts"`
	TotalCount int64           `json:"totalCount"`
	HasMore    bool            `json:"hasMore"`
}

// PromptVersionList represents a list of prompt versions
type PromptVersionList struct {
	Versions   []PromptVersion `json:"versions"`
	TotalCount int64           `json:"totalCount"`
	HasMore    bool            `json:"hasMore"`
}

// ForkInput represents input for forking a prompt
type ForkInput struct {
	Name       string           `json:"name,omitempty"`       // Optional new name
	Visibility PromptVisibility `json:"visibility,omitempty"` // Optional visibility
}

// BenchmarkInput represents input for running a benchmark
type BenchmarkInput struct {
	Model       string         `json:"model" validate:"required"`
	DatasetID   *uuid.UUID     `json:"datasetId,omitempty"`
	SampleCount int            `json:"sampleCount,omitempty"` // If no dataset, generate samples
	Variables   map[string]any `json:"variables,omitempty"`   // Default variable values
	Evaluators  []string       `json:"evaluators,omitempty"`  // Evaluator IDs to use
}

// PromptUsageRecord tracks usage of a library prompt
type PromptUsageRecord struct {
	ID        uuid.UUID `json:"id"`
	PromptID  uuid.UUID `json:"promptId"`
	Version   int       `json:"version"`
	UserID    uuid.UUID `json:"userId"`
	ProjectID uuid.UUID `json:"projectId"`
	TraceID   *uuid.UUID `json:"traceId,omitempty"`
	UsedAt    time.Time `json:"usedAt"`
}
