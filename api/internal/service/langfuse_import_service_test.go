package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// langfuseImportRepositoryStub mirrors the PostgreSQL ledger: item records are
// upserted by (job, sourceType, sourceID) and reads are project scoped.
type langfuseImportRepositoryStub struct {
	mu           sync.Mutex
	jobs         map[uuid.UUID]*domain.MigrationJob
	items        map[string]string
	recordErrors map[string]error
	recordCalls  int
	conflictSave bool

	// batchLocks provides real per-(projectID,jobID) mutual exclusion, mirroring
	// the PostgreSQL transaction-scoped advisory lock used in production.
	locksMu    sync.Mutex
	batchLocks map[string]*sync.Mutex
	lockErr    error
}

func newLangfuseImportRepositoryStub() *langfuseImportRepositoryStub {
	return &langfuseImportRepositoryStub{
		jobs:         map[uuid.UUID]*domain.MigrationJob{},
		items:        map[string]string{},
		recordErrors: map[string]error{},
		batchLocks:   map[string]*sync.Mutex{},
	}
}

func (r *langfuseImportRepositoryStub) WithBatchLock(
	ctx context.Context,
	projectID, jobID uuid.UUID,
	fn func(ctx context.Context) error,
) error {
	if r.lockErr != nil {
		return r.lockErr
	}
	r.locksMu.Lock()
	key := projectID.String() + ":" + jobID.String()
	lock, ok := r.batchLocks[key]
	if !ok {
		lock = &sync.Mutex{}
		r.batchLocks[key] = lock
	}
	r.locksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()
	return fn(ctx)
}

func (r *langfuseImportRepositoryStub) Save(
	_ context.Context,
	job *domain.MigrationJob,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.jobs[job.ID]; exists {
		return apperrors.Conflict("migration job already exists")
	}
	copy := *job
	r.jobs[job.ID] = &copy
	if r.conflictSave {
		r.conflictSave = false
		return apperrors.Conflict("migration job already exists")
	}
	return nil
}

func (r *langfuseImportRepositoryStub) GetByID(
	_ context.Context,
	projectID, id uuid.UUID,
) (*domain.MigrationJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok || job.ProjectID != projectID {
		return nil, apperrors.NotFound("migration job")
	}
	copy := *job
	copy.Errors = append([]string(nil), job.Errors...)
	return &copy, nil
}

func (r *langfuseImportRepositoryStub) UpdateJob(
	_ context.Context,
	job *domain.MigrationJob,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *job
	copy.Errors = append([]string(nil), job.Errors...)
	r.jobs[job.ID] = &copy
	return nil
}

func (r *langfuseImportRepositoryStub) FindImportedItem(
	_ context.Context,
	projectID, jobID uuid.UUID,
	sourceType, sourceID string,
) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	importedID, ok := r.items[ledgerKey(projectID, jobID, sourceType, sourceID)]
	return importedID, ok, nil
}

func (r *langfuseImportRepositoryStub) RecordItem(
	_ context.Context,
	projectID, jobID uuid.UUID,
	sourceType, sourceID, _, importedID, status, _ string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := ledgerKey(projectID, jobID, sourceType, sourceID)
	r.recordCalls++
	if err, failing := r.recordErrors[sourceType]; failing {
		return err
	}
	if status == "imported" {
		r.items[key] = importedID
	} else {
		delete(r.items, key)
	}
	return nil
}

func (r *langfuseImportRepositoryStub) importedItem(
	projectID, jobID uuid.UUID,
	sourceType, sourceID string,
) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	importedID, ok := r.items[ledgerKey(projectID, jobID, sourceType, sourceID)]
	return importedID, ok
}

func ledgerKey(projectID, jobID uuid.UUID, sourceType, sourceID string) string {
	return projectID.String() + ":" + jobID.String() + ":" + sourceType + ":" + sourceID
}

