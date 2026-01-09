package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// ============ TRACE RESOLVERS ============

// Trace returns a single trace by ID
func (r *Resolver) Trace(ctx context.Context, id string) (*domain.Trace, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	return r.queryService.GetTrace(ctx, projectID, id)
}

// TracesInput for traces query
type TracesInput struct {
	Limit         *int       `json:"limit,omitempty"`
	Cursor        *string    `json:"cursor,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	SessionID     *string    `json:"sessionId,omitempty"`
	Name          *string    `json:"name,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
	Version       *string    `json:"version,omitempty"`
	Release       *string    `json:"release,omitempty"`
	OrderBy       *string    `json:"orderBy,omitempty"`
	Order         *string    `json:"order,omitempty"`
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

// PageInfo represents pagination info
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}

// Traces returns paginated traces
func (r *Resolver) Traces(ctx context.Context, input TracesInput) (*TraceConnection, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}
	if limit > 100 {
		limit = 100
	}

	filter := &domain.TraceFilter{
		ProjectID: projectID,
		UserID:    input.UserID,
		SessionID: input.SessionID,
		Name:      input.Name,
		Tags:      input.Tags,
		FromTime:  input.FromTimestamp,
		ToTime:    input.ToTimestamp,
		Version:   input.Version,
		Release:   input.Release,
	}

	list, err := r.queryService.ListTraces(ctx, filter, limit, 0)
	if err != nil {
		return nil, err
	}

	edges := make([]*TraceEdge, len(list.Traces))
	for i, trace := range list.Traces {
		traceCopy := trace
		edges[i] = &TraceEdge{
			Node:   &traceCopy,
			Cursor: trace.ID,
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &TraceConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     int64(len(edges)) < list.TotalCount,
			HasPreviousPage: input.Cursor != nil,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(list.TotalCount),
	}, nil
}

// ============ OBSERVATION RESOLVERS ============

// Observation returns a single observation by ID
func (r *Resolver) Observation(ctx context.Context, id string) (*domain.Observation, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	return r.queryService.GetObservation(ctx, projectID, id)
}

