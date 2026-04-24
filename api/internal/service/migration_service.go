package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MigrationRepository defines repository operations for migration jobs
type MigrationRepository interface {
	Save(ctx context.Context, job *domain.MigrationJob) error
	GetByID(ctx context.Context, projectID, id uuid.UUID) (*domain.MigrationJob, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.MigrationJob, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress domain.MigrationProgress) error
}

// LocalMigrationDSN marks the only supported source that reads a local export
// file instead of contacting a remote platform.
const LocalMigrationDSN = "json-export"

// MigrationService handles data migration from external platforms
type MigrationService struct {
	logger        *zap.Logger
	migrationRepo MigrationRepository
	ingestionSvc  *IngestionService
	promptSvc     *PromptService
	datasetSvc    *DatasetService
	guard         OutboundGuard
}

// NewMigrationService creates a new migration service.
// The outbound guard rejects remote-source migrations in no-egress mode; local
// JSON export imports stay available.
func NewMigrationService(
	logger *zap.Logger,
	migrationRepo MigrationRepository,
	ingestionSvc *IngestionService,
	promptSvc *PromptService,
	datasetSvc *DatasetService,
	guards ...OutboundGuard,
) *MigrationService {
	var guard OutboundGuard
	if len(guards) > 0 {
		guard = guards[0]
	}
	return &MigrationService{
		logger:        logger,
		migrationRepo: migrationRepo,
		ingestionSvc:  ingestionSvc,
		promptSvc:     promptSvc,
		datasetSvc:    datasetSvc,
		guard:         guard,
	}
}

// StartMigration creates and starts a new migration job
func (s *MigrationService) StartMigration(ctx context.Context, projectID uuid.UUID, input *domain.MigrationInput) (*domain.MigrationJob, error) {
	if !isLocalMigrationSource(input.Config.SourceDSN) {
		if err := RequireOutbound(s.guard, EgressRemoteImport); err != nil {
			return nil, err
		}
	}

	job := &domain.MigrationJob{
		ID:        uuid.New(),
		ProjectID: projectID,
		Source:    input.Source,
		Status:    domain.MigrationStatusPending,
		Config:    input.Config,
		Progress:  domain.MigrationProgress{},
		Errors:    []string{},
		CreatedAt: time.Now(),
	}

	if err := s.migrationRepo.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to save migration job: %w", err)
	}

	s.logger.Info("started migration job",
		zap.String("jobId", job.ID.String()),
		zap.String("source", job.Source),
		zap.String("projectId", projectID.String()),
	)

	return job, nil
}

// GetMigration retrieves a migration job by ID
func (s *MigrationService) GetMigration(ctx context.Context, projectID, jobID uuid.UUID) (*domain.MigrationJob, error) {
	job, err := s.migrationRepo.GetByID(ctx, projectID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration job: %w", err)
	}
	return job, nil
}

// ListMigrations retrieves all migration jobs for a project
func (s *MigrationService) ListMigrations(ctx context.Context, projectID uuid.UUID) ([]domain.MigrationJob, error) {
	jobs, err := s.migrationRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list migration jobs: %w", err)
	}
	return jobs, nil
}

// ValidateSource checks if a migration source connection is valid
func (s *MigrationService) ValidateSource(ctx context.Context, source string, dsn string) (bool, string, error) {
	if !isLocalMigrationSource(dsn) {
		if err := RequireOutbound(s.guard, EgressRemoteImport); err != nil {
			return false, "", err
		}
	}

	switch source {
	case "langfuse":
		if isLocalMigrationSource(dsn) {
			return true, "Langfuse JSON export mode is supported", nil
		}
		return false, "Server-side Langfuse DSN access is not supported; use agenttrace migrate --source-file with a JSON export", nil
	default:
		return false, fmt.Sprintf("unsupported migration source: %s", source), nil
	}
}

// isLocalMigrationSource reports whether a migration reads a local export file.
func isLocalMigrationSource(dsn string) bool {
	return strings.TrimSpace(dsn) == "" || strings.TrimSpace(dsn) == LocalMigrationDSN
}
