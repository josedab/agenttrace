package domain

import (
	"time"

	"github.com/google/uuid"
)

// Trace represents a single trace (request/response cycle)
type Trace struct {
	ID            string     `json:"id" ch:"id"`
	ProjectID     uuid.UUID  `json:"projectId" ch:"project_id"`
	Name          string     `json:"name" ch:"name"`
	UserID        string     `json:"userId,omitempty" ch:"user_id"`
	SessionID     string     `json:"sessionId,omitempty" ch:"session_id"`
	Release       string     `json:"release,omitempty" ch:"release"`
	Version       string     `json:"version,omitempty" ch:"version"`
	Tags          []string   `json:"tags" ch:"tags"`
	Metadata      string     `json:"metadata,omitempty" ch:"metadata"`
	Public        bool       `json:"public" ch:"public"`
	Bookmarked    bool       `json:"bookmarked" ch:"bookmarked"`
	StartTime     time.Time  `json:"startTime" ch:"start_time"`
	EndTime       *time.Time `json:"endTime,omitempty" ch:"end_time"`
	DurationMs    float64    `json:"durationMs" ch:"duration_ms"`
	Input         string     `json:"input,omitempty" ch:"input"`
	Output        string     `json:"output,omitempty" ch:"output"`
	Level         Level      `json:"level" ch:"level"`
	StatusMessage string     `json:"statusMessage,omitempty" ch:"status_message"`

	// Aggregated costs
	TotalCost  float64 `json:"totalCost" ch:"total_cost"`
	InputCost  float64 `json:"inputCost" ch:"input_cost"`
	OutputCost float64 `json:"outputCost" ch:"output_cost"`

	// Aggregated tokens
	TotalTokens  uint64 `json:"totalTokens" ch:"total_tokens"`
	InputTokens  uint64 `json:"inputTokens" ch:"input_tokens"`
	OutputTokens uint64 `json:"outputTokens" ch:"output_tokens"`

	// Git integration
	GitCommitSha string `json:"gitCommitSha,omitempty" ch:"git_commit_sha"`
	GitBranch    string `json:"gitBranch,omitempty" ch:"git_branch"`
	GitRepoURL   string `json:"gitRepoUrl,omitempty" ch:"git_repo_url"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" ch:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" ch:"updated_at"`

	// Related data (populated by resolvers)
	Observations []Observation `json:"observations,omitempty" ch:"-"`
	Scores       []Score       `json:"scores,omitempty" ch:"-"`
}

// TraceInput represents input for creating/updating a trace
type TraceInput struct {
	ID            string     `json:"id,omitempty"`
	Name          string     `json:"name,omitempty"`
	UserID        string     `json:"userId,omitempty"`
	SessionID     string     `json:"sessionId,omitempty"`
	Release       string     `json:"release,omitempty"`
	Version       string     `json:"version,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	Metadata      any        `json:"metadata,omitempty"`
	Public        bool       `json:"public,omitempty"`
	Input         any        `json:"input,omitempty"`
	Output        any        `json:"output,omitempty"`
	Level         Level      `json:"level,omitempty"`
	StatusMessage string     `json:"statusMessage,omitempty"`
	Timestamp     *time.Time `json:"timestamp,omitempty"`
	StartTime     *time.Time `json:"startTime,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`

	// Git integration
	GitCommitSha string `json:"gitCommitSha,omitempty"`
	GitBranch    string `json:"gitBranch,omitempty"`
	GitRepoURL   string `json:"gitRepoUrl,omitempty"`
}

// IngestionBatch represents a batch of ingestion items
type IngestionBatch struct {
	Traces       []*TraceInput       `json:"traces"`
	Observations []*ObservationInput `json:"observations"`
	Generations  []*GenerationInput  `json:"generations"`
}

// TraceUpdateInput represents input for updating a trace
type TraceUpdateInput struct {
	Name          *string    `json:"name,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	SessionID     *string    `json:"sessionId,omitempty"`
	Release       *string    `json:"release,omitempty"`
	Version       *string    `json:"version,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	Metadata      any        `json:"metadata,omitempty"`
	Public        *bool      `json:"public,omitempty"`
	Input         any        `json:"input,omitempty"`
	Output        any        `json:"output,omitempty"`
	Level         *Level     `json:"level,omitempty"`
	StatusMessage *string    `json:"statusMessage,omitempty"`
	EndTime       *time.Time `json:"endTime,omitempty"`
	Bookmarked    *bool      `json:"bookmarked,omitempty"`

	// Git integration
	GitCommitSha *string `json:"gitCommitSha,omitempty"`
	GitBranch    *string `json:"gitBranch,omitempty"`
	GitRepoURL   *string `json:"gitRepoUrl,omitempty"`
}

// TraceFilter represents filter options for querying traces
type TraceFilter struct {
	ProjectID   uuid.UUID
	IDs         []string
	UserID      *string
	SessionID   *string
	Name        *string
	Release     *string
	Version     *string
	Tags        []string
	Level       *Level
	FromTime    *time.Time
	ToTime      *time.Time
	Bookmarked  *bool
	HasError    *bool
	MinCost     *float64
	MaxCost     *float64
	MinDuration *float64
	MaxDuration *float64
	Search      *string

	// Git correlation filters
	GitCommitSha *string
	GitBranch    *string
	GitRepoURL   *string
}

// TraceOrderBy represents ordering options for traces
type TraceOrderBy struct {
	Field     string
	Direction string // "asc" or "desc"
}

// ValidOrderByFields for traces
var ValidTraceOrderByFields = map[string]bool{
	"start_time":  true,
	"end_time":    true,
	"duration_ms": true,
	"total_cost":  true,
	"name":        true,
	"level":       true,
	"created_at":  true,
}

// TraceList represents a paginated list of traces
type TraceList struct {
	Traces     []Trace `json:"traces"`
	TotalCount int64   `json:"totalCount"`
	HasMore    bool    `json:"hasMore"`
}

// Session represents a session (group of related traces)
type Session struct {
	ID             string    `json:"id" ch:"id"`
	ProjectID      uuid.UUID `json:"projectId" ch:"project_id"`
	UserID         string    `json:"userId,omitempty" ch:"user_id"`
	Bookmarked     bool      `json:"bookmarked" ch:"bookmarked"`
	Public         bool      `json:"public" ch:"public"`
	CreatedAt      time.Time `json:"createdAt" ch:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" ch:"updated_at"`
	TraceCount     uint64    `json:"traceCount" ch:"trace_count"`
	TotalCost      float64   `json:"totalCost" ch:"total_cost"`
	TotalTokens    uint64    `json:"totalTokens" ch:"total_tokens"`
	FirstTraceTime time.Time `json:"firstTraceTime" ch:"first_trace_time"`
	LastTraceTime  time.Time `json:"lastTraceTime" ch:"last_trace_time"`

	// Related data
	Traces []Trace `json:"traces,omitempty" ch:"-"`
}

// SessionFilter represents filter options for querying sessions
type SessionFilter struct {
	ProjectID  uuid.UUID
	UserID     *string
	FromTime   *time.Time
	ToTime     *time.Time
	Bookmarked *bool
}

// SessionList represents a paginated list of sessions
type SessionList struct {
	Sessions   []Session `json:"sessions"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}
