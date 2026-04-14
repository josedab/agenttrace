package domain

import (
	"time"

	"github.com/google/uuid"
)

// GitHubOutcomeReportInput selects an optional GitHub delivery target.
type GitHubOutcomeReportInput struct {
	RepositoryID      uuid.UUID `json:"repositoryId"`
	CommitSHA         string    `json:"commitSha,omitempty"`
	PullRequestNumber int       `json:"pullRequestNumber,omitempty"`
	Window            string    `json:"window,omitempty"`
}

// GitHubOutcomeReportResult confirms an optional delivery.
type GitHubOutcomeReportResult struct {
	RepositoryID uuid.UUID `json:"repositoryId"`
	Target       string    `json:"target"`
	DeliveredAt  time.Time `json:"deliveredAt"`
}
