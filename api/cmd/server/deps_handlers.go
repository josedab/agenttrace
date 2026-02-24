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
	// Core
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
	AgentGraph       *handler.AgentGraphHandler
	Benchmarks       *handler.BenchmarksHandler
	Collaboration    *handler.CollaborationHandler
	Migration        *handler.MigrationHandler
	OTelReceiver     *handler.OTelReceiverHandler
	CollaborationWS  *handler.CollaborationWSHandler
	Billing          *handler.BillingHandler
	Reasoning        *handler.ReasoningHandler
	Instrumentation  *handler.InstrumentationHandler
	Scorecard        *handler.ScorecardHandler
	Tickets          *handler.TicketHandler
	Streaming        *handler.StreamingHandler
	DiffIntelligence *handler.DiffIntelligenceHandler
	Anomaly          *handler.AnomalyHandler
	TeamIntelligence *handler.TeamIntelligenceHandler
	SemanticSearch   *handler.SemanticSearchHandler
	TrainingPipeline *handler.TrainingPipelineHandler
	Prediction       *handler.PredictionHandler
	Annotation       *handler.AnnotationHandler
	Handoff          *handler.HandoffHandler
	MultiModal       *handler.MultiModalHandler

	// Cost
	CostOptimizer   *handler.CostOptimizerHandler
	CostBudget      *handler.CostBudgetHandler
	PredictiveCost  *handler.PredictiveCostHandler
	CostAttribution *handler.CostAttributionHandler
	Carbon          *handler.CarbonHandler

	// Compliance & Security
	Compliance        *handler.ComplianceHandler
	ComplianceExport  *handler.ComplianceExportHandler
	ComplianceReport  *handler.ComplianceReportHandler
	ComplianceMonitor *handler.ComplianceMonitorHandler
	Guardrails        *handler.GuardrailsHandler
	Privacy           *handler.PrivacyHandler
	RBAC              *handler.RBACHandler

	// Agents
	AgentBuilder    *handler.AgentBuilderHandler
	AgentVersion    *handler.AgentVersionHandler
	AgentMemory     *handler.AgentMemoryHandler
	Autonomy        *handler.AutonomyHandler
	Fleet           *handler.FleetHandler
	Copilot         *handler.CopilotHandler
	KnowledgeGraph  *handler.KnowledgeGraphHandler
	Intent          *handler.IntentHandler
	SLO             *handler.SLOHandler

	// Collaboration
	CollabPattern *handler.CollabPatternHandler
	CrossOrg      *handler.CrossOrgHandler

	// Infrastructure
	Federation           *handler.FederationHandler
	FederatedLearning    *handler.FederatedLearningHandler
	WebhookOrchestration *handler.WebhookOrchestrationHandler
	Marketplace          *handler.MarketplaceHandler
	Mobile               *handler.MobileHandler
	Plugin               *handler.PluginHandler
	OrchestrationDebugger *handler.OrchestrationDebuggerHandler
	RCA                   *handler.RCAHandler
	Embed                 *handler.EmbedHandler
	DistributedTrace      *handler.DistributedTraceHandler
	PromptCache           *handler.PromptCacheHandler
	Chaos                 *handler.ChaosHandler
	CustomMetrics         *handler.CustomMetricsHandler
	SyntheticData         *handler.SyntheticDataHandler

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

	// Next-Gen Features (v8)
	AgentComparison     *handler.AgentComparisonHandler
	StreamingWS         *handler.StreamingWSHandler
	RegressionDetection *handler.RegressionDetectionHandler
	CodeQuality         *handler.CodeQualityHandler
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
	h := &Handlers{}

	initCoreHandlers(h, logger, svcs, repos, pgDB, chDB, redisClient, asynqClient, version)
	initCostHandlers(h, logger, svcs)
	initComplianceHandlers(h, logger, svcs)
	initAgentHandlers(h, logger, svcs)
	initCollabHandlers(h, logger, svcs)
	initInfraHandlers(h, logger, svcs)
	initV6Handlers(h, logger, svcs)
	initV7Handlers(h, logger, svcs)
	initV8Handlers(h, logger, svcs)

	return h
}

