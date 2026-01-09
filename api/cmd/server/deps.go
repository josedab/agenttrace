package main

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/graphql/resolver"
	"github.com/agenttrace/agenttrace/api/internal/handler"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
	chrepo "github.com/agenttrace/agenttrace/api/internal/repository/clickhouse"
	pgrepo "github.com/agenttrace/agenttrace/api/internal/repository/postgres"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// Dependencies holds all application dependencies
type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger

	// Database connections
	Postgres   *database.PostgresDB
	ClickHouse *database.ClickHouseDB
	Redis      *redis.Client
	Minio      *minio.Client

	// Repositories
	TraceRepo           *chrepo.TraceRepository
	ObservationRepo     *chrepo.ObservationRepository
	ScoreRepo           *chrepo.ScoreRepository
	SessionRepo         *chrepo.SessionRepository
	CheckpointRepo      *chrepo.CheckpointRepository
	GitLinkRepo         *chrepo.GitLinkRepository
	FileOperationRepo   *chrepo.FileOperationRepository
	TerminalCommandRepo *chrepo.TerminalCommandRepository
	CIRunRepo           *chrepo.CIRunRepository
	UserRepo            *pgrepo.UserRepository
	OrgRepo             *pgrepo.OrgRepository
	ProjectRepo         *pgrepo.ProjectRepository
	APIKeyRepo          *pgrepo.APIKeyRepository
	PromptRepo          *pgrepo.PromptRepository
	DatasetRepo         *pgrepo.DatasetRepository
	EvaluatorRepo       *pgrepo.EvaluatorRepository

	// Services
	QueryService           *service.QueryService
	IngestionService       *service.IngestionService
	ScoreService           *service.ScoreService
	PromptService          *service.PromptService
	DatasetService         *service.DatasetService
	EvalService            *service.EvalService
	AuthService            *service.AuthService
	OrgService             *service.OrgService
	ProjectService         *service.ProjectService
	CostService            *service.CostService
	RealtimeService        *service.RealtimeService
	CheckpointService      *service.CheckpointService
	GitLinkService         *service.GitLinkService
	FileOperationService   *service.FileOperationService
	TerminalCommandService *service.TerminalCommandService
	CIRunService           *service.CIRunService

	// Handlers
	HealthHandler           *handler.HealthHandler
	IngestionHandler        *handler.IngestionHandler
	TracesHandler           *handler.TracesHandler
	ScoresHandler           *handler.ScoresHandler
	PromptsHandler          *handler.PromptsHandler
	DatasetsHandler         *handler.DatasetsHandler
	EvaluatorsHandler       *handler.EvaluatorsHandler
	EventsHandler           *handler.EventsHandler
	APIKeysHandler          *handler.APIKeysHandler
	ProjectsHandler         *handler.ProjectsHandler
	OrganizationsHandler    *handler.OrganizationsHandler
	AuthHandler             *handler.AuthHandler
	CheckpointsHandler      *handler.CheckpointsHandler
	GitLinksHandler         *handler.GitLinksHandler
	FileOperationsHandler   *handler.FileOperationsHandler
	TerminalCommandsHandler *handler.TerminalCommandsHandler
	CIRunsHandler           *handler.CIRunsHandler
	ExportHandler           *handler.ExportHandler
	ImportHandler           *handler.ImportHandler
	DocsHandler             *handler.DocsHandler

	// Middleware
	AuthMiddleware      *middleware.AuthMiddleware
	RateLimitMiddleware *middleware.RateLimitMiddleware

	// GraphQL resolver
	Resolver *resolver.Resolver

	// Asynq client
	AsynqClient *asynq.Client
}

