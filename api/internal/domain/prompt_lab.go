package domain

import (
	"time"

	"github.com/google/uuid"
)

// PromptExperimentStatus represents the lifecycle state of a prompt experiment.
type PromptExperimentStatus string

const (
	PromptExpStatusDraft     PromptExperimentStatus = "draft"
	PromptExpStatusRunning   PromptExperimentStatus = "running"
	PromptExpStatusCompleted PromptExperimentStatus = "completed"
	PromptExpStatusCancelled PromptExperimentStatus = "cancelled"
)

// IsValid reports whether s is a recognized prompt experiment status value.
func (s PromptExperimentStatus) IsValid() bool {
	switch s {
	case PromptExpStatusDraft, PromptExpStatusRunning, PromptExpStatusCompleted, PromptExpStatusCancelled:
		return true
	}
	return false
}

// PromptExperiment represents an A/B test comparing multiple prompt variants
// for a given prompt name within a project.
type PromptExperiment struct {
	ID          uuid.UUID        `json:"id"`
	ProjectID   uuid.UUID        `json:"projectId"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	PromptName  string           `json:"promptName"`
	Status      PromptExperimentStatus `json:"status"`
	Variants    []PromptVariant  `json:"variants"`
	WinnerID    *uuid.UUID       `json:"winnerId,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`
}

// PromptVariant is a single variant in a prompt experiment, carrying its
// content, traffic allocation, and observed metrics.
type PromptVariant struct {
	ID            uuid.UUID            `json:"id"`
	ExperimentID  uuid.UUID            `json:"experimentId"`
	Name          string               `json:"name"`
	PromptContent string               `json:"promptContent"`
	TrafficWeight float64              `json:"trafficWeight"` // 0-1
	Metrics       PromptVariantMetrics `json:"metrics"`
	IsControl     bool                 `json:"isControl"`
}

// PromptVariantMetrics holds the aggregated performance metrics for a prompt variant.
type PromptVariantMetrics struct {
	Traces       int     `json:"traces"`
	AvgQuality   float64 `json:"avgQuality"`
	AvgCost      float64 `json:"avgCost"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	AvgTokens    float64 `json:"avgTokens"`
	ErrorRate    float64 `json:"errorRate"`
}

// OptimizationSuggestion describes a suggested prompt optimization, including
// the technique used and estimated token savings.
type OptimizationSuggestion struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"projectId"`
	PromptName      string    `json:"promptName"`
	Technique       string    `json:"technique"` // compression, restructure, few_shot, caching
	OriginalTokens  int       `json:"originalTokens"`
	OptimizedTokens int       `json:"optimizedTokens"`
	SavingsPercent  float64   `json:"savingsPercent"`
	Description     string    `json:"description"`
	Suggestion      string    `json:"suggestion"`
	Confidence      float64   `json:"confidence"`
	CreatedAt       time.Time `json:"createdAt"`
}

// PromptExperimentInput is the input for creating a new prompt experiment.
type PromptExperimentInput struct {
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description,omitempty"`
	PromptName  string         `json:"promptName" validate:"required"`
	Variants    []PromptVariantInput `json:"variants" validate:"required,min=2"`
}

// PromptVariantInput is the input for defining a variant within a prompt experiment.
type PromptVariantInput struct {
	Name          string  `json:"name" validate:"required"`
	PromptContent string  `json:"promptContent" validate:"required"`
	TrafficWeight float64 `json:"trafficWeight"`
	IsControl     bool    `json:"isControl"`
}