// initCoreHandlers initializes core platform handlers
func initCoreHandlers(
	h *Handlers,
	logger *zap.Logger,
	svcs *Services,
	repos *Repositories,
	pgDB *database.PostgresDB,
	chDB *database.ClickHouseDB,
	redisClient *redis.Client,
	asynqClient *asynq.Client,
	version string,
) {
	h.Health = handler.NewHealthHandler(pgDB.Pool, chDB.Conn, redisClient, version)
	h.Ingestion = handler.NewIngestionHandler(svcs.Ingestion, svcs.Score, logger)
	h.Traces = handler.NewTracesHandler(svcs.Query, logger)
	h.Scores = handler.NewScoresHandler(svcs.Score, logger)
	h.Prompts = handler.NewPromptsHandler(svcs.Prompt, logger)
	h.Datasets = handler.NewDatasetsHandler(svcs.Dataset, logger)
	h.Evaluators = handler.NewEvaluatorsHandler(svcs.Eval, logger)
	h.Events = handler.NewEventsHandler(svcs.Realtime, logger)
	h.APIKeys = handler.NewAPIKeysHandler(svcs.Auth, logger)
	h.Projects = handler.NewProjectsHandler(svcs.Project, logger)
	h.Organizations = handler.NewOrganizationsHandler(svcs.Org, logger)
	h.Auth = handler.NewAuthHandler(svcs.Auth, logger)
	h.Checkpoints = handler.NewCheckpointsHandler(svcs.Checkpoint, logger)
	h.GitLinks = handler.NewGitLinksHandler(svcs.GitLink, logger)
	h.FileOperations = handler.NewFileOperationsHandler(svcs.FileOperation, logger)
	h.TerminalCommands = handler.NewTerminalCommandsHandler(svcs.TerminalCommand, logger)
	h.CIRuns = handler.NewCIRunsHandler(svcs.CIRun, logger)
	h.Export = handler.NewExportHandler(asynqClient, logger)
	h.Import = handler.NewImportHandler(svcs.Dataset, svcs.Prompt, logger)
	h.Docs = handler.NewDocsHandler()
	h.Webhook = handler.NewWebhookHandler(logger, repos.Webhook, nil)
	h.Replay = handler.NewReplayHandler(logger, svcs.Replay)
	h.Experiment = handler.NewExperimentHandler(logger, svcs.Experiment)
	h.Debug = handler.NewDebugHandler(svcs.Debug, logger)
	h.Regression = handler.NewRegressionHandler(svcs.Regression, logger)
	h.AgentGraph = handler.NewAgentGraphHandler(svcs.AgentGraph, logger)
	h.Benchmarks = handler.NewBenchmarksHandler(svcs.Benchmark, logger)
	h.Collaboration = handler.NewCollaborationHandler(svcs.Collaboration, logger)
	h.Migration = handler.NewMigrationHandler(svcs.Migration, logger)
	h.OTelReceiver = handler.NewOTelReceiverHandler(svcs.OTelReceiver, logger)
	h.CollaborationWS = handler.NewCollaborationWSHandler(logger, svcs.Collaboration)
	h.Billing = handler.NewBillingHandler(svcs.Billing, logger)
	h.Reasoning = handler.NewReasoningHandler(svcs.Reasoning, logger)
	h.Instrumentation = handler.NewInstrumentationHandler(svcs.Instrumentation, logger)
	h.Scorecard = handler.NewScorecardHandler(svcs.Scorecard, logger)
	h.Tickets = handler.NewTicketHandler(svcs.Ticket, logger)
	h.DiffIntelligence = handler.NewDiffIntelligenceHandler(svcs.DiffIntelligence, logger)
	h.Streaming = handler.NewStreamingHandler(svcs.Streaming, logger)
	h.Anomaly = handler.NewAnomalyHandler(logger, svcs.Anomaly)
	h.TeamIntelligence = handler.NewTeamIntelligenceHandler(logger, svcs.TeamIntelligence)
	h.SemanticSearch = handler.NewSemanticSearchHandler(logger, svcs.SemanticSearch)
	h.TrainingPipeline = handler.NewTrainingPipelineHandler(logger, svcs.TrainingPipeline)
	h.Prediction = handler.NewPredictionHandler(svcs.Prediction, logger)
	h.Annotation = handler.NewAnnotationHandler(svcs.Annotation, logger)
	h.Handoff = handler.NewHandoffHandler(svcs.Handoff, logger)
	h.MultiModal = handler.NewMultiModalHandler(svcs.MultiModal, logger)
}

