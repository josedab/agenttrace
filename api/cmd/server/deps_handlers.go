package main

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/handler"
	"github.com/agenttrace/agenttrace/api/internal/pkg/database"
)

// Handlers holds all handler instances
type Handlers struct {
	Health           *handler.HealthHandler
	Ingestion        *handler.IngestionHandler
	Traces           *handler.TracesHandler
	Scores           *handler.ScoresHandler
	Prompts          *handler.PromptsHandler
	Datasets         *handler.DatasetsHandler
	Evaluators       *handler.EvaluatorsHandler
	Events           *handler.EventsHandler
	APIKeys          *handler.APIKeysHandler
	Projects         *handler.ProjectsHandler
	Organizations    *handler.OrganizationsHandler
	Auth             *handler.AuthHandler
	Checkpoints      *handler.CheckpointsHandler
	GitLinks         *handler.GitLinksHandler
	FileOperations   *handler.FileOperationsHandler
	TerminalCommands *handler.TerminalCommandsHandler
	CIRuns           *handler.CIRunsHandler
	Export           *handler.ExportHandler
	Import           *handler.ImportHandler
	Docs             *handler.DocsHandler
	Webhook          *handler.WebhookHandler
	Replay           *handler.ReplayHandler
	Experiment       *handler.ExperimentHandler
}

// initHandlers initializes all handlers
func initHandlers(
	logger *zap.Logger,
	svcs *Services,
	repos *Repositories,
	pgDB *database.PostgresDB,
	chDB *database.ClickHouseDB,
	redisClient *redis.Client,
	asynqClient *asynq.Client,
	version string,
) *Handlers {
	return &Handlers{
		Health: handler.NewHealthHandler(
			pgDB.Pool,
			chDB.Conn,
			redisClient,
			version,
		),
		Ingestion: handler.NewIngestionHandler(
			svcs.Ingestion,
			svcs.Score,
			logger,
		),
		Traces: handler.NewTracesHandler(
			svcs.Query,
			logger,
		),
		Scores: handler.NewScoresHandler(
			svcs.Score,
			logger,
		),
		Prompts: handler.NewPromptsHandler(
			svcs.Prompt,
			logger,
		),
		Datasets: handler.NewDatasetsHandler(
			svcs.Dataset,
			logger,
		),
		Evaluators: handler.NewEvaluatorsHandler(
			svcs.Eval,
			logger,
		),
		Events: handler.NewEventsHandler(
			svcs.Realtime,
			logger,
		),
		APIKeys: handler.NewAPIKeysHandler(
			svcs.Auth,
			logger,
		),
		Projects: handler.NewProjectsHandler(
			svcs.Project,
			logger,
		),
		Organizations: handler.NewOrganizationsHandler(
			svcs.Org,
			logger,
		),
		Auth: handler.NewAuthHandler(
			svcs.Auth,
			logger,
		),
		Checkpoints: handler.NewCheckpointsHandler(
			svcs.Checkpoint,
			logger,
		),
		GitLinks: handler.NewGitLinksHandler(
			svcs.GitLink,
			logger,
		),
		FileOperations: handler.NewFileOperationsHandler(
			svcs.FileOperation,
			logger,
		),
		TerminalCommands: handler.NewTerminalCommandsHandler(
			svcs.TerminalCommand,
			logger,
		),
		CIRuns: handler.NewCIRunsHandler(
			svcs.CIRun,
			logger,
		),
		Export: handler.NewExportHandler(
			asynqClient,
			logger,
		),
		Import: handler.NewImportHandler(
			svcs.Dataset,
			svcs.Prompt,
			logger,
		),
		Docs:    handler.NewDocsHandler(),
		Webhook: handler.NewWebhookHandler(
			logger,
			repos.Webhook,
			nil, // NotificationService
		),
		Replay: handler.NewReplayHandler(
			logger,
			svcs.Replay,
		),
		Experiment: handler.NewExperimentHandler(
			logger,
			svcs.Experiment,
		),
	}
}
