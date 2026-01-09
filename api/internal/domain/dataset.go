package domain

import (
	"time"

	"github.com/google/uuid"
)

// Dataset represents a dataset for evaluation
type Dataset struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Metadata    string    `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Aggregated fields (populated by resolver)
	ItemCount int64 `json:"itemCount,omitempty"`
	RunCount  int64 `json:"runCount,omitempty"`

	// Related data
	Items []DatasetItem `json:"items,omitempty"`
	Runs  []DatasetRun  `json:"runs,omitempty"`
}

// DatasetInput represents input for creating a dataset
type DatasetInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Metadata    any    `json:"metadata,omitempty"`
}

// DatasetItem represents an item in a dataset
type DatasetItem struct {
	ID                  uuid.UUID         `json:"id"`
	DatasetID           uuid.UUID         `json:"datasetId"`
	Input               string            `json:"input"`
	ExpectedOutput      *string           `json:"expectedOutput,omitempty"`
	Metadata            string            `json:"metadata,omitempty"`
	SourceTraceID       *string           `json:"sourceTraceId,omitempty"`
	SourceObservationID *string           `json:"sourceObservationId,omitempty"`
	Status              DatasetItemStatus `json:"status"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// DatasetItemInput represents input for creating a dataset item
type DatasetItemInput struct {
	Input               any     `json:"input" validate:"required"`
	ExpectedOutput      any     `json:"expectedOutput,omitempty"`
	Metadata            any     `json:"metadata,omitempty"`
	SourceTraceID       *string `json:"sourceTraceId,omitempty"`
	SourceObservationID *string `json:"sourceObservationId,omitempty"`
}

// DatasetItemUpdateInput represents input for updating a dataset item
type DatasetItemUpdateInput struct {
	Input          any                `json:"input,omitempty"`
	ExpectedOutput any                `json:"expectedOutput,omitempty"`
	Metadata       any                `json:"metadata,omitempty"`
	Status         *DatasetItemStatus `json:"status,omitempty"`
}

// DatasetRun represents a run of a dataset
type DatasetRun struct {
	ID          uuid.UUID `json:"id"`
	DatasetID   uuid.UUID `json:"datasetId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Metadata    string    `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Aggregated fields
	ItemCount      int64 `json:"itemCount,omitempty"`
	CompletedCount int64 `json:"completedCount,omitempty"`

	// Related data
	Items []DatasetRunItem `json:"items,omitempty"`
}

// DatasetRunInput represents input for creating a dataset run
type DatasetRunInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Metadata    any    `json:"metadata,omitempty"`
}

// DatasetRunItem represents the link between a dataset item and a trace
type DatasetRunItem struct {
	ID            uuid.UUID `json:"id"`
	DatasetRunID  uuid.UUID `json:"datasetRunId"`
	DatasetItemID uuid.UUID `json:"datasetItemId"`
	TraceID       string    `json:"traceId"`
	ObservationID *string   `json:"observationId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`

	// Related data (populated by resolver)
	DatasetItem *DatasetItem `json:"datasetItem,omitempty"`
	Trace       *Trace       `json:"trace,omitempty"`
	Scores      []Score      `json:"scores,omitempty"`
}

// DatasetRunItemInput represents input for creating a dataset run item
type DatasetRunItemInput struct {
	DatasetItemID string  `json:"datasetItemId" validate:"required"`
	TraceID       string  `json:"traceId" validate:"required"`
	ObservationID *string `json:"observationId,omitempty"`
}

// DatasetFilter represents filter options for querying datasets
type DatasetFilter struct {
	ProjectID uuid.UUID
	Name      *string
}

// DatasetItemFilter represents filter options for querying dataset items
type DatasetItemFilter struct {
	DatasetID uuid.UUID
	Status    *DatasetItemStatus
}

// DatasetList represents a paginated list of datasets
type DatasetList struct {
	Datasets   []Dataset `json:"datasets"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}

// DatasetRunResults represents the results of a dataset run with comparison
type DatasetRunResults struct {
	Run   *DatasetRun              `json:"run"`
	Items []DatasetRunItemWithDiff `json:"items"`
}

// DatasetRunItemWithDiff represents a run item with comparison to expected output
type DatasetRunItemWithDiff struct {
	DatasetRunItem
	ExpectedOutput *string  `json:"expectedOutput,omitempty"`
	ActualOutput   *string  `json:"actualOutput,omitempty"`
	Match          *bool    `json:"match,omitempty"`
	Similarity     *float64 `json:"similarity,omitempty"`
}
