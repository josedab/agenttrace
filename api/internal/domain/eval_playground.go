package domain

import (
	"time"

	"github.com/google/uuid"
)

// PlaygroundSessionStatus represents the state of a playground session
type PlaygroundSessionStatus string

const (
	PlaygroundStatusDraft   PlaygroundSessionStatus = "draft"
	PlaygroundStatusRunning PlaygroundSessionStatus = "running"
	PlaygroundStatusDone    PlaygroundSessionStatus = "done"
	PlaygroundStatusError   PlaygroundSessionStatus = "error"
)

// PlaygroundSession represents a browser-based eval playground session
type PlaygroundSession struct {
	ID          uuid.UUID               `json:"id"`
	ProjectID   uuid.UUID               `json:"projectId"`
	UserID      uuid.UUID               `json:"userId"`
	Name        string                  `json:"name"`
	Status      PlaygroundSessionStatus `json:"status"`
	Code        string                  `json:"code"`
	Language    string                  `json:"language"` // javascript, python
	TraceIDs    []string                `json:"traceIds,omitempty"`
	Results     []PlaygroundResult      `json:"results,omitempty"`
	ShareToken  string                  `json:"shareToken,omitempty"`
	IsShared    bool                    `json:"isShared"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
}

// PlaygroundResult represents the output of running an evaluator function
type PlaygroundResult struct {
	TraceID     string         `json:"traceId"`
	Score       *float64       `json:"score,omitempty"`
	Label       string         `json:"label,omitempty"`
	Reasoning   string         `json:"reasoning,omitempty"`
	Error       string         `json:"error,omitempty"`
	DurationMs  int64          `json:"durationMs"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	ExecutedAt  time.Time      `json:"executedAt"`
}

// PlaygroundExecuteInput represents input for executing evaluator code
type PlaygroundExecuteInput struct {
	Code     string   `json:"code" validate:"required"`
	Language string   `json:"language" validate:"required,oneof=javascript python"`
	TraceIDs []string `json:"traceIds" validate:"required,min=1,max=50"`
	Timeout  int      `json:"timeout,omitempty"` // seconds, default 30
}

// PlaygroundCreateInput represents input for creating a session
type PlaygroundCreateInput struct {
	Name     string   `json:"name"`
	Code     string   `json:"code"`
	Language string   `json:"language"`
	TraceIDs []string `json:"traceIds,omitempty"`
}

// PlaygroundShareInput represents input for sharing a session
type PlaygroundShareInput struct {
	SessionID uuid.UUID `json:"sessionId" validate:"required"`
}

// PlaygroundTemplate provides a pre-built evaluator template
type PlaygroundTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Category    string `json:"category"`
}