type langfuseIngestionStub struct {
	mu             sync.Mutex
	traces         int
	observations   int
	traceInput     *domain.TraceInput
	traceIDs       []string
	lastGeneration *domain.GenerationInput

	blockTraceName string
	traceStarted   chan struct{}
	releaseTrace   chan struct{}
	traceStartOnce sync.Once
}

func (s *langfuseIngestionStub) IngestTrace(
	_ context.Context,
	projectID uuid.UUID,
	input *domain.TraceInput,
) (*domain.Trace, error) {
	if s.blockTraceName != "" && input.Name == s.blockTraceName {
		s.traceStartOnce.Do(func() {
			close(s.traceStarted)
		})
		<-s.releaseTrace
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traces++
	s.traceInput = input
	s.traceIDs = append(s.traceIDs, input.ID)
	return &domain.Trace{ID: input.ID, ProjectID: projectID}, nil
}

func (s *langfuseIngestionStub) IngestObservation(
	_ context.Context,
	projectID uuid.UUID,
	input *domain.ObservationInput,
) (*domain.Observation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observations++
	return &domain.Observation{ID: *input.ID, ProjectID: projectID}, nil
}

func (s *langfuseIngestionStub) IngestGeneration(
	_ context.Context,
	projectID uuid.UUID,
	input *domain.GenerationInput,
) (*domain.Observation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observations++
	s.lastGeneration = input
	return &domain.Observation{ID: *input.ID, ProjectID: projectID}, nil
}

func (s *langfuseIngestionStub) traceCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.traces
}

type langfuseScoreStub struct {
	mu       sync.Mutex
	count    int
	err      error
	created  []uuid.UUID
	inputIDs []*uuid.UUID
}

func (s *langfuseScoreStub) Create(
	_ context.Context,
	projectID uuid.UUID,
	input *domain.ScoreInput,
) (*domain.Score, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return nil, s.err
	}
	s.count++
	s.inputIDs = append(s.inputIDs, input.ID)
	// The stub honors the caller-provided identity exactly like the score
	// service does, so tests observe the deterministic import identifier.
	scoreID := uuid.New()
	if input.ID != nil {
		scoreID = *input.ID
	}
	s.created = append(s.created, scoreID)
	return &domain.Score{ID: scoreID, ProjectID: projectID}, nil
}

func (s *langfuseScoreStub) createdIDs() []uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]uuid.UUID(nil), s.created...)
}

type langfusePromptStub struct {
	count    int
	existing *domain.Prompt
}

func (s *langfusePromptStub) GetByName(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) (*domain.Prompt, error) {
	if s.existing != nil {
		return s.existing, nil
	}
	return nil, apperrors.NotFound("prompt")
}

func (s *langfusePromptStub) Create(
	_ context.Context,
	projectID uuid.UUID,
	_ *domain.PromptInput,
	_ uuid.UUID,
) (*domain.Prompt, error) {
	s.count++
	return &domain.Prompt{ID: uuid.New(), ProjectID: projectID}, nil
}

func (s *langfusePromptStub) CreateVersion(
	_ context.Context,
	_ uuid.UUID,
	_ *domain.PromptVersionInput,
	_ uuid.UUID,
) (*domain.PromptVersion, error) {
	s.count++
	return &domain.PromptVersion{ID: uuid.New()}, nil
}