// initDependencies initializes all dependencies
func initDependencies(cfg *config.Config, logger *zap.Logger) (*Dependencies, error) {
	deps := &Dependencies{
		Config: cfg,
		Logger: logger,
	}

	ctx := context.Background()

	// Initialize PostgreSQL using database wrapper
	pgDB, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	deps.Postgres = pgDB

	// Initialize ClickHouse using database wrapper
	chDB, err := database.NewClickHouse(ctx, cfg.ClickHouse)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ClickHouse: %w", err)
	}
	deps.ClickHouse = chDB

	// Initialize Redis
	redisClient, err := initRedis(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}
	deps.Redis = redisClient

	// Initialize MinIO
	minioClient, err := initMinio(cfg)
	if err != nil {
		logger.Warn("failed to initialize MinIO, file storage will be unavailable", zap.Error(err))
	}
	deps.Minio = minioClient

	// Initialize repositories
	deps.TraceRepo = chrepo.NewTraceRepository(chDB)
	deps.ObservationRepo = chrepo.NewObservationRepository(chDB)
	deps.ScoreRepo = chrepo.NewScoreRepository(chDB)
	deps.SessionRepo = chrepo.NewSessionRepository(chDB)
	deps.CheckpointRepo = chrepo.NewCheckpointRepository(chDB)
	deps.GitLinkRepo = chrepo.NewGitLinkRepository(chDB)
	deps.FileOperationRepo = chrepo.NewFileOperationRepository(chDB)
	deps.TerminalCommandRepo = chrepo.NewTerminalCommandRepository(chDB)
	deps.CIRunRepo = chrepo.NewCIRunRepository(chDB)
	deps.UserRepo = pgrepo.NewUserRepository(pgDB)
	deps.OrgRepo = pgrepo.NewOrgRepository(pgDB)
	deps.ProjectRepo = pgrepo.NewProjectRepository(pgDB)
	deps.APIKeyRepo = pgrepo.NewAPIKeyRepository(pgDB)
	deps.PromptRepo = pgrepo.NewPromptRepository(pgDB)
	deps.DatasetRepo = pgrepo.NewDatasetRepository(pgDB)
	deps.EvaluatorRepo = pgrepo.NewEvaluatorRepository(pgDB)

	// Initialize Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	deps.AsynqClient = asynqClient

	// Initialize services
	deps.CostService = service.NewCostService()
	deps.QueryService = service.NewQueryService(
		deps.TraceRepo,
		deps.ObservationRepo,
		deps.ScoreRepo,
		deps.SessionRepo,
	)
	deps.ScoreService = service.NewScoreService(
		deps.ScoreRepo,
		deps.TraceRepo,
		deps.ObservationRepo,
	)
	deps.IngestionService = service.NewIngestionService(
		deps.TraceRepo,
		deps.ObservationRepo,
		deps.CostService,
		nil, // evalService - created later
	)
	deps.PromptService = service.NewPromptService(
		deps.PromptRepo,
	)
	deps.DatasetService = service.NewDatasetService(
		deps.DatasetRepo,
		deps.TraceRepo,
		deps.ScoreRepo,
	)
	deps.EvalService = service.NewEvalService(
		deps.EvaluatorRepo,
		deps.ScoreService,
	)
	deps.AuthService = service.NewAuthService(
		cfg,
		deps.UserRepo,
		deps.APIKeyRepo,
		deps.OrgRepo,
		deps.ProjectRepo,
	)
	deps.OrgService = service.NewOrgService(
		deps.OrgRepo,
	)
	deps.ProjectService = service.NewProjectService(
		deps.ProjectRepo,
		deps.OrgRepo,
	)
	deps.RealtimeService = service.NewRealtimeService()
	deps.CheckpointService = service.NewCheckpointService(
		deps.CheckpointRepo,
		deps.TraceRepo,
	)
	deps.GitLinkService = service.NewGitLinkService(
		deps.GitLinkRepo,
		deps.TraceRepo,
	)
	deps.FileOperationService = service.NewFileOperationService(
		deps.FileOperationRepo,
		deps.TraceRepo,
	)
	deps.TerminalCommandService = service.NewTerminalCommandService(
		deps.TerminalCommandRepo,
		deps.TraceRepo,
	)
	deps.CIRunService = service.NewCIRunService(
		deps.CIRunRepo,
	)

	// Initialize handlers
	deps.HealthHandler = handler.NewHealthHandler(
		pgDB.Pool,
		chDB.Conn,
		redisClient,
		"0.1.0", // Version is hardcoded for now
	)
	deps.IngestionHandler = handler.NewIngestionHandler(
		deps.IngestionService,
		deps.ScoreService,
		logger,
	)
	deps.TracesHandler = handler.NewTracesHandler(
		deps.QueryService,
		logger,
	)
	deps.ScoresHandler = handler.NewScoresHandler(
		deps.ScoreService,
		logger,
	)
	deps.PromptsHandler = handler.NewPromptsHandler(
		deps.PromptService,
		logger,
	)
	deps.DatasetsHandler = handler.NewDatasetsHandler(
		deps.DatasetService,
		logger,
	)
	deps.EvaluatorsHandler = handler.NewEvaluatorsHandler(
		deps.EvalService,
		logger,
	)
	deps.EventsHandler = handler.NewEventsHandler(
		deps.RealtimeService,
		logger,
	)
	deps.APIKeysHandler = handler.NewAPIKeysHandler(
		deps.AuthService,
		logger,
	)
	deps.ProjectsHandler = handler.NewProjectsHandler(
		deps.ProjectService,
		logger,
	)
	deps.OrganizationsHandler = handler.NewOrganizationsHandler(
		deps.OrgService,
		logger,
	)
	deps.AuthHandler = handler.NewAuthHandler(
		deps.AuthService,
		logger,
	)
	deps.CheckpointsHandler = handler.NewCheckpointsHandler(
		deps.CheckpointService,
		logger,
	)
	deps.GitLinksHandler = handler.NewGitLinksHandler(
		deps.GitLinkService,
		logger,
	)
	deps.FileOperationsHandler = handler.NewFileOperationsHandler(
		deps.FileOperationService,
		logger,
	)
	deps.TerminalCommandsHandler = handler.NewTerminalCommandsHandler(
		deps.TerminalCommandService,
		logger,
	)
	deps.CIRunsHandler = handler.NewCIRunsHandler(
		deps.CIRunService,
		logger,
	)
	deps.ExportHandler = handler.NewExportHandler(
		deps.AsynqClient,
		logger,
	)
	deps.ImportHandler = handler.NewImportHandler(
		deps.DatasetService,
		deps.PromptService,
		logger,
	)
	deps.DocsHandler = handler.NewDocsHandler()

	// Initialize middleware
	deps.AuthMiddleware = middleware.NewAuthMiddleware(deps.AuthService)
	deps.RateLimitMiddleware = middleware.NewRateLimitMiddleware(redisClient)

	// Initialize GraphQL resolver
	deps.Resolver = resolver.NewResolver(
		logger,
		deps.QueryService,
		deps.IngestionService,
		deps.ScoreService,
		deps.PromptService,
		deps.DatasetService,
		deps.EvalService,
		deps.AuthService,
		deps.OrgService,
		deps.ProjectService,
		deps.CostService,
	)

	return deps, nil
}

// Close closes all dependencies
func (d *Dependencies) Close() {
	if d.Postgres != nil {
		d.Postgres.Close()
	}
	if d.ClickHouse != nil {
		_ = d.ClickHouse.Close()
	}
	if d.Redis != nil {
		d.Redis.Close()
	}
	if d.AsynqClient != nil {
		d.AsynqClient.Close()
	}
}

// initRedis initializes Redis client
func initRedis(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

// initMinio initializes MinIO client
func initMinio(cfg *config.Config) (*minio.Client, error) {
	if cfg.MinIO.Endpoint == "" {
		return nil, nil // MinIO not configured
	}

	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIO.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIO.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return client, nil
}
