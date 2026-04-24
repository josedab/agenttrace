package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// evalHubRepositoryStub mirrors the PostgreSQL repository: run creation enforces
// the unique (projectId, idempotencyKey) constraint and every read and write is
// project scoped.
type evalHubRepositoryStub struct {
	mu        sync.Mutex
	packages  map[uuid.UUID]*domain.EvalHubPackage
	versions  map[uuid.UUID]map[int]*domain.EvalHubVersion
	runs      map[uuid.UUID]*domain.EvalHubRun
	runKeys   map[string]uuid.UUID
	updateErr error
}

func newEvalHubRepositoryStub() *evalHubRepositoryStub {
	return &evalHubRepositoryStub{
		packages: map[uuid.UUID]*domain.EvalHubPackage{},
		versions: map[uuid.UUID]map[int]*domain.EvalHubVersion{},
		runs:     map[uuid.UUID]*domain.EvalHubRun{},
		runKeys:  map[string]uuid.UUID{},
	}
}

func (r *evalHubRepositoryStub) SavePackageVersion(
	_ context.Context,
	pkg *domain.EvalHubPackage,
	version *domain.EvalHubVersion,
	_ bool,
) error {
	pkgCopy := *pkg
	versionCopy := *version
	pkgCopy.Version = &versionCopy
	r.packages[pkg.ID] = &pkgCopy
	if r.versions[pkg.ID] == nil {
		r.versions[pkg.ID] = map[int]*domain.EvalHubVersion{}
	}
	r.versions[pkg.ID][version.Version] = &versionCopy
	return nil
}

func (r *evalHubRepositoryStub) GetOwnedPackage(
	_ context.Context,
	projectID, packageID uuid.UUID,
) (*domain.EvalHubPackage, error) {
	pkg, ok := r.packages[packageID]
	if !ok || pkg.OwnerProjectID != projectID {
		return nil, apperrors.NotFound("eval hub package")
	}
	copy := *pkg
	return &copy, nil
}

func (r *evalHubRepositoryStub) GetAccessiblePackage(
	_ context.Context,
	packageID, requesterProjectID, organizationID uuid.UUID,
) (*domain.EvalHubPackage, error) {
	pkg, ok := r.packages[packageID]
	if !ok {
		return nil, apperrors.NotFound("eval hub package")
	}
	accessible := pkg.OwnerProjectID == requesterProjectID ||
		pkg.Visibility == domain.EvalHubVisibilityPublic ||
		(pkg.Visibility == domain.EvalHubVisibilityOrganization &&
			pkg.OrganizationID == organizationID)
	if !accessible {
		return nil, apperrors.NotFound("eval hub package")
	}
	copy := *pkg
	return &copy, nil
}

func (r *evalHubRepositoryStub) GetVersion(
	_ context.Context,
	packageID uuid.UUID,
	version int,
) (*domain.EvalHubVersion, error) {
	item, ok := r.versions[packageID][version]
	if !ok {
		return nil, apperrors.NotFound("eval hub package version")
	}
	copy := *item
	return &copy, nil
}

func (r *evalHubRepositoryStub) ListAccessiblePackages(
	_ context.Context,
	filter domain.EvalHubPackageFilter,
) (*domain.EvalHubPackageList, error) {
	packages := []domain.EvalHubPackage{}
	for _, pkg := range r.packages {
		accessible := pkg.OwnerProjectID == filter.RequesterProjectID ||
			pkg.Visibility == domain.EvalHubVisibilityPublic ||
			(pkg.Visibility == domain.EvalHubVisibilityOrganization &&
				pkg.OrganizationID == filter.OrganizationID)
		if accessible && (filter.Kind == nil || pkg.Kind == *filter.Kind) {
			packages = append(packages, *pkg)
		}
	}
	return &domain.EvalHubPackageList{
		Packages:   packages,
		TotalCount: int64(len(packages)),
	}, nil
}

func (r *evalHubRepositoryStub) CreateRun(_ context.Context, run *domain.EvalHubRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := run.ProjectID.String() + ":" + run.IdempotencyKey
	if run.IdempotencyKey != "" {
		if _, exists := r.runKeys[key]; exists {
			return apperrors.Conflict("eval hub run already exists for this idempotency key")
		}
	}
	stored := *run
	r.runs[run.ID] = &stored
	if run.IdempotencyKey != "" {
		r.runKeys[key] = run.ID
	}
	return nil
}

func (r *evalHubRepositoryStub) UpdateRun(_ context.Context, run *domain.EvalHubRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.updateErr != nil {
		return r.updateErr
	}
	existing, ok := r.runs[run.ID]
	if !ok || existing.ProjectID != run.ProjectID {
		return apperrors.NotFound("eval hub run")
	}
	stored := *run
	r.runs[run.ID] = &stored
	return nil
}

