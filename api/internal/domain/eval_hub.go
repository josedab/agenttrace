package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EvalHubAssetKind identifies the existing domain represented by a package.
type EvalHubAssetKind string

// Eval Hub asset kinds.
const (
	EvalHubDataset    EvalHubAssetKind = "dataset"
	EvalHubEvaluator  EvalHubAssetKind = "evaluator"
	EvalHubPrompt     EvalHubAssetKind = "prompt"
	EvalHubExperiment EvalHubAssetKind = "experiment"
	EvalHubBenchmark  EvalHubAssetKind = "benchmark"
)

// IsValid reports whether an Eval Hub asset kind is supported.
func (k EvalHubAssetKind) IsValid() bool {
	switch k {
	case EvalHubDataset, EvalHubEvaluator, EvalHubPrompt, EvalHubExperiment, EvalHubBenchmark:
		return true
	default:
		return false
	}
}

// EvalHubVisibility defines who may discover and fork a package.
type EvalHubVisibility string

// Eval Hub visibility levels.
const (
	EvalHubVisibilityPrivate      EvalHubVisibility = "private"
	EvalHubVisibilityOrganization EvalHubVisibility = "organization"
	EvalHubVisibilityPublic       EvalHubVisibility = "public"
)

// IsValid reports whether an Eval Hub visibility is supported.
func (v EvalHubVisibility) IsValid() bool {
	switch v {
	case EvalHubVisibilityPrivate, EvalHubVisibilityOrganization, EvalHubVisibilityPublic:
		return true
	default:
		return false
	}
}

// EvalHubPackage is the canonical publishable Eval Hub asset.
type EvalHubPackage struct {
	ID                  uuid.UUID         `json:"id"`
	OwnerProjectID      uuid.UUID         `json:"ownerProjectId"`
	OrganizationID      uuid.UUID         `json:"organizationId"`
	Kind                EvalHubAssetKind  `json:"kind"`
	Name                string            `json:"name"`
	Description         string            `json:"description,omitempty"`
	Visibility          EvalHubVisibility `json:"visibility"`
	LatestVersion       int               `json:"latestVersion"`
	ForkedFromPackageID *uuid.UUID        `json:"forkedFromPackageId,omitempty"`
	ForkedFromVersion   *int              `json:"forkedFromVersion,omitempty"`
	PublishedBy         uuid.UUID         `json:"publishedBy"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
	Version             *EvalHubVersion   `json:"version,omitempty"`
}

// EvalHubVersion is an immutable package manifest with provenance.
type EvalHubVersion struct {
	ID               uuid.UUID       `json:"id"`
	PackageID        uuid.UUID       `json:"packageId"`
	Version          int             `json:"version"`
	SourceResourceID uuid.UUID       `json:"sourceResourceId"`
	Manifest         json.RawMessage `json:"manifest"`
	Checksum         string          `json:"checksum"`
	VersionNote      string          `json:"versionNote,omitempty"`
	CreatedBy        uuid.UUID       `json:"createdBy"`
	CreatedAt        time.Time       `json:"createdAt"`
}

// EvalHubPublishInput publishes a new package or an immutable version.
type EvalHubPublishInput struct {
	PackageID        *uuid.UUID        `json:"packageId,omitempty"`
	Kind             EvalHubAssetKind  `json:"kind"`
	SourceResourceID uuid.UUID         `json:"sourceResourceId"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
	Visibility       EvalHubVisibility `json:"visibility"`
	VersionNote      string            `json:"versionNote,omitempty"`
}

// EvalHubForkInput controls the target package metadata.
type EvalHubForkInput struct {
	Name       string            `json:"name,omitempty"`
	Visibility EvalHubVisibility `json:"visibility,omitempty"`
}

// EvalHubRunInput starts a project-scoped package execution.
type EvalHubRunInput struct {
	Name           string            `json:"name,omitempty"`
	Version        *int              `json:"version,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
}

// EvalHubRunStatus represents execution capability and state.
type EvalHubRunStatus string

// Eval Hub run states.
const (
	EvalHubRunReady       EvalHubRunStatus = "ready"
	EvalHubRunRunning     EvalHubRunStatus = "running"
	EvalHubRunCompleted   EvalHubRunStatus = "completed"
	EvalHubRunUnsupported EvalHubRunStatus = "unsupported"
	EvalHubRunFailed      EvalHubRunStatus = "failed"
)

// EvalHubRun records a project-scoped execution or explicit prerequisite state.
type EvalHubRun struct {
	ID                uuid.UUID        `json:"id"`
	ProjectID         uuid.UUID        `json:"projectId"`
	PackageID         uuid.UUID        `json:"packageId"`
	PackageVersion    int              `json:"packageVersion"`
	Status            EvalHubRunStatus `json:"status"`
	DatasetRunID      *uuid.UUID       `json:"datasetRunId,omitempty"`
	ExperimentID      *uuid.UUID       `json:"experimentId,omitempty"`
	Result            json.RawMessage  `json:"result,omitempty"`
	CapabilityMessage string           `json:"capabilityMessage,omitempty"`
	IdempotencyKey    string           `json:"idempotencyKey,omitempty"`
	CreatedBy         uuid.UUID        `json:"createdBy"`
	StartedAt         time.Time        `json:"startedAt"`
	CompletedAt       *time.Time       `json:"completedAt,omitempty"`
}

// EvalHubRunList is a paginated project execution list.
type EvalHubRunList struct {
	Runs       []EvalHubRun `json:"runs"`
	TotalCount int64        `json:"totalCount"`
	HasMore    bool         `json:"hasMore"`
}

// EvalHubPackageFilter controls accessible package discovery.
type EvalHubPackageFilter struct {
	RequesterProjectID uuid.UUID
	OrganizationID     uuid.UUID
	Kind               *EvalHubAssetKind
	Visibility         *EvalHubVisibility
	Query              string
	Limit              int
	Offset             int
}

// EvalHubPackageList is a paginated accessible package result.
type EvalHubPackageList struct {
	Packages   []EvalHubPackage `json:"packages"`
	TotalCount int64            `json:"totalCount"`
	HasMore    bool             `json:"hasMore"`
}
