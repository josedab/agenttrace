package domain

import (
	"time"

	"github.com/google/uuid"
)

// CodeImpactMap represents a trace-linked code impact map
type CodeImpactMap struct {
	ID        uuid.UUID         `json:"id"`
	ProjectID uuid.UUID         `json:"projectId"`
	TraceID   string            `json:"traceId"`
	SessionID string            `json:"sessionId"`
	RepoURL   string            `json:"repoUrl"`
	Branch    string            `json:"branch"`
	Files     []FileImpact      `json:"files"`
	Summary   CodeImpactSummary `json:"summary"`
	CreatedAt time.Time         `json:"createdAt"`
}

// FileImpact represents the impact on a single file
type FileImpact struct {
	FilePath       string   `json:"filePath"`
	OperationType  string   `json:"operationType"`  // created, modified, deleted
	LinesAdded     int      `json:"linesAdded"`
	LinesRemoved   int      `json:"linesRemoved"`
	DiffURL        string   `json:"diffUrl"`
	ObservationIDs []string `json:"observationIds"`
	Language       string   `json:"language"`
	Complexity     string   `json:"complexity"` // low, medium, high
}

// CodeImpactSummary represents aggregated impact metrics
type CodeImpactSummary struct {
	TotalFiles       int            `json:"totalFiles"`
	FilesCreated     int            `json:"filesCreated"`
	FilesModified    int            `json:"filesModified"`
	FilesDeleted     int            `json:"filesDeleted"`
	TotalLinesAdded  int            `json:"totalLinesAdded"`
	TotalLinesRemoved int           `json:"totalLinesRemoved"`
	Languages        map[string]int `json:"languages"`
	MostImpactedFile string         `json:"mostImpactedFile"`
}

// CodeImpactFilter represents filter options for querying code impact maps
type CodeImpactFilter struct {
	TraceID   *string `json:"traceId,omitempty"`
	SessionID *string `json:"sessionId,omitempty"`
	RepoURL   *string `json:"repoUrl,omitempty"`
	Branch    *string `json:"branch,omitempty"`
}
