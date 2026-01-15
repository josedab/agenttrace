package main

import (
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// Services holds all service instances
type Services struct {
	Query           *service.QueryService
	Ingestion       *service.IngestionService
	Score           *service.ScoreService
	Prompt          *service.PromptService
	Dataset         *service.DatasetService
	Eval            *service.EvalService
	Auth            *service.AuthService
	Org             *service.OrgService
	Project         *service.ProjectService
	Cost            *service.CostService
	Realtime        *service.RealtimeService
	Checkpoint      *service.CheckpointService
	GitLink         *service.GitLinkService
	FileOperation   *service.FileOperationService
	TerminalCommand *service.TerminalCommandService
	CIRun           *service.CIRunService
	Replay          *service.ReplayService
	Experiment      *service.ExperimentService
}

// initServices initializes all services
func initServices(cfg *config.Config, logger *zap.Logger, repos *Repositories) *Services {
	svcs := &Services{}

	// Cost service (no dependencies)
	svcs.Cost = service.NewCostService(logger)

	// Query service
	svcs.Query = service.NewQueryService(
		repos.Trace,
		repos.Observation,
		repos.Score,
		repos.Session,
	)

	// Score service
	svcs.Score = service.NewScoreService(
		repos.Score,
		repos.Trace,
		repos.Observation,
	)

	// Ingestion service (eval service set later due to circular dependency)
	svcs.Ingestion = service.NewIngestionService(
		logger,
		repos.Trace,
		repos.Observation,
		svcs.Cost,
		nil, // evalService - set below
	)

	// Prompt service
	svcs.Prompt = service.NewPromptService(repos.Prompt)

	// Dataset service
	svcs.Dataset = service.NewDatasetService(
		repos.Dataset,
		repos.Trace,
		repos.Score,
	)

	// Eval service
	svcs.Eval = service.NewEvalService(
		repos.Evaluator,
		svcs.Score,
	)

	// Auth service
	svcs.Auth = service.NewAuthService(
		cfg,
		repos.User,
		repos.APIKey,
		repos.Org,
		repos.Project,
	)

	// Org service
	svcs.Org = service.NewOrgService(repos.Org)

	// Project service
	svcs.Project = service.NewProjectService(
		repos.Project,
		repos.Org,
	)

	// Realtime service
	svcs.Realtime = service.NewRealtimeService()

	// Checkpoint service
	svcs.Checkpoint = service.NewCheckpointService(
		repos.Checkpoint,
		repos.Trace,
	)

	// GitLink service
	svcs.GitLink = service.NewGitLinkService(
		repos.GitLink,
		repos.Trace,
	)

	// FileOperation service
	svcs.FileOperation = service.NewFileOperationService(
		repos.FileOperation,
		repos.Trace,
	)

	// TerminalCommand service
	svcs.TerminalCommand = service.NewTerminalCommandService(
		repos.TerminalCommand,
		repos.Trace,
	)

	// CIRun service
	svcs.CIRun = service.NewCIRunService(repos.CIRun)

	// Replay service
	svcs.Replay = service.NewReplayService(
		logger,
		repos.Trace,
		repos.Observation,
		repos.FileOperation,
		repos.TerminalCommand,
		repos.Checkpoint,
		repos.GitLink,
	)

	// Experiment service
	svcs.Experiment = service.NewExperimentService(
		logger,
		repos.Experiment,
	)

	return svcs
}
