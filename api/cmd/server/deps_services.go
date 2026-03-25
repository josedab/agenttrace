package main

import (
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/config"
	"github.com/agenttrace/agenttrace/api/internal/service"
)

// Services holds all service instances
type Services struct {
	// Core
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
	AgentGraph      *service.AgentGraphService
	Migration       *service.MigrationService
	Benchmark       *service.BenchmarkService
	Collaboration   *service.CollaborationService
	Guardrail       *service.GuardrailService
	OTelReceiver    *service.OTelReceiverService
	Billing         *service.BillingService
	Reasoning       *service.ReasoningService
	Instrumentation *service.InstrumentationService
	Scorecard       *service.ScorecardService
	Ticket          *service.TicketService
	Audit           *service.AuditService
	Streaming       *service.StreamingService
	DiffIntelligence *service.DiffIntelligenceService
	Anomaly          *service.AnomalyService
	TeamIntelligence *service.TeamIntelligenceService
	SemanticSearch   *service.SemanticSearchService
	TrainingPipeline *service.TrainingPipelineService
	Prediction       *service.PredictionService
	Annotation       *service.AnnotationService
	Handoff          *service.HandoffService
	MultiModal       *service.MultiModalService
	Embedding        *service.EmbeddingService

	// Cost
	CostOptimizer   *service.CostOptimizerService
	CostBudget      *service.CostBudgetService
	PredictiveCost  *service.PredictiveCostService
	CostAttribution *service.CostAttributionService
	Carbon          *service.CarbonService

	// Compliance & Security
	Compliance       *service.ComplianceService
	ComplianceExport *service.ComplianceExportService
	ComplianceReport *service.ComplianceReportService
	ComplianceMonitor *service.ComplianceMonitorService
	Privacy           *service.PrivacyService
	RBAC              *service.RBACService

	// Agents
	AgentBuilder    *service.AgentBuilderService
	AgentVersion    *service.AgentVersionService
	AgentMemory     *service.AgentMemoryService
	Autonomy        *service.AutonomyService
	Fleet           *service.FleetService
	Copilot         *service.CopilotService
	KnowledgeGraph  *service.KnowledgeGraphService
	Intent          *service.IntentService
	SLO             *service.SLOService

	// Collaboration
	CollabPattern     *service.CollabPatternService
	CrossOrg          *service.CrossOrgService

	// Infrastructure
	Federation            *service.FederationService
	FederatedLearning     *service.FederatedLearningService
	WebhookOrchestration  *service.WebhookOrchestrationService
	Marketplace           *service.MarketplaceService
	Mobile                *service.MobileService
	Plugin                *service.PluginService
	OrchestrationDebugger *service.OrchestrationDebuggerService
	RCA                   *service.RCAService
	Embed                 *service.EmbedService
	DistributedTrace      *service.DistributedTraceService
	PromptCache           *service.PromptCacheService
	Chaos                 *service.ChaosService
	CustomMetrics         *service.CustomMetricsService
	SyntheticData         *service.SyntheticDataService

	LLM *service.LLMClient

	// Next-Gen Features (v6)
	ReplaySession        *service.ReplaySessionService
	CostGuardrail        *service.CostGuardrailService
	MultiAgentGraph      *service.MultiAgentGraphService
	PromptCI             *service.PromptCIService
	AgentBenchmark       *service.AgentBenchmarkService
	SemanticTraceSearch  *service.SemanticTraceSearchService
	AgentKnowledgeGraph  *service.AgentKnowledgeGraphService
	IDETraceView         *service.IDETraceViewService
	FederatedAggregation *service.FederatedAggregationService
	EmbeddingPipeline    *service.EmbeddingPipelineService
	StripeWebhook        *service.StripeWebhookService
	UsageMetering        *service.UsageMeteringService

	// Next-Gen Features (v7)
	WorkflowSimulator  *service.WorkflowSimulatorService
	AutoDiscovery      *service.AutoDiscoveryService
	CloudOnboarding    *service.CloudOnboardingService
	AIDebugger         *service.AIDebuggerService
	PromptOptimization *service.PromptOptimizationService
	CostAlerting       *service.CostAlertingService
	RegressionSuite    *service.RegressionSuiteService
	CollabHub          *service.CollabHubService
	OTelCompat         *service.OTelCompatService
	SecurityScanner    *service.SecurityScannerService

	// Next-Gen Features (v8)
	AgentComparison     *service.AgentComparisonService
	RegressionDetection *service.RegressionDetectionService
	CodeQuality         *service.CodeQualityService

	// Next-Gen Features (v9)
	SkillProfile *service.SkillProfileService
	NLQuery      *service.NLQueryService
	Sandbox      *service.SandboxService

	// Next-Gen Features (v10)
	Gateway          *service.GatewayService
	TestGenerator    *service.TestGeneratorService
	CanaryDeployment *service.CanaryDeploymentService
	CertExport       *service.CertificationExportService
	WarehouseSync    *service.WarehouseSyncService
	TraceReview      *service.TraceReviewService
	RunbookEngine    *service.RunbookEngineService
	EdgeIngest       *service.EdgeIngestService
	PromptImpact     *service.PromptImpactService

	// Next-Gen Features (v11)
	TraceDiff          *service.TraceDiffService
	EvalPlayground     *service.EvalPlaygroundService
	CodeImpact         *service.CodeImpactService
	StreamingDashboard *service.StreamingDashboardService
	EvalMarketplace    *service.EvalMarketplaceService
	SessionJourney     *service.SessionJourneyService
	TraceEnrichment    *service.TraceEnrichmentService
	CostForecast       *service.CostForecastService
	TraceKB            *service.TraceKBService
	OTelBridge         *service.OTelBridgeService

	// Next-Gen Features (v12)
	Adapter   *service.AdapterService
	ABTesting *service.ABTestingService
}

