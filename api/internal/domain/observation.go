package domain

import (
	"time"

	"github.com/google/uuid"
)

// Observation represents an observation within a trace
type Observation struct {
	ID                  string          `json:"id" ch:"id"`
	TraceID             string          `json:"traceId" ch:"trace_id"`
	ProjectID           uuid.UUID       `json:"projectId" ch:"project_id"`
	ParentObservationID *string         `json:"parentObservationId,omitempty" ch:"parent_observation_id"`
	Type                ObservationType `json:"type" ch:"type"`
	Name                string          `json:"name" ch:"name"`
	Level               Level           `json:"level" ch:"level"`
	StatusMessage       string          `json:"statusMessage,omitempty" ch:"status_message"`
	Metadata            string          `json:"metadata,omitempty" ch:"metadata"`
	StartTime           time.Time       `json:"startTime" ch:"start_time"`
	EndTime             *time.Time      `json:"endTime,omitempty" ch:"end_time"`
	CompletionStartTime *time.Time      `json:"completionStartTime,omitempty" ch:"completion_start_time"`
	DurationMs          float64         `json:"durationMs" ch:"duration_ms"`
	TimeToFirstTokenMs  float64         `json:"timeToFirstTokenMs" ch:"time_to_first_token_ms"`
	Input               string          `json:"input,omitempty" ch:"input"`
	Output              string          `json:"output,omitempty" ch:"output"`

	// Generation-specific fields
	Model           string `json:"model,omitempty" ch:"model"`
	ModelParameters string `json:"modelParameters,omitempty" ch:"model_parameters"`

	// Token usage
	UsageDetails UsageDetails `json:"usageDetails" ch:"-"`

	// Costs
	CostDetails CostDetails `json:"costDetails" ch:"-"`

	// Prompt tracking
	PromptID      *uuid.UUID `json:"promptId,omitempty" ch:"prompt_id"`
	PromptVersion *uint32    `json:"promptVersion,omitempty" ch:"prompt_version"`
	PromptName    *string    `json:"promptName,omitempty" ch:"prompt_name"`

	// Version
	Version string `json:"version,omitempty" ch:"version"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" ch:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" ch:"updated_at"`

	// Related data (populated by resolvers)
	Children []Observation `json:"children,omitempty" ch:"-"`
	Scores   []Score       `json:"scores,omitempty" ch:"-"`
}

// UsageDetails contains token usage information
type UsageDetails struct {
	InputTokens         uint64 `json:"inputTokens" ch:"usage_input_tokens"`
	OutputTokens        uint64 `json:"outputTokens" ch:"usage_output_tokens"`
	TotalTokens         uint64 `json:"totalTokens" ch:"usage_total_tokens"`
	CacheReadTokens     uint64 `json:"cacheReadTokens,omitempty" ch:"usage_cache_read_tokens"`
	CacheCreationTokens uint64 `json:"cacheCreationTokens,omitempty" ch:"usage_cache_creation_tokens"`
}

// CostDetails contains cost information
type CostDetails struct {
	InputCost  float64 `json:"inputCost" ch:"input_cost"`
	OutputCost float64 `json:"outputCost" ch:"output_cost"`
	TotalCost  float64 `json:"totalCost" ch:"total_cost"`
	Currency   string  `json:"currency,omitempty" ch:"currency"`
}

// ObservationInput represents input for creating/updating an observation
type ObservationInput struct {
	ID                  *string          `json:"id,omitempty"`
	TraceID             *string          `json:"traceId,omitempty"`
	ParentObservationID *string          `json:"parentObservationId,omitempty"`
	Type                *ObservationType `json:"type,omitempty"`
	Name                *string          `json:"name,omitempty"`
	Level               *Level           `json:"level,omitempty"`
	StatusMessage       *string          `json:"statusMessage,omitempty"`
	Metadata            any              `json:"metadata,omitempty"`
	StartTime           *time.Time       `json:"startTime,omitempty"`
	EndTime             *time.Time       `json:"endTime,omitempty"`
	CompletionStartTime *time.Time       `json:"completionStartTime,omitempty"`
	Input               any              `json:"input,omitempty"`
	Output              any              `json:"output,omitempty"`

	// Generation-specific
	Model           *string `json:"model,omitempty"`
	ModelParameters any     `json:"modelParameters,omitempty"`
	Usage           any     `json:"usage,omitempty"`

	// Prompt tracking
	PromptID      *string `json:"promptId,omitempty"`
	PromptVersion *int    `json:"promptVersion,omitempty"`
	PromptName    *string `json:"promptName,omitempty"`

	// Version
	Version *string `json:"version,omitempty"`
}

