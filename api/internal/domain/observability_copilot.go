package domain

import (
	"time"

	"github.com/google/uuid"
)

// CopilotQuery represents a query to the observability copilot
type CopilotQuery struct {
	ID          uuid.UUID          `json:"id"`
	ProjectID   uuid.UUID          `json:"projectId"`
	Question    string             `json:"question"`
	Answer      string             `json:"answer"`
	Sources     []CopilotSource    `json:"sources"`
	Suggestions []CopilotSuggestion `json:"suggestions"`
	QueryTimeMs int64              `json:"queryTimeMs"`
	CreatedAt   time.Time          `json:"createdAt"`
}

// CopilotSource represents a source referenced in a copilot answer
type CopilotSource struct {
	Type      string  `json:"type"` // trace, metric, config, documentation
	Reference string  `json:"reference"`
	Relevance float64 `json:"relevance"`
}

// CopilotSuggestion represents a suggestion from the copilot
type CopilotSuggestion struct {
	Category    string  `json:"category"` // cost, quality, performance, security
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
	Automated   bool    `json:"automated"`
}

// ProactiveInsight represents a proactive insight from the copilot
type ProactiveInsight struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"projectId"`
	Category    string         `json:"category"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Severity    string         `json:"severity"` // info, warning, action
	Data        map[string]any `json:"data"`
	CreatedAt   time.Time      `json:"createdAt"`
}

// CopilotQueryInput represents input for asking the copilot a question
type CopilotQueryInput struct {
	Question string `json:"question"`
}