// initCostHandlers initializes cost-related handlers
func initCostHandlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.CostOptimizer = handler.NewCostOptimizerHandler(svcs.CostOptimizer, logger)
	h.CostBudget = handler.NewCostBudgetHandler(svcs.CostBudget, logger)
	h.PredictiveCost = handler.NewPredictiveCostHandler(svcs.PredictiveCost, logger)
	h.CostAttribution = handler.NewCostAttributionHandler(svcs.CostAttribution, logger)
	h.Carbon = handler.NewCarbonHandler(svcs.Carbon, logger)
}

// initComplianceHandlers initializes compliance and security handlers
func initComplianceHandlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.Compliance = handler.NewComplianceHandler(svcs.Compliance, logger)
	h.ComplianceExport = handler.NewComplianceExportHandler(svcs.ComplianceExport, logger)
	h.ComplianceReport = handler.NewComplianceReportHandler(svcs.ComplianceReport, logger)
	h.ComplianceMonitor = handler.NewComplianceMonitorHandler(svcs.ComplianceMonitor, logger)
	h.Guardrails = handler.NewGuardrailsHandler(svcs.Guardrail, logger)
	h.Privacy = handler.NewPrivacyHandler(svcs.Privacy, logger)
	h.RBAC = handler.NewRBACHandler(logger, svcs.RBAC)
}

// initAgentHandlers initializes agent management handlers
func initAgentHandlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.AgentBuilder = handler.NewAgentBuilderHandler(svcs.AgentBuilder, logger)
	h.AgentVersion = handler.NewAgentVersionHandler(svcs.AgentVersion, logger)
	h.AgentMemory = handler.NewAgentMemoryHandler(svcs.AgentMemory, logger)
	h.Autonomy = handler.NewAutonomyHandler(svcs.Autonomy, logger)
	h.Fleet = handler.NewFleetHandler(svcs.Fleet, logger)
	h.Copilot = handler.NewCopilotHandler(svcs.Copilot, logger)
	h.KnowledgeGraph = handler.NewKnowledgeGraphHandler(svcs.KnowledgeGraph, logger)
	h.Intent = handler.NewIntentHandler(svcs.Intent, logger)
	h.SLO = handler.NewSLOHandler(svcs.SLO, logger)
}

// initCollabHandlers initializes collaboration handlers
func initCollabHandlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.CollabPattern = handler.NewCollabPatternHandler(svcs.CollabPattern, logger)
	h.CrossOrg = handler.NewCrossOrgHandler(svcs.CrossOrg, logger)
}