// ObservationsInput for observations query
type ObservationsInput struct {
	TraceID             *string                  `json:"traceId,omitempty"`
	ParentObservationID *string                  `json:"parentObservationId,omitempty"`
	Type                *domain.ObservationType  `json:"type,omitempty"`
	Name                *string                  `json:"name,omitempty"`
	Limit               *int                     `json:"limit,omitempty"`
	Cursor              *string                  `json:"cursor,omitempty"`
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

// Observations returns paginated observations
func (r *Resolver) Observations(ctx context.Context, input ObservationsInput) (*ObservationConnection, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}

	filter := &domain.ObservationFilter{
		ProjectID:           projectID,
		TraceID:             input.TraceID,
		ParentObservationID: input.ParentObservationID,
		Type:                input.Type,
		Name:                input.Name,
	}

	observations, totalCount, err := r.queryService.ListObservations(ctx, filter, limit, 0)
	if err != nil {
		return nil, err
	}

	edges := make([]*ObservationEdge, len(observations))
	for i, obs := range observations {
		obsCopy := obs
		edges[i] = &ObservationEdge{
			Node:   &obsCopy,
			Cursor: obs.ID,
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &ObservationConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     len(edges) < int(totalCount),
			HasPreviousPage: input.Cursor != nil,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(totalCount),
	}, nil
}

// ============ SCORE RESOLVERS ============

// Score returns a single score by ID
func (r *Resolver) Score(ctx context.Context, id string) (*domain.Score, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	return r.scoreService.Get(ctx, projectID, id)
}

// ScoresInput for scores query
type ScoresInput struct {
	TraceID       *string             `json:"traceId,omitempty"`
	ObservationID *string             `json:"observationId,omitempty"`
	Name          *string             `json:"name,omitempty"`
	Source        *domain.ScoreSource `json:"source,omitempty"`
	Limit         *int                `json:"limit,omitempty"`
	Cursor        *string             `json:"cursor,omitempty"`
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

// Scores returns paginated scores
func (r *Resolver) Scores(ctx context.Context, input ScoresInput) (*ScoreConnection, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}

	filter := &domain.ScoreFilter{
		ProjectID:     projectID,
		TraceID:       input.TraceID,
		ObservationID: input.ObservationID,
		Name:          input.Name,
		Source:        input.Source,
	}

	scoreList, err := r.scoreService.List(ctx, filter, limit, 0)
	if err != nil {
		return nil, err
	}

	edges := make([]*ScoreEdge, len(scoreList.Scores))
	for i, score := range scoreList.Scores {
		scoreCopy := score
		edges[i] = &ScoreEdge{
			Node:   &scoreCopy,
			Cursor: scoreCopy.ID.String(),
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &ScoreConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     len(edges) < int(scoreList.TotalCount),
			HasPreviousPage: input.Cursor != nil,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(scoreList.TotalCount),
	}, nil
}

// ============ SESSION RESOLVERS ============

// Session returns a single session by ID
func (r *Resolver) Session(ctx context.Context, id string) (*domain.Session, error) {
	// Session queries are derived from traces with sessionID
	// Return a basic session structure from trace data
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	traces, err := r.queryService.GetSessionTraces(ctx, projectID, id)
	if err != nil {
		return nil, err
	}
	if len(traces) == 0 {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return &domain.Session{
		ID:        id,
		ProjectID: projectID,
	}, nil
}

// SessionsInput for sessions query
type SessionsInput struct {
	Limit         *int       `json:"limit,omitempty"`
	Cursor        *string    `json:"cursor,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// SessionConnection represents a paginated list of sessions
type SessionConnection struct {
	Edges      []*SessionEdge `json:"edges"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

// SessionEdge represents a session in the connection
type SessionEdge struct {
	Node   *domain.Session `json:"node"`
	Cursor string          `json:"cursor"`
}

// Sessions returns paginated sessions
// Note: Sessions are derived from traces with sessionID - this is a placeholder
func (r *Resolver) Sessions(ctx context.Context, input SessionsInput) (*SessionConnection, error) {
	_ = ctx.Value(ContextKeyProjectID).(uuid.UUID)

	// Session listing is not directly supported - return empty list
	// In a full implementation, this would aggregate unique sessionIDs from traces
	return &SessionConnection{
		Edges: []*SessionEdge{},
		PageInfo: &PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		},
		TotalCount: 0,
	}, nil
}

// ============ PROMPT RESOLVERS ============

// Prompt returns a prompt by name and optional version/label
func (r *Resolver) Prompt(ctx context.Context, name string, version *int, label *string) (*domain.Prompt, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	if version != nil {
		return r.promptService.GetByNameAndVersion(ctx, projectID, name, *version)
	}
	if label != nil {
		return r.promptService.GetByNameAndLabel(ctx, projectID, name, *label)
	}
	return r.promptService.GetByName(ctx, projectID, name)
}

// PromptsInput for prompts query
type PromptsInput struct {
	Name   *string  `json:"name,omitempty"`
	Label  *string  `json:"label,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Limit  *int     `json:"limit,omitempty"`
	Cursor *string  `json:"cursor,omitempty"`
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

// Prompts returns paginated prompts
func (r *Resolver) Prompts(ctx context.Context, input PromptsInput) (*PromptConnection, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}

	filter := &domain.PromptFilter{
		ProjectID: projectID,
		Name:      input.Name,
		Label:     input.Label,
		Tags:      input.Tags,
	}

	list, err := r.promptService.List(ctx, filter, limit, 0)
	if err != nil {
		return nil, err
	}

	edges := make([]*PromptEdge, len(list.Prompts))
	for i, prompt := range list.Prompts {
		promptCopy := prompt
		edges[i] = &PromptEdge{
			Node:   &promptCopy,
			Cursor: prompt.ID.String(),
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &PromptConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     int64(len(edges)) < list.TotalCount,
			HasPreviousPage: input.Cursor != nil,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(list.TotalCount),
	}, nil
}

// ============ DATASET RESOLVERS ============

// Dataset returns a dataset by ID
func (r *Resolver) Dataset(ctx context.Context, id uuid.UUID) (*domain.Dataset, error) {
	return r.datasetService.Get(ctx, id)
}

// DatasetByName returns a dataset by name
func (r *Resolver) DatasetByName(ctx context.Context, name string) (*domain.Dataset, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)
	return r.datasetService.GetByName(ctx, projectID, name)
}

// DatasetsInput for datasets query
type DatasetsInput struct {
	Name   *string `json:"name,omitempty"`
	Limit  *int    `json:"limit,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
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

// Datasets returns paginated datasets
func (r *Resolver) Datasets(ctx context.Context, input DatasetsInput) (*DatasetConnection, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}

	filter := &domain.DatasetFilter{
		ProjectID: projectID,
	}

	list, err := r.datasetService.List(ctx, filter, limit, 0)
	if err != nil {
		return nil, err
	}

	edges := make([]*DatasetEdge, len(list.Datasets))
	for i, dataset := range list.Datasets {
		datasetCopy := dataset
		edges[i] = &DatasetEdge{
			Node:   &datasetCopy,
			Cursor: dataset.ID.String(),
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &DatasetConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     int64(len(edges)) < list.TotalCount,
			HasPreviousPage: input.Cursor != nil,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(list.TotalCount),
	}, nil
}

// ============ EVALUATOR RESOLVERS ============

// Evaluator returns an evaluator by ID
func (r *Resolver) Evaluator(ctx context.Context, id uuid.UUID) (*domain.Evaluator, error) {
	return r.evalService.Get(ctx, id)
}

// EvaluatorsInput for evaluators query
type EvaluatorsInput struct {
	Type    *domain.EvaluatorType `json:"type,omitempty"`
	Enabled *bool                 `json:"enabled,omitempty"`
	Limit   *int                  `json:"limit,omitempty"`
	Cursor  *string               `json:"cursor,omitempty"`
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

// Evaluators returns paginated evaluators
func (r *Resolver) Evaluators(ctx context.Context, input EvaluatorsInput) (*EvaluatorConnection, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}

	filter := &domain.EvaluatorFilter{
		ProjectID: projectID,
		Type:      input.Type,
		Enabled:   input.Enabled,
	}

	list, err := r.evalService.List(ctx, filter, limit, 0)
	if err != nil {
		return nil, err
	}

	edges := make([]*EvaluatorEdge, len(list.Evaluators))
	for i, eval := range list.Evaluators {
		evalCopy := eval
		edges[i] = &EvaluatorEdge{
			Node:   &evalCopy,
			Cursor: eval.ID.String(),
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &EvaluatorConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     int64(len(edges)) < list.TotalCount,
			HasPreviousPage: input.Cursor != nil,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(list.TotalCount),
	}, nil
}

// EvaluatorTemplates returns all evaluator templates
func (r *Resolver) EvaluatorTemplates(ctx context.Context) ([]domain.EvaluatorTemplate, error) {
	return r.evalService.ListTemplates(ctx)
}

// ============ ORGANIZATION RESOLVERS ============

// Organization returns an organization by ID
func (r *Resolver) Organization(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return r.orgService.Get(ctx, id)
}

// Organizations returns all organizations for the current user
func (r *Resolver) Organizations(ctx context.Context) ([]domain.Organization, error) {
	userID := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return r.orgService.ListByUser(ctx, userID)
}

// Project returns a project by ID
func (r *Resolver) Project(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return r.projectService.Get(ctx, id)
}

// Projects returns projects for the current user or organization
func (r *Resolver) Projects(ctx context.Context, organizationID *uuid.UUID) ([]domain.Project, error) {
	if organizationID != nil {
		return r.projectService.ListByOrganization(ctx, *organizationID)
	}
	userID := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return r.projectService.ListByUser(ctx, userID)
}

// Me returns the current user
func (r *Resolver) Me(ctx context.Context) (*domain.User, error) {
	userID := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return r.authService.GetUserByID(ctx, userID)
}

// ============ METRICS RESOLVERS ============

// MetricsInput for metrics query
type MetricsInput struct {
	FromTimestamp time.Time `json:"fromTimestamp"`
	ToTimestamp   time.Time `json:"toTimestamp"`
	UserID        *string   `json:"userId,omitempty"`
	SessionID     *string   `json:"sessionId,omitempty"`
	Name          *string   `json:"name,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
}

// Metrics represents aggregated metrics
type Metrics struct {
	TraceCount       int           `json:"traceCount"`
	ObservationCount int           `json:"observationCount"`
	TotalCost        float64       `json:"totalCost"`
	TotalTokens      int           `json:"totalTokens"`
	AvgLatency       *float64      `json:"avgLatency,omitempty"`
	P50Latency       *float64      `json:"p50Latency,omitempty"`
	P95Latency       *float64      `json:"p95Latency,omitempty"`
	P99Latency       *float64      `json:"p99Latency,omitempty"`
	ModelUsage       []*ModelUsage `json:"modelUsage"`
}

// ModelUsage represents usage per model
type ModelUsage struct {
	Model            string  `json:"model"`
	Count            int     `json:"count"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	TotalTokens      int     `json:"totalTokens"`
	Cost             float64 `json:"cost"`
}

// Metrics returns aggregated metrics
func (r *Resolver) Metrics(ctx context.Context, input MetricsInput) (*Metrics, error) {
	projectID := ctx.Value(ContextKeyProjectID).(uuid.UUID)

	filter := &domain.TraceFilter{
		ProjectID: projectID,
		FromTime:  &input.FromTimestamp,
		ToTime:    &input.ToTimestamp,
		UserID:    input.UserID,
		SessionID: input.SessionID,
		Name:      input.Name,
		Tags:      input.Tags,
	}

	stats, err := r.queryService.GetTraceStats(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		TraceCount:  int(stats.TotalCount),
		TotalCost:   stats.TotalCost,
		TotalTokens: int(stats.TotalTokens),
		ModelUsage:  []*ModelUsage{},
	}, nil
}

// DailyCostsInput for daily costs query
type DailyCostsInput struct {
	FromDate string  `json:"fromDate"`
	ToDate   string  `json:"toDate"`
	GroupBy  *string `json:"groupBy,omitempty"`
}

// DailyCost represents daily cost breakdown
type DailyCost struct {
	Date       string       `json:"date"`
	TotalCost  float64      `json:"totalCost"`
	TraceCount int          `json:"traceCount"`
	ModelCosts []*ModelCost `json:"modelCosts"`
}

// ModelCost represents cost per model
type ModelCost struct {
	Model string  `json:"model"`
	Cost  float64 `json:"cost"`
	Count int     `json:"count"`
}

// DailyCosts returns daily cost breakdown
func (r *Resolver) DailyCosts(ctx context.Context, input DailyCostsInput) ([]*DailyCost, error) {
	// This would need a dedicated service method to query daily costs
	// For now, return empty slice
	return []*DailyCost{}, nil
}