// initServices initializes all services
func initServices(cfg *config.Config, logger *zap.Logger, repos *Repositories) *Services {
	svcs := &Services{}

	initCoreServices(svcs, cfg, logger, repos)
	initCostServices(svcs, logger, repos)
	initComplianceServices(svcs, logger)
	initAgentServices(svcs, logger)
	initCollabServices(svcs, logger)
	initInfraServices(svcs, logger)
	initV6Services(svcs, cfg, logger)
	initV7Services(svcs, logger)
	initV8Services(svcs, logger)
	initV10Services(svcs, logger)
	initV11Services(svcs, logger)
	initV12Services(svcs, logger)

	return svcs
}

// initCoreServices initializes core platform services
func initCoreServices(svcs *Services, cfg *config.Config, logger *zap.Logger, repos *Repositories) {
	svcs.LLM = service.NewLLMClient(logger)
	svcs.Cost = service.NewCostService(logger)

	svcs.Query = service.NewQueryService(repos.Trace, repos.Observation, repos.Score, repos.Session)
	svcs.Score = service.NewScoreService(repos.Score, repos.Trace, repos.Observation)
	svcs.Ingestion = service.NewIngestionService(logger, repos.Trace, repos.Observation, svcs.Cost, nil)
	svcs.Prompt = service.NewPromptService(repos.Prompt)
	svcs.Dataset = service.NewDatasetService(repos.Dataset, repos.Trace, repos.Score)
	svcs.Eval = service.NewEvalService(repos.Evaluator, svcs.Score)
	svcs.Auth = service.NewAuthService(cfg, repos.User, repos.APIKey, repos.Org, repos.Project)
	svcs.Auth.SetLogger(logger)
	svcs.Org = service.NewOrgService(repos.Org)
	svcs.Project = service.NewProjectService(repos.Project, repos.Org)
	svcs.Realtime = service.NewRealtimeService()
	svcs.Checkpoint = service.NewCheckpointService(repos.Checkpoint, repos.Trace)
	svcs.GitLink = service.NewGitLinkService(repos.GitLink, repos.Trace)
	svcs.FileOperation = service.NewFileOperationService(repos.FileOperation, repos.Trace)
	svcs.TerminalCommand = service.NewTerminalCommandService(repos.TerminalCommand, repos.Trace)
	svcs.CIRun = service.NewCIRunService(repos.CIRun)
	svcs.Replay = service.NewReplayService(logger, repos.Trace, repos.Observation, repos.FileOperation, repos.TerminalCommand, repos.Checkpoint, repos.GitLink)
	svcs.Experiment = service.NewExperimentService(logger, repos.Experiment)
	svcs.Debug = service.NewDebugService(logger, repos.DebugSession, svcs.Replay)
	svcs.Regression = service.NewRegressionService(logger, repos.Regression, svcs.Dataset, svcs.Eval)
	svcs.Tenant = service.NewTenantService(logger, repos.Tenant)
	svcs.AgentGraph = service.NewAgentGraphService(logger, svcs.Query)
	svcs.Migration = service.NewMigrationService(logger, repos.Migration, svcs.Ingestion, svcs.Prompt, svcs.Dataset)
	svcs.Benchmark = service.NewBenchmarkService(logger, repos.Benchmark, svcs.Dataset, svcs.Eval)
	svcs.Collaboration = service.NewCollaborationService(logger, repos.Collaboration, svcs.Realtime)
	svcs.Guardrail = service.NewGuardrailService(logger, repos.Guardrail, nil)
	svcs.Ingestion.SetGuardrailService(svcs.Guardrail)
	svcs.OTelReceiver = service.NewOTelReceiverService(logger, svcs.Ingestion)
	svcs.Audit = service.NewAuditService(nil)
	svcs.Billing = service.NewBillingService(logger, nil, svcs.Tenant)
	svcs.Reasoning = service.NewReasoningService(logger, svcs.Query)
	svcs.Instrumentation = service.NewInstrumentationService(logger)
	svcs.Scorecard = service.NewScorecardService(logger, nil, svcs.Query)
	svcs.DiffIntelligence = service.NewDiffIntelligenceService(logger, repos.DiffAnalysis)
	svcs.Ticket = service.NewTicketService(logger, nil, svcs.Query)
	svcs.Streaming = service.NewStreamingService(logger, svcs.Realtime)
	svcs.Anomaly = service.NewAnomalyService(logger)
	svcs.TeamIntelligence = service.NewTeamIntelligenceService(logger, svcs.Query, svcs.Cost)
	svcs.SemanticSearch = service.NewSemanticSearchService(logger, svcs.Query)
	svcs.TrainingPipeline = service.NewTrainingPipelineService(logger)
	svcs.Prediction = service.NewPredictionService(logger, nil, svcs.Query, svcs.Cost)
	svcs.Annotation = service.NewAnnotationService(logger)
	svcs.Handoff = service.NewHandoffService(logger)
	svcs.MultiModal = service.NewMultiModalService(logger)
	svcs.Embedding = service.NewEmbeddingService(logger)
}