func TestLangfuseImportBatchIsResumableAndIdempotent(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{}
	scores := &langfuseScoreStub{}
	prompts := &langfusePromptStub{}
	service := NewLangfuseImportService(
		repository,
		ingestion,
		scores,
		prompts,
		NewSensitiveDataRedactor(),
	)
	service.clock = func() time.Time {
		return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}

	batch := domain.LangfuseImportBatch{
		JobID:       jobID,
		Fingerprint: "0123456789abcdef0123456789abcdef",
		TotalItems:  4,
		FinalBatch:  true,
		Records: domain.LangfuseExport{
			Traces: []domain.LangfuseTrace{{
				ID:        "trace-source-id",
				Name:      "agent",
				StartTime: "2026-07-25T10:00:00Z",
			}},
			Observations: []domain.LangfuseObservation{{
				ID:        "observation-source-id",
				TraceID:   "trace-source-id",
				Type:      "GENERATION",
				StartTime: "2026-07-25T10:00:01Z",
				Model:     "gpt-4.1",
			}},
			Scores: []domain.LangfuseScore{{
				ID:      "score-source-id",
				TraceID: "trace-source-id",
				Name:    "quality",
				Value:   floatPointer(0.9),
			}},
			Prompts: []domain.LangfusePrompt{{
				ID:      "prompt-source-id",
				Name:    "review",
				Content: "Review {{code}}",
			}},
		},
	}

	job, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batch)
	require.NoError(t, err)
	assert.Equal(t, domain.MigrationStatusCompleted, job.Status)
	assert.Equal(t, int64(4), job.Progress.ProcessedItems)
	assert.Equal(t, int64(1), job.Progress.TracesMigrated)
	assert.Equal(t, int64(1), job.Progress.ScoresMigrated)
	assert.Equal(t, int64(1), job.Progress.PromptsMigrated)
	assert.Equal(t, 1, ingestion.traces)
	assert.Equal(t, 1, ingestion.observations)
	assert.Equal(t, 1, scores.count)
	assert.Equal(t, 1, prompts.count)
	assert.Len(t, ingestion.traceInput.ID, 32)

	repeated, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batch)
	require.NoError(t, err)
	assert.Equal(t, int64(4), repeated.Progress.ProcessedItems)
	assert.Equal(t, 1, ingestion.traces)
	assert.Equal(t, 1, ingestion.observations)
	assert.Equal(t, 1, scores.count)
	assert.Equal(t, 1, prompts.count)
}

func TestLangfuseImportDryRunDoesNotWrite(t *testing.T) {
	projectID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{}
	service := NewLangfuseImportService(
		repository,
		ingestion,
		&langfuseScoreStub{},
		&langfusePromptStub{},
		NewSensitiveDataRedactor(),
	)

	job, err := service.ImportBatch(
		context.Background(),
		projectID,
		uuid.New(),
		domain.LangfuseImportBatch{
			JobID:       uuid.New(),
			Fingerprint: "dryrun-0123456789abcdef",
			DryRun:      true,
			TotalItems:  1,
			FinalBatch:  true,
			Records: domain.LangfuseExport{
				Traces: []domain.LangfuseTrace{{
					ID:        "trace",
					StartTime: "2026-07-25T10:00:00Z",
				}},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, domain.MigrationStatusCompleted, job.Status)
	assert.Zero(t, ingestion.traces)
	assert.Empty(t, repository.items)
}

func TestLangfuseImportRedactsErrorsAndCanRetry(t *testing.T) {
	projectID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	scoreService := &langfuseScoreStub{
		err: errors.New("provider failed for sk-at-supersecret user@example.com"),
	}
	service := NewLangfuseImportService(
		repository,
		&langfuseIngestionStub{},
		scoreService,
		&langfusePromptStub{},
		NewSensitiveDataRedactor(),
	)
	batch := domain.LangfuseImportBatch{
		JobID:       uuid.New(),
		Fingerprint: "errors-0123456789abcdef",
		TotalItems:  1,
		FinalBatch:  true,
		Records: domain.LangfuseExport{
			Scores: []domain.LangfuseScore{{
				ID:      "score-1",
				TraceID: "trace-1",
				Name:    "quality",
				Value:   floatPointer(0.5),
			}},
		},
	}

	job, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batch)
	require.NoError(t, err)
	assert.Equal(t, domain.MigrationStatusFailed, job.Status)
	require.Len(t, job.Errors, 1)
	assert.Contains(t, job.Errors[0], "[REDACTED:api-key]")
	assert.Contains(t, job.Errors[0], "[REDACTED:email]")
	assert.NotContains(t, job.Errors[0], "supersecret")

	scoreService.err = nil
	retried, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batch)
	require.NoError(t, err)
	assert.Equal(t, domain.MigrationStatusCompleted, retried.Status)
	assert.Empty(t, retried.Errors)
	assert.Equal(t, 1, scoreService.count)
}