// initInfraHandlers initializes infrastructure handlers
func initInfraHandlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.Federation = handler.NewFederationHandler(svcs.Federation, logger)
	h.FederatedLearning = handler.NewFederatedLearningHandler(svcs.FederatedLearning, logger)
	h.WebhookOrchestration = handler.NewWebhookOrchestrationHandler(svcs.WebhookOrchestration, logger)
	h.Marketplace = handler.NewMarketplaceHandler(svcs.Marketplace, logger)
	h.Mobile = handler.NewMobileHandler(svcs.Mobile, logger)
	h.Plugin = handler.NewPluginHandler(svcs.Plugin, logger)
	h.OrchestrationDebugger = handler.NewOrchestrationDebuggerHandler(svcs.OrchestrationDebugger, logger)
	h.RCA = handler.NewRCAHandler(svcs.RCA, logger)
	h.Embed = handler.NewEmbedHandler(svcs.Embed, logger)
	h.DistributedTrace = handler.NewDistributedTraceHandler(svcs.DistributedTrace, logger)
	h.PromptCache = handler.NewPromptCacheHandler(svcs.PromptCache, logger)
	h.Chaos = handler.NewChaosHandler(svcs.Chaos, logger)
	h.CustomMetrics = handler.NewCustomMetricsHandler(svcs.CustomMetrics, logger)
	h.SyntheticData = handler.NewSyntheticDataHandler(svcs.SyntheticData, logger)
}

// initV6Handlers initializes Next-Gen v6 feature handlers
func initV6Handlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.ReplaySession = handler.NewReplaySessionHandler(svcs.ReplaySession, logger)
	h.CostGuardrail = handler.NewCostGuardrailHandler(svcs.CostGuardrail, logger)
	h.MultiAgentGraph = handler.NewMultiAgentGraphHandler(svcs.MultiAgentGraph, logger)
	h.PromptCI = handler.NewPromptCIHandler(svcs.PromptCI, logger)
	h.AgentBenchmark = handler.NewAgentBenchmarkHandler(svcs.AgentBenchmark, logger)
	h.SemanticTraceSearch = handler.NewSemanticTraceSearchHandler(svcs.SemanticTraceSearch, logger)
	h.AgentKnowledgeGraph = handler.NewAgentKnowledgeGraphHandler(svcs.AgentKnowledgeGraph, logger)
	h.IDETraceView = handler.NewIDETraceViewHandler(svcs.IDETraceView, logger)
	h.FederatedAggregation = handler.NewFederatedAggregationHandler(svcs.FederatedAggregation, logger)
}

// initV7Handlers initializes Next-Gen v7 feature handlers
func initV7Handlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.WorkflowSimulator = handler.NewWorkflowSimulatorHandler(svcs.WorkflowSimulator, logger)
	h.AutoDiscovery = handler.NewAutoDiscoveryHandler(svcs.AutoDiscovery, logger)
	h.CloudOnboarding = handler.NewCloudOnboardingHandler(svcs.CloudOnboarding, logger)
	h.AIDebugger = handler.NewAIDebuggerHandler(svcs.AIDebugger, logger)
	h.PromptOptimization = handler.NewPromptOptimizationHandler(svcs.PromptOptimization, logger)
	h.CostAlerting = handler.NewCostAlertingHandler(svcs.CostAlerting, logger)
	h.RegressionSuite = handler.NewRegressionSuiteHandler(svcs.RegressionSuite, logger)
	h.CollabHub = handler.NewCollabHubHandler(svcs.CollabHub, logger)
	h.OTelCompat = handler.NewOTelCompatHandler(svcs.OTelCompat, logger)
	h.SecurityScanner = handler.NewSecurityScannerHandler(svcs.SecurityScanner, logger)
}

// initV8Handlers initializes Next-Gen v8 feature handlers
func initV8Handlers(h *Handlers, logger *zap.Logger, svcs *Services) {
	h.AgentComparison = handler.NewAgentComparisonHandler(svcs.AgentComparison, logger)
	h.StreamingWS = handler.NewStreamingWSHandler(logger, svcs.Streaming)
	h.RegressionDetection = handler.NewRegressionDetectionHandler(svcs.RegressionDetection, logger)
	h.CodeQuality = handler.NewCodeQualityHandler(logger, svcs.CodeQuality)
}
