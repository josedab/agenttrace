package domain

import (
	"time"

	"github.com/google/uuid"
)

// GitOpsPromptSyncStatus represents the sync status
type GitOpsPromptSyncStatus string

const (
	GitOpsSyncPending    GitOpsPromptSyncStatus = "pending"
	GitOpsSyncSyncing    GitOpsPromptSyncStatus = "syncing"
	GitOpsSyncSynced     GitOpsPromptSyncStatus = "synced"
	GitOpsSyncFailed     GitOpsPromptSyncStatus = "failed"
	GitOpsSyncDiverged   GitOpsPromptSyncStatus = "diverged"
)

// GitOpsEnvironment represents a branch-based deployment environment
type GitOpsEnvironment string

const (
	GitOpsEnvProduction  GitOpsEnvironment = "production"
	GitOpsEnvStaging     GitOpsEnvironment = "staging"
	GitOpsEnvDevelopment GitOpsEnvironment = "development"
)

// GitOpsPromptConfig defines a GitOps prompt pipeline configuration
type GitOpsPromptConfig struct {
	ID           uuid.UUID              `json:"id"`
	ProjectID    uuid.UUID              `json:"projectId"`
	Name         string                 `json:"name"`
	RepoURL      string                 `json:"repoUrl"`
	RepoOwner    string                 `json:"repoOwner"`
	RepoName     string                 `json:"repoName"`
	BasePath     string                 `json:"basePath"` // e.g., "prompts/"
	Enabled      bool                   `json:"enabled"`
	SyncStatus   GitOpsPromptSyncStatus `json:"syncStatus"`
	LastSyncAt   *time.Time             `json:"lastSyncAt,omitempty"`
	LastSyncErr  string                 `json:"lastSyncError,omitempty"`
	WebhookSecret string               `json:"-"` // Secret for git webhook

	// Branch → Environment mapping
	BranchMapping []BranchEnvironmentMap `json:"branchMapping"`

	// Promotion settings
	AutoPromote      bool   `json:"autoPromote"` // Auto-promote on merge
	EvalGateEnabled  bool   `json:"evalGateEnabled"` // Run evals before promotion
	EvalGateConfigID *uuid.UUID `json:"evalGateConfigId,omitempty"`

	// ArgoCD/FluxCD annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	CreatedBy uuid.UUID `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BranchEnvironmentMap maps a git branch pattern to an environment
type BranchEnvironmentMap struct {
	BranchPattern string            `json:"branchPattern"` // e.g., "main", "staging", "feature/*"
	Environment   GitOpsEnvironment `json:"environment"`
	Label         string            `json:"label,omitempty"` // Prompt version label to apply
	AutoSync      bool              `json:"autoSync"`
}

// PromptFileSpec defines the YAML/JSON schema for a prompt file in git
type PromptFileSpec struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"` // "agenttrace.io/v1"
	Kind       string            `json:"kind" yaml:"kind"`             // "Prompt"
	Metadata   PromptFileMeta    `json:"metadata" yaml:"metadata"`
	Spec       PromptFileContent `json:"spec" yaml:"spec"`
}

// PromptFileMeta contains prompt file metadata
type PromptFileMeta struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations"`
}

// PromptFileContent contains the prompt content
type PromptFileContent struct {
	Type      string                 `json:"type" yaml:"type"` // "text" or "chat"
	Content   string                 `json:"content,omitempty" yaml:"content"`
	Messages  []PromptFileMessage    `json:"messages,omitempty" yaml:"messages"`
	Config    map[string]interface{} `json:"config,omitempty" yaml:"config"`
	Variables []PromptFileVariable   `json:"variables,omitempty" yaml:"variables"`
}

// PromptFileMessage represents a message in a chat prompt file
type PromptFileMessage struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

// PromptFileVariable defines a prompt variable with metadata
type PromptFileVariable struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description"`
	Required    bool   `json:"required,omitempty" yaml:"required"`
	Default     string `json:"default,omitempty" yaml:"default"`
}

// GitOpsSyncEvent represents a sync event in the reconciliation loop
type GitOpsSyncEvent struct {
	ID         uuid.UUID              `json:"id"`
	ConfigID   uuid.UUID              `json:"configId"`
	ProjectID  uuid.UUID              `json:"projectId"`
	Status     GitOpsPromptSyncStatus `json:"status"`
	Branch     string                 `json:"branch"`
	CommitSHA  string                 `json:"commitSha"`
	Environment GitOpsEnvironment     `json:"environment"`
	Changes    []SyncChange           `json:"changes"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  time.Time              `json:"startedAt"`
	CompletedAt *time.Time            `json:"completedAt,omitempty"`
}

// SyncChange represents a single change during a sync
type SyncChange struct {
	PromptName string `json:"promptName"`
	Action     string `json:"action"` // "created", "updated", "deleted"
	FilePath   string `json:"filePath"`
	OldVersion *int   `json:"oldVersion,omitempty"`
	NewVersion *int   `json:"newVersion,omitempty"`
}

// GitOpsPromptConfigInput for creating/updating configs
type GitOpsPromptConfigInput struct {
	Name            string                `json:"name" validate:"required"`
	RepoURL         string                `json:"repoUrl" validate:"required"`
	BasePath        string                `json:"basePath,omitempty"`
	BranchMapping   []BranchEnvironmentMap `json:"branchMapping" validate:"required,min=1"`
	AutoPromote     *bool                 `json:"autoPromote,omitempty"`
	EvalGateEnabled *bool                 `json:"evalGateEnabled,omitempty"`
	Annotations     map[string]string     `json:"annotations,omitempty"`
}

// PromptFileExample provides a reference YAML prompt file
const PromptFileExample = `apiVersion: agenttrace.io/v1
kind: Prompt
metadata:
  name: code-review-prompt
  description: Prompt for automated code review
  tags:
    - code-review
    - automation
  labels:
    team: platform
    priority: high
spec:
  type: chat
  messages:
    - role: system
      content: |
        You are an expert code reviewer. Analyze the provided code
        and give constructive feedback on quality, security, and performance.
    - role: user
      content: |
        Please review this code change:
        {{code_diff}}

        Context: {{context}}
  config:
    model: gpt-4
    temperature: 0.3
    max_tokens: 2000
  variables:
    - name: code_diff
      description: The code diff to review
      required: true
    - name: context
      description: Additional context about the change
      required: false
      default: "No additional context provided"
`