func (r *evalHubRepositoryStub) ListRuns(
	_ context.Context,
	projectID uuid.UUID,
	_, _ int,
) (*domain.EvalHubRunList, error) {
	runs := []domain.EvalHubRun{}
	for _, run := range r.runs {
		if run.ProjectID == projectID {
			runs = append(runs, *run)
		}
	}
	return &domain.EvalHubRunList{Runs: runs, TotalCount: int64(len(runs))}, nil
}

func (r *evalHubRepositoryStub) GetRunByID(
	_ context.Context,
	projectID, runID uuid.UUID,
) (*domain.EvalHubRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRun(projectID, runID)
}

func (r *evalHubRepositoryStub) getRun(
	projectID, runID uuid.UUID,
) (*domain.EvalHubRun, error) {
	run, ok := r.runs[runID]
	if !ok || run.ProjectID != projectID {
		return nil, apperrors.NotFound("eval hub run")
	}
	copy := *run
	return &copy, nil
}

func (r *evalHubRepositoryStub) GetRunByIdempotencyKey(
	_ context.Context,
	projectID uuid.UUID,
	key string,
) (*domain.EvalHubRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	runID, ok := r.runKeys[projectID.String()+":"+key]
	if !ok {
		return nil, apperrors.NotFound("eval hub run")
	}
	return r.getRun(projectID, runID)
}

type evalHubProjectReaderStub struct {
	projects map[uuid.UUID]*domain.Project
}

func (r *evalHubProjectReaderStub) Get(
	_ context.Context,
	id uuid.UUID,
) (*domain.Project, error) {
	project, ok := r.projects[id]
	if !ok {
		return nil, apperrors.NotFound("project")
	}
	return project, nil
}

type evalHubAssetManagerStub struct {
	mu        sync.Mutex
	snapshot  *EvalHubAssetSnapshot
	forkID    uuid.UUID
	outcome   *EvalHubAssetRunOutcome
	runCalls  int
	runErr    error
	beforeRun func()
}

func (m *evalHubAssetManagerStub) Snapshot(
	_ context.Context,
	_ uuid.UUID,
	_ domain.EvalHubAssetKind,
	_ uuid.UUID,
) (*EvalHubAssetSnapshot, error) {
	return m.snapshot, nil
}

func (m *evalHubAssetManagerStub) Fork(
	_ context.Context,
	_, _ uuid.UUID,
	_ domain.EvalHubAssetKind,
	_ string,
	manifest json.RawMessage,
) (uuid.UUID, json.RawMessage, error) {
	return m.forkID, manifest, nil
}

func (m *evalHubAssetManagerStub) Run(
	_ context.Context,
	_, _ uuid.UUID,
	_ domain.EvalHubAssetKind,
	_ uuid.UUID,
	_ json.RawMessage,
	_ domain.EvalHubRunInput,
) (*EvalHubAssetRunOutcome, error) {
	m.mu.Lock()
	m.runCalls++
	before := m.beforeRun
	runErr := m.runErr
	outcome := m.outcome
	m.mu.Unlock()

	if before != nil {
		before()
	}
	if runErr != nil {
		return nil, runErr
	}
	return outcome, nil
}

func (m *evalHubAssetManagerStub) calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.runCalls
}

func newEvalHubTestService() (
	*EvalHubService,
	*evalHubRepositoryStub,
	*evalHubAssetManagerStub,
	uuid.UUID,
	uuid.UUID,
) {
	projectID := uuid.New()
	organizationID := uuid.New()
	repository := newEvalHubRepositoryStub()
	assets := &evalHubAssetManagerStub{
		snapshot: &EvalHubAssetSnapshot{
			Name:        "Quality dataset",
			Description: "Real dataset snapshot",
			Manifest:    json.RawMessage(`{"items":2}`),
		},
		forkID: uuid.New(),
		outcome: &EvalHubAssetRunOutcome{
			Status: domain.EvalHubRunReady,
		},
	}
	service := NewEvalHubService(
		repository,
		&evalHubProjectReaderStub{
			projects: map[uuid.UUID]*domain.Project{
				projectID: {
					ID:             projectID,
					OrganizationID: organizationID,
				},
			},
		},
		assets,
	)
	service.clock = func() time.Time {
		return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}
	return service, repository, assets, projectID, organizationID
}

