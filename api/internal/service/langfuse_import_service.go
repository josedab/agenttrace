package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

const maxLangfuseBatchItems = 500

// LangfuseImportRepository defines resumable job and item-ledger persistence.
type LangfuseImportRepository interface {
	Save(ctx context.Context, job *domain.MigrationJob) error
	GetByID(ctx context.Context, projectID, id uuid.UUID) (*domain.MigrationJob, error)
	UpdateJob(ctx context.Context, job *domain.MigrationJob) error
	// FindImportedItem returns the identifier recorded for a previously
	// imported source item so a retry can return the prior result instead of
	// importing it a second time.
	FindImportedItem(
		ctx context.Context,
		projectID, jobID uuid.UUID,
		sourceType, sourceID string,
	) (string, bool, error)
	RecordItem(
		ctx context.Context,
		projectID, jobID uuid.UUID,
		sourceType, sourceID, checksum, importedID, status, errorMessage string,
	) error
	// WithBatchLock runs fn while holding a serialization lock scoped to the
	// (projectID, jobID) pair, so at most one ImportBatch for a given job runs
	// at a time — even across processes. Implementations back this with a
	// PostgreSQL transaction-scoped advisory lock. Acquisition and release
	// errors are propagated to the caller.
	WithBatchLock(
		ctx context.Context,
		projectID, jobID uuid.UUID,
		fn func(ctx context.Context) error,
	) error
}

// LangfuseIngestion writes trace and observation records.
type LangfuseIngestion interface {
	IngestTrace(
		ctx context.Context,
		projectID uuid.UUID,
		input *domain.TraceInput,
	) (*domain.Trace, error)
	IngestObservation(
		ctx context.Context,
		projectID uuid.UUID,
		input *domain.ObservationInput,
	) (*domain.Observation, error)
	IngestGeneration(
		ctx context.Context,
		projectID uuid.UUID,
		input *domain.GenerationInput,
	) (*domain.Observation, error)
}

// LangfuseScores writes score records.
type LangfuseScores interface {
	Create(
		ctx context.Context,
		projectID uuid.UUID,
		input *domain.ScoreInput,
	) (*domain.Score, error)
}

// LangfusePrompts writes prompt records and versions.
type LangfusePrompts interface {
	GetByName(
		ctx context.Context,
		projectID uuid.UUID,
		name string,
	) (*domain.Prompt, error)
	Create(
		ctx context.Context,
		projectID uuid.UUID,
		input *domain.PromptInput,
		userID uuid.UUID,
	) (*domain.Prompt, error)
	CreateVersion(
		ctx context.Context,
		promptID uuid.UUID,
		input *domain.PromptVersionInput,
		userID uuid.UUID,
	) (*domain.PromptVersion, error)
}

// LangfuseImportService imports the documented JSON-export subset.
type LangfuseImportService struct {
	repository LangfuseImportRepository
	ingestion  LangfuseIngestion
	scores     LangfuseScores
	prompts    LangfusePrompts
	redactor   *SensitiveDataRedactor
	clock      func() time.Time
}

// NewLangfuseImportService creates a resumable Langfuse importer.
func NewLangfuseImportService(
	repository LangfuseImportRepository,
	ingestion LangfuseIngestion,
	scores LangfuseScores,
	prompts LangfusePrompts,
	redactor *SensitiveDataRedactor,
) *LangfuseImportService {
	return &LangfuseImportService{
		repository: repository,
		ingestion:  ingestion,
		scores:     scores,
		prompts:    prompts,
		redactor:   redactor,
		clock:      time.Now,
	}
}

