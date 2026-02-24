package domain

import (
	"time"

	"github.com/google/uuid"
)

// SearchResultType represents the type of a semantic search result
type SearchResultType string

const (
	SearchResultTypeTrace       SearchResultType = "trace"
	SearchResultTypeObservation SearchResultType = "observation"
	SearchResultTypeGeneration  SearchResultType = "generation"
	SearchResultTypeSession     SearchResultType = "session"
)

// SemanticTraceSearchQuery represents a semantic trace search request
type SemanticTraceSearchQuery struct {
	Query     string                     `json:"query"`
	ProjectID uuid.UUID                  `json:"projectId"`
	Filters   SemanticTraceSearchFilters `json:"filters"`
	Limit     int                        `json:"limit"`
	Offset    int                        `json:"offset"`
}

// SemanticTraceSearchFilters represents filters for semantic trace search
type SemanticTraceSearchFilters struct {
	StartTime  *time.Time `json:"startTime,omitempty"`
	EndTime    *time.Time `json:"endTime,omitempty"`
	MinScore   *float64   `json:"minScore,omitempty"`
	Models     []string   `json:"models,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	TraceNames []string   `json:"traceNames,omitempty"`
}

// SemanticTraceSearchResult represents a single result from semantic trace search
type SemanticTraceSearchResult struct {
	ID         uuid.UUID              `json:"id"`
	Type       SearchResultType       `json:"type"`
	Score      float64                `json:"score"`
	Snippet    string                 `json:"snippet"`
	Highlights []string               `json:"highlights"`
	TraceID    uuid.UUID              `json:"traceId"`
	TraceName  string                 `json:"traceName"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TraceCluster represents a cluster of related traces
type TraceCluster struct {
	ID                     uuid.UUID   `json:"id"`
	Label                  string      `json:"label"`
	Description            string      `json:"description"`
	TraceCount             int         `json:"traceCount"`
	CommonPatterns         []string    `json:"commonPatterns"`
	AvgScore               float64     `json:"avgScore"`
	RepresentativeTraceIDs []uuid.UUID `json:"representativeTraceIds"`
}

// TraceAnomalyPattern represents a detected anomaly pattern across traces
type TraceAnomalyPattern struct {
	ID             uuid.UUID `json:"id"`
	Pattern        string    `json:"pattern"`
	Frequency      int       `json:"frequency"`
	FirstSeen      time.Time `json:"firstSeen"`
	LastSeen       time.Time `json:"lastSeen"`
	AffectedTraces int       `json:"affectedTraces"`
	Severity       string    `json:"severity"`
}

// SemanticTraceSearchDashboard provides a dashboard view for semantic trace search
type SemanticTraceSearchDashboard struct {
	RecentSearches     []string         `json:"recentSearches"`
	TopClusters        []TraceCluster   `json:"topClusters"`
	AnomalyPatterns    []TraceAnomalyPattern `json:"anomalyPatterns"`
	TotalIndexedTraces int64            `json:"totalIndexedTraces"`
}
