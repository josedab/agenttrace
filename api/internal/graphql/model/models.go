package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ======== CONNECTION TYPES ========

// PageInfo for cursor-based pagination
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
}

// TraceConnection represents a paginated list of traces
type TraceConnection struct {
	Edges      []*TraceEdge `json:"edges"`
	PageInfo   *PageInfo    `json:"pageInfo"`
	TotalCount int          `json:"totalCount"`
}

// TraceEdge represents a trace in the connection
type TraceEdge struct {
	Node   *domain.Trace `json:"node"`
	Cursor string        `json:"cursor"`
}

// ObservationConnection represents a paginated list of observations
type ObservationConnection struct {
	Edges      []*ObservationEdge `json:"edges"`
	PageInfo   *PageInfo          `json:"pageInfo"`
	TotalCount int                `json:"totalCount"`
}

// ObservationEdge represents an observation in the connection
type ObservationEdge struct {
	Node   *domain.Observation `json:"node"`
	Cursor string              `json:"cursor"`
}

// ScoreConnection represents a paginated list of scores
type ScoreConnection struct {
	Edges      []*ScoreEdge `json:"edges"`
	PageInfo   *PageInfo    `json:"pageInfo"`
	TotalCount int          `json:"totalCount"`
}

// ScoreEdge represents a score in the connection
type ScoreEdge struct {
	Node   *domain.Score `json:"node"`
	Cursor string        `json:"cursor"`
}

