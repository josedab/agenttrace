package domain

import (
	"time"

	"github.com/google/uuid"
)

// TrainingDatasetFormat represents the export format for training data
type TrainingDatasetFormat string

const (
	TrainingFormatOpenAIFinetune TrainingDatasetFormat = "openai_finetune"
	TrainingFormatAnthropicRLHF  TrainingDatasetFormat = "anthropic_rlhf"
	TrainingFormatJSONL          TrainingDatasetFormat = "jsonl"
)

// TrainingDatasetStatus represents the status of a training dataset
type TrainingDatasetStatus string

const (
	TrainingDatasetBuilding TrainingDatasetStatus = "building"
	TrainingDatasetReady    TrainingDatasetStatus = "ready"
	TrainingDatasetExported TrainingDatasetStatus = "exported"
)

// TrainingDataset represents a curated training dataset
type TrainingDataset struct {
	ID           uuid.UUID             `json:"id"`
	ProjectID    uuid.UUID             `json:"projectId"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Format       TrainingDatasetFormat `json:"format"`
	SourceFilter DatasetSourceFilter   `json:"sourceFilter"`
	Items        int                   `json:"items"`
	Status       TrainingDatasetStatus `json:"status"`
	CreatedAt    time.Time             `json:"createdAt"`
}

// TrainingItem represents a single training data item
type TrainingItem struct {
	TraceID string  `json:"traceId"`
	Input   string  `json:"input"`
	Output  string  `json:"output"`
	Score   float64 `json:"score"`
	Label   string  `json:"label"`
}

// TrainingExport represents an exported training dataset
type TrainingExport struct {
	ID         uuid.UUID             `json:"id"`
	DatasetID  uuid.UUID             `json:"datasetId"`
	Format     TrainingDatasetFormat `json:"format"`
	URL        string                `json:"url"`
	LineCount  int                   `json:"lineCount"`
	TokenCount int64                 `json:"tokenCount"`
	CreatedAt  time.Time             `json:"createdAt"`
}

// FailurePattern represents a detected failure pattern
type FailurePattern struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"projectId"`
	Pattern         string    `json:"pattern"`
	Description     string    `json:"description"`
	Frequency       int       `json:"frequency"`
	ExampleTraceIDs []string  `json:"exampleTraceIds"`
	SuggestedFix    string    `json:"suggestedFix"`
}

// TrainingDatasetInput represents input for creating a training dataset
type TrainingDatasetInput struct {
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Format       TrainingDatasetFormat `json:"format"`
	SourceFilter DatasetSourceFilter   `json:"sourceFilter"`
	MinScore     float64               `json:"minScore"`
}

// DatasetSourceFilter represents filters for selecting training data
type DatasetSourceFilter struct {
	MinQualityScore    float64  `json:"minQualityScore"`
	AgentNames         []string `json:"agentNames,omitempty"`
	DateRange          []string `json:"dateRange,omitempty"`
	IncludeSuccessOnly bool     `json:"includeSuccessOnly"`
}