// GenerationInput represents input for creating a generation observation
type GenerationInput struct {
	ObservationInput

	// LLM-specific fields
	Model           string `json:"model"`
	ModelParameters any    `json:"modelParameters,omitempty"`

	// Request/Response
	Messages         any `json:"messages,omitempty"`
	Prompt           any `json:"prompt,omitempty"`
	Completion       any `json:"completion,omitempty"`
	CompletionTokens any `json:"completionTokens,omitempty"`
	PromptTokens     any `json:"promptTokens,omitempty"`
	TotalTokens      any `json:"totalTokens,omitempty"`

	// Usage object (alternative format)
	Usage *UsageDetailsInput `json:"usage,omitempty"`
}

// UsageDetailsInput represents input for usage details
type UsageDetailsInput struct {
	InputTokens         *int64 `json:"inputTokens,omitempty"`
	OutputTokens        *int64 `json:"outputTokens,omitempty"`
	TotalTokens         *int64 `json:"totalTokens,omitempty"`
	CacheReadTokens     *int64 `json:"cacheReadTokens,omitempty"`
	CacheCreationTokens *int64 `json:"cacheCreationTokens,omitempty"`

	// Alternative field names
	PromptTokens     *int64 `json:"promptTokens,omitempty"`
	CompletionTokens *int64 `json:"completionTokens,omitempty"`
}

// Normalize normalizes usage details input
func (u *UsageDetailsInput) Normalize() UsageDetails {
	var details UsageDetails

	if u.InputTokens != nil {
		details.InputTokens = uint64(*u.InputTokens)
	} else if u.PromptTokens != nil {
		details.InputTokens = uint64(*u.PromptTokens)
	}

	if u.OutputTokens != nil {
		details.OutputTokens = uint64(*u.OutputTokens)
	} else if u.CompletionTokens != nil {
		details.OutputTokens = uint64(*u.CompletionTokens)
	}

	if u.TotalTokens != nil {
		details.TotalTokens = uint64(*u.TotalTokens)
	} else {
		details.TotalTokens = details.InputTokens + details.OutputTokens
	}

	if u.CacheReadTokens != nil {
		details.CacheReadTokens = uint64(*u.CacheReadTokens)
	}

	if u.CacheCreationTokens != nil {
		details.CacheCreationTokens = uint64(*u.CacheCreationTokens)
	}

	return details
}

// ObservationFilter represents filter options for querying observations
type ObservationFilter struct {
	ProjectID           uuid.UUID
	TraceID             *string
	ParentObservationID *string
	Type                *ObservationType
	Name                *string
	Model               *string
	Level               *Level
	FromTime            *time.Time
	ToTime              *time.Time
}

// ObservationTree represents observations organized in a tree structure
type ObservationTree struct {
	Observation *Observation       `json:"observation"`
	Children    []*ObservationTree `json:"children,omitempty"`
}

// BuildObservationTree builds a tree from flat observations
func BuildObservationTree(observations []Observation) []*ObservationTree {
	// Create a map for quick lookup
	nodeMap := make(map[string]*ObservationTree)
	var roots []*ObservationTree

	// First pass: create all nodes
	for i := range observations {
		obs := &observations[i]
		nodeMap[obs.ID] = &ObservationTree{
			Observation: obs,
			Children:    []*ObservationTree{},
		}
	}

	// Second pass: build tree
	for i := range observations {
		obs := &observations[i]
		node := nodeMap[obs.ID]

		if obs.ParentObservationID == nil || *obs.ParentObservationID == "" {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*obs.ParentObservationID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent not found, treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots
}
