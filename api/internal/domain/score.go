package domain

import (
	"time"

	"github.com/google/uuid"
)

// Score represents an evaluation score for a trace or observation
type Score struct {
	ID            uuid.UUID     `json:"id" ch:"id"`
	ProjectID     uuid.UUID     `json:"projectId" ch:"project_id"`
	TraceID       string        `json:"traceId" ch:"trace_id"`
	ObservationID *string       `json:"observationId,omitempty" ch:"observation_id"`
	Name          string        `json:"name" ch:"name"`
	Source        ScoreSource   `json:"source" ch:"source"`
	DataType      ScoreDataType `json:"dataType" ch:"data_type"`
	Value         *float64      `json:"value,omitempty" ch:"value"`
	StringValue   *string       `json:"stringValue,omitempty" ch:"string_value"`
	Comment       string        `json:"comment,omitempty" ch:"comment"`
	ConfigID      *uuid.UUID    `json:"configId,omitempty" ch:"config_id"`
	AuthorUserID  *uuid.UUID    `json:"authorUserId,omitempty" ch:"author_user_id"`
	CreatedAt     time.Time     `json:"createdAt" ch:"created_at"`
	UpdatedAt     time.Time     `json:"updatedAt" ch:"updated_at"`
}

// ScoreInput represents input for creating a score
type ScoreInput struct {
	TraceID       string        `json:"traceId" validate:"required"`
	ObservationID *string       `json:"observationId,omitempty"`
	Name          string        `json:"name" validate:"required"`
	Source        ScoreSource   `json:"source,omitempty"`
	DataType      ScoreDataType `json:"dataType,omitempty"`
	Value         *float64      `json:"value,omitempty"`
	StringValue   *string       `json:"stringValue,omitempty"`
	Comment       *string       `json:"comment,omitempty"`
	ConfigID      *string       `json:"configId,omitempty"`
}

// ScoreFilter represents filter options for querying scores
type ScoreFilter struct {
	ProjectID     uuid.UUID
	TraceID       *string
	ObservationID *string
	Name          *string
	Source        *ScoreSource
	DataType      *ScoreDataType
	ConfigID      *uuid.UUID
	FromTime      *time.Time
	ToTime        *time.Time
}

// ScoreList represents a paginated list of scores
type ScoreList struct {
	Scores     []Score `json:"scores"`
	TotalCount int64   `json:"totalCount"`
	HasMore    bool    `json:"hasMore"`
}

// ScoreStats represents statistics for scores
type ScoreStats struct {
	Name        string   `json:"name"`
	Count       int64    `json:"count"`
	AvgValue    *float64 `json:"avgValue,omitempty"`
	MinValue    *float64 `json:"minValue,omitempty"`
	MaxValue    *float64 `json:"maxValue,omitempty"`
	MedianValue *float64 `json:"medianValue,omitempty"`
}

// ScoreConfig represents a score configuration (for evaluators)
type ScoreConfig struct {
	ID             uuid.UUID     `json:"id"`
	ProjectID      uuid.UUID     `json:"projectId"`
	Name           string        `json:"name"`
	DataType       ScoreDataType `json:"dataType"`
	Categories     []string      `json:"categories,omitempty"`
	Description    string        `json:"description,omitempty"`
	MinValue       *float64      `json:"minValue,omitempty"`
	MaxValue       *float64      `json:"maxValue,omitempty"`
	IsArchived     bool          `json:"isArchived"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// ScoreConfigInput represents input for creating a score config
type ScoreConfigInput struct {
	Name        string        `json:"name" validate:"required"`
	DataType    ScoreDataType `json:"dataType" validate:"required"`
	Categories  []string      `json:"categories,omitempty"`
	Description *string       `json:"description,omitempty"`
	MinValue    *float64      `json:"minValue,omitempty"`
	MaxValue    *float64      `json:"maxValue,omitempty"`
}

// ValidateScore validates a score value based on data type
func ValidateScore(dataType ScoreDataType, value *float64, stringValue *string, categories []string) bool {
	switch dataType {
	case ScoreDataTypeNumeric:
		return value != nil
	case ScoreDataTypeBoolean:
		if value != nil {
			return *value == 0 || *value == 1
		}
		return false
	case ScoreDataTypeCategorical:
		if stringValue == nil {
			return false
		}
		for _, cat := range categories {
			if cat == *stringValue {
				return true
			}
		}
		return false
	}
	return false
}
