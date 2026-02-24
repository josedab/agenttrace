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
	Debug            *handler.DebugHandler
	Regression       *handler.RegressionHandler
	CostOptimizer    *handler.CostOptimizerHandler
	AgentGraph       *handler.AgentGraphHandler
	Guardrails       *handler.GuardrailsHandler
	Benchmarks       *handler.BenchmarksHandler
	Collaboration    *handler.CollaborationHandler
	Migration        *handler.MigrationHandler
	OTelReceiver     *handler.OTelReceiverHandler
	CollaborationWS    *handler.CollaborationWSHandler
	Compliance         *handler.ComplianceHandler
	Billing            *handler.BillingHandler
	Prediction         *handler.PredictionHandler
	Reasoning          *handler.ReasoningHandler
	CostBudget         *handler.CostBudgetHandler
	Instrumentation    *handler.InstrumentationHandler
	ComplianceExport   *handler.ComplianceExportHandler
	Scorecard          *handler.ScorecardHandler
	Tickets            *handler.TicketHandler
	Streaming          *handler.StreamingHandler
	DiffIntelligence   *handler.DiffIntelligenceHandler
	Anomaly            *handler.AnomalyHandler
	Federation         *handler.FederationHandler
	TeamIntelligence   *handler.TeamIntelligenceHandler
	SemanticSearch     *handler.SemanticSearchHandler
	TrainingPipeline   *handler.TrainingPipelineHandler
	RBAC                 *handler.RBACHandler
	WebhookOrchestration *handler.WebhookOrchestrationHandler
	Marketplace          *handler.MarketplaceHandler
	ComplianceReport     *handler.ComplianceReportHandler
	AgentBuilder         *handler.AgentBuilderHandler
	Fleet                *handler.FleetHandler
	Privacy              *handler.PrivacyHandler
	Mobile               *handler.MobileHandler
	Plugin                *handler.PluginHandler
	OrchestrationDebugger *handler.OrchestrationDebuggerHandler
	RCA                   *handler.RCAHandler
	AgentVersion          *handler.AgentVersionHandler
	PredictiveCost        *handler.PredictiveCostHandler
	Embed                 *handler.EmbedHandler
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
		Debug: handler.NewDebugHandler(
			svcs.Debug,
			logger,
		),
		Regression: handler.NewRegressionHandler(
			svcs.Regression,
			logger,
		),
		CostOptimizer: handler.NewCostOptimizerHandler(
			svcs.CostOptimizer,
			logger,
		),
		AgentGraph: handler.NewAgentGraphHandler(
			svcs.AgentGraph,
			logger,
		),
		Guardrails: handler.NewGuardrailsHandler(
			svcs.Guardrail,
			logger,
		),
		Benchmarks: handler.NewBenchmarksHandler(
			svcs.Benchmark,
			logger,
		),
		Collaboration: handler.NewCollaborationHandler(
			svcs.Collaboration,
			logger,
		),
		Migration: handler.NewMigrationHandler(
			svcs.Migration,
			logger,
		),
		OTelReceiver: handler.NewOTelReceiverHandler(
			svcs.OTelReceiver,
			logger,
		),
		CollaborationWS: handler.NewCollaborationWSHandler(
			logger,
			svcs.Collaboration,
		),
		Compliance: handler.NewComplianceHandler(
			svcs.Compliance,
			logger,
		),
		Billing: handler.NewBillingHandler(
			svcs.Billing,
			logger,
		),
		Prediction: handler.NewPredictionHandler(
			svcs.Prediction,
			logger,
		),
		Reasoning: handler.NewReasoningHandler(
			svcs.Reasoning,
			logger,
		),
		CostBudget: handler.NewCostBudgetHandler(
			svcs.CostBudget,
			logger,
		),
		Instrumentation: handler.NewInstrumentationHandler(
			svcs.Instrumentation,
			logger,
		),
		ComplianceExport: handler.NewComplianceExportHandler(
			svcs.ComplianceExport,
			logger,
		),
		Scorecard: handler.NewScorecardHandler(
			svcs.Scorecard,
			logger,
		),
		Tickets: handler.NewTicketHandler(
			svcs.Ticket,
			logger,
		),
		DiffIntelligence: handler.NewDiffIntelligenceHandler(
			svcs.DiffIntelligence,
			logger,
		),
		Streaming: handler.NewStreamingHandler(
			svcs.Streaming,
			logger,
		),
		Anomaly: handler.NewAnomalyHandler(
			logger,
			svcs.Anomaly,
		),
		Federation: handler.NewFederationHandler(
			svcs.Federation,
			logger,
		),
		TeamIntelligence: handler.NewTeamIntelligenceHandler(
			logger,
			svcs.TeamIntelligence,
		),
		SemanticSearch: handler.NewSemanticSearchHandler(
			logger,
			svcs.SemanticSearch,
		),
		TrainingPipeline: handler.NewTrainingPipelineHandler(
			logger,
			svcs.TrainingPipeline,
		),
		RBAC: handler.NewRBACHandler(
			logger,
			svcs.RBAC,
		),
		WebhookOrchestration: handler.NewWebhookOrchestrationHandler(
			svcs.WebhookOrchestration,
			logger,
		),
		Marketplace: handler.NewMarketplaceHandler(
			svcs.Marketplace,
			logger,
		),
		ComplianceReport: handler.NewComplianceReportHandler(
			svcs.ComplianceReport,
			logger,
		),
		AgentBuilder: handler.NewAgentBuilderHandler(
			svcs.AgentBuilder,
			logger,
		),
		Fleet: handler.NewFleetHandler(
			svcs.Fleet,
			logger,
		),
		Privacy: handler.NewPrivacyHandler(
			svcs.Privacy,
			logger,
		),
		Mobile: handler.NewMobileHandler(
			svcs.Mobile,
			logger,
		),
		Plugin: handler.NewPluginHandler(
			svcs.Plugin,
			logger,
		),
		OrchestrationDebugger: handler.NewOrchestrationDebuggerHandler(svcs.OrchestrationDebugger, logger),
		RCA:                   handler.NewRCAHandler(svcs.RCA, logger),
		AgentVersion:          handler.NewAgentVersionHandler(svcs.AgentVersion, logger),
		PredictiveCost:        handler.NewPredictiveCostHandler(svcs.PredictiveCost, logger),
		Embed:                 handler.NewEmbedHandler(svcs.Embed, logger),
	}
}