// ImportBatch imports a bounded batch and updates durable progress.
func (s *LangfuseImportService) ImportBatch(
	ctx context.Context,
	projectID, actorID uuid.UUID,
	batch domain.LangfuseImportBatch,
) (*domain.MigrationJob, error) {
	if batch.JobID == uuid.Nil {
		return nil, apperrors.Validation("jobId is required")
	}
	if len(batch.Fingerprint) < 16 || len(batch.Fingerprint) > 128 {
		return nil, apperrors.Validation("fingerprint must be between 16 and 128 characters")
	}
	itemCount := langfuseBatchItemCount(batch.Records)
	if itemCount == 0 {
		return nil, apperrors.Validation("import batch contains no records")
	}
	if itemCount > maxLangfuseBatchItems {
		return nil, apperrors.Validation(
			fmt.Sprintf("import batch cannot exceed %d records", maxLangfuseBatchItems),
		)
	}

	// Serialize batches for the same job across processes so concurrent batches
	// cannot lose each other's progress via a read-modify-write race. The lock
	// is acquired before get-or-create and released when the callback returns.
	var result *domain.MigrationJob
	if lockErr := s.repository.WithBatchLock(
		ctx,
		projectID,
		batch.JobID,
		func(ctx context.Context) error {
			job, err := s.importBatchLocked(ctx, projectID, actorID, batch)
			if err != nil {
				return err
			}
			result = job
			return nil
		},
	); lockErr != nil {
		return nil, lockErr
	}
	return result, nil
}