// SessionConnection represents a paginated list of sessions
type SessionConnection struct {
	Edges      []*SessionEdge `json:"edges"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

// SessionEdge represents a session in the connection
type SessionEdge struct {
	Node   *Session `json:"node"`
	Cursor string   `json:"cursor"`
}

// Session represents a session
type Session struct {
	ID            string          `json:"id"`
	ProjectID     uuid.UUID       `json:"projectId"`
	CreatedAt     time.Time       `json:"createdAt"`
	TraceCount    int             `json:"traceCount"`
	TotalDuration *float64        `json:"totalDuration"`
	TotalCost     *float64        `json:"totalCost"`
	Traces        []*domain.Trace `json:"traces"`
}

// PromptConnection represents a paginated list of prompts
type PromptConnection struct {
	Edges      []*PromptEdge `json:"edges"`
	PageInfo   *PageInfo     `json:"pageInfo"`
	TotalCount int           `json:"totalCount"`
}

// PromptEdge represents a prompt in the connection
type PromptEdge struct {
	Node   *domain.Prompt `json:"node"`
	Cursor string         `json:"cursor"`
}

// DatasetConnection represents a paginated list of datasets
type DatasetConnection struct {
	Edges      []*DatasetEdge `json:"edges"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

// DatasetEdge represents a dataset in the connection
type DatasetEdge struct {
	Node   *domain.Dataset `json:"node"`
	Cursor string          `json:"cursor"`
}

// DatasetItemConnection represents a paginated list of dataset items
type DatasetItemConnection struct {
	Edges      []*DatasetItemEdge `json:"edges"`
	PageInfo   *PageInfo          `json:"pageInfo"`
	TotalCount int                `json:"totalCount"`
}

// DatasetItemEdge represents a dataset item in the connection
type DatasetItemEdge struct {
	Node   *domain.DatasetItem `json:"node"`
	Cursor string              `json:"cursor"`
}

// EvaluatorConnection represents a paginated list of evaluators
type EvaluatorConnection struct {
	Edges      []*EvaluatorEdge `json:"edges"`
	PageInfo   *PageInfo        `json:"pageInfo"`
	TotalCount int              `json:"totalCount"`
}

// EvaluatorEdge represents an evaluator in the connection
type EvaluatorEdge struct {
	Node   *domain.Evaluator `json:"node"`
	Cursor string            `json:"cursor"`
}

// ======== INPUT TYPES ========

// TracesInput for listing traces
type TracesInput struct {
	Limit         *int                 `json:"limit"`
	Cursor        *string              `json:"cursor"`
	UserID        *string              `json:"userId"`
	SessionID     *string              `json:"sessionId"`
	Name          *string              `json:"name"`
	Tags          []string             `json:"tags"`
	FromTimestamp *time.Time           `json:"fromTimestamp"`
	ToTimestamp   *time.Time           `json:"toTimestamp"`
	Version       *string              `json:"version"`
	Release       *string              `json:"release"`
	OrderBy       *string              `json:"orderBy"`
	Order         *domain.SortOrder    `json:"order"`
}

// ObservationsInput for listing observations
type ObservationsInput struct {
	TraceID             *string                   `json:"traceId"`
	ParentObservationID *string                   `json:"parentObservationId"`
	Type                *domain.ObservationType   `json:"type"`
	Name                *string                   `json:"name"`
	Limit               *int                      `json:"limit"`
	Cursor              *string                   `json:"cursor"`
}

// ScoresInput for listing scores
type ScoresInput struct {
	TraceID       *string            `json:"traceId"`
	ObservationID *string            `json:"observationId"`
	Name          *string            `json:"name"`
	Source        *domain.ScoreSource `json:"source"`
	Limit         *int               `json:"limit"`
	Cursor        *string            `json:"cursor"`
}

// SessionsInput for listing sessions
type SessionsInput struct {
	Limit         *int       `json:"limit"`
	Cursor        *string    `json:"cursor"`
	FromTimestamp *time.Time `json:"fromTimestamp"`
	ToTimestamp   *time.Time `json:"toTimestamp"`
}

// PromptsInput for listing prompts
type PromptsInput struct {
	Name   *string  `json:"name"`
	Label  *string  `json:"label"`
	Tags   []string `json:"tags"`
	Limit  *int     `json:"limit"`
	Cursor *string  `json:"cursor"`
}

// DatasetsInput for listing datasets
type DatasetsInput struct {
	Name   *string `json:"name"`
	Limit  *int    `json:"limit"`
	Cursor *string `json:"cursor"`
}

// EvaluatorsInput for listing evaluators
type EvaluatorsInput struct {
	Type    *domain.EvaluatorType `json:"type"`
	Enabled *bool                 `json:"enabled"`
	Limit   *int                  `json:"limit"`
	Cursor  *string               `json:"cursor"`
}

// ======== CREATE INPUT TYPES ========

// CreateTraceInput for creating a trace
type CreateTraceInput struct {
	ID        *string                `json:"id"`
	Name      *string                `json:"name"`
	Timestamp *time.Time             `json:"timestamp"`
	Input     map[string]interface{} `json:"input"`
	Output    map[string]interface{} `json:"output"`
	Metadata  map[string]interface{} `json:"metadata"`
	Tags      []string               `json:"tags"`
	UserId    *string                `json:"userId"`
	SessionId *string                `json:"sessionId"`
	Release   *string                `json:"release"`
	Version   *string                `json:"version"`
	Public    *bool                  `json:"public"`
}

// UpdateTraceInput for updating a trace
type UpdateTraceInput struct {
	Name          *string                `json:"name"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	Metadata      map[string]interface{} `json:"metadata"`
	Tags          []string               `json:"tags"`
	UserId        *string                `json:"userId"`
	Level         *domain.Level          `json:"level"`
	StatusMessage *string                `json:"statusMessage"`
	Public        *bool                  `json:"public"`
}

// CreateObservationInput for creating a span or event
type CreateObservationInput struct {
	ID                  *string                `json:"id"`
	TraceId             string                 `json:"traceId"`
	ParentObservationId *string                `json:"parentObservationId"`
	Name                *string                `json:"name"`
	StartTime           *time.Time             `json:"startTime"`
	EndTime             *time.Time             `json:"endTime"`
	Input               map[string]interface{} `json:"input"`
	Output              map[string]interface{} `json:"output"`
	Metadata            map[string]interface{} `json:"metadata"`
	Level               *domain.Level          `json:"level"`
	StatusMessage       *string                `json:"statusMessage"`
	Version             *string                `json:"version"`
}

// CreateGenerationInput for creating a generation
type CreateGenerationInput struct {
	ID                  *string                `json:"id"`
	TraceId             string                 `json:"traceId"`
	ParentObservationId *string                `json:"parentObservationId"`
	Name                *string                `json:"name"`
	StartTime           *time.Time             `json:"startTime"`
	EndTime             *time.Time             `json:"endTime"`
	Input               map[string]interface{} `json:"input"`
	Output              map[string]interface{} `json:"output"`
	Metadata            map[string]interface{} `json:"metadata"`
	Level               *domain.Level          `json:"level"`
	StatusMessage       *string                `json:"statusMessage"`
	Version             *string                `json:"version"`
	Model               *string                `json:"model"`
	ModelParameters     map[string]interface{} `json:"modelParameters"`
	Usage               *UsageInput            `json:"usage"`
	PromptId            *uuid.UUID             `json:"promptId"`
}

// UsageInput for token usage
type UsageInput struct {
	PromptTokens     *int `json:"promptTokens"`
	CompletionTokens *int `json:"completionTokens"`
	TotalTokens      *int `json:"totalTokens"`
}

// UpdateObservationInput for updating an observation
type UpdateObservationInput struct {
	Name            *string                `json:"name"`
	EndTime         *time.Time             `json:"endTime"`
	Input           map[string]interface{} `json:"input"`
	Output          map[string]interface{} `json:"output"`
	Metadata        map[string]interface{} `json:"metadata"`
	Level           *domain.Level          `json:"level"`
	StatusMessage   *string                `json:"statusMessage"`
	Model           *string                `json:"model"`
	ModelParameters map[string]interface{} `json:"modelParameters"`
	Usage           *UsageInput            `json:"usage"`
}

// CreateScoreInput for creating a score
type CreateScoreInput struct {
	ID            *string               `json:"id"`
	TraceId       string                `json:"traceId"`
	ObservationId *string               `json:"observationId"`
	Name          string                `json:"name"`
	Value         *float64              `json:"value"`
	StringValue   *string               `json:"stringValue"`
	DataType      *domain.ScoreDataType `json:"dataType"`
	Source        *domain.ScoreSource   `json:"source"`
	Comment       *string               `json:"comment"`
}

// UpdateScoreInput for updating a score
type UpdateScoreInput struct {
	Value       *float64 `json:"value"`
	StringValue *string  `json:"stringValue"`
	Comment     *string  `json:"comment"`
}

// CreatePromptInput for creating a prompt
type CreatePromptInput struct {
	Name        string                 `json:"name"`
	Type        domain.PromptType      `json:"type"`
	Description *string                `json:"description"`
	Prompt      *string                `json:"prompt"`
	Messages    []*PromptMessageInput  `json:"messages"`
	Config      map[string]interface{} `json:"config"`
	Labels      []string               `json:"labels"`
}

// PromptMessageInput for prompt messages
type PromptMessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// UpdatePromptInput for updating a prompt
type UpdatePromptInput struct {
	Description *string                `json:"description"`
	Prompt      *string                `json:"prompt"`
	Messages    []*PromptMessageInput  `json:"messages"`
	Config      map[string]interface{} `json:"config"`
	Labels      []string               `json:"labels"`
	IsActive    *bool                  `json:"isActive"`
}

// CreateDatasetInput for creating a dataset
type CreateDatasetInput struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateDatasetInput for updating a dataset
type UpdateDatasetInput struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CreateDatasetItemInput for creating a dataset item
type CreateDatasetItemInput struct {
	Input               map[string]interface{} `json:"input"`
	ExpectedOutput      map[string]interface{} `json:"expectedOutput"`
	Metadata            map[string]interface{} `json:"metadata"`
	SourceTraceId       *string                `json:"sourceTraceId"`
	SourceObservationId *string                `json:"sourceObservationId"`
}

// UpdateDatasetItemInput for updating a dataset item
type UpdateDatasetItemInput struct {
	Input          map[string]interface{}     `json:"input"`
	ExpectedOutput map[string]interface{}     `json:"expectedOutput"`
	Metadata       map[string]interface{}     `json:"metadata"`
	Status         *domain.DatasetItemStatus  `json:"status"`
}

// CreateDatasetRunInput for creating a dataset run
type CreateDatasetRunInput struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AddDatasetRunItemInput for adding a run item
type AddDatasetRunItemInput struct {
	DatasetItemId uuid.UUID              `json:"datasetItemId"`
	TraceId       *string                `json:"traceId"`
	ObservationId *string                `json:"observationId"`
	Output        map[string]interface{} `json:"output"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CreateEvaluatorInput for creating an evaluator
type CreateEvaluatorInput struct {
	Name            string                 `json:"name"`
	Description     *string                `json:"description"`
	Type            domain.EvaluatorType   `json:"type"`
	ScoreName       string                 `json:"scoreName"`
	ScoreDataType   domain.ScoreDataType   `json:"scoreDataType"`
	ScoreCategories []string               `json:"scoreCategories"`
	PromptTemplate  *string                `json:"promptTemplate"`
	Variables       []string               `json:"variables"`
	Config          map[string]interface{} `json:"config"`
	TargetFilter    map[string]interface{} `json:"targetFilter"`
	SamplingRate    *float64               `json:"samplingRate"`
	Enabled         *bool                  `json:"enabled"`
	TemplateId      *uuid.UUID             `json:"templateId"`
}

// UpdateEvaluatorInput for updating an evaluator
type UpdateEvaluatorInput struct {
	Name            *string                `json:"name"`
	Description     *string                `json:"description"`
	PromptTemplate  *string                `json:"promptTemplate"`
	Variables       []string               `json:"variables"`
	ScoreCategories []string               `json:"scoreCategories"`
	Config          map[string]interface{} `json:"config"`
	TargetFilter    map[string]interface{} `json:"targetFilter"`
	SamplingRate    *float64               `json:"samplingRate"`
	IsActive        *bool                  `json:"isActive"`
}

// CreateProjectInput for creating a project
type CreateProjectInput struct {
	OrganizationId  uuid.UUID              `json:"organizationId"`
	Name            string                 `json:"name"`
	Description     *string                `json:"description"`
	Settings        map[string]interface{} `json:"settings"`
	RetentionDays   *int                   `json:"retentionDays"`
	RateLimitPerMin *int                   `json:"rateLimitPerMin"`
}

// UpdateProjectInput for updating a project
type UpdateProjectInput struct {
	Name            *string                `json:"name"`
	Description     *string                `json:"description"`
	Settings        map[string]interface{} `json:"settings"`
	RetentionDays   *int                   `json:"retentionDays"`
	RateLimitPerMin *int                   `json:"rateLimitPerMin"`
}

// CreateAPIKeyInput for creating an API key
type CreateAPIKeyInput struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// APIKeyWithSecret represents an API key with its secret (only returned on creation)
type APIKeyWithSecret struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	DisplayKey string     `json:"displayKey"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// MetricsInput for querying metrics
type MetricsInput struct {
	FromTimestamp time.Time `json:"fromTimestamp"`
	ToTimestamp   time.Time `json:"toTimestamp"`
	UserId        *string   `json:"userId"`
	SessionId     *string   `json:"sessionId"`
	Name          *string   `json:"name"`
	Tags          []string  `json:"tags"`
}

// DailyCostsInput for querying daily costs
type DailyCostsInput struct {
	FromDate string  `json:"fromDate"`
	ToDate   string  `json:"toDate"`
	GroupBy  *string `json:"groupBy"`
}

// Metrics represents project metrics
type Metrics struct {
	TraceCount       int           `json:"traceCount"`
	ObservationCount int           `json:"observationCount"`
	TotalCost        float64       `json:"totalCost"`
	TotalTokens      int           `json:"totalTokens"`
	AvgLatency       *float64      `json:"avgLatency"`
	P50Latency       *float64      `json:"p50Latency"`
	P95Latency       *float64      `json:"p95Latency"`
	P99Latency       *float64      `json:"p99Latency"`
	ModelUsage       []*ModelUsage `json:"modelUsage"`
}

// ModelUsage represents usage by model
type ModelUsage struct {
	Model            string  `json:"model"`
	Count            int     `json:"count"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	TotalTokens      int     `json:"totalTokens"`
	Cost             float64 `json:"cost"`
}

// DailyCost represents daily cost data
type DailyCost struct {
	Date       string       `json:"date"`
	TotalCost  float64      `json:"totalCost"`
	TraceCount int          `json:"traceCount"`
	ModelCosts []*ModelCost `json:"modelCosts"`
}

// ModelCost represents cost by model
type ModelCost struct {
	Model string  `json:"model"`
	Cost  float64 `json:"cost"`
	Count int     `json:"count"`
}

// TokenUsage represents token usage
type TokenUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}
