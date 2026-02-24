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
	Handoff        *handler.HandoffHandler
	Annotation     *handler.AnnotationHandler
	Carbon         *handler.CarbonHandler
	SyntheticData  *handler.SyntheticDataHandler
	SLO            *handler.SLOHandler
	AgentMemory      *handler.AgentMemoryHandler
	DistributedTrace *handler.DistributedTraceHandler
	PromptCache      *handler.PromptCacheHandler
	Chaos            *handler.ChaosHandler
	CustomMetrics    *handler.CustomMetricsHandler
	Autonomy        *handler.AutonomyHandler
	CrossOrg        *handler.CrossOrgHandler
	Intent          *handler.IntentHandler
	CostAttribution *handler.CostAttributionHandler
	KnowledgeGraph  *handler.KnowledgeGraphHandler
	ComplianceMonitor  *handler.ComplianceMonitorHandler
	MultiModal         *handler.MultiModalHandler
	CollabPattern      *handler.CollabPatternHandler
	FederatedLearning  *handler.FederatedLearningHandler
	Copilot            *handler.CopilotHandler

	// Next-Gen Features (v6)
	ReplaySession        *handler.ReplaySessionHandler
	CostGuardrail        *handler.CostGuardrailHandler
	MultiAgentGraph      *handler.MultiAgentGraphHandler
	PromptCI             *handler.PromptCIHandler
	AgentBenchmark       *handler.AgentBenchmarkHandler
	SemanticTraceSearch  *handler.SemanticTraceSearchHandler
	AgentKnowledgeGraph  *handler.AgentKnowledgeGraphHandler
	IDETraceView         *handler.IDETraceViewHandler
	FederatedAggregation *handler.FederatedAggregationHandler

	// Next-Gen Features (v7)
	WorkflowSimulator  *handler.WorkflowSimulatorHandler
	AutoDiscovery      *handler.AutoDiscoveryHandler
	CloudOnboarding    *handler.CloudOnboardingHandler
	AIDebugger         *handler.AIDebuggerHandler
	PromptOptimization *handler.PromptOptimizationHandler
	CostAlerting       *handler.CostAlertingHandler
	RegressionSuite    *handler.RegressionSuiteHandler
	CollabHub          *handler.CollabHubHandler
	OTelCompat         *handler.OTelCompatHandler
	SecurityScanner    *handler.SecurityScannerHandler
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
		Handoff:       handler.NewHandoffHandler(svcs.Handoff, logger),
		Annotation:    handler.NewAnnotationHandler(svcs.Annotation, logger),
		Carbon:        handler.NewCarbonHandler(svcs.Carbon, logger),
		SyntheticData: handler.NewSyntheticDataHandler(svcs.SyntheticData, logger),
		SLO:           handler.NewSLOHandler(svcs.SLO, logger),
		AgentMemory:      handler.NewAgentMemoryHandler(svcs.AgentMemory, logger),
		DistributedTrace: handler.NewDistributedTraceHandler(svcs.DistributedTrace, logger),
		PromptCache:      handler.NewPromptCacheHandler(svcs.PromptCache, logger),
		Chaos:            handler.NewChaosHandler(svcs.Chaos, logger),
		CustomMetrics:    handler.NewCustomMetricsHandler(svcs.CustomMetrics, logger),
		Autonomy:        handler.NewAutonomyHandler(svcs.Autonomy, logger),
		CrossOrg:        handler.NewCrossOrgHandler(svcs.CrossOrg, logger),
		Intent:          handler.NewIntentHandler(svcs.Intent, logger),
		CostAttribution: handler.NewCostAttributionHandler(svcs.CostAttribution, logger),
		KnowledgeGraph:  handler.NewKnowledgeGraphHandler(svcs.KnowledgeGraph, logger),
		ComplianceMonitor:  handler.NewComplianceMonitorHandler(svcs.ComplianceMonitor, logger),
		MultiModal:         handler.NewMultiModalHandler(svcs.MultiModal, logger),
		CollabPattern:      handler.NewCollabPatternHandler(svcs.CollabPattern, logger),
		FederatedLearning:  handler.NewFederatedLearningHandler(svcs.FederatedLearning, logger),
		Copilot:            handler.NewCopilotHandler(svcs.Copilot, logger),

		// Next-Gen Features (v6)
		ReplaySession:        handler.NewReplaySessionHandler(svcs.ReplaySession, logger),
		CostGuardrail:        handler.NewCostGuardrailHandler(svcs.CostGuardrail, logger),
		MultiAgentGraph:      handler.NewMultiAgentGraphHandler(svcs.MultiAgentGraph, logger),
		PromptCI:             handler.NewPromptCIHandler(svcs.PromptCI, logger),
		AgentBenchmark:       handler.NewAgentBenchmarkHandler(svcs.AgentBenchmark, logger),
		SemanticTraceSearch:  handler.NewSemanticTraceSearchHandler(svcs.SemanticTraceSearch, logger),
		AgentKnowledgeGraph:  handler.NewAgentKnowledgeGraphHandler(svcs.AgentKnowledgeGraph, logger),
		IDETraceView:         handler.NewIDETraceViewHandler(svcs.IDETraceView, logger),
		FederatedAggregation: handler.NewFederatedAggregationHandler(svcs.FederatedAggregation, logger),

		// Next-Gen Features (v7)
		WorkflowSimulator:  handler.NewWorkflowSimulatorHandler(svcs.WorkflowSimulator, logger),
		AutoDiscovery:      handler.NewAutoDiscoveryHandler(svcs.AutoDiscovery, logger),
		CloudOnboarding:    handler.NewCloudOnboardingHandler(svcs.CloudOnboarding, logger),
		AIDebugger:         handler.NewAIDebuggerHandler(svcs.AIDebugger, logger),
		PromptOptimization: handler.NewPromptOptimizationHandler(svcs.PromptOptimization, logger),
		CostAlerting:       handler.NewCostAlertingHandler(svcs.CostAlerting, logger),
		RegressionSuite:    handler.NewRegressionSuiteHandler(svcs.RegressionSuite, logger),
		CollabHub:          handler.NewCollabHubHandler(svcs.CollabHub, logger),
		OTelCompat:         handler.NewOTelCompatHandler(svcs.OTelCompat, logger),
		SecurityScanner:    handler.NewSecurityScannerHandler(svcs.SecurityScanner, logger),
	}
}
