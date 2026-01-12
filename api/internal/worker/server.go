package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	chrepo "github.com/agenttrace/agenttrace/api/internal/repository/clickhouse"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// Server is the worker server
type Server struct {
	logger     *zap.Logger
	config     *config.Config
	server     *asynq.Server
	mux        *asynq.ServeMux
	scheduler  *asynq.Scheduler
	client     *asynq.Client
}

// WorkerDependencies holds dependencies for workers
type WorkerDependencies struct {
	CostService      *service.CostService
	EvalService      *service.EvalService
	ScoreService     *service.ScoreService
	QueryService     *service.QueryService
	IngestionService *service.IngestionService
	DatasetService   *service.DatasetService
	ProjectService   *service.ProjectService
	MinioClient      *minio.Client
	MinioBucket      string
	// Repositories for cleanup worker
	TraceRepo       *chrepo.TraceRepository
	ObservationRepo *chrepo.ObservationRepository
	ScoreRepo       *chrepo.ScoreRepository
}

// NewServer creates a new worker server
func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	deps *WorkerDependencies,
) (*Server, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Create asynq server
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("task processing failed",
					zap.String("type", task.Type()),
					zap.Error(err),
				)
			}),
			Logger: &asynqLogger{logger: logger},
		},
	)

	// Create workers
	costWorker := NewCostWorker(
		logger,
		deps.CostService,
		deps.QueryService,
		deps.IngestionService,
	)

	evalWorker := NewEvalWorker(
		logger,
		cfg,
		deps.EvalService,
		deps.ScoreService,
		deps.QueryService,
	)

	exportWorker := NewExportWorker(
		logger,
		deps.QueryService,
		deps.ScoreService,
		deps.DatasetService,
		deps.MinioClient,
		deps.MinioBucket,
	)

	cleanupWorker := NewCleanupWorker(
		logger,
		deps.QueryService,
		deps.IngestionService,
		deps.ProjectService,
		deps.TraceRepo,
		deps.ObservationRepo,
		deps.ScoreRepo,
	)

	// Create mux and register handlers
	mux := asynq.NewServeMux()

	// Cost workers
	mux.HandleFunc(TypeCostCalculation, costWorker.ProcessTask)
	mux.HandleFunc(TypeBatchCostCalculation, costWorker.ProcessBatchCostTask)
	mux.HandleFunc(TypeDailyAggregation, costWorker.ProcessDailyAggregationTask)

	// Eval workers
	mux.HandleFunc(TypeEvaluation, evalWorker.ProcessTask)
	mux.HandleFunc(TypeBatchEvaluation, evalWorker.ProcessBatchTask)

	// Export workers
	mux.HandleFunc(TypeDataExport, exportWorker.ProcessTask)
	mux.HandleFunc(TypeDatasetExport, exportWorker.ProcessDatasetTask)

	// Cleanup workers
	mux.HandleFunc(TypeDataCleanup, cleanupWorker.ProcessTask)
	mux.HandleFunc(TypeProjectCleanup, cleanupWorker.ProcessProjectCleanupTask)
	mux.HandleFunc(TypeOrphanCleanup, cleanupWorker.ProcessOrphanCleanupTask)

	// Create scheduler for periodic tasks
	scheduler := asynq.NewScheduler(redisOpt, nil)

	// Create client for enqueuing tasks
	client := asynq.NewClient(redisOpt)

	return &Server{
		logger:    logger,
		config:    cfg,
		server:    server,
		mux:       mux,
		scheduler: scheduler,
		client:    client,
	}, nil
}

// Start starts the worker server
func (s *Server) Start() error {
	// Register scheduled tasks
	if err := s.registerScheduledTasks(); err != nil {
		return fmt.Errorf("failed to register scheduled tasks: %w", err)
	}

	// Start scheduler
	go func() {
		if err := s.scheduler.Run(); err != nil {
			s.logger.Error("scheduler stopped", zap.Error(err))
		}
	}()

	// Start server
	s.logger.Info("starting worker server",
		zap.Int("concurrency", s.config.Worker.Concurrency),
	)

	return s.server.Run(s.mux)
}

// Stop stops the worker server
func (s *Server) Stop() {
	s.server.Shutdown()
	s.scheduler.Shutdown()
	s.client.Close()
}

// Client returns the asynq client for enqueuing tasks
func (s *Server) Client() *asynq.Client {
	return s.client
}

// registerScheduledTasks registers periodic tasks with the scheduler
func (s *Server) registerScheduledTasks() error {
	// Daily cleanup at 3 AM UTC
	_, err := s.scheduler.Register(
		"0 3 * * *", // Cron expression
		asynq.NewTask(TypeOrphanCleanup, []byte(`{"dry_run":false}`)),
		asynq.Queue("low"),
	)
	if err != nil {
		return fmt.Errorf("failed to register orphan cleanup task: %w", err)
	}

	// Daily cost aggregation at 1 AM UTC
	_, err = s.scheduler.Register(
		"0 1 * * *",
		asynq.NewTask(TypeDailyAggregation, []byte(`{}`)),
		asynq.Queue("low"),
	)
	if err != nil {
		return fmt.Errorf("failed to register daily aggregation task: %w", err)
	}

	return nil
}

// asynqLogger adapts zap.Logger to asynq.Logger
type asynqLogger struct {
	logger *zap.Logger
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

// EnqueueCostCalculation enqueues a cost calculation task
func EnqueueCostCalculation(client *asynq.Client, payload *CostCalculationPayload) error {
	task, err := NewCostCalculationTask(payload)
	if err != nil {
		return err
	}
	_, err = client.Enqueue(task, asynq.Queue("default"))
	return err
}

// EnqueueEvaluation enqueues an evaluation task
func EnqueueEvaluation(client *asynq.Client, payload *EvaluationPayload) error {
	task, err := NewEvaluationTask(payload)
	if err != nil {
		return err
	}
	_, err = client.Enqueue(task, asynq.Queue("default"))
	return err
}

// EnqueueDataExport enqueues a data export task
func EnqueueDataExport(client *asynq.Client, payload *DataExportPayload) error {
	task, err := NewDataExportTask(payload)
	if err != nil {
		return err
	}
	_, err = client.Enqueue(task, asynq.Queue("low"))
	return err
}

// EnqueueDataCleanup enqueues a data cleanup task
func EnqueueDataCleanup(client *asynq.Client, payload *DataCleanupPayload) error {
	task, err := NewDataCleanupTask(payload)
	if err != nil {
		return err
	}
	// Use ProcessIn to delay cleanup tasks
	_, err = client.Enqueue(task, asynq.Queue("low"), asynq.ProcessIn(time.Hour))
	return err
}