// initCostServices initializes cost-related services
func initCostServices(svcs *Services, logger *zap.Logger, repos *Repositories) {
	svcs.CostOptimizer = service.NewCostOptimizerService(logger, repos.CostRecommendation, svcs.Cost, svcs.Query)
	svcs.CostBudget = service.NewCostBudgetService(logger, nil, svcs.Query, svcs.Cost)
	svcs.PredictiveCost = service.NewPredictiveCostService(logger)
	svcs.CostAttribution = service.NewCostAttributionService(logger)
	svcs.Carbon = service.NewCarbonService(logger)
}

// initComplianceServices initializes compliance and security services
func initComplianceServices(svcs *Services, logger *zap.Logger) {
	svcs.Compliance = service.NewComplianceService(logger, nil, svcs.Audit)
	svcs.ComplianceExport = service.NewComplianceExportService(logger, nil, svcs.Compliance, svcs.Audit)
	svcs.ComplianceReport = service.NewComplianceReportService(logger)
	svcs.ComplianceMonitor = service.NewComplianceMonitorService(logger)
	svcs.Privacy = service.NewPrivacyService(logger)
	svcs.RBAC = service.NewRBACService(logger)
}

// initAgentServices initializes agent management services
func initAgentServices(svcs *Services, logger *zap.Logger) {
	svcs.AgentBuilder = service.NewAgentBuilderService(logger)
	svcs.AgentVersion = service.NewAgentVersionService(logger)
	svcs.AgentMemory = service.NewAgentMemoryService(logger)
	svcs.Autonomy = service.NewAutonomyService(logger)
	svcs.Fleet = service.NewFleetService(logger)
	svcs.Copilot = service.NewCopilotService(logger)
	svcs.KnowledgeGraph = service.NewKnowledgeGraphService(logger)
	svcs.Intent = service.NewIntentService(logger)
	svcs.SLO = service.NewSLOService(logger)
}

// initCollabServices initializes collaboration services
func initCollabServices(svcs *Services, logger *zap.Logger) {
	svcs.CollabPattern = service.NewCollabPatternService(logger)
	svcs.CrossOrg = service.NewCrossOrgService(logger)
}

// initInfraServices initializes infrastructure services
func initInfraServices(svcs *Services, logger *zap.Logger) {
	svcs.Federation = service.NewFederationService(logger)
	svcs.FederatedLearning = service.NewFederatedLearningService(logger)
	svcs.WebhookOrchestration = service.NewWebhookOrchestrationService(logger)
	svcs.Marketplace = service.NewMarketplaceService(logger)
	svcs.Mobile = service.NewMobileService(logger)
	svcs.Plugin = service.NewPluginService(logger)
	svcs.OrchestrationDebugger = service.NewOrchestrationDebuggerService(logger)
	svcs.RCA = service.NewRCAService(logger)
	svcs.Embed = service.NewEmbedService(logger)
	svcs.DistributedTrace = service.NewDistributedTraceService(logger)
	svcs.PromptCache = service.NewPromptCacheService(logger)
	svcs.Chaos = service.NewChaosService(logger)
	svcs.CustomMetrics = service.NewCustomMetricsService(logger)
	svcs.SyntheticData = service.NewSyntheticDataService(logger)
}

