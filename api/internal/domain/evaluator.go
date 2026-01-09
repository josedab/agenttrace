package domain

import (
	"time"

	"github.com/google/uuid"
)

// Evaluator represents an LLM-as-judge evaluator
type Evaluator struct {
	ID             uuid.UUID     `json:"id"`
	ProjectID      uuid.UUID     `json:"projectId"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	Type           EvaluatorType `json:"type"`
	Config         string        `json:"config,omitempty"`
	PromptTemplate string        `json:"promptTemplate,omitempty"`
	Variables      []string      `json:"variables"`
	TargetFilter   string        `json:"targetFilter,omitempty"`
	SamplingRate   float64       `json:"samplingRate"`
	ScoreName      string        `json:"scoreName"`
	ScoreDataType  ScoreDataType `json:"scoreDataType"`
	ScoreCategories []string     `json:"scoreCategories,omitempty"`
	Enabled        bool          `json:"enabled"`
	CreatedBy      *uuid.UUID    `json:"createdBy,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`

	// Stats (populated by resolver)
	EvalCount    int64    `json:"evalCount,omitempty"`
	AvgScore     *float64 `json:"avgScore,omitempty"`
	LastEvalTime *time.Time `json:"lastEvalTime,omitempty"`
}

// EvaluatorInput represents input for creating an evaluator
type EvaluatorInput struct {
	Name            string         `json:"name" validate:"required"`
	Description     *string        `json:"description,omitempty"`
	Type            *EvaluatorType `json:"type,omitempty"`
	Config          any            `json:"config,omitempty"`
	PromptTemplate  string         `json:"promptTemplate,omitempty"`
	Variables       []string       `json:"variables,omitempty"`
	TargetFilter    any            `json:"targetFilter,omitempty"`
	SamplingRate    *float64       `json:"samplingRate,omitempty"`
	ScoreName       string         `json:"scoreName" validate:"required"`
	ScoreDataType   *ScoreDataType `json:"scoreDataType,omitempty"`
	ScoreCategories []string       `json:"scoreCategories,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	TemplateID      *string        `json:"templateId,omitempty"`
}

// EvaluatorUpdateInput represents input for updating an evaluator
type EvaluatorUpdateInput struct {
	Name            *string        `json:"name,omitempty"`
	Description     *string        `json:"description,omitempty"`
	Config          any            `json:"config,omitempty"`
	PromptTemplate  *string        `json:"promptTemplate,omitempty"`
	Variables       []string       `json:"variables,omitempty"`
	TargetFilter    any            `json:"targetFilter,omitempty"`
	SamplingRate    *float64       `json:"samplingRate,omitempty"`
	ScoreCategories []string       `json:"scoreCategories,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
}

// EvaluatorTemplate represents a built-in evaluator template
type EvaluatorTemplate struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	PromptTemplate  string        `json:"promptTemplate"`
	Variables       []string      `json:"variables"`
	ScoreDataType   ScoreDataType `json:"scoreDataType"`
	ScoreCategories []string      `json:"scoreCategories,omitempty"`
	Config          string        `json:"config,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
}

// EvaluatorFilter represents filter options for querying evaluators
type EvaluatorFilter struct {
	ProjectID uuid.UUID
	Name      *string
	Type      *EvaluatorType
	Enabled   *bool
}

// EvaluatorList represents a paginated list of evaluators
type EvaluatorList struct {
	Evaluators []Evaluator `json:"evaluators"`
	TotalCount int64       `json:"totalCount"`
	HasMore    bool        `json:"hasMore"`
}

// EvaluationJob represents a pending or completed evaluation job
type EvaluationJob struct {
	ID            uuid.UUID  `json:"id"`
	EvaluatorID   uuid.UUID  `json:"evaluatorId"`
	TraceID       string     `json:"traceId"`
	ObservationID *string    `json:"observationId,omitempty"`
	Status        JobStatus  `json:"status"`
	Result        *string    `json:"result,omitempty"`
	Error         *string    `json:"error,omitempty"`
	Attempts      int        `json:"attempts"`
	ScheduledAt   time.Time  `json:"scheduledAt"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// EvaluationResult represents the result of an evaluation
type EvaluationResult struct {
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning,omitempty"`
}

// TargetFilter represents a filter for matching traces/observations
type TargetFilter struct {
	TraceFilters       map[string]interface{} `json:"traceFilters,omitempty"`
	ObservationFilters map[string]interface{} `json:"observationFilters,omitempty"`
	Models             []string               `json:"models,omitempty"`
	Names              []string               `json:"names,omitempty"`
}

// MatchesTrace checks if a trace matches the target filter
func (f *TargetFilter) MatchesTrace(trace *Trace) bool {
	if f == nil {
		return true
	}

	// Check trace filters
	if len(f.Names) > 0 {
		found := false
		for _, name := range f.Names {
			if trace.Name == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// MatchesObservation checks if an observation matches the target filter
func (f *TargetFilter) MatchesObservation(obs *Observation) bool {
	if f == nil {
		return true
	}

	// Check model filter
	if len(f.Models) > 0 {
		found := false
		for _, model := range f.Models {
			if obs.Model == model {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check name filter
	if len(f.Names) > 0 {
		found := false
		for _, name := range f.Names {
			if obs.Name == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// AnnotationQueue represents a queue for human annotation
type AnnotationQueue struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ScoreName   string    `json:"scoreName"`
	ScoreConfig string    `json:"scoreConfig,omitempty"`
	Filters     string    `json:"filters,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Stats
	PendingCount   int64 `json:"pendingCount,omitempty"`
	CompletedCount int64 `json:"completedCount,omitempty"`
}

// AnnotationQueueItem represents an item in an annotation queue
type AnnotationQueueItem struct {
	ID            uuid.UUID  `json:"id"`
	QueueID       uuid.UUID  `json:"queueId"`
	TraceID       string     `json:"traceId"`
	ObservationID *string    `json:"observationId,omitempty"`
	Status        JobStatus  `json:"status"`
	CompletedBy   *uuid.UUID `json:"completedBy,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}
