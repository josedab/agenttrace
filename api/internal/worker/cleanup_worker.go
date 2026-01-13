package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	chrepo "github.com/agenttrace/agenttrace/api/internal/repository/clickhouse"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

const (
	// TypeDataCleanup is the task type for data cleanup
	TypeDataCleanup = "cleanup:data"
	// TypeProjectCleanup is the task type for project cleanup
	TypeProjectCleanup = "cleanup:project"
	// TypeOrphanCleanup is the task type for orphan data cleanup
	TypeOrphanCleanup = "cleanup:orphans"
)

// DataCleanupPayload is the payload for data cleanup tasks
type DataCleanupPayload struct {
	ProjectID     uuid.UUID `json:"project_id"`
	RetentionDays int       `json:"retention_days"`
	DryRun        bool      `json:"dry_run"`
}

// NewDataCleanupTask creates a data cleanup task
func NewDataCleanupTask(payload *DataCleanupPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data cleanup payload: %w", err)
	}
	return asynq.NewTask(TypeDataCleanup, data, asynq.MaxRetry(3), asynq.Timeout(1*time.Hour)), nil
}

// ProjectCleanupPayload is the payload for project cleanup tasks
type ProjectCleanupPayload struct {
	ProjectID uuid.UUID `json:"project_id"`
}

// NewProjectCleanupTask creates a project cleanup task
func NewProjectCleanupTask(payload *ProjectCleanupPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project cleanup payload: %w", err)
	}
	return asynq.NewTask(TypeProjectCleanup, data, asynq.MaxRetry(3), asynq.Timeout(30*time.Minute)), nil
}

// OrphanCleanupPayload is the payload for orphan cleanup tasks
type OrphanCleanupPayload struct {
	DryRun bool `json:"dry_run"`
}

// NewOrphanCleanupTask creates an orphan cleanup task
func NewOrphanCleanupTask(payload *OrphanCleanupPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal orphan cleanup payload: %w", err)
	}
	return asynq.NewTask(TypeOrphanCleanup, data, asynq.MaxRetry(3), asynq.Timeout(1*time.Hour)), nil
}

// CleanupWorker handles cleanup tasks
type CleanupWorker struct {
	logger           *zap.Logger
	queryService     *service.QueryService
	ingestionService *service.IngestionService
	projectService   *service.ProjectService
	traceRepo        *chrepo.TraceRepository
	observationRepo  *chrepo.ObservationRepository
	scoreRepo        *chrepo.ScoreRepository
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(
	logger *zap.Logger,
	queryService *service.QueryService,
	ingestionService *service.IngestionService,
	projectService *service.ProjectService,
	traceRepo *chrepo.TraceRepository,
	observationRepo *chrepo.ObservationRepository,
	scoreRepo *chrepo.ScoreRepository,
) *CleanupWorker {
	return &CleanupWorker{
		logger:           logger,
		queryService:     queryService,
		ingestionService: ingestionService,
		projectService:   projectService,
		traceRepo:        traceRepo,
		observationRepo:  observationRepo,
		scoreRepo:        scoreRepo,
	}
}

// ProcessTask processes a data cleanup task
func (w *CleanupWorker) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload DataCleanupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal data cleanup payload: %w", err)
	}

	w.logger.Info("processing data cleanup",
		zap.String("project_id", payload.ProjectID.String()),
		zap.Int("retention_days", payload.RetentionDays),
		zap.Bool("dry_run", payload.DryRun),
	)

	// Calculate cutoff date
	cutoff := time.Now().AddDate(0, 0, -payload.RetentionDays)

	// Count records to delete
	traceCount, err := w.countTracesBeforeCutoff(ctx, payload.ProjectID, cutoff)
	if err != nil {
		return fmt.Errorf("failed to count traces: %w", err)
	}

	w.logger.Info("found traces to clean up",
		zap.String("project_id", payload.ProjectID.String()),
		zap.Int64("trace_count", traceCount),
		zap.Time("cutoff", cutoff),
	)

	if payload.DryRun {
		w.logger.Info("dry run - skipping actual deletion",
			zap.Int64("would_delete_traces", traceCount),
		)
		return nil
	}

	// Delete traces (this will cascade to observations, scores)
	deleted, err := w.deleteTracesBeforeCutoff(ctx, payload.ProjectID, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete traces: %w", err)
	}

	w.logger.Info("data cleanup completed",
		zap.String("project_id", payload.ProjectID.String()),
		zap.Int64("deleted_traces", deleted),
	)

	return nil
}