func TestEvalHubPublishAndVersion(t *testing.T) {
	service, _, _, projectID, _ := newEvalHubTestService()
	userID := uuid.New()
	sourceID := uuid.New()

	pkg, err := service.Publish(context.Background(), projectID, userID, domain.EvalHubPublishInput{
		Kind:             domain.EvalHubDataset,
		SourceResourceID: sourceID,
		Visibility:       domain.EvalHubVisibilityOrganization,
		VersionNote:      "Initial release",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, pkg.LatestVersion)
	assert.Equal(t, "Quality dataset", pkg.Name)
	require.NotNil(t, pkg.Version)
	assert.Len(t, pkg.Version.Checksum, 64)

	packageID := pkg.ID
	updated, err := service.Publish(
		context.Background(),
		projectID,
		userID,
		domain.EvalHubPublishInput{
			PackageID:        &packageID,
			Kind:             domain.EvalHubDataset,
			SourceResourceID: sourceID,
			Visibility:       domain.EvalHubVisibilityPublic,
			VersionNote:      "Second release",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 2, updated.LatestVersion)
	assert.Equal(t, domain.EvalHubVisibilityPublic, updated.Visibility)
}

func TestEvalHubVisibilityAndForkProvenance(t *testing.T) {
	service, repository, assets, projectID, organizationID := newEvalHubTestService()
	userID := uuid.New()
	source, err := service.Publish(
		context.Background(),
		projectID,
		userID,
		domain.EvalHubPublishInput{
			Kind:             domain.EvalHubDataset,
			SourceResourceID: uuid.New(),
			Visibility:       domain.EvalHubVisibilityPublic,
		},
	)
	require.NoError(t, err)

	targetProjectID := uuid.New()
	targetOrganizationID := uuid.New()
	projectReader, ok := service.projects.(*evalHubProjectReaderStub)
	require.True(t, ok)
	projectReader.projects[targetProjectID] = &domain.Project{
		ID:             targetProjectID,
		OrganizationID: targetOrganizationID,
	}
	forked, err := service.Fork(
		context.Background(),
		targetProjectID,
		uuid.New(),
		source.ID,
		domain.EvalHubForkInput{},
	)

	require.NoError(t, err)
	assert.Equal(t, targetProjectID, forked.OwnerProjectID)
	assert.Equal(t, domain.EvalHubVisibilityPrivate, forked.Visibility)
	require.NotNil(t, forked.ForkedFromPackageID)
	assert.Equal(t, source.ID, *forked.ForkedFromPackageID)
	require.NotNil(t, forked.Version)
	assert.Equal(t, assets.forkID, forked.Version.SourceResourceID)

	privatePackage := *source
	privatePackage.ID = uuid.New()
	privatePackage.Visibility = domain.EvalHubVisibilityPrivate
	repository.packages[privatePackage.ID] = &privatePackage
	_, err = service.GetPackage(context.Background(), targetProjectID, privatePackage.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.NotEqual(t, organizationID, targetOrganizationID)
}

func TestEvalHubRunRequiresForkAndIsIdempotent(t *testing.T) {
	service, _, assets, projectID, _ := newEvalHubTestService()
	userID := uuid.New()
	source, err := service.Publish(
		context.Background(),
		projectID,
		userID,
		domain.EvalHubPublishInput{
			Kind:             domain.EvalHubDataset,
			SourceResourceID: uuid.New(),
			Visibility:       domain.EvalHubVisibilityPublic,
		},
	)
	require.NoError(t, err)

	run, err := service.Run(
		context.Background(),
		projectID,
		userID,
		source.ID,
		domain.EvalHubRunInput{IdempotencyKey: "run-1"},
	)
	require.NoError(t, err)
	assert.Equal(t, domain.EvalHubRunReady, run.Status)
	assert.Equal(t, 1, assets.runCalls)

	repeated, err := service.Run(
		context.Background(),
		projectID,
		userID,
		source.ID,
		domain.EvalHubRunInput{IdempotencyKey: "run-1"},
	)
	require.NoError(t, err)
	assert.Equal(t, run.ID, repeated.ID)
	assert.Equal(t, 1, assets.runCalls)

	otherProjectID := uuid.New()
	projectReader, ok := service.projects.(*evalHubProjectReaderStub)
	require.True(t, ok)
	projectReader.projects[otherProjectID] = &domain.Project{
		ID:             otherProjectID,
		OrganizationID: uuid.New(),
	}

	unsupported, err := service.Run(
		context.Background(),
		otherProjectID,
		uuid.New(),
		source.ID,
		domain.EvalHubRunInput{},
	)
	require.NoError(t, err)
	assert.Equal(t, domain.EvalHubRunUnsupported, unsupported.Status)
	assert.Contains(t, unsupported.CapabilityMessage, "Fork")
}

func TestEvalHubRunRejectsOversizedIdempotencyKeyBeforeSideEffects(t *testing.T) {
	service, repository, assets, projectID, _ := newEvalHubTestService()

	_, err := service.Run(
		context.Background(),
		projectID,
		uuid.New(),
		uuid.New(),
		domain.EvalHubRunInput{IdempotencyKey: strings.Repeat("x", 201)},
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsValidation(err))
	assert.Empty(t, repository.runs)
	assert.Zero(t, assets.calls())
}

func TestEvalHubRunPersistsDurableRunBeforeExecution(t *testing.T) {
	service, repository, assets, projectID, _ := newEvalHubTestService()
	userID := uuid.New()
	pkg, err := service.Publish(context.Background(), projectID, userID, domain.EvalHubPublishInput{
		Kind:             domain.EvalHubDataset,
		SourceResourceID: uuid.New(),
	})
	require.NoError(t, err)

	var statusDuringExecution domain.EvalHubRunStatus
	var runsDuringExecution int
	assets.beforeRun = func() {
		repository.mu.Lock()
		defer repository.mu.Unlock()
		runsDuringExecution = len(repository.runs)
		for _, run := range repository.runs {
			statusDuringExecution = run.Status
		}
	}
	assets.outcome = &EvalHubAssetRunOutcome{Status: domain.EvalHubRunRunning}

	run, err := service.Run(context.Background(), projectID, userID, pkg.ID, domain.EvalHubRunInput{
		IdempotencyKey: "durable-1",
	})
	require.NoError(t, err)

	// The run row existed with a non-terminal status while the asset executed.
	assert.Equal(t, 1, runsDuringExecution)
	assert.Equal(t, domain.EvalHubRunReady, statusDuringExecution)
	assert.Equal(t, domain.EvalHubRunRunning, run.Status)

	stored, err := service.GetRun(context.Background(), projectID, run.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EvalHubRunRunning, stored.Status)
}

func TestEvalHubRunRecordsFailureWhenExecutionFails(t *testing.T) {
	service, _, assets, projectID, _ := newEvalHubTestService()
	userID := uuid.New()
	pkg, err := service.Publish(context.Background(), projectID, userID, domain.EvalHubPublishInput{
		Kind:             domain.EvalHubExperiment,
		SourceResourceID: uuid.New(),
	})
	require.NoError(t, err)

	assets.runErr = apperrors.Unprocessable("experiment is already completed")

	_, err = service.Run(context.Background(), projectID, userID, pkg.ID, domain.EvalHubRunInput{
		IdempotencyKey: "failing-run",
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsUnprocessable(err))

	// The attempted work is visible instead of silently disappearing.
	stored, err := service.repository.GetRunByIdempotencyKey(
		context.Background(),
		projectID,
		"failing-run",
	)
	require.NoError(t, err)
	assert.Equal(t, domain.EvalHubRunFailed, stored.Status)
	assert.Equal(t, "experiment is already completed", stored.CapabilityMessage)
	assert.NotNil(t, stored.CompletedAt)
}

func TestEvalHubRunExecutesOnceForConcurrentIdempotencyKey(t *testing.T) {
	service, _, assets, projectID, _ := newEvalHubTestService()
	userID := uuid.New()
	pkg, err := service.Publish(context.Background(), projectID, userID, domain.EvalHubPublishInput{
		Kind:             domain.EvalHubDataset,
		SourceResourceID: uuid.New(),
	})
	require.NoError(t, err)

	start := make(chan struct{})
	results := make(chan *domain.EvalHubRun, 2)
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			run, runErr := service.Run(
				context.Background(),
				projectID,
				userID,
				pkg.ID,
				domain.EvalHubRunInput{IdempotencyKey: "concurrent-key"},
			)
			results <- run
			errs <- runErr
		}()
	}
	close(start)

	first, second := <-results, <-results
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	require.NotNil(t, first)
	require.NotNil(t, second)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, 1, assets.calls())
}

func TestEvalHubRejectsBenchmarkPackaging(t *testing.T) {
	manager := NewDefaultEvalHubAssetManager(nil, nil, nil, nil)

	_, err := manager.Snapshot(
		context.Background(),
		uuid.New(),
		domain.EvalHubBenchmark,
		uuid.New(),
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsUnprocessable(err))
	assert.Contains(t, err.Error(), "not project-owned")

	_, _, err = manager.Fork(
		context.Background(),
		uuid.New(),
		uuid.New(),
		domain.EvalHubBenchmark,
		"copy",
		json.RawMessage(`{}`),
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsUnprocessable(err))
}
