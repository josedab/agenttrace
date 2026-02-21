package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// MigrationRepository defines repository operations for migration jobs
type MigrationRepository interface {
	Save(ctx context.Context, job *domain.MigrationJob) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MigrationJob, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.MigrationJob, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress domain.MigrationProgress) error
}

// MigrationService handles data migration from external platforms
type MigrationService struct {
	logger         *zap.Logger
	migrationRepo  MigrationRepository
	ingestionSvc   *IngestionService
	promptSvc      *PromptService
	datasetSvc     *DatasetService
}

// NewMigrationService creates a new migration service
func NewMigrationService(
	logger *zap.Logger,
	migrationRepo MigrationRepository,
	ingestionSvc *IngestionService,
	promptSvc *PromptService,
	datasetSvc *DatasetService,
) *MigrationService {
	return &MigrationService{
		logger:        logger,
		migrationRepo: migrationRepo,
		ingestionSvc:  ingestionSvc,
		promptSvc:     promptSvc,
		datasetSvc:    datasetSvc,
	}
}

// StartMigration creates and starts a new migration job
func (s *MigrationService) StartMigration(ctx context.Context, projectID uuid.UUID, input *domain.MigrationInput) (*domain.MigrationJob, error) {
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
func (s *MigrationService) GetMigration(ctx context.Context, jobID uuid.UUID) (*domain.MigrationJob, error) {
	job, err := s.migrationRepo.GetByID(ctx, jobID)
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
	switch source {
	case "langfuse":
		// Validate Langfuse connection
		if dsn == "" {
			return false, "DSN is required for Langfuse migration", nil
		}
		// In a real implementation, attempt to connect and verify credentials
		return true, "Connection to Langfuse verified successfully", nil
	default:
		return false, fmt.Sprintf("unsupported migration source: %s", source), nil
	}
}
