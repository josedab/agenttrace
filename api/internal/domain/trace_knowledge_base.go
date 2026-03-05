package domain

import (
	"time"

	"github.com/google/uuid"
)

// KBEntry represents a knowledge base entry for trace annotations
type KBEntry struct {
	ID            uuid.UUID        `json:"id"`
	ProjectID     uuid.UUID        `json:"projectId"`
	TraceID       string           `json:"traceId"`
	ObservationID *string          `json:"observationId,omitempty"`
	Category      string           `json:"category"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	Tags          []string         `json:"tags"`
	RootCause     *RootCauseDetail `json:"rootCause,omitempty"`
	Pattern       *PatternDetail   `json:"pattern,omitempty"`
	Fix           *FixDetail       `json:"fix,omitempty"`
	CreatedBy     uuid.UUID        `json:"createdBy"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
}

// RootCauseDetail represents details about a root cause analysis
type RootCauseDetail struct {
	Category          string `json:"category"`
	Severity          string `json:"severity"`
	AffectedComponent string `json:"affectedComponent"`
	Description       string `json:"description"`
}

// PatternDetail represents details about a recognized pattern
type PatternDetail struct {
	Frequency       int       `json:"frequency"`
	FirstSeen       time.Time `json:"firstSeen"`
	LastSeen        time.Time `json:"lastSeen"`
	ExampleTraceIDs []string  `json:"exampleTraceIds"`
}

// FixDetail represents details about a fix or remediation
type FixDetail struct {
	Solution string `json:"solution"`
	Effort   string `json:"effort"`
	Impact   string `json:"impact"`
}

// KBSearchResult represents the result of a knowledge base search
type KBSearchResult struct {
	Entries    []KBEntry `json:"entries"`
	TotalCount int64     `json:"totalCount"`
	Query      string    `json:"query"`
}

// KBSuggestion represents a knowledge base suggestion for a trace
type KBSuggestion struct {
	EntryID   uuid.UUID `json:"entryId"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Relevance float64   `json:"relevance"`
	Reason    string    `json:"reason"`
}

// KBEntryInput represents input for creating or updating a knowledge base entry
type KBEntryInput struct {
	TraceID       string           `json:"traceId"`
	ObservationID *string          `json:"observationId,omitempty"`
	Category      string           `json:"category" validate:"required"`
	Title         string           `json:"title" validate:"required"`
	Description   string           `json:"description" validate:"required"`
	Tags          []string         `json:"tags,omitempty"`
	RootCause     *RootCauseDetail `json:"rootCause,omitempty"`
	Pattern       *PatternDetail   `json:"pattern,omitempty"`
	Fix           *FixDetail       `json:"fix,omitempty"`
}

// KBSearchInput represents input for searching the knowledge base
type KBSearchInput struct {
	Query    string   `json:"query"`
	Category *string  `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
}
