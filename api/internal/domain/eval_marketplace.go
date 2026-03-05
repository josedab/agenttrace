package domain

import (
	"time"

	"github.com/google/uuid"
)

// EvalDatasetListing represents an evaluation dataset in the marketplace
type EvalDatasetListing struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Author         string         `json:"author"`
	AuthorID       uuid.UUID      `json:"authorId"`
	Category       string         `json:"category"`
	TaskType       string         `json:"taskType"` // coding, qa, summarization, classification, custom
	SampleCount    int            `json:"sampleCount"`
	ScoringRubric  ScoringRubric  `json:"scoringRubric"`
	BaselineScores []BaselineScore `json:"baselineScores"`
	Tags           []string       `json:"tags"`
	Downloads      int            `json:"downloads"`
	Rating         float64        `json:"rating"`
	RatingCount    int            `json:"ratingCount"`
	IsVerified     bool           `json:"isVerified"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// ScoringRubric defines the scoring criteria for an evaluation dataset
type ScoringRubric struct {
	Criteria     []RubricCriterion `json:"criteria"`
	MaxScore     float64           `json:"maxScore"`
	PassingScore float64           `json:"passingScore"`
}

// RubricCriterion represents a single criterion in a scoring rubric
type RubricCriterion struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	MaxPoints   float64 `json:"maxPoints"`
}

// BaselineScore represents a baseline evaluation score for a model
type BaselineScore struct {
	Model       string    `json:"model"`
	Score       float64   `json:"score"`
	EvaluatedAt time.Time `json:"evaluatedAt"`
	SampleSize  int       `json:"sampleSize"`
}

// EvalDatasetPublishInput represents input for publishing an evaluation dataset
type EvalDatasetPublishInput struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Category      string        `json:"category"`
	TaskType      string        `json:"taskType"`
	Content       string        `json:"content"`
	Tags          []string      `json:"tags"`
	ScoringRubric ScoringRubric `json:"scoringRubric"`
}

// EvalDatasetImportInput represents input for importing a marketplace dataset
type EvalDatasetImportInput struct {
	DatasetID uuid.UUID `json:"datasetId"`
	ProjectID uuid.UUID `json:"projectId"`
}

// EvalMarketplaceSearch represents search parameters for the marketplace
type EvalMarketplaceSearch struct {
	Query     string   `json:"query"`
	Category  *string  `json:"category,omitempty"`
	TaskType  *string  `json:"taskType,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	SortBy    string   `json:"sortBy"`
	MinRating *float64 `json:"minRating,omitempty"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
}
