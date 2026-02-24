package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerRoutes registers all HTTP routes
func registerRoutes(app *fiber.App, deps *Dependencies) {
	h := deps.Handlers // Shorthand for handlers

	// Health check routes (no auth required)
	app.Get("/health", h.Health.Health)
	app.Get("/healthz", h.Health.Health)
	app.Get("/livez", h.Health.Liveness)
	app.Get("/live", h.Health.Liveness)
	app.Get("/readyz", h.Health.Readiness)
	app.Get("/ready", h.Health.Readiness)
	app.Get("/version", h.Health.Version)

	// API Documentation routes (no auth required)
	h.Docs.RegisterRoutes(app)

	// Public API routes (API key auth)
	public := app.Group("/api/public")
	public.Use(deps.AuthMiddleware.RequireAPIKey())
	public.Use(deps.RateLimitMiddleware.Handler())
	{
		// Ingestion endpoints
		public.Post("/ingestion", h.Ingestion.BatchIngestion)
		public.Post("/traces", h.Ingestion.CreateTrace)
		public.Post("/spans", h.Ingestion.CreateSpan)
		public.Post("/generations", h.Ingestion.CreateGeneration)
		public.Post("/events", h.Ingestion.CreateEvent)

		// OTLP-compatible ingestion (also uses BatchIngestion)
		public.Post("/v1/traces", h.Ingestion.BatchIngestion)

		// Trace queries
		public.Get("/traces", h.Traces.ListTraces)
		public.Get("/traces/search", h.Traces.SearchTraces)
		public.Get("/traces/:id", h.Traces.GetTrace)
		public.Get("/traces/:id/observations", h.Traces.GetTraceObservations)
		public.Get("/traces/:id/stats", h.Traces.GetTraceStats)
		public.Delete("/traces/:id", h.Traces.DeleteTrace)

		// Sessions
		public.Get("/sessions", h.Traces.GetSessions)
		public.Get("/sessions/:id", h.Traces.GetSession)

		// Scores
		public.Get("/scores", h.Scores.ListScores)
		public.Get("/scores/stats", h.Scores.GetScoreStats)
		public.Get("/scores/:id", h.Scores.GetScore)
		public.Post("/scores", h.Scores.CreateScore)
		public.Post("/scores/batch", h.Scores.BatchCreateScores)
		public.Put("/scores/:id", h.Scores.UpdateScore)
		public.Get("/traces/:traceId/scores", h.Scores.GetTraceScores)

		// Prompts
		public.Get("/prompts", h.Prompts.ListPrompts)
		public.Get("/prompts/:name", h.Prompts.GetPrompt)
		public.Post("/prompts", h.Prompts.CreatePrompt)
		public.Put("/prompts/:name", h.Prompts.UpdatePrompt)
		public.Delete("/prompts/:name", h.Prompts.DeletePrompt)
		public.Get("/prompts/:name/versions", h.Prompts.ListVersions)
		public.Post("/prompts/:name/labels", h.Prompts.SetLabel)
		public.Delete("/prompts/:name/labels/:label", h.Prompts.RemoveLabel)
		public.Post("/prompts/:name/compile", h.Prompts.CompilePrompt)

		// Datasets
		public.Get("/datasets", h.Datasets.ListDatasets)
		public.Get("/datasets/:id", h.Datasets.GetDataset)
		public.Post("/datasets", h.Datasets.CreateDataset)
		public.Put("/datasets/:id", h.Datasets.UpdateDataset)
		public.Delete("/datasets/:id", h.Datasets.DeleteDataset)
		public.Get("/datasets/:id/items", h.Datasets.ListItems)
		public.Post("/datasets/:id/items", h.Datasets.CreateItem)
		public.Put("/datasets/:datasetId/items/:id", h.Datasets.UpdateItem)
		public.Delete("/datasets/:datasetId/items/:id", h.Datasets.DeleteItem)
		public.Get("/datasets/:id/runs", h.Datasets.ListRuns)
		public.Post("/datasets/:id/runs", h.Datasets.CreateRun)
		public.Get("/datasets/:datasetId/runs/:id", h.Datasets.GetRun)
		public.Post("/datasets/:datasetId/runs/:id/items", h.Datasets.AddRunItem)

		// Evaluators
		public.Get("/evaluators", h.Evaluators.ListEvaluators)
		public.Get("/evaluators/:id", h.Evaluators.GetEvaluator)
		public.Post("/evaluators", h.Evaluators.CreateEvaluator)
		public.Put("/evaluators/:id", h.Evaluators.UpdateEvaluator)
		public.Delete("/evaluators/:id", h.Evaluators.DeleteEvaluator)
		public.Get("/evaluator-templates", h.Evaluators.ListTemplates)

		// Metrics
		public.Get("/metrics/project", h.Traces.GetMetrics)

		// Real-time events (SSE)
		public.Get("/events", h.Events.StreamEvents)

		// Checkpoints (agent-specific)
		public.Get("/checkpoints", h.Checkpoints.ListCheckpoints)
		public.Get("/checkpoints/:checkpointId", h.Checkpoints.GetCheckpoint)
		public.Post("/checkpoints", h.Checkpoints.CreateCheckpoint)
		public.Post("/checkpoints/:checkpointId/restore", h.Checkpoints.RestoreCheckpoint)
		public.Get("/traces/:traceId/checkpoints", h.Checkpoints.GetTraceCheckpoints)

		// Git Links (agent-specific)
		public.Get("/git-links", h.GitLinks.ListGitLinks)
		public.Get("/git-links/timeline", h.GitLinks.GetTimeline)
		public.Get("/git-links/commit/:commitSha", h.GitLinks.GetByCommit)
		public.Get("/git-links/:gitLinkId", h.GitLinks.GetGitLink)
		public.Post("/git-links", h.GitLinks.CreateGitLink)
		public.Get("/traces/:traceId/git-links", h.GitLinks.GetTraceGitLinks)

		// File Operations (agent-specific)
		public.Get("/file-operations", h.FileOperations.ListFileOperations)
		public.Get("/file-operations/stats", h.FileOperations.GetFileOperationStats)
		public.Post("/file-operations", h.FileOperations.CreateFileOperation)
		public.Post("/file-operations/batch", h.FileOperations.BatchCreateFileOperations)
		public.Get("/traces/:traceId/file-operations", h.FileOperations.GetTraceFileOperations)

		// Terminal Commands (agent-specific)
		public.Get("/terminal-commands", h.TerminalCommands.ListTerminalCommands)
		public.Get("/terminal-commands/stats", h.TerminalCommands.GetTerminalCommandStats)
		public.Post("/terminal-commands", h.TerminalCommands.CreateTerminalCommand)
		public.Post("/terminal-commands/batch", h.TerminalCommands.BatchCreateTerminalCommands)
		public.Get("/traces/:traceId/terminal-commands", h.TerminalCommands.GetTraceTerminalCommands)

		// CI Runs (agent-specific)
		public.Get("/ci-runs", h.CIRuns.ListCIRuns)
		public.Get("/ci-runs/stats", h.CIRuns.GetCIRunStats)
		public.Get("/ci-runs/provider/:providerRunId", h.CIRuns.GetCIRunByProviderID)
		public.Get("/ci-runs/:ciRunId", h.CIRuns.GetCIRun)
		public.Post("/ci-runs", h.CIRuns.CreateCIRun)
		public.Patch("/ci-runs/:ciRunId", h.CIRuns.UpdateCIRun)
		public.Post("/ci-runs/:ciRunId/traces", h.CIRuns.AddTraceToCIRun)
		public.Post("/ci-runs/:ciRunId/complete", h.CIRuns.CompleteCIRun)

		// Export endpoints
		public.Post("/export/data", h.Export.ExportData)
		public.Post("/export/dataset", h.Export.ExportDataset)

		// Import endpoints
		public.Post("/import/dataset", h.Import.ImportDataset)
		public.Post("/import/dataset/csv", h.Import.ImportDatasetCSV)
		public.Post("/import/dataset/openai-finetune", h.Import.ImportOpenAIFinetune)
		public.Post("/import/prompt", h.Import.ImportPrompt)

		// Webhooks
		public.Get("/webhooks", h.Webhook.ListWebhooks)
		public.Get("/webhooks/:id", h.Webhook.GetWebhook)
		public.Post("/webhooks", h.Webhook.CreateWebhook)
		public.Patch("/webhooks/:id", h.Webhook.UpdateWebhook)
		public.Delete("/webhooks/:id", h.Webhook.DeleteWebhook)
		public.Post("/webhooks/:id/test", h.Webhook.TestWebhook)
		public.Get("/webhooks/:id/deliveries", h.Webhook.ListWebhookDeliveries)

		// Replay
		public.Get("/traces/:traceId/replay", h.Replay.GetTimeline)
		public.Get("/traces/:traceId/replay/export", h.Replay.ExportTimeline)
		public.Get("/traces/:traceId/replay/events", h.Replay.GetTimelineEvents)
		public.Get("/traces/:traceId/replay/events/:eventId", h.Replay.GetEventDetails)
		public.Post("/replay/compare", h.Replay.CompareTimelines)
		public.Post("/traces/:traceId/reproduce", h.Replay.GenerateReproduction)
		public.Post("/replay/compare-ab", h.Replay.CompareReplaysAB)

		// Debug Sessions (extends replay with interactive debugging)
		public.Post("/debug/sessions", h.Debug.CreateSession)
		public.Get("/debug/sessions/:sessionId", h.Debug.GetSession)
		public.Get("/traces/:traceId/debug/step/:stepIndex", h.Debug.GetStepState)
		public.Post("/debug/sessions/:sessionId/annotations", h.Debug.AddAnnotation)

		// Regression Detection
		public.Get("/regression/tests", h.Regression.ListTests)
		public.Post("/regression/tests", h.Regression.CreateTest)
		public.Post("/regression/tests/:testId/run", h.Regression.RunTest)
		public.Get("/regression/tests/:testId/results/:resultId", h.Regression.GetResult)
		public.Post("/regression/gate", h.Regression.CheckGate)

		// Cost Optimization
		public.Get("/cost-optimizer/analyze", h.CostOptimizer.Analyze)
		public.Get("/cost-optimizer/recommendations", h.CostOptimizer.GetRecommendations)
		public.Post("/cost-optimizer/recommendations/:id/apply", h.CostOptimizer.ApplyRecommendation)
		public.Post("/cost-optimizer/recommendations/:id/dismiss", h.CostOptimizer.DismissRecommendation)
		public.Get("/cost-optimizer/forecast", h.CostOptimizer.GetForecast)
		public.Post("/cost-optimizer/report", h.CostOptimizer.GenerateReport)
		public.Post("/cost-optimizer/autopilot", h.CostOptimizer.ConfigureAutopilot)

		// Agent Graph (multi-agent visualization)
		public.Get("/traces/:traceId/graph", h.AgentGraph.BuildGraph)
		public.Post("/agent-graph/compare", h.AgentGraph.CompareGraphs)

		// Guardrails
		public.Get("/guardrails", h.Guardrails.ListRules)
		public.Post("/guardrails", h.Guardrails.CreateRule)
		public.Put("/guardrails/:ruleId", h.Guardrails.UpdateRule)
		public.Delete("/guardrails/:ruleId", h.Guardrails.DeleteRule)
		public.Get("/guardrails/violations", h.Guardrails.ListViolations)
		public.Get("/guardrails/violations/stats", h.Guardrails.GetViolationStats)
		public.Get("/guardrails/templates", h.Guardrails.GetPlaybookTemplates)
		public.Post("/guardrails/playbooks", h.Guardrails.CreatePlaybook)

		// Benchmarks
		public.Get("/benchmarks", h.Benchmarks.ListBenchmarks)
		public.Get("/benchmarks/:benchmarkId", h.Benchmarks.GetBenchmark)
		public.Post("/benchmarks", h.Benchmarks.CreateBenchmark)
		public.Post("/benchmarks/:benchmarkId/submit", h.Benchmarks.Submit)
		public.Get("/benchmarks/:benchmarkId/leaderboard", h.Benchmarks.GetLeaderboard)
		public.Post("/benchmarks/:benchmarkId/compare", h.Benchmarks.CompareSubmissions)
		public.Get("/benchmarks/:benchmarkId/stats", h.Benchmarks.GetStats)

		// Anomaly Detection & Alerting
		public.Get("/anomaly/dashboard", h.Anomaly.GetDashboard)
		public.Post("/anomaly/channels", h.Anomaly.CreateAlertChannel)
		public.Get("/anomaly/anomalies/:anomalyId/root-cause", h.Anomaly.GetRootCause)

		// Collaboration
		public.Get("/collaboration/traces/:traceId/presence", h.Collaboration.GetPresence)
		public.Post("/collaboration/traces/:traceId/annotations", h.Collaboration.AddAnnotation)
		public.Get("/collaboration/traces/:traceId/annotations", h.Collaboration.ListAnnotations)
		public.Post("/collaboration/annotations/:annotationId/resolve", h.Collaboration.ResolveAnnotation)
		public.Post("/collaboration/sessions", h.Collaboration.CreateSharedSession)

		// Collaboration - Discussions
		public.Post("/collaboration/discussions", h.Collaboration.CreateDiscussion)
		public.Post("/collaboration/discussions/:threadId/messages", h.Collaboration.AddMessage)

		// Collaboration - Evaluation Queues
		public.Post("/collaboration/eval-queues", h.Collaboration.CreateEvalQueue)

		// Migration
		public.Get("/migrations", h.Migration.ListMigrations)
		public.Post("/migrations", h.Migration.StartMigration)
		public.Get("/migrations/:jobId", h.Migration.GetMigration)
		public.Post("/migrations/validate", h.Migration.ValidateSource)

		// EU AI Act Compliance
		public.Get("/compliance/assess", h.Compliance.AssessProject)
		public.Get("/compliance/status", h.Compliance.GetStatus)
		public.Get("/compliance/audit-trail", h.Compliance.GetAuditTrail)
		public.Post("/compliance/assessments", h.Compliance.CreateAssessment)
		public.Get("/compliance/assessments/:id", h.Compliance.GetAssessment)
		public.Post("/compliance/reports", h.Compliance.GenerateReport)

		// Compliance Export
		public.Post("/compliance/exports", h.ComplianceExport.StartExport)
		public.Get("/compliance/exports", h.ComplianceExport.ListExports)
		public.Get("/compliance/exports/:id", h.ComplianceExport.GetExport)
		public.Get("/compliance/templates", h.ComplianceExport.GetTemplates)

		// Predictive Agent Health
		public.Get("/health/analyze", h.Prediction.AnalyzeHealth)
		public.Get("/health/predictions", h.Prediction.GetPredictions)
		public.Get("/health/trends/:metricName", h.Prediction.GetTrend)

		// Agent Reasoning Explorer
		public.Get("/traces/:traceId/reasoning", h.Reasoning.GetReasoningTree)
		public.Get("/traces/:traceId/reasoning/:nodeId", h.Reasoning.GetNode)

		// Cost Budgets & Forecasting
		public.Get("/budgets", h.CostBudget.ListBudgets)
		public.Post("/budgets", h.CostBudget.CreateBudget)
		public.Get("/budgets/forecast", h.CostBudget.GetForecast)
		public.Post("/budgets/check", h.CostBudget.CheckBudget)
		public.Get("/budgets/:id", h.CostBudget.GetBudget)
		public.Put("/budgets/:id", h.CostBudget.UpdateBudget)
		public.Delete("/budgets/:id", h.CostBudget.DeleteBudget)

		// Framework Auto-Instrumentation
		public.Get("/instrumentation/frameworks", h.Instrumentation.ListFrameworks)
		public.Get("/instrumentation/setup/:framework", h.Instrumentation.GetSetup)

		// Agent Performance Scorecards
		public.Get("/scorecards", h.Scorecard.ListScorecards)
		public.Post("/scorecards", h.Scorecard.Generate)
		public.Get("/scorecards/config", h.Scorecard.GetConfig)
		public.Post("/scorecards/config", h.Scorecard.ConfigureAuto)
		public.Get("/scorecards/:id", h.Scorecard.GetScorecard)

		// Diff Intelligence
		public.Get("/diff-analysis", h.DiffIntelligence.ListAnalyses)
		public.Post("/diff-analysis", h.DiffIntelligence.AnalyzeDiff)
		public.Get("/diff-analysis/trend", h.DiffIntelligence.GetQualityTrend)
		public.Get("/diff-analysis/:id", h.DiffIntelligence.GetAnalysis)
		public.Get("/traces/:traceId/diff-analysis", h.DiffIntelligence.GetTraceAnalyses)

		// Trace-to-Ticket Pipeline
		public.Get("/tickets", h.Tickets.ListTickets)
		public.Post("/tickets", h.Tickets.CreateTicket)
		public.Post("/tickets/preview", h.Tickets.PreviewTicket)
		public.Get("/tickets/integrations", h.Tickets.GetIntegrations)
		public.Post("/tickets/integrations", h.Tickets.ConfigureIntegration)

		// Real-time streaming
		public.Get("/streams", h.Streaming.GetActiveStreams)
		public.Get("/traces/:traceId/stream", h.Streaming.StreamTrace)
		public.Get("/traces/:traceId/live-metrics", h.Streaming.GetLiveMetrics)
		public.Get("/traces/:traceId/activities", h.Streaming.GetRecentActivities)
		public.Post("/traces/:traceId/intervene", h.Streaming.RequestIntervention)
		public.Get("/traces/:traceId/interventions", h.Streaming.GetPendingInterventions)
		public.Post("/traces/:traceId/interventions/:interventionId/ack", h.Streaming.AcknowledgeIntervention)

		// Team Intelligence
		public.Get("/team/dashboard", h.TeamIntelligence.GetDashboard)
		public.Get("/team/roi", h.TeamIntelligence.CalculateROI)

		// Semantic Search
		public.Post("/search", h.SemanticSearch.Search)
		public.Get("/search/suggestions", h.SemanticSearch.GetSuggestions)

		// Training Pipeline
		public.Get("/training/datasets", h.TrainingPipeline.ListDatasets)
		public.Post("/training/datasets", h.TrainingPipeline.CreateDataset)
		public.Post("/training/datasets/:datasetId/export", h.TrainingPipeline.ExportDataset)
		public.Get("/training/failure-patterns", h.TrainingPipeline.DetectFailures)

		// RBAC & SSO
		public.Get("/rbac/permissions", h.RBAC.GetPermissions)
		public.Post("/rbac/roles", h.RBAC.AssignRole)
		public.Post("/rbac/check", h.RBAC.CheckPermission)
		public.Get("/rbac/sso", h.RBAC.GetSSOConfig)
		public.Post("/rbac/sso", h.RBAC.ConfigureSSO)
		public.Post("/rbac/api-key-scope", h.RBAC.ScopeAPIKey)

		// Federation & OTLP Export
		public.Get("/federation/peers", h.Federation.ListPeers)
		public.Post("/federation/peers", h.Federation.AddPeer)
		public.Delete("/federation/peers/:peerId", h.Federation.RemovePeer)
		public.Post("/federation/query", h.Federation.FederatedQuery)
		public.Get("/federation/destinations", h.Federation.ListExportDestinations)
		public.Post("/federation/destinations", h.Federation.CreateExportDestination)

		// Webhook Orchestration
		public.Get("/webhook-rules", h.WebhookOrchestration.ListRules)
		public.Post("/webhook-rules", h.WebhookOrchestration.CreateRule)
		public.Delete("/webhook-rules/:ruleId", h.WebhookOrchestration.DeleteRule)
		public.Get("/webhook-rules/templates", h.WebhookOrchestration.GetTemplates)
		public.Get("/webhook-rules/deliveries", h.WebhookOrchestration.ListDeliveries)
		public.Post("/webhook-rules/:ruleId/test", h.WebhookOrchestration.TestRule)

		// Agent Marketplace
		public.Get("/marketplace", h.Marketplace.Search)
		public.Get("/marketplace/featured", h.Marketplace.Featured)
		public.Get("/marketplace/:packageId", h.Marketplace.Get)
		public.Post("/marketplace", h.Marketplace.Publish)
		public.Post("/marketplace/:packageId/install", h.Marketplace.Install)
		public.Post("/marketplace/:packageId/rate", h.Marketplace.Rate)

		// Compliance Reports
		public.Get("/compliance-reports", h.ComplianceReport.List)
		public.Post("/compliance-reports", h.ComplianceReport.Generate)
		public.Get("/compliance-reports/templates", h.ComplianceReport.GetTemplates)
		public.Get("/compliance-reports/:reportId", h.ComplianceReport.Get)

		// NL Agent Builder
		public.Post("/agent-builder/generate", h.AgentBuilder.Generate)
		public.Get("/agent-builder/blueprints", h.AgentBuilder.List)
		public.Get("/agent-builder/blueprints/:blueprintId", h.AgentBuilder.Get)
		public.Post("/agent-builder/blueprints/:blueprintId/deploy", h.AgentBuilder.Deploy)

		// Fleet Management
		public.Get("/fleet/dashboard", h.Fleet.GetDashboard)
		public.Get("/fleet/agents", h.Fleet.ListAgents)
		public.Get("/fleet/policies", h.Fleet.ListPolicies)
		public.Post("/fleet/policies", h.Fleet.CreatePolicy)
		public.Post("/fleet/bulk-update", h.Fleet.BulkUpdate)
		public.Get("/fleet/scaling", h.Fleet.GetScalingRecommendations)

		// Privacy
		public.Post("/privacy/scan", h.Privacy.ScanPII)
		public.Get("/privacy/config", h.Privacy.GetConfig)
		public.Put("/privacy/config", h.Privacy.UpdateConfig)
		public.Post("/privacy/deletion-requests", h.Privacy.RequestDeletion)
		public.Get("/privacy/deletion-requests", h.Privacy.ListDeletionRequests)

		// Mobile API
		public.Post("/mobile/devices", h.Mobile.RegisterDevice)
		public.Get("/mobile/dashboard", h.Mobile.GetDashboard)
		public.Get("/mobile/notifications", h.Mobile.ListNotifications)

		// Plugin System
		public.Get("/plugins", h.Plugin.List)
		public.Post("/plugins", h.Plugin.Install)
		public.Get("/plugins/:pluginId", h.Plugin.Get)
		public.Post("/plugins/:pluginId/activate", h.Plugin.Activate)
		public.Post("/plugins/:pluginId/disable", h.Plugin.Disable)
		public.Post("/plugins/:pluginId/execute", h.Plugin.Execute)
		public.Delete("/plugins/:pluginId", h.Plugin.Uninstall)

		// Multi-Agent Orchestration Debugger
		public.Get("/orchestration/sessions", h.OrchestrationDebugger.ListSessions)
		public.Post("/orchestration/sessions", h.OrchestrationDebugger.CreateSession)
		public.Get("/orchestration/sessions/:sessionId", h.OrchestrationDebugger.GetSession)
		public.Post("/orchestration/sessions/:sessionId/command", h.OrchestrationDebugger.ExecuteCommand)
		public.Post("/orchestration/sessions/:sessionId/breakpoints", h.OrchestrationDebugger.AddBreakpoint)

		// AI-Powered Root Cause Analysis
		public.Post("/rca/analyze", h.RCA.Analyze)
		public.Get("/rca/reports", h.RCA.ListReports)
		public.Get("/rca/reports/:reportId", h.RCA.GetReport)

		// Agent Rollback & Versioning
		public.Get("/agent-versions", h.AgentVersion.ListVersions)
		public.Post("/agent-versions", h.AgentVersion.CreateVersion)
		public.Get("/agent-versions/active", h.AgentVersion.GetActive)
		public.Get("/agent-versions/:versionId", h.AgentVersion.GetVersion)
		public.Post("/agent-versions/:versionId/rollback", h.AgentVersion.Rollback)
		public.Post("/agent-versions/diff", h.AgentVersion.DiffVersions)

		// Predictive Cost Modeling
		public.Post("/predictions/cost", h.PredictiveCost.Predict)
		public.Get("/predictions", h.PredictiveCost.ListPredictions)
		public.Post("/predictions/:predictionId/approve", h.PredictiveCost.RequestApproval)
		public.Post("/approvals/:approvalId/decide", h.PredictiveCost.DecideApproval)

		// White-Label & Embedding
		public.Get("/embed/config", h.Embed.GetConfig)
		public.Post("/embed/config", h.Embed.CreateConfig)
		public.Put("/embed/config", h.Embed.UpdateConfig)
		public.Post("/embed/token", h.Embed.GenerateToken)

		// Agent Handoff Protocol
		public.Post("/handoffs", h.Handoff.Initiate)
		public.Post("/handoffs/:handoffId/accept", h.Handoff.Accept)
		public.Post("/handoffs/:handoffId/complete", h.Handoff.Complete)
		public.Get("/handoffs/chain/:traceId", h.Handoff.GetChain)
		public.Get("/handoffs/stats", h.Handoff.GetStats)

		// Collaborative Annotations
		public.Get("/annotations/traces/:traceId", h.Annotation.List)
		public.Post("/annotations", h.Annotation.Create)
		public.Post("/annotations/:annotationId/reply", h.Annotation.Reply)
		public.Post("/annotations/:annotationId/resolve", h.Annotation.Resolve)
		public.Get("/annotations/presence/:traceId", h.Annotation.GetPresence)

		// Energy & Carbon
		public.Get("/carbon/footprint", h.Carbon.GetFootprint)
		public.Get("/carbon/config", h.Carbon.GetConfig)
		public.Put("/carbon/config", h.Carbon.UpdateConfig)
		public.Get("/carbon/suggestions", h.Carbon.GetSuggestions)

		// Synthetic Data
		public.Post("/synthetic-data/generate", h.SyntheticData.Generate)
		public.Get("/synthetic-data/datasets", h.SyntheticData.List)
		public.Get("/synthetic-data/datasets/:datasetId", h.SyntheticData.Get)
		public.Get("/synthetic-data/stats", h.SyntheticData.GetStats)

		// Agent SLOs
		public.Get("/slos", h.SLO.List)
		public.Post("/slos", h.SLO.Create)
		public.Get("/slos/:sloId/status", h.SLO.GetStatus)
		public.Get("/slos/report", h.SLO.GetReport)
		public.Get("/slos/:sloId/history", h.SLO.GetHistory)
		// Agent Memory & Context
		public.Post("/memory/analyze", h.AgentMemory.AnalyzeMemory)
		public.Get("/memory/traces/:traceId/snapshots/:stepIndex", h.AgentMemory.GetSnapshot)
		public.Get("/memory/optimizations", h.AgentMemory.GetOptimizations)

		// Distributed Tracing
		public.Get("/distributed/traces/:traceId", h.DistributedTrace.GetTrace)
		public.Get("/distributed/service-map", h.DistributedTrace.GetServiceMap)
		public.Post("/distributed/correlate", h.DistributedTrace.Correlate)

		// Prompt Caching
		public.Get("/prompt-cache/analyze", h.PromptCache.Analyze)
		public.Get("/prompt-cache/config", h.PromptCache.GetConfig)
		public.Put("/prompt-cache/config", h.PromptCache.UpdateConfig)
		public.Get("/prompt-cache/stats", h.PromptCache.GetStats)
		public.Post("/prompt-cache/invalidate", h.PromptCache.Invalidate)

		// Chaos Testing
		public.Get("/chaos/experiments", h.Chaos.List)
		public.Post("/chaos/experiments", h.Chaos.Create)
		public.Get("/chaos/experiments/:experimentId", h.Chaos.Get)
		public.Post("/chaos/experiments/:experimentId/run", h.Chaos.Run)
		public.Get("/chaos/scorecard/:agentName", h.Chaos.GetScorecard)

		// Custom Metrics
		public.Get("/custom-metrics", h.CustomMetrics.ListMetrics)
		public.Post("/custom-metrics", h.CustomMetrics.CreateMetric)
		public.Get("/custom-metrics/:metricId/values", h.CustomMetrics.GetValues)
		public.Get("/custom-metrics/dashboards", h.CustomMetrics.ListDashboards)
		public.Post("/custom-metrics/dashboards", h.CustomMetrics.CreateDashboard)
		public.Get("/custom-metrics/alerts", h.CustomMetrics.ListAlerts)
		public.Post("/custom-metrics/alerts", h.CustomMetrics.CreateAlert)

		// Autonomy Gradient
		public.Get("/autonomy/dashboard", h.Autonomy.GetDashboard)
		public.Get("/autonomy/:agentName", h.Autonomy.GetConfig)
		public.Post("/autonomy", h.Autonomy.SetAutonomy)
		public.Get("/autonomy/:agentName/trust", h.Autonomy.GetTrustEvolution)

		// Cross-Organization Benchmarking
		public.Post("/cross-org/submit", h.CrossOrg.Submit)
		public.Get("/cross-org/report", h.CrossOrg.GetReport)
		public.Get("/cross-org/industry/:category", h.CrossOrg.GetIndustryStats)

		// Intent Verification
		public.Post("/intents", h.Intent.Declare)
		public.Post("/intents/:intentId/verify", h.Intent.Verify)
		public.Get("/intents/:intentId", h.Intent.Get)
		public.Get("/intents/stats", h.Intent.GetStats)

		// Cost Attribution
		public.Post("/cost-attribution", h.CostAttribution.Attribute)
		public.Get("/cost-attribution/report", h.CostAttribution.GetReport)
		public.Get("/cost-attribution", h.CostAttribution.List)

		// Knowledge Graph
		public.Get("/knowledge-graph", h.KnowledgeGraph.Build)
		public.Post("/knowledge-graph/query", h.KnowledgeGraph.Query)
		public.Get("/knowledge-graph/stats", h.KnowledgeGraph.GetStats)
		// Compliance Monitoring
		public.Get("/compliance-monitor/policies", h.ComplianceMonitor.ListPolicies)
		public.Post("/compliance-monitor/policies", h.ComplianceMonitor.CreatePolicy)
		public.Post("/compliance-monitor/evaluate", h.ComplianceMonitor.Evaluate)
		public.Get("/compliance-monitor/score/:framework", h.ComplianceMonitor.GetScore)
		public.Post("/compliance-monitor/configure", h.ComplianceMonitor.Configure)

		// Multi-Modal Traces
		public.Post("/multimodal/attachments", h.MultiModal.Register)
		public.Get("/multimodal/traces/:traceId", h.MultiModal.GetTraceAttachments)
		public.Get("/multimodal/attachments/:attachmentId", h.MultiModal.GetAttachment)
		public.Get("/multimodal/traces/:traceId/summary", h.MultiModal.GetSummary)
		public.Get("/multimodal/attachments", h.MultiModal.List)

		// Collaboration Patterns
		public.Get("/collab-patterns", h.CollabPattern.List)
		public.Get("/collab-patterns/:patternId", h.CollabPattern.Get)
		public.Post("/collab-patterns/:patternId/deploy", h.CollabPattern.Deploy)
		public.Get("/collab-patterns/deployments", h.CollabPattern.GetDeployments)
		public.Get("/collab-patterns/:patternId/analytics", h.CollabPattern.GetAnalytics)

		// Federated Learning
		public.Get("/federated/rings", h.FederatedLearning.ListRings)
		public.Post("/federated/rings/join", h.FederatedLearning.JoinRing)
		public.Get("/federated/rings/:ringId/insights", h.FederatedLearning.GetInsights)
		public.Get("/federated/config", h.FederatedLearning.GetConfig)
		public.Put("/federated/config", h.FederatedLearning.UpdateConfig)

		// Observability Copilot
		public.Post("/copilot/ask", h.Copilot.Ask)
		public.Get("/copilot/suggestions", h.Copilot.GetSuggestions)
		public.Get("/copilot/insights", h.Copilot.GetInsights)

		// ===== Next-Gen Features (v6) =====

		// Real-Time Agent Replay
		public.Get("/replay-sessions", h.ReplaySession.ListSessions)
		public.Post("/replay-sessions", h.ReplaySession.CreateSession)
		public.Get("/replay-sessions/:sessionId", h.ReplaySession.GetSession)
		public.Get("/replay-sessions/:sessionId/timeline", h.ReplaySession.GetTimeline)
		public.Get("/replay-sessions/:sessionId/playback", h.ReplaySession.GetPlaybackState)
		public.Post("/replay-sessions/:sessionId/branch", h.ReplaySession.BranchSession)
		public.Post("/replay-sessions/:sessionId/share", h.ReplaySession.ShareSession)

		// Intelligent Cost Guardrails
		public.Get("/cost-guardrails/dashboard", h.CostGuardrail.GetDashboard)
		public.Get("/cost-guardrails/policies", h.CostGuardrail.ListPolicies)
		public.Post("/cost-guardrails/policies", h.CostGuardrail.CreatePolicy)
		public.Post("/cost-guardrails/check", h.CostGuardrail.CheckBudget)
		public.Get("/cost-guardrails/forecast", h.CostGuardrail.GetForecast)
		public.Get("/cost-guardrails/violations", h.CostGuardrail.ListViolations)

		// Multi-Agent Collaboration Graph
		public.Get("/multi-agent/sessions", h.MultiAgentGraph.ListSessions)
		public.Post("/multi-agent/analyze", h.MultiAgentGraph.AnalyzeSession)
		public.Get("/multi-agent/sessions/:sessionId", h.MultiAgentGraph.GetSession)

		// Prompt Regression Testing in CI
		public.Get("/prompt-ci/baselines", h.PromptCI.ListBaselines)
		public.Post("/prompt-ci/baselines", h.PromptCI.CreateBaseline)
		public.Get("/prompt-ci/baselines/:baselineId", h.PromptCI.GetBaseline)
		public.Post("/prompt-ci/compare", h.PromptCI.RunComparison)
		public.Get("/prompt-ci/runs", h.PromptCI.ListRuns)

		// Agent Performance Benchmarks (next-gen)
		public.Get("/agent-benchmarks/suites", h.AgentBenchmark.ListSuites)
		public.Post("/agent-benchmarks/suites", h.AgentBenchmark.CreateSuite)
		public.Get("/agent-benchmarks/suites/:suiteId", h.AgentBenchmark.GetSuite)
		public.Post("/agent-benchmarks/run", h.AgentBenchmark.RunBenchmark)
		public.Get("/agent-benchmarks/suites/:suiteId/leaderboard", h.AgentBenchmark.GetLeaderboard)

		// Semantic Trace Search
		public.Post("/semantic-search", h.SemanticTraceSearch.Search)
		public.Get("/semantic-search/clusters", h.SemanticTraceSearch.GetClusters)
		public.Get("/semantic-search/anomaly-patterns", h.SemanticTraceSearch.GetAnomalyPatterns)
		public.Get("/semantic-search/dashboard", h.SemanticTraceSearch.GetDashboard)

		// Agent Knowledge Graph (next-gen)
		public.Get("/agent-knowledge-graph", h.AgentKnowledgeGraph.BuildGraph)
		public.Post("/agent-knowledge-graph/query", h.AgentKnowledgeGraph.QueryGraph)
		public.Get("/agent-knowledge-graph/evolution", h.AgentKnowledgeGraph.GetEvolution)
		public.Get("/agent-knowledge-graph/stats", h.AgentKnowledgeGraph.GetStats)

		// IDE Inline Trace Viewer
		public.Get("/ide/file-mapping", h.IDETraceView.GetFileMapping)
		public.Post("/ide/batch-mappings", h.IDETraceView.GetBatchMappings)
		public.Get("/ide/trace-context/:traceId", h.IDETraceView.GetTraceContext)

		// Federated Trace Aggregation
		public.Get("/federated-aggregation/dashboard", h.FederatedAggregation.GetDashboard)
		public.Get("/federated-aggregation/instances", h.FederatedAggregation.ListInstances)
		public.Post("/federated-aggregation/instances", h.FederatedAggregation.RegisterInstance)
		public.Post("/federated-aggregation/metrics", h.FederatedAggregation.SubmitMetrics)
		public.Get("/federated-aggregation/benchmarks", h.FederatedAggregation.GetBenchmarks)
		public.Get("/federated-aggregation/insights", h.FederatedAggregation.GetInsights)

		// ===== Next-Gen Features (v7) =====

		// Agent Workflow Simulator
		public.Get("/workflows", h.WorkflowSimulator.ListWorkflows)
		public.Post("/workflows", h.WorkflowSimulator.CreateWorkflow)
		public.Get("/workflows/:workflowId", h.WorkflowSimulator.GetWorkflow)
		public.Put("/workflows/:workflowId", h.WorkflowSimulator.UpdateWorkflow)
		public.Delete("/workflows/:workflowId", h.WorkflowSimulator.DeleteWorkflow)
		public.Post("/workflows/validate", h.WorkflowSimulator.ValidateWorkflow)
		public.Post("/workflows/simulate", h.WorkflowSimulator.RunSimulation)
		public.Get("/workflows/simulations/:simulationId", h.WorkflowSimulator.GetSimulation)
		public.Get("/workflows/:workflowId/simulations", h.WorkflowSimulator.ListSimulations)

		// Zero-Config Auto-Discovery
		public.Post("/discovery/scan", h.AutoDiscovery.ScanProject)
		public.Get("/discovery/frameworks/:frameworkId", h.AutoDiscovery.GetFramework)
		public.Put("/discovery/config", h.AutoDiscovery.UpdateConfig)
		public.Post("/discovery/frameworks/:frameworkId/toggle", h.AutoDiscovery.ToggleInstrumentation)

		// Cloud Onboarding
		public.Get("/onboarding", h.CloudOnboarding.GetOnboarding)
		public.Post("/onboarding/step", h.CloudOnboarding.CompleteStep)
		public.Post("/onboarding/quickstart", h.CloudOnboarding.GenerateQuickstart)
		public.Get("/onboarding/usage", h.CloudOnboarding.GetUsage)
		public.Post("/onboarding/quota-check", h.CloudOnboarding.CheckQuota)

		// AI-Powered Trace Debugger
		public.Post("/debug", h.AIDebugger.DebugTrace)
		public.Get("/traces/:traceId/debug-history", h.AIDebugger.GetDebugHistory)
		public.Get("/traces/:traceId/debug-context", h.AIDebugger.BuildContext)

		// Continuous Prompt Optimization
		public.Post("/prompt-optimization", h.PromptOptimization.StartOptimization)
		public.Get("/prompt-optimization/:optimizationId", h.PromptOptimization.GetOptimization)
		public.Get("/prompt-optimization", h.PromptOptimization.ListOptimizations)
		public.Get("/prompt-optimization/config", h.PromptOptimization.GetOptConfig)
		public.Put("/prompt-optimization/config", h.PromptOptimization.UpdateOptConfig)
		public.Post("/prompt-optimization/variants/:variantId/approve", h.PromptOptimization.ApproveVariant)
		public.Post("/prompt-optimization/variants/:variantId/reject", h.PromptOptimization.RejectVariant)

		// Real-Time Cost Anomaly Alerting
		public.Post("/cost-alerts/rules", h.CostAlerting.CreateAlertRule)
		public.Get("/cost-alerts/rules", h.CostAlerting.ListAlertRules)
		public.Delete("/cost-alerts/rules/:ruleId", h.CostAlerting.DeleteAlertRule)
		public.Get("/cost-alerts", h.CostAlerting.ListCostAlerts)
		public.Post("/cost-alerts/:alertId/acknowledge", h.CostAlerting.AcknowledgeCostAlert)
		public.Get("/cost-alerts/circuit-breaker", h.CostAlerting.GetCircuitBreakerConfig)
		public.Put("/cost-alerts/circuit-breaker", h.CostAlerting.UpdateCircuitBreakerConfig)
		public.Post("/cost-alerts/check", h.CostAlerting.CheckCost)

		// Agent Regression Test Suite
		public.Post("/regression/golden-datasets", h.RegressionSuite.CreateGoldenDataset)
		public.Get("/regression/golden-datasets/:datasetId", h.RegressionSuite.GetGoldenDataset)
		public.Get("/regression/golden-datasets", h.RegressionSuite.ListGoldenDatasets)
		public.Post("/regression/run", h.RegressionSuite.RunRegression)
		public.Get("/regression/runs/:runId", h.RegressionSuite.GetRegressionRun)
		public.Get("/regression/runs", h.RegressionSuite.ListRegressionRuns)

		// Collaboration Hub
		public.Post("/collab/queues", h.CollabHub.CreateReviewQueue)
		public.Get("/collab/queues", h.CollabHub.ListReviewQueues)
		public.Post("/collab/reviews", h.CollabHub.AssignReview)
		public.Post("/collab/reviews/:assignmentId/complete", h.CollabHub.CompleteReview)
		public.Post("/collab/standards", h.CollabHub.CreateQualityStandard)
		public.Get("/collab/standards", h.CollabHub.ListQualityStandards)
		public.Get("/collab/activity", h.CollabHub.GetActivityFeed)

		// OpenTelemetry Native Compatibility
		public.Post("/otel/destinations", h.OTelCompat.CreateExportDestination)
		public.Get("/otel/destinations", h.OTelCompat.ListExportDestinations)
		public.Delete("/otel/destinations/:destinationId", h.OTelCompat.DeleteExportDestination)
		public.Get("/otel/mappings", h.OTelCompat.GetOTelMappings)
		public.Get("/otel/dashboard", h.OTelCompat.GetOTelDashboard)
		public.Post("/otel/collector-config", h.OTelCompat.GenerateCollectorConfig)

		// Agent Security Scanner
		public.Post("/security/scan", h.SecurityScanner.ScanTrace)
		public.Post("/security/policies", h.SecurityScanner.CreateSecurityPolicy)
		public.Get("/security/policies", h.SecurityScanner.ListSecurityPolicies)
		public.Get("/security/dashboard", h.SecurityScanner.GetSecurityDashboard)
		public.Post("/security/findings/:findingId/acknowledge", h.SecurityScanner.AcknowledgeSecurityFinding)
	}

	// Internal API routes (JWT auth)
	internal := app.Group("/api/v1")
	internal.Use(deps.AuthMiddleware.RequireJWT())
	internal.Use(deps.RateLimitMiddleware.UserRateLimit(100)) // 100 requests per minute per user
	internal.Use(deps.CSRFMiddleware.Handler())
	{
		// CSRF token endpoint for SPAs
		internal.Get("/csrf-token", deps.CSRFMiddleware.GetToken())

		// Current user
		internal.Get("/me", h.Auth.GetCurrentUser)

		// Organizations
		internal.Get("/organizations", h.Organizations.ListOrganizations)
		internal.Get("/organizations/slug/:slug", h.Organizations.GetOrganizationBySlug)
		internal.Get("/organizations/:id", h.Organizations.GetOrganization)
		internal.Post("/organizations", h.Organizations.CreateOrganization)
		internal.Put("/organizations/:id", h.Organizations.UpdateOrganization)
		internal.Delete("/organizations/:id", h.Organizations.DeleteOrganization)
		internal.Get("/organizations/:orgId/members/:userId", h.Organizations.GetMember)

		// Projects
		internal.Get("/projects", h.Projects.ListProjects)
		internal.Get("/projects/:id", h.Projects.GetProject)
		internal.Post("/projects", h.Projects.CreateProject)
		internal.Put("/projects/:id", h.Projects.UpdateProject)
		internal.Delete("/projects/:id", h.Projects.DeleteProject)
		internal.Post("/projects/:id/members", h.Projects.AddMember)
		internal.Delete("/projects/:projectId/members/:userId", h.Projects.RemoveMember)
		internal.Get("/projects/:id/role", h.Projects.GetUserRole)

		// API Keys
		internal.Get("/projects/:id/api-keys", h.APIKeys.ListAPIKeys)
		internal.Post("/projects/:id/api-keys", h.APIKeys.CreateAPIKey)
		internal.Delete("/api-keys/:id", h.APIKeys.DeleteAPIKey)

		// Dashboard metrics
		internal.Get("/projects/:id/metrics", h.Traces.GetMetrics)
	}

	// Auth routes (no auth required)
	auth := app.Group("/api/auth")
	{
		auth.Post("/login", h.Auth.Login)
		auth.Post("/register", h.Auth.Register)
		auth.Post("/refresh", h.Auth.RefreshToken)
		auth.Post("/logout", h.Auth.Logout)
		auth.Get("/callback/:provider", h.Auth.OAuthCallback)
	}

	// User feedback endpoint (special auth - accepts both API key and user token)
	app.Post("/api/public/feedback", deps.AuthMiddleware.RequireAuth(), h.Scores.SubmitFeedback)

	// OTLP receiver endpoints (API key auth)
	otlp := app.Group("/v1")
	otlp.Use(deps.AuthMiddleware.RequireAPIKey())
	{
		otlp.Post("/traces", h.OTelReceiver.ReceiveTraces)
		otlp.Post("/metrics", h.OTelReceiver.ReceiveMetrics)
		otlp.Post("/logs", h.OTelReceiver.ReceiveLogs)
	}

	// WebSocket routes for real-time collaboration
	app.Use("/ws", h.CollaborationWS.UpgradeCheck())
	app.Get("/ws/collaboration/:traceId", h.CollaborationWS.HandleWebSocket())

	// Billing webhook (no auth — Stripe sends directly with signature)
	app.Post("/api/billing/webhook", h.Billing.HandleWebhook)
}