func TestLangfuseImportPreventsCrossProjectResume(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	service := NewLangfuseImportService(
		repository,
		&langfuseIngestionStub{},
		&langfuseScoreStub{},
		&langfusePromptStub{},
		NewSensitiveDataRedactor(),
	)
	batch := domain.LangfuseImportBatch{
		JobID:       jobID,
		Fingerprint: "project-0123456789abcdef",
		DryRun:      true,
		TotalItems:  1,
		FinalBatch:  true,
		Records: domain.LangfuseExport{
			Traces: []domain.LangfuseTrace{{
				ID:        "trace",
				StartTime: "2026-07-25T10:00:00Z",
			}},
		},
	}
	_, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batch)
	require.NoError(t, err)

	_, err = service.ImportBatch(context.Background(), uuid.New(), uuid.New(), batch)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func floatPointer(value float64) *float64 {
	return &value
}

func newLangfuseTestService(
	repository *langfuseImportRepositoryStub,
	ingestion *langfuseIngestionStub,
	scores *langfuseScoreStub,
	prompts *langfusePromptStub,
) *LangfuseImportService {
	service := NewLangfuseImportService(
		repository,
		ingestion,
		scores,
		prompts,
		NewSensitiveDataRedactor(),
	)
	service.clock = func() time.Time {
		return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	}
	return service
}

func langfuseScoreBatch(jobID uuid.UUID) domain.LangfuseImportBatch {
	return domain.LangfuseImportBatch{
		JobID:       jobID,
		Fingerprint: "0123456789abcdef0123456789abcdef",
		TotalItems:  2,
		Records: domain.LangfuseExport{
			Traces: []domain.LangfuseTrace{{
				ID:        "trace-source-id",
				Name:      "agent",
				StartTime: "2026-07-25T10:00:00Z",
			}},
			Scores: []domain.LangfuseScore{{
				ID:      "score-source-id",
				TraceID: "trace-source-id",
				Name:    "quality",
				Value:   floatPointer(0.9),
			}},
		},
	}
}

func TestLangfuseImportRetriesScoreWithSameIdentityAfterLedgerFailure(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{}
	scores := &langfuseScoreStub{}
	service := newLangfuseTestService(repository, ingestion, scores, &langfusePromptStub{})

	// The score reaches storage but the ledger write fails, which is the window
	// where a non-deterministic import would create a duplicate score.
	repository.recordErrors["score"] = errors.New("ledger unavailable")
	job, err := service.ImportBatch(
		context.Background(),
		projectID,
		uuid.New(),
		langfuseScoreBatch(jobID),
	)
	require.NoError(t, err)
	assert.Len(t, scores.createdIDs(), 1)
	require.NotEmpty(t, job.Errors)
	assert.Contains(t, job.Errors[0], "score/")

	_, recorded := repository.importedItem(projectID, jobID, "score", "score-source-id")
	assert.False(t, recorded, "the failed ledger write must not mark the item imported")

	// Retrying writes the same deterministic row instead of a second score.
	delete(repository.recordErrors, "score")
	retried, err := service.ImportBatch(
		context.Background(),
		projectID,
		uuid.New(),
		langfuseScoreBatch(jobID),
	)
	require.NoError(t, err)

	created := scores.createdIDs()
	require.Len(t, created, 2)
	assert.Equal(t, created[0], created[1], "retried import must reuse the score identity")
	assert.Empty(t, retried.Errors)
	assert.Equal(t, int64(1), retried.Progress.ScoresMigrated)

	importedID, recorded := repository.importedItem(projectID, jobID, "score", "score-source-id")
	assert.True(t, recorded)
	assert.Equal(t, created[0].String(), importedID)

	// A third attempt returns the prior imported item without touching storage.
	third, err := service.ImportBatch(
		context.Background(),
		projectID,
		uuid.New(),
		langfuseScoreBatch(jobID),
	)
	require.NoError(t, err)
	assert.Len(t, scores.createdIDs(), 2)
	assert.Positive(t, third.Progress.SkippedItems)
}

