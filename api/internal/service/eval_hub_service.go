package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// EvalHubRepository defines package, version, and run persistence.
type EvalHubRepository interface {
	SavePackageVersion(
		ctx context.Context,
		pkg *domain.EvalHubPackage,
		version *domain.EvalHubVersion,
		createPackage bool,
	) error
	GetOwnedPackage(
		ctx context.Context,
		projectID, packageID uuid.UUID,
	) (*domain.EvalHubPackage, error)
	GetAccessiblePackage(
		ctx context.Context,
		packageID, requesterProjectID, organizationID uuid.UUID,
	) (*domain.EvalHubPackage, error)
	GetVersion(
		ctx context.Context,
		packageID uuid.UUID,
		version int,
	) (*domain.EvalHubVersion, error)
	ListAccessiblePackages(
		ctx context.Context,
		filter domain.EvalHubPackageFilter,
	) (*domain.EvalHubPackageList, error)
	// CreateRun persists a durable run and returns a conflict when the
	// (projectId, idempotencyKey) pair already exists.
	CreateRun(ctx context.Context, run *domain.EvalHubRun) error
	// UpdateRun persists the outcome of a durable run within its project.
	UpdateRun(ctx context.Context, run *domain.EvalHubRun) error
	ListRuns(
		ctx context.Context,
		projectID uuid.UUID,
		limit, offset int,
	) (*domain.EvalHubRunList, error)
	GetRunByID(
		ctx context.Context,
		projectID, runID uuid.UUID,
	) (*domain.EvalHubRun, error)
	GetRunByIdempotencyKey(
		ctx context.Context,
		projectID uuid.UUID,
		key string,
	) (*domain.EvalHubRun, error)
}

// EvalHubProjectReader resolves organization visibility boundaries.
type EvalHubProjectReader interface {
	Get(ctx context.Context, id uuid.UUID) (*domain.Project, error)
}

// EvalHubService owns publish, version, fork, discovery, and run use cases.
type EvalHubService struct {
	repository EvalHubRepository
	projects   EvalHubProjectReader
	assets     EvalHubAssetManager
	clock      func() time.Time
}

// NewEvalHubService creates the canonical Eval Hub service.
func NewEvalHubService(
	repository EvalHubRepository,
	projects EvalHubProjectReader,
	assets EvalHubAssetManager,
) *EvalHubService {
	return &EvalHubService{
		repository: repository,
		projects:   projects,
		assets:     assets,
		clock:      time.Now,
	}
}

// Publish creates a package or an immutable next version.
func (s *EvalHubService) Publish(
	ctx context.Context,
	projectID, userID uuid.UUID,
	input domain.EvalHubPublishInput,
) (*domain.EvalHubPackage, error) {
	if !input.Kind.IsValid() {
		return nil, apperrors.Validation("invalid Eval Hub asset kind")
	}
	if input.SourceResourceID == uuid.Nil {
		return nil, apperrors.Validation("sourceResourceId is required")
	}
	if input.Visibility == "" {
		input.Visibility = domain.EvalHubVisibilityPrivate
	}
	if !input.Visibility.IsValid() {
		return nil, apperrors.Validation("invalid Eval Hub visibility")
	}

	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	snapshot, err := s.assets.Snapshot(ctx, projectID, input.Kind, input.SourceResourceID)
	if err != nil {
		return nil, err
	}

	now := s.clock().UTC()
	createPackage := input.PackageID == nil
	var pkg *domain.EvalHubPackage
	versionNumber := 1
	if createPackage {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			name = snapshot.Name
		}
		pkg = &domain.EvalHubPackage{
			ID:             uuid.New(),
			OwnerProjectID: projectID,
			OrganizationID: project.OrganizationID,
			Kind:           input.Kind,
			Name:           name,
			Description:    firstNonEmpty(input.Description, snapshot.Description),
			Visibility:     input.Visibility,
			LatestVersion:  1,
			PublishedBy:    userID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	} else {
		pkg, err = s.repository.GetOwnedPackage(ctx, projectID, *input.PackageID)
		if err != nil {
			return nil, err
		}
		if pkg.Kind != input.Kind {
			return nil, apperrors.Validation("package kind cannot change between versions")
		}
		versionNumber = pkg.LatestVersion + 1
		pkg.LatestVersion = versionNumber
		pkg.Visibility = input.Visibility
		if strings.TrimSpace(input.Name) != "" {
			pkg.Name = strings.TrimSpace(input.Name)
		}
		if input.Description != "" {
			pkg.Description = input.Description
		}
		pkg.UpdatedAt = now
	}

	version := newEvalHubVersion(
		pkg.ID,
		versionNumber,
		input.SourceResourceID,
		snapshot.Manifest,
		input.VersionNote,
		userID,
		now,
	)
	if err := s.repository.SavePackageVersion(ctx, pkg, version, createPackage); err != nil {
		return nil, err
	}
	pkg.Version = version
	return pkg, nil
}

