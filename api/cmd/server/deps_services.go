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
	Debug           *service.DebugService
	Regression      *service.RegressionService
	Tenant          *service.TenantService
	CostOptimizer   *service.CostOptimizerService
	AgentGraph      *service.AgentGraphService
	Migration       *service.MigrationService
	Benchmark       *service.BenchmarkService
	Collaboration   *service.CollaborationService
	Guardrail       *service.GuardrailService
	OTelReceiver    *service.OTelReceiverService
	Compliance      *service.ComplianceService
	Billing         *service.BillingService
	Prediction      *service.PredictionService
	Reasoning       *service.ReasoningService
	CostBudget      *service.CostBudgetService
	Instrumentation *service.InstrumentationService
	ComplianceExport *service.ComplianceExportService
	Scorecard       *service.ScorecardService
	Ticket          *service.TicketService
	Audit           *service.AuditService
	Streaming       *service.StreamingService
	DiffIntelligence *service.DiffIntelligenceService
	Anomaly         *service.AnomalyService
	Federation      *service.FederationService
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

	// Debug service (extends replay with interactive debugging)
	svcs.Debug = service.NewDebugService(logger, repos.DebugSession, svcs.Replay)

	// Regression detection service
	svcs.Regression = service.NewRegressionService(logger, repos.Regression, svcs.Dataset, svcs.Eval)

	// Tenant service (multi-tenancy and usage metering)
	svcs.Tenant = service.NewTenantService(logger, repos.Tenant)

	// Cost optimizer service
	svcs.CostOptimizer = service.NewCostOptimizerService(logger, repos.CostRecommendation, svcs.Cost, svcs.Query)

	// Agent graph service (multi-agent visualization)
	svcs.AgentGraph = service.NewAgentGraphService(logger, svcs.Query)

	// Migration service
	svcs.Migration = service.NewMigrationService(logger, repos.Migration, svcs.Ingestion, svcs.Prompt, svcs.Dataset)

	// Benchmark service
	svcs.Benchmark = service.NewBenchmarkService(logger, repos.Benchmark, svcs.Dataset, svcs.Eval)

	// Collaboration service (real-time)
	svcs.Collaboration = service.NewCollaborationService(logger, repos.Collaboration, svcs.Realtime)

	// Guardrail service
	svcs.Guardrail = service.NewGuardrailService(logger, repos.Guardrail, nil)

	// Wire guardrails into the ingestion pipeline
	svcs.Ingestion.SetGuardrailService(svcs.Guardrail)

	// OTel receiver service
	svcs.OTelReceiver = service.NewOTelReceiverService(logger, svcs.Ingestion)

	// Audit service (required by compliance)
	// Note: AuditRepository uses sqlx.DB — pass nil until migration to pgx
	svcs.Audit = service.NewAuditService(nil)

	// Compliance service (EU AI Act)
	svcs.Compliance = service.NewComplianceService(logger, nil, svcs.Audit)

	// Billing service (managed cloud)
	svcs.Billing = service.NewBillingService(logger, nil, svcs.Tenant)

	// Prediction service (predictive agent health)
	svcs.Prediction = service.NewPredictionService(logger, nil, svcs.Query, svcs.Cost)

	// Reasoning service (decision tree explorer)
	svcs.Reasoning = service.NewReasoningService(logger, svcs.Query)

	// Cost budget service
	svcs.CostBudget = service.NewCostBudgetService(logger, nil, svcs.Query, svcs.Cost)

	// Instrumentation service (framework setup)
	svcs.Instrumentation = service.NewInstrumentationService(logger)

	// Compliance export service
	svcs.ComplianceExport = service.NewComplianceExportService(logger, nil, svcs.Compliance, svcs.Audit)

	// Scorecard service
	svcs.Scorecard = service.NewScorecardService(logger, nil, svcs.Query)

	// Diff intelligence service
	svcs.DiffIntelligence = service.NewDiffIntelligenceService(logger, repos.DiffAnalysis)

	// Ticket service (trace-to-ticket pipeline)
	svcs.Ticket = service.NewTicketService(logger, nil, svcs.Query)

	// Streaming service (per-trace real-time streaming)
	svcs.Streaming = service.NewStreamingService(logger, svcs.Realtime)

	// Anomaly detection service
	svcs.Anomaly = service.NewAnomalyService(logger)

	// Federation service (cross-instance federation and export)
	svcs.Federation = service.NewFederationService(logger)

	return svcs
}