// ProcessProjectCleanupTask processes a project cleanup task (delete all project data)
func (w *CleanupWorker) ProcessProjectCleanupTask(ctx context.Context, t *asynq.Task) error {
	var payload ProjectCleanupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project cleanup payload: %w", err)
	}

	w.logger.Info("processing project cleanup",
		zap.String("project_id", payload.ProjectID.String()),
	)

	// Delete all traces for the project
	if err := w.deleteAllProjectData(ctx, payload.ProjectID); err != nil {
		return fmt.Errorf("failed to delete project data: %w", err)
	}

	w.logger.Info("project cleanup completed",
		zap.String("project_id", payload.ProjectID.String()),
	)

	return nil
}

// ProcessOrphanCleanupTask processes an orphan cleanup task
func (w *CleanupWorker) ProcessOrphanCleanupTask(ctx context.Context, t *asynq.Task) error {
	var payload OrphanCleanupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal orphan cleanup payload: %w", err)
	}

	w.logger.Info("processing orphan cleanup",
		zap.Bool("dry_run", payload.DryRun),
	)

	// Find orphan observations (observations without valid traces)
	orphanObservations, err := w.findOrphanObservations(ctx)
	if err != nil {
		return fmt.Errorf("failed to find orphan observations: %w", err)
	}

	// Find orphan scores (scores without valid traces/observations)
	orphanScores, err := w.findOrphanScores(ctx)
	if err != nil {
		return fmt.Errorf("failed to find orphan scores: %w", err)
	}

	w.logger.Info("found orphan records",
		zap.Int64("orphan_observations", orphanObservations),
		zap.Int64("orphan_scores", orphanScores),
	)

	if payload.DryRun {
		w.logger.Info("dry run - skipping actual deletion")
		return nil
	}

	// Delete orphan records
	if err := w.deleteOrphanRecords(ctx); err != nil {
		return fmt.Errorf("failed to delete orphan records: %w", err)
	}

	w.logger.Info("orphan cleanup completed")

	return nil
}

// countTracesBeforeCutoff counts traces before the cutoff date
func (w *CleanupWorker) countTracesBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	return w.traceRepo.CountBeforeCutoff(ctx, projectID, cutoff)
}

// deleteTracesBeforeCutoff deletes traces before the cutoff date
// Also deletes related observations and scores
func (w *CleanupWorker) deleteTracesBeforeCutoff(ctx context.Context, projectID uuid.UUID, cutoff time.Time) (int64, error) {
	// Delete observations and scores first (they reference traces)
	obsDeleted, err := w.observationRepo.DeleteBeforeCutoff(ctx, projectID, cutoff)
	if err != nil {
		w.logger.Error("failed to delete observations before cutoff",
			zap.String("project_id", projectID.String()),
			zap.Time("cutoff", cutoff),
			zap.Error(err),
		)
		// Continue with trace deletion even if observation deletion fails
	} else {
		w.logger.Info("deleted observations before cutoff",
			zap.String("project_id", projectID.String()),
			zap.Int64("count", obsDeleted),
		)
	}

	scoresDeleted, err := w.scoreRepo.DeleteBeforeCutoff(ctx, projectID, cutoff)
	if err != nil {
		w.logger.Error("failed to delete scores before cutoff",
			zap.String("project_id", projectID.String()),
			zap.Time("cutoff", cutoff),
			zap.Error(err),
		)
		// Continue with trace deletion even if score deletion fails
	} else {
		w.logger.Info("deleted scores before cutoff",
			zap.String("project_id", projectID.String()),
			zap.Int64("count", scoresDeleted),
		)
	}

	// Delete traces
	return w.traceRepo.DeleteBeforeCutoff(ctx, projectID, cutoff)
}

