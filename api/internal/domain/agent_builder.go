package domain

import (
	"time"

	"github.com/google/uuid"
)

// BlueprintStatus represents the status of an agent blueprint
type BlueprintStatus string

const (
	BlueprintStatusGenerating BlueprintStatus = "generating"
	BlueprintStatusReady      BlueprintStatus = "ready"
	BlueprintStatusDeployed   BlueprintStatus = "deployed"
)

// BuilderComplexity represents task complexity levels
type BuilderComplexity string

const (
	ComplexitySimple   BuilderComplexity = "simple"
	ComplexityModerate BuilderComplexity = "moderate"
	ComplexityComplex  BuilderComplexity = "complex"
)

// AgentBlueprint represents a generated agent configuration blueprint
type AgentBlueprint struct {
	ID               uuid.UUID       `json:"id"`
	ProjectID        uuid.UUID       `json:"projectId"`
	TaskDescription  string          `json:"taskDescription"`
	GeneratedConfig  BlueprintConfig `json:"generatedConfig"`
	EstimatedCost    float64         `json:"estimatedCost"`
	EstimatedLatency float64         `json:"estimatedLatency"`
	EstimatedQuality float64         `json:"estimatedQuality"`
	Status           BlueprintStatus `json:"status"`
	CreatedAt        time.Time       `json:"createdAt"`
}

// BlueprintConfig represents the generated agent configuration for a blueprint
type BlueprintConfig struct {
	Model       string            `json:"model"`
	MaxTokens   int               `json:"maxTokens"`
	Temperature float64           `json:"temperature"`
	Tools       []string          `json:"tools"`
	SystemPrompt string           `json:"systemPrompt"`
	Parameters  map[string]any    `json:"parameters,omitempty"`
}

// BlueprintSuggestion represents a suggestion for improving a blueprint
type BlueprintSuggestion struct {
	Field          string  `json:"field"`
	CurrentValue   string  `json:"currentValue"`
	SuggestedValue string  `json:"suggestedValue"`
	Reason         string  `json:"reason"`
	Confidence     float64 `json:"confidence"`
}

// BuilderInput represents input for generating a blueprint
type BuilderInput struct {
	TaskDescription string            `json:"taskDescription" validate:"required"`
	TargetLanguage  string            `json:"targetLanguage,omitempty"`
	Complexity      BuilderComplexity `json:"complexity,omitempty"`
	Constraints     map[string]any    `json:"constraints,omitempty"`
}