// initV6Services initializes Next-Gen v6 feature services
func initV6Services(svcs *Services, cfg *config.Config, logger *zap.Logger) {
	svcs.ReplaySession = service.NewReplaySessionService(logger)
	svcs.CostGuardrail = service.NewCostGuardrailService(logger)
	svcs.MultiAgentGraph = service.NewMultiAgentGraphService(logger)
	svcs.PromptCI = service.NewPromptCIService(logger)
	svcs.AgentBenchmark = service.NewAgentBenchmarkService(logger)
	svcs.SemanticTraceSearch = service.NewSemanticTraceSearchService(logger)
	svcs.AgentKnowledgeGraph = service.NewAgentKnowledgeGraphService(logger)
	svcs.IDETraceView = service.NewIDETraceViewService(logger)
	svcs.FederatedAggregation = service.NewFederatedAggregationService(logger)
	svcs.EmbeddingPipeline = service.NewEmbeddingPipelineService(logger)
	svcs.UsageMetering = service.NewUsageMeteringService(logger)
	svcs.StripeWebhook = service.NewStripeWebhookService(
		logger,
		cfg.Server.StripeWebhookSecret,
		svcs.Billing,
		svcs.Tenant,
	)
}

// initV7Services initializes Next-Gen v7 feature services
func initV7Services(svcs *Services, logger *zap.Logger) {
	svcs.WorkflowSimulator = service.NewWorkflowSimulatorService(logger)
	svcs.AutoDiscovery = service.NewAutoDiscoveryService(logger)
	svcs.CloudOnboarding = service.NewCloudOnboardingService(logger)
	svcs.AIDebugger = service.NewAIDebuggerService(logger)
	svcs.PromptOptimization = service.NewPromptOptimizationService(logger)
	svcs.CostAlerting = service.NewCostAlertingService(logger)
	svcs.RegressionSuite = service.NewRegressionSuiteService(logger)
	svcs.CollabHub = service.NewCollabHubService(logger)
	svcs.OTelCompat = service.NewOTelCompatService(logger)
	svcs.SecurityScanner = service.NewSecurityScannerService(logger)
}

// initV8Services initializes Next-Gen v8 feature services
func initV8Services(svcs *Services, logger *zap.Logger) {
	svcs.AgentComparison = service.NewAgentComparisonService(logger)
	svcs.RegressionDetection = service.NewRegressionDetectionService(logger)
	svcs.CodeQuality = service.NewCodeQualityService(logger)
	svcs.SkillProfile = service.NewSkillProfileService(logger, svcs.Query)
	svcs.NLQuery = service.NewNLQueryService(nil, svcs.Query, logger)
	svcs.Sandbox = service.NewSandboxService(logger)
}

// initV10Services initializes Next-Gen v10 feature services
func initV10Services(svcs *Services, logger *zap.Logger) {
	svcs.Gateway = service.NewGatewayService(logger)
	svcs.TestGenerator = service.NewTestGeneratorService(logger)
	svcs.CanaryDeployment = service.NewCanaryDeploymentService(logger)
	svcs.CertExport = service.NewCertificationExportService(logger)
	svcs.WarehouseSync = service.NewWarehouseSyncService(logger)
	svcs.TraceReview = service.NewTraceReviewService(logger)
	svcs.RunbookEngine = service.NewRunbookEngineService(logger)
	svcs.EdgeIngest = service.NewEdgeIngestService(logger)
	svcs.PromptImpact = service.NewPromptImpactService(logger)
}

// initV11Services initializes Next-Gen v11 feature services
func initV11Services(svcs *Services, logger *zap.Logger) {
	svcs.TraceDiff = service.NewTraceDiffService(logger, svcs.Query)
	svcs.EvalPlayground = service.NewEvalPlaygroundService(logger)
	svcs.CodeImpact = service.NewCodeImpactService(logger, svcs.Query, svcs.FileOperation)
	svcs.StreamingDashboard = service.NewStreamingDashboardService(logger, svcs.Streaming)
	svcs.EvalMarketplace = service.NewEvalMarketplaceService(logger, svcs.Dataset)
	svcs.SessionJourney = service.NewSessionJourneyService(logger, svcs.Query)
	svcs.TraceEnrichment = service.NewTraceEnrichmentService(logger)
	svcs.CostForecast = service.NewCostForecastService(logger, svcs.Cost, svcs.Query)
	svcs.TraceKB = service.NewTraceKBService(logger)
	svcs.OTelBridge = service.NewOTelBridgeService(logger, svcs.Ingestion)
}

// initV12Services initializes Next-Gen v12 feature services
func initV12Services(svcs *Services, logger *zap.Logger) {
	svcs.Adapter = service.NewAdapterService(logger)
	svcs.ABTesting = service.NewABTestingService(logger)
}
