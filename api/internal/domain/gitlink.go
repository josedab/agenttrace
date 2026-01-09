package domain

import (
	"time"

	"github.com/google/uuid"
)

// GitLink represents an association between a trace and a git commit
type GitLink struct {
	ID      uuid.UUID   `json:"id" ch:"id"`
	ProjectID uuid.UUID `json:"projectId" ch:"project_id"`
	TraceID   string    `json:"traceId" ch:"trace_id"`

	// Git info
	CommitSha   string `json:"commitSha" ch:"commit_sha"`
	ParentSha   string `json:"parentSha,omitempty" ch:"parent_sha"`
	Branch      string `json:"branch,omitempty" ch:"branch"`
	Tag         string `json:"tag,omitempty" ch:"tag"`
	RepoURL     string `json:"repoUrl,omitempty" ch:"repo_url"`

	// Commit metadata
	CommitMessage     string    `json:"commitMessage,omitempty" ch:"commit_message"`
	CommitAuthor      string    `json:"commitAuthor,omitempty" ch:"commit_author"`
	CommitAuthorEmail string    `json:"commitAuthorEmail,omitempty" ch:"commit_author_email"`
	CommitTimestamp   time.Time `json:"commitTimestamp" ch:"commit_timestamp"`

	// File changes
	FilesAdded        []string `json:"filesAdded,omitempty" ch:"files_added"`
	FilesModified     []string `json:"filesModified,omitempty" ch:"files_modified"`
	FilesDeleted      []string `json:"filesDeleted,omitempty" ch:"files_deleted"`
	FilesChangedCount uint32   `json:"filesChangedCount" ch:"files_changed_count"`
	Additions         uint32   `json:"additions" ch:"additions"`
	Deletions         uint32   `json:"deletions" ch:"deletions"`

	// Link type
	LinkType GitLinkType `json:"linkType" ch:"link_type"`

	// CI context
	CIRunID *uuid.UUID `json:"ciRunId,omitempty" ch:"ci_run_id"`

	CreatedAt time.Time `json:"createdAt" ch:"created_at"`
}

// GitLinkInput represents input for creating a git link
type GitLinkInput struct {
	TraceID   string `json:"traceId" validate:"required"`
	CommitSha string `json:"commitSha" validate:"required"`

	ParentSha string `json:"parentSha,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Tag       string `json:"tag,omitempty"`
	RepoURL   string `json:"repoUrl,omitempty"`

	// Commit metadata
	CommitMessage     *string    `json:"commitMessage,omitempty"`
	CommitAuthor      *string    `json:"commitAuthor,omitempty"`
	CommitAuthorEmail *string    `json:"commitAuthorEmail,omitempty"`
	CommitTimestamp   *time.Time `json:"commitTimestamp,omitempty"`

	// File changes
	FilesAdded    []string `json:"filesAdded,omitempty"`
	FilesModified []string `json:"filesModified,omitempty"`
	FilesDeleted  []string `json:"filesDeleted,omitempty"`
	Additions     *uint32  `json:"additions,omitempty"`
	Deletions     *uint32  `json:"deletions,omitempty"`

	LinkType GitLinkType `json:"linkType,omitempty"`
	CIRunID  *string     `json:"ciRunId,omitempty"`
}

// GitLinkFilter represents filter options for querying git links
type GitLinkFilter struct {
	ProjectID uuid.UUID
	TraceID   *string
	CommitSha *string
	Branch    *string
	RepoURL   *string
	LinkType  *GitLinkType
	CIRunID   *uuid.UUID
	FromTime  *time.Time
	ToTime    *time.Time
}

// GitLinkList represents a paginated list of git links
type GitLinkList struct {
	GitLinks   []GitLink `json:"gitLinks"`
	TotalCount int64     `json:"totalCount"`
	HasMore    bool      `json:"hasMore"`
}

// GitTimeline represents a timeline of git commits with associated traces
type GitTimeline struct {
	Commits []GitTimelineEntry `json:"commits"`
}

// GitTimelineEntry represents a single entry in the git timeline
type GitTimelineEntry struct {
	CommitSha     string    `json:"commitSha"`
	CommitMessage string    `json:"commitMessage"`
	CommitAuthor  string    `json:"commitAuthor"`
	CommitTime    time.Time `json:"commitTime"`
	Branch        string    `json:"branch"`
	TraceCount    int64     `json:"traceCount"`
	TraceIDs      []string  `json:"traceIds"`
}
