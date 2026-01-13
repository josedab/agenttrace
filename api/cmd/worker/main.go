package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	chrepo "github.com/agenttrace/agenttrace/api/internal/repository/clickhouse"
	pgrepo "github.com/agenttrace/agenttrace/api/internal/repository/postgres"
	"github.com/agenttrace/agenttrace/api/internal/service"
	"github.com/agenttrace/agenttrace/api/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Server.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("starting worker service")

	// Initialize dependencies
	deps, cleanup, err := initWorkerDependencies(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize dependencies", zap.Error(err))
	}
	defer cleanup()

	// Create worker server
	workerServer, err := worker.NewServer(logger, cfg, deps)
	if err != nil {
		logger.Fatal("failed to create worker server", zap.Error(err))
	}

	// Start worker in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- workerServer.Start()
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("shutting down worker...")
		workerServer.Stop()
	case err := <-errCh:
		if err != nil {
			logger.Error("worker server error", zap.Error(err))
		}
	}

	logger.Info("worker stopped")
}

// initWorkerDependencies initializes dependencies for the worker
func initWorkerDependencies(cfg *config.Config, logger *zap.Logger) (*worker.WorkerDependencies, func(), error) {
	ctx := context.Background()

	// Initialize PostgreSQL using database wrapper
	pgDB, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Initialize ClickHouse using database wrapper
	chDB, err := database.NewClickHouse(ctx, cfg.ClickHouse)
	if err != nil {
		pgDB.Close()
		return nil, nil, fmt.Errorf("failed to initialize ClickHouse: %w", err)
	}

	// Initialize MinIO
	minioClient, err := initMinio(cfg)
	if err != nil {
		logger.Warn("failed to initialize MinIO", zap.Error(err))
	}

	// Initialize repositories
	traceRepo := chrepo.NewTraceRepository(chDB, logger)
	observationRepo := chrepo.NewObservationRepository(chDB, logger)
	scoreRepo := chrepo.NewScoreRepository(chDB, logger)
	sessionRepo := chrepo.NewSessionRepository(chDB)
	projectRepo := pgrepo.NewProjectRepository(pgDB)
	datasetRepo := pgrepo.NewDatasetRepository(pgDB)
	evaluatorRepo := pgrepo.NewEvaluatorRepository(pgDB)
	orgRepo := pgrepo.NewOrgRepository(pgDB)
	webhookRepo := pgrepo.NewWebhookRepository(pgDB)

	// Initialize services
	costService := service.NewCostService(logger)
	queryService := service.NewQueryService(traceRepo, observationRepo, scoreRepo, sessionRepo)
	ingestionService := service.NewIngestionService(logger, traceRepo, observationRepo, costService, nil)
	scoreService := service.NewScoreService(scoreRepo, traceRepo, observationRepo)
	datasetService := service.NewDatasetService(datasetRepo, traceRepo, scoreRepo)
	evalService := service.NewEvalService(evaluatorRepo, scoreService)
	projectService := service.NewProjectService(projectRepo, orgRepo)
	notificationService := service.NewNotificationService(logger, "") // Dashboard URL not configured

	// Create dependencies
	deps := &worker.WorkerDependencies{
		CostService:         costService,
		EvalService:         evalService,
		ScoreService:        scoreService,
		QueryService:        queryService,
		IngestionService:    ingestionService,
		DatasetService:      datasetService,
		ProjectService:      projectService,
		NotificationService: notificationService,
		MinioClient:         minioClient,
		MinioBucket:         cfg.MinIO.Bucket,
		// Repositories for cleanup worker
		TraceRepo:       traceRepo,
		ObservationRepo: observationRepo,
		ScoreRepo:       scoreRepo,
		// Repositories for notification worker
		WebhookRepo: webhookRepo,
	}

	// Cleanup function
	cleanup := func() {
		pgDB.Close()
		chDB.Close()
	}

	return deps, cleanup, nil
}

// initMinio initializes MinIO client
func initMinio(cfg *config.Config) (*minio.Client, error) {
	if cfg.MinIO.Endpoint == "" {
		return nil, nil
	}

	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return client, nil
}
