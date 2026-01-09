package domain

import (
	"time"

	"github.com/google/uuid"
)

// GitHubInstallation represents a GitHub App installation
type GitHubInstallation struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organizationId" db:"organization_id"`
	InstallationID   int64      `json:"installationId" db:"installation_id"`
	AccountID        int64      `json:"accountId" db:"account_id"`
	AccountLogin     string     `json:"accountLogin" db:"account_login"`
	AccountType      string     `json:"accountType" db:"account_type"` // "User" or "Organization"
	TargetType       string     `json:"targetType" db:"target_type"`
	AppID            int64      `json:"appId" db:"app_id"`
	AppSlug          string     `json:"appSlug" db:"app_slug"`
	RepositorySelection string  `json:"repositorySelection" db:"repository_selection"` // "all" or "selected"
	AccessTokensURL  string     `json:"accessTokensUrl" db:"access_tokens_url"`
	RepositoriesURL  string     `json:"repositoriesUrl" db:"repositories_url"`
	HTMLURL          string     `json:"htmlUrl" db:"html_url"`
	Permissions      JSONMap    `json:"permissions" db:"permissions"`
	Events           []string   `json:"events" db:"events"`
	SuspendedAt      *time.Time `json:"suspendedAt,omitempty" db:"suspended_at"`
	SuspendedBy      *string    `json:"suspendedBy,omitempty" db:"suspended_by"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" db:"updated_at"`
}

// GitHubRepository represents a repository linked to an installation
type GitHubRepository struct {
	ID             uuid.UUID `json:"id" db:"id"`
	InstallationID uuid.UUID `json:"installationId" db:"installation_id"`
	ProjectID      uuid.UUID `json:"projectId" db:"project_id"`
	RepoID         int64     `json:"repoId" db:"repo_id"`
	RepoFullName   string    `json:"repoFullName" db:"repo_full_name"`
	RepoName       string    `json:"repoName" db:"repo_name"`
	Owner          string    `json:"owner" db:"owner"`
	Private        bool      `json:"private" db:"private"`
	DefaultBranch  string    `json:"defaultBranch" db:"default_branch"`
	HTMLURL        string    `json:"htmlUrl" db:"html_url"`
	CloneURL       string    `json:"cloneUrl" db:"clone_url"`
	SyncEnabled    bool      `json:"syncEnabled" db:"sync_enabled"`
	AutoLink       bool      `json:"autoLink" db:"auto_link"` // Auto-link commits to traces
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// GitHubWebhookEvent represents a webhook event from GitHub
type GitHubWebhookEvent struct {
	ID             uuid.UUID `json:"id" db:"id"`
	InstallationID int64     `json:"installationId" db:"installation_id"`
	EventType      string    `json:"eventType" db:"event_type"`
	Action         string    `json:"action,omitempty" db:"action"`
	DeliveryID     string    `json:"deliveryId" db:"delivery_id"`
	Payload        JSONMap   `json:"payload" db:"payload"`
	Processed      bool      `json:"processed" db:"processed"`
	ProcessedAt    *time.Time `json:"processedAt,omitempty" db:"processed_at"`
	Error          *string   `json:"error,omitempty" db:"error"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}

// GitHubPushPayload represents the parsed push event payload
type GitHubPushPayload struct {
	Ref        string          `json:"ref"`
	Before     string          `json:"before"`
	After      string          `json:"after"`
	Created    bool            `json:"created"`
	Deleted    bool            `json:"deleted"`
	Forced     bool            `json:"forced"`
	HeadCommit *GitHubCommit   `json:"head_commit"`
	Commits    []GitHubCommit  `json:"commits"`
	Repository GitHubRepo      `json:"repository"`
	Pusher     GitHubPusher    `json:"pusher"`
}

// GitHubCommit represents a commit in a push event
type GitHubCommit struct {
	ID        string   `json:"id"`
	TreeID    string   `json:"tree_id"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Author    GitHubAuthor `json:"author"`
	Committer GitHubAuthor `json:"committer"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Modified  []string `json:"modified"`
}

// GitHubAuthor represents the author/committer of a commit
type GitHubAuthor struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

// GitHubRepo represents repository info in webhook payloads
type GitHubRepo struct {
	ID            int64  `json:"id"`
	NodeID        string `json:"node_id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	Owner         GitHubOwner `json:"owner"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// GitHubOwner represents repository owner info
type GitHubOwner struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
}

// GitHubPusher represents the pusher in a push event
type GitHubPusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubInstallationPayload represents installation webhook payload
type GitHubInstallationPayload struct {
	Action       string `json:"action"`
	Installation struct {
		ID              int64  `json:"id"`
		Account         GitHubOwner `json:"account"`
		AppID           int64  `json:"app_id"`
		AppSlug         string `json:"app_slug"`
		TargetType      string `json:"target_type"`
		Permissions     map[string]string `json:"permissions"`
		Events          []string `json:"events"`
		AccessTokensURL string `json:"access_tokens_url"`
		RepositoriesURL string `json:"repositories_url"`
		HTMLURL         string `json:"html_url"`
		RepositorySelection string `json:"repository_selection"`
	} `json:"installation"`
	Repositories []struct {
		ID       int64  `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
	} `json:"repositories,omitempty"`
	Sender GitHubOwner `json:"sender"`
}

// GitHubInstallationFilter represents filter options for installations
type GitHubInstallationFilter struct {
	OrganizationID *uuid.UUID
	InstallationID *int64
	AccountLogin   *string
}

// GitHubRepositoryFilter represents filter options for repositories
type GitHubRepositoryFilter struct {
	InstallationID *uuid.UUID
	ProjectID      *uuid.UUID
	RepoFullName   *string
	SyncEnabled    *bool
	AutoLink       *bool
}

// LinkRepoToProjectInput represents input for linking a repo to a project
type LinkRepoToProjectInput struct {
	RepositoryID uuid.UUID `json:"repositoryId" validate:"required"`
	ProjectID    uuid.UUID `json:"projectId" validate:"required"`
	AutoLink     bool      `json:"autoLink"`
	SyncEnabled  bool      `json:"syncEnabled"`
}

// JSONMap is a map[string]any alias for JSON fields
type JSONMap map[string]any