// importBatchLocked performs the durable get-or-create, item processing, and
// progress persistence for a single batch. Callers must hold the per-job batch
// lock so this runs without interleaving against a concurrent batch.
func (s *LangfuseImportService) importBatchLocked(
	ctx context.Context,
	projectID, actorID uuid.UUID,
	batch domain.LangfuseImportBatch,
) (*domain.MigrationJob, error) {
	job, err := s.getOrCreateJob(ctx, projectID, batch)
	if err != nil {
		return nil, err
	}
	if job.Config.SourceFingerprint != batch.Fingerprint || job.Config.DryRun != batch.DryRun {
		return nil, apperrors.Conflict("migration job fingerprint or dry-run mode does not match")
	}
	if batch.TotalItems > job.Progress.TotalItems {
		job.Progress.TotalItems = batch.TotalItems
	}

	for _, trace := range batch.Records.Traces {
		s.processItem(ctx, job, "trace", trace.ID, trace, func() (string, error) {
			return s.importTrace(ctx, projectID, trace, batch.DryRun)
		})
	}
	for _, observation := range batch.Records.Observations {
		s.processItem(ctx, job, "observation", observation.ID, observation, func() (string, error) {
			return s.importObservation(ctx, projectID, observation, batch.DryRun)
		})
	}
	for _, score := range batch.Records.Scores {
		s.processItem(ctx, job, "score", score.ID, score, func() (string, error) {
			return s.importScore(ctx, projectID, score, batch.DryRun)
		})
	}
	for _, prompt := range batch.Records.Prompts {
		s.processItem(ctx, job, "prompt", prompt.ID, prompt, func() (string, error) {
			return s.importPrompt(ctx, projectID, actorID, prompt, batch.DryRun)
		})
	}

	if batch.FinalBatch && job.CompletedAt == nil {
		completedAt := s.clock().UTC()
		job.CompletedAt = &completedAt
	}
	// Once the final-batch marker has been observed, later batches that were
	// already in flight must not downgrade the job back to RUNNING. Recompute
	// the terminal state after every serialized batch so a later error changes
	// COMPLETED to FAILED and a successful retry can clear FAILED.
	if job.CompletedAt != nil {
		if len(job.Errors) > 0 {
			job.Status = domain.MigrationStatusFailed
		} else {
			job.Status = domain.MigrationStatusCompleted
		}
	} else {
		job.Status = domain.MigrationStatusRunning
	}
	if err := s.repository.UpdateJob(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *LangfuseImportService) getOrCreateJob(
	ctx context.Context,
	projectID uuid.UUID,
	batch domain.LangfuseImportBatch,
) (*domain.MigrationJob, error) {
	job, err := s.repository.GetByID(ctx, projectID, batch.JobID)
	if err == nil {
		return job, nil
	}
	if !apperrors.IsNotFound(err) {
		return nil, err
	}

	now := s.clock().UTC()
	job = &domain.MigrationJob{
		ID:        batch.JobID,
		ProjectID: projectID,
		Source:    "langfuse",
		Status:    domain.MigrationStatusRunning,
		Config: domain.MigrationConfig{
			DryRun:            batch.DryRun,
			IncrementalMode:   true,
			IncludeTraces:     true,
			IncludePrompts:    true,
			IncludeScores:     true,
			SourceFingerprint: batch.Fingerprint,
		},
		Progress:  domain.MigrationProgress{TotalItems: batch.TotalItems},
		Errors:    []string{},
		CreatedAt: now,
	}
	if err := s.repository.Save(ctx, job); err != nil {
		if apperrors.IsConflict(err) {
			existing, readErr := s.repository.GetByID(ctx, projectID, batch.JobID)
			if readErr == nil {
				return existing, nil
			}
		}
		return nil, err
	}
	return job, nil
}

func (s *LangfuseImportService) processItem(
	ctx context.Context,
	job *domain.MigrationJob,
	sourceType, sourceID string,
	source interface{},
	importItem func() (string, error),
) {
	if strings.TrimSpace(sourceID) == "" {
		sourceID = checksumValue(source)[:16]
	}
	errorKey := sourceType + "/" + normalizeExternalID(sourceID, 16)
	wasRetry := hasMigrationError(job.Errors, errorKey)
	wasLedgerRetry := hasMigrationError(job.Errors, errorKey+"/ledger")
	job.Errors = removeMigrationError(job.Errors, errorKey)

	if !job.Config.DryRun {
		importedID, imported, err := s.repository.FindImportedItem(
			ctx,
			job.ProjectID,
			job.ID,
			sourceType,
			sourceID,
		)
		if err != nil {
			s.recordMigrationError(job, errorKey, err)
			return
		}
		// A ledger row without an identifier means the previous attempt did not
		// finish recording, so the item is imported again; deterministic ids
		// make that rewrite safe.
		if imported && importedID != "" {
			job.Progress.SkippedItems++
			return
		}
	}

	// Imports derive deterministic identifiers from source identifiers, so a
	// retry after a ledger write failure rewrites the same records instead of
	// creating duplicates.
	importedID, err := importItem()
	if !wasRetry {
		job.Progress.ProcessedItems++
	}
	checksum := checksumValue(source)
	if err != nil {
		s.recordMigrationError(job, errorKey, err)
		if !job.Config.DryRun {
			if recordErr := s.repository.RecordItem(
				ctx,
				job.ProjectID,
				job.ID,
				sourceType,
				sourceID,
				checksum,
				"",
				"failed",
				s.redactor.RedactText(err.Error()),
			); recordErr != nil {
				s.recordMigrationError(job, errorKey+"/ledger", recordErr)
			}
		}
		return
	}

	if !wasLedgerRetry {
		switch sourceType {
		case "trace":
			job.Progress.TracesMigrated++
		case "prompt":
			job.Progress.PromptsMigrated++
		case "score":
			job.Progress.ScoresMigrated++
		}
	}
	if !job.Config.DryRun {
		if recordErr := s.repository.RecordItem(
			ctx,
			job.ProjectID,
			job.ID,
			sourceType,
			sourceID,
			checksum,
			importedID,
			"imported",
			"",
		); recordErr != nil {
			s.recordMigrationError(job, errorKey+"/ledger", recordErr)
		}
	}
}

func (s *LangfuseImportService) recordMigrationError(
	job *domain.MigrationJob,
	key string,
	err error,
) {
	message := key + ": " + s.redactor.RedactText(err.Error())
	job.Errors = appendUniqueBounded(job.Errors, message, 100)
}

func langfuseBatchItemCount(export domain.LangfuseExport) int {
	return len(export.Traces) +
		len(export.Observations) +
		len(export.Scores) +
		len(export.Prompts)
}

// removeMigrationError clears the item error and any sub-step error, such as a
// ledger write failure, so a successful retry no longer reports the old failure.
func removeMigrationError(errors []string, key string) []string {
	result := errors[:0]
	for _, message := range errors {
		if !isMigrationErrorFor(message, key) {
			result = append(result, message)
		}
	}
	return result
}

func hasMigrationError(errors []string, key string) bool {
	for _, message := range errors {
		if isMigrationErrorFor(message, key) {
			return true
		}
	}
	return false
}

func isMigrationErrorFor(message, key string) bool {
	return strings.HasPrefix(message, key+":") || strings.HasPrefix(message, key+"/")
}

func appendUniqueBounded(values []string, value string, limit int) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	if len(values) >= limit {
		return values
	}
	return append(values, value)
}