func TestLangfuseImportReimportsLedgerRowWithoutImportedID(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{}
	scores := &langfuseScoreStub{}
	service := newLangfuseTestService(repository, ingestion, scores, &langfusePromptStub{})

	// A ledger row that never captured an identifier is not trusted as complete.
	repository.mu.Lock()
	repository.items[ledgerKey(projectID, jobID, "score", "score-source-id")] = ""
	repository.mu.Unlock()

	_, err := service.ImportBatch(
		context.Background(),
		projectID,
		uuid.New(),
		langfuseScoreBatch(jobID),
	)
	require.NoError(t, err)

	assert.Len(t, scores.createdIDs(), 1)
	importedID, recorded := repository.importedItem(projectID, jobID, "score", "score-source-id")
	assert.True(t, recorded)
	assert.NotEmpty(t, importedID)
}

func TestLangfuseImportConcurrentDuplicateBatchesStayIdempotent(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{}
	scores := &langfuseScoreStub{}
	service := newLangfuseTestService(repository, ingestion, scores, &langfusePromptStub{})

	start := make(chan struct{})
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := service.ImportBatch(
				context.Background(),
				projectID,
				uuid.New(),
				langfuseScoreBatch(jobID),
			)
			errs <- err
		}()
	}
	close(start)
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)

	// Concurrent duplicates may both reach storage, but every write targets the
	// same deterministic identity so no duplicate record can survive.
	created := scores.createdIDs()
	require.NotEmpty(t, created)
	for _, id := range created[1:] {
		assert.Equal(t, created[0], id)
	}
	for _, traceID := range ingestion.traceIDs[1:] {
		assert.Equal(t, ingestion.traceIDs[0], traceID)
	}
	assert.LessOrEqual(t, ingestion.traceCount(), 2)

	importedID, recorded := repository.importedItem(projectID, jobID, "score", "score-source-id")
	assert.True(t, recorded)
	assert.Equal(t, created[0].String(), importedID)
}

func TestLangfuseImportRereadsJobAfterConcurrentCreateConflict(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	repository := newLangfuseImportRepositoryStub()
	repository.conflictSave = true
	scores := &langfuseScoreStub{}
	service := newLangfuseTestService(
		repository,
		&langfuseIngestionStub{},
		scores,
		&langfusePromptStub{},
	)

	job, err := service.ImportBatch(
		context.Background(),
		projectID,
		uuid.New(),
		langfuseScoreBatch(jobID),
	)

	require.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Len(t, scores.createdIDs(), 1)
}

func TestLangfuseImportReusesIdenticalPromptVersion(t *testing.T) {
	projectID := uuid.New()
	existingVersion := domain.PromptVersion{ID: uuid.New(), Content: "Review {{code}}"}
	prompts := &langfusePromptStub{
		existing: &domain.Prompt{
			ID:            uuid.New(),
			ProjectID:     projectID,
			Name:          "review",
			LatestVersion: &existingVersion,
		},
	}
	service := newLangfuseTestService(
		newLangfuseImportRepositoryStub(),
		&langfuseIngestionStub{},
		&langfuseScoreStub{},
		prompts,
	)

	importedID, err := service.importPrompt(
		context.Background(),
		projectID,
		uuid.New(),
		domain.LangfusePrompt{ID: "prompt-source-id", Name: "review", Content: "Review {{code}}"},
		false,
	)

	require.NoError(t, err)
	assert.Equal(t, existingVersion.ID.String(), importedID)
	assert.Zero(t, prompts.count, "an identical version must not be appended again")
}

