package domain

import "time"

// SemanticSearchQuery represents a semantic search request
type SemanticSearchQuery struct {
	Query         string        `json:"query"`
	Filters       SearchFilters `json:"filters"`
	Limit         int           `json:"limit"`
	UseEmbeddings bool          `json:"useEmbeddings"`
}

// SemanticSearchResult represents the result of a semantic search
type SemanticSearchResult struct {
	Results     []SearchHit `json:"results"`
	TotalCount  int         `json:"totalCount"`
	QueryTimeMs int64       `json:"queryTimeMs"`
}

// SearchHit represents a single search result
type SearchHit struct {
	TraceID   string    `json:"traceId"`
	TraceName string    `json:"traceName"`
	Score     float64   `json:"score"`
	MatchType string    `json:"matchType"` // "semantic" or "keyword"
	Highlight string    `json:"highlight"`
	Timestamp time.Time `json:"timestamp"`
	Cost      float64   `json:"cost"`
	Model     string    `json:"model"`
}

// SearchFilters represents filters for semantic search
type SearchFilters struct {
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
	MinCost   *float64   `json:"minCost,omitempty"`
	MaxCost   *float64   `json:"maxCost,omitempty"`
	Models    []string   `json:"models,omitempty"`
	Levels    []string   `json:"levels,omitempty"`
}