// deleteAllProjectData deletes all data for a project
func (w *CleanupWorker) deleteAllProjectData(ctx context.Context, projectID uuid.UUID) error {
	// Delete observations first (they reference traces)
	if err := w.observationRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete observations: %w", err)
	}
	w.logger.Info("deleted all observations for project",
		zap.String("project_id", projectID.String()),
	)

	// Delete scores
	if err := w.scoreRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete scores: %w", err)
	}
	w.logger.Info("deleted all scores for project",
		zap.String("project_id", projectID.String()),
	)

	// Delete traces
	if err := w.traceRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete traces: %w", err)
	}
	w.logger.Info("deleted all traces for project",
		zap.String("project_id", projectID.String()),
	)

	return nil
}

// findOrphanObservations finds observations without valid parent traces
func (w *CleanupWorker) findOrphanObservations(ctx context.Context) (int64, error) {
	return w.observationRepo.CountOrphans(ctx)
}

// findOrphanScores finds scores without valid parent traces/observations
func (w *CleanupWorker) findOrphanScores(ctx context.Context) (int64, error) {
	return w.scoreRepo.CountOrphans(ctx)
}

// deleteOrphanRecords deletes orphan observations and scores
func (w *CleanupWorker) deleteOrphanRecords(ctx context.Context) error {
	// Delete orphan observations
	obsDeleted, err := w.observationRepo.DeleteOrphans(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete orphan observations: %w", err)
	}
	w.logger.Info("deleted orphan observations", zap.Int64("count", obsDeleted))

	// Delete orphan scores
	scoresDeleted, err := w.scoreRepo.DeleteOrphans(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete orphan scores: %w", err)
	}
	w.logger.Info("deleted orphan scores", zap.Int64("count", scoresDeleted))

	return nil
}

// ScheduledCleanupConfig holds configuration for scheduled cleanup
type ScheduledCleanupConfig struct {
	DefaultRetentionDays int
	CleanupHour          int // Hour of day to run cleanup (0-23)
	BatchSize            int // Number of projects to process per batch
}

// ScheduleCleanupTasks schedules cleanup tasks for all projects
func ScheduleCleanupTasks(
	ctx context.Context,
	client *asynq.Client,
	projectService *service.ProjectService,
	config *ScheduledCleanupConfig,
	logger *zap.Logger,
) error {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}

	offset := 0
	totalScheduled := 0

	for {
		projects, err := projectService.ListAll(ctx, config.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		if len(projects) == 0 {
			break
		}

		for _, project := range projects {
			// Use project-specific retention or default
			retentionDays := project.RetentionDays
			if retentionDays <= 0 {
				retentionDays = config.DefaultRetentionDays
			}

			task, err := NewDataCleanupTask(&DataCleanupPayload{
				ProjectID:     project.ID,
				RetentionDays: retentionDays,
				DryRun:        false,
			})
			if err != nil {
				logger.Error("failed to create cleanup task",
					zap.String("project_id", project.ID.String()),
					zap.Error(err),
				)
				continue
			}

			if _, err := client.Enqueue(task); err != nil {
				logger.Error("failed to enqueue cleanup task",
					zap.String("project_id", project.ID.String()),
					zap.Error(err),
				)
				continue
			}

			totalScheduled++
		}

		offset += len(projects)

		// If we got fewer than batch size, we've reached the end
		if len(projects) < config.BatchSize {
			break
		}
	}

	logger.Info("scheduled data cleanup tasks",
		zap.Int("total_projects", totalScheduled),
	)

	// Schedule orphan cleanup
	orphanTask, err := NewOrphanCleanupTask(&OrphanCleanupPayload{
		DryRun: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create orphan cleanup task: %w", err)
	}

	if _, err := client.Enqueue(orphanTask); err != nil {
		return fmt.Errorf("failed to enqueue orphan cleanup task: %w", err)
	}

	logger.Info("scheduled orphan cleanup task")

	return nil
}