// TestLangfuseImportConcurrentDistinctBatchesPreserveProgress runs two distinct
// batches for the SAME job concurrently. The transaction-scoped advisory lock
// (modeled here by the stub's per-job mutex) must serialize them so neither
// batch's progress, errors, nor final status is lost to a read-modify-write
// race. Without the lock, the two ImportBatch calls would read the same job,
// each add only their own progress, and the last UpdateJob would clobber the
// other's — leaving TracesMigrated at 1 instead of 2.
func TestLangfuseImportConcurrentDistinctBatchesPreserveProgress(t *testing.T) {
	projectID := uuid.New()
	jobID := uuid.New()
	fingerprint := "0123456789abcdef0123456789abcdef"
	repository := newLangfuseImportRepositoryStub()
	ingestion := &langfuseIngestionStub{
		blockTraceName: "agent-a",
		traceStarted:   make(chan struct{}),
		releaseTrace:   make(chan struct{}),
	}
	service := newLangfuseTestService(repository, ingestion, &langfuseScoreStub{}, &langfusePromptStub{})

	// Batch A is the final batch. Its trace is blocked so it deterministically
	// acquires the per-job lock before the non-final batch starts.
	batchA := domain.LangfuseImportBatch{
		JobID:       jobID,
		Fingerprint: fingerprint,
		TotalItems:  3,
		FinalBatch:  true,
		Records: domain.LangfuseExport{
			Traces: []domain.LangfuseTrace{{
				ID:        "trace-a",
				Name:      "agent-a",
				StartTime: "2026-07-25T10:00:00Z",
			}},
			Observations: []domain.LangfuseObservation{{
				ID:        "obs-a",
				TraceID:   "trace-a",
				Type:      "SPAN",
				StartTime: "not-a-valid-time",
			}},
		},
	}

	// Batch B imports a second, distinct valid trace. It is intentionally
	// non-final: after batch A records the final marker, this later in-flight
	// batch must preserve (and recompute) the terminal status rather than
	// downgrading the job to RUNNING.
	batchB := domain.LangfuseImportBatch{
		JobID:       jobID,
		Fingerprint: fingerprint,
		TotalItems:  3,
		FinalBatch:  false,
		Records: domain.LangfuseExport{
			Traces: []domain.LangfuseTrace{{
				ID:        "trace-b",
				Name:      "agent-b",
				StartTime: "2026-07-25T11:00:00Z",
			}},
		},
	}

	errs := make(chan error, 2)
	go func() {
		_, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batchA)
		errs <- err
	}()
	select {
	case <-ingestion.traceStarted:
	case <-time.After(time.Second):
		t.Fatal("final batch did not start")
	}
	go func() {
		_, err := service.ImportBatch(context.Background(), projectID, uuid.New(), batchB)
		errs <- err
	}()
	close(ingestion.releaseTrace)
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)

	final, err := repository.GetByID(context.Background(), projectID, jobID)
	require.NoError(t, err)

	// Both batches' trace imports survived: neither overwrote the other.
	assert.Equal(t, int64(2), final.Progress.TracesMigrated, "both distinct traces must be counted")
	assert.Equal(t, int64(3), final.Progress.ProcessedItems, "every processed item is counted once")

	// The failing observation's error is preserved and drives the final status.
	require.NotEmpty(t, final.Errors, "the failed item error must be preserved")
	assert.Equal(t, domain.MigrationStatusFailed, final.Status)
	require.NotNil(t, final.CompletedAt)

	// Both traces were actually ingested.
	assert.GreaterOrEqual(t, ingestion.traceCount(), 2)
}