// Fork creates a private-by-default package with materialized project-owned resources.
func (s *EvalHubService) Fork(
	ctx context.Context,
	projectID, userID, packageID uuid.UUID,
	input domain.EvalHubForkInput,
) (*domain.EvalHubPackage, error) {
	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	source, err := s.repository.GetAccessiblePackage(
		ctx,
		packageID,
		projectID,
		project.OrganizationID,
	)
	if err != nil {
		return nil, err
	}
	if source.Version == nil {
		return nil, apperrors.NotFound("eval hub package version")
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = domain.EvalHubVisibilityPrivate
	}
	if !visibility.IsValid() {
		return nil, apperrors.Validation("invalid Eval Hub visibility")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = source.Name + " fork"
	}

	resourceID, manifest, err := s.assets.Fork(
		ctx,
		projectID,
		userID,
		source.Kind,
		name,
		source.Version.Manifest,
	)
	if err != nil {
		return nil, err
	}

	now := s.clock().UTC()
	sourceVersion := source.LatestVersion
	pkg := &domain.EvalHubPackage{
		ID:                  uuid.New(),
		OwnerProjectID:      projectID,
		OrganizationID:      project.OrganizationID,
		Kind:                source.Kind,
		Name:                name,
		Description:         source.Description,
		Visibility:          visibility,
		LatestVersion:       1,
		ForkedFromPackageID: &source.ID,
		ForkedFromVersion:   &sourceVersion,
		PublishedBy:         userID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	version := newEvalHubVersion(
		pkg.ID,
		1,
		resourceID,
		manifest,
		"Forked from "+source.ID.String(),
		userID,
		now,
	)
	if err := s.repository.SavePackageVersion(ctx, pkg, version, true); err != nil {
		return nil, err
	}
	pkg.Version = version
	return pkg, nil
}

// Run starts a project-scoped execution with idempotency.
// The run row is persisted before any experiment or dataset side effect, so a
// crash mid-execution leaves an honest running record instead of silent work.
func (s *EvalHubService) Run(
	ctx context.Context,
	projectID, userID, packageID uuid.UUID,
	input domain.EvalHubRunInput,
) (*domain.EvalHubRun, error) {
	if len(input.IdempotencyKey) > 200 {
		return nil, apperrors.Validation("idempotencyKey cannot exceed 200 characters")
	}
	if input.IdempotencyKey != "" {
		existing, err := s.repository.GetRunByIdempotencyKey(
			ctx,
			projectID,
			input.IdempotencyKey,
		)
		if err == nil {
			return existing, nil
		}
		if !apperrors.IsNotFound(err) {
			return nil, err
		}
	}

	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	pkg, err := s.repository.GetAccessiblePackage(
		ctx,
		packageID,
		projectID,
		project.OrganizationID,
	)
	if err != nil {
		return nil, err
	}

	version := pkg.Version
	if input.Version != nil && *input.Version != pkg.LatestVersion {
		version, err = s.repository.GetVersion(ctx, pkg.ID, *input.Version)
		if err != nil {
			return nil, err
		}
	}
	if version == nil {
		return nil, apperrors.NotFound("eval hub package version")
	}

	now := s.clock().UTC()
	run := &domain.EvalHubRun{
		ID:             uuid.New(),
		ProjectID:      projectID,
		PackageID:      pkg.ID,
		PackageVersion: version.Version,
		Status:         domain.EvalHubRunReady,
		IdempotencyKey: input.IdempotencyKey,
		CreatedBy:      userID,
		StartedAt:      now,
	}
	if pkg.OwnerProjectID != projectID {
		run.Status = domain.EvalHubRunUnsupported
		run.CapabilityMessage = "Fork this package into the project before running it."
		run.CompletedAt = &now
	}

	existing, err := s.persistRun(ctx, run)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	if run.Status == domain.EvalHubRunUnsupported {
		return run, nil
	}

	outcome, err := s.assets.Run(
		ctx,
		projectID,
		userID,
		pkg.Kind,
		version.SourceResourceID,
		version.Manifest,
		input,
	)
	if err != nil {
		return nil, s.failRun(ctx, run, err)
	}

	run.Status = outcome.Status
	run.DatasetRunID = outcome.DatasetRunID
	run.ExperimentID = outcome.ExperimentID
	run.Result = outcome.Result
	run.CapabilityMessage = outcome.CapabilityMessage
	run.CompletedAt = outcome.CompletedAt
	if err := s.repository.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

// persistRun stores the durable run. When a concurrent request already won the
// (projectId, idempotencyKey) race, the run it created is returned instead so
// the same execution is never started twice.
func (s *EvalHubService) persistRun(
	ctx context.Context,
	run *domain.EvalHubRun,
) (*domain.EvalHubRun, error) {
	err := s.repository.CreateRun(ctx, run)
	if err == nil {
		return nil, nil
	}
	if !apperrors.IsConflict(err) || run.IdempotencyKey == "" {
		return nil, err
	}
	existing, readErr := s.repository.GetRunByIdempotencyKey(
		ctx,
		run.ProjectID,
		run.IdempotencyKey,
	)
	if readErr != nil {
		return nil, err
	}
	return existing, nil
}

// failRun records a failed execution so a caller never sees a missing run for
// work that was already attempted.
func (s *EvalHubService) failRun(
	ctx context.Context,
	run *domain.EvalHubRun,
	cause error,
) error {
	completedAt := s.clock().UTC()
	run.Status = domain.EvalHubRunFailed
	run.CompletedAt = &completedAt
	run.CapabilityMessage = safeEvalHubFailure(cause)
	if updateErr := s.repository.UpdateRun(ctx, run); updateErr != nil {
		return fmt.Errorf("eval hub run failed: %v; persist failure state: %w", cause, updateErr)
	}
	return cause
}

// safeEvalHubFailure keeps internal error details out of stored run records.
func safeEvalHubFailure(err error) string {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		return appErr.Message
	}
	return "Run failed; see server logs for details."
}

// GetPackage retrieves an accessible package with its latest immutable version.
func (s *EvalHubService) GetPackage(
	ctx context.Context,
	projectID, packageID uuid.UUID,
) (*domain.EvalHubPackage, error) {
	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return s.repository.GetAccessiblePackage(
		ctx,
		packageID,
		projectID,
		project.OrganizationID,
	)
}

// ListPackages lists packages visible to the current project.
func (s *EvalHubService) ListPackages(
	ctx context.Context,
	projectID uuid.UUID,
	filter domain.EvalHubPackageFilter,
) (*domain.EvalHubPackageList, error) {
	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}
	filter.RequesterProjectID = projectID
	filter.OrganizationID = project.OrganizationID
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repository.ListAccessiblePackages(ctx, filter)
}

// GetRun retrieves an execution only within its project.
func (s *EvalHubService) GetRun(
	ctx context.Context,
	projectID, runID uuid.UUID,
) (*domain.EvalHubRun, error) {
	return s.repository.GetRunByID(ctx, projectID, runID)
}

// ListRuns lists project-scoped executions.
func (s *EvalHubService) ListRuns(
	ctx context.Context,
	projectID uuid.UUID,
	limit, offset int,
) (*domain.EvalHubRunList, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repository.ListRuns(ctx, projectID, limit, offset)
}

func newEvalHubVersion(
	packageID uuid.UUID,
	version int,
	sourceResourceID uuid.UUID,
	manifest json.RawMessage,
	note string,
	userID uuid.UUID,
	createdAt time.Time,
) *domain.EvalHubVersion {
	checksum := sha256.Sum256(manifest)
	return &domain.EvalHubVersion{
		ID:               uuid.New(),
		PackageID:        packageID,
		Version:          version,
		SourceResourceID: sourceResourceID,
		Manifest:         manifest,
		Checksum:         hex.EncodeToString(checksum[:]),
		VersionNote:      note,
		CreatedBy:        userID,
		CreatedAt:        createdAt,
	}
}

func firstNonEmpty(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}
