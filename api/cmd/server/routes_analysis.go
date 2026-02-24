package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerAnalysisRoutes registers analysis, debugging, and regression routes
func registerAnalysisRoutes(public fiber.Router, h *Handlers) {
	// Diff Intelligence
	public.Get("/diff-analysis", h.DiffIntelligence.ListAnalyses)
	public.Post("/diff-analysis", h.DiffIntelligence.AnalyzeDiff)
	public.Get("/diff-analysis/trend", h.DiffIntelligence.GetQualityTrend)
	public.Get("/diff-analysis/:id", h.DiffIntelligence.GetAnalysis)
	public.Get("/traces/:traceId/diff-analysis", h.DiffIntelligence.GetTraceAnalyses)

	// Anomaly Detection & Alerting
	public.Get("/anomaly/dashboard", h.Anomaly.GetDashboard)
	public.Post("/anomaly/channels", h.Anomaly.CreateAlertChannel)
	public.Get("/anomaly/anomalies/:anomalyId/root-cause", h.Anomaly.GetRootCause)

	// Replay
	public.Get("/traces/:traceId/replay", h.Replay.GetTimeline)
	public.Get("/traces/:traceId/replay/export", h.Replay.ExportTimeline)
	public.Get("/traces/:traceId/replay/events", h.Replay.GetTimelineEvents)
	public.Get("/traces/:traceId/replay/events/:eventId", h.Replay.GetEventDetails)
	public.Post("/replay/compare", h.Replay.CompareTimelines)
	public.Post("/traces/:traceId/reproduce", h.Replay.GenerateReproduction)
	public.Post("/replay/compare-ab", h.Replay.CompareReplaysAB)

	// Real-Time Agent Replay
	public.Get("/replay-sessions", h.ReplaySession.ListSessions)
	public.Post("/replay-sessions", h.ReplaySession.CreateSession)
	public.Get("/replay-sessions/:sessionId", h.ReplaySession.GetSession)
	public.Get("/replay-sessions/:sessionId/timeline", h.ReplaySession.GetTimeline)
	public.Get("/replay-sessions/:sessionId/playback", h.ReplaySession.GetPlaybackState)
	public.Post("/replay-sessions/:sessionId/branch", h.ReplaySession.BranchSession)
	public.Post("/replay-sessions/:sessionId/share", h.ReplaySession.ShareSession)

	// Debug Sessions
	public.Post("/debug/sessions", h.Debug.CreateSession)
	public.Get("/debug/sessions/:sessionId", h.Debug.GetSession)
	public.Get("/traces/:traceId/debug/step/:stepIndex", h.Debug.GetStepState)
	public.Post("/debug/sessions/:sessionId/annotations", h.Debug.AddAnnotation)

	// AI-Powered Trace Debugger
	public.Post("/debug", h.AIDebugger.DebugTrace)
	public.Get("/traces/:traceId/debug-history", h.AIDebugger.GetDebugHistory)
	public.Get("/traces/:traceId/debug-context", h.AIDebugger.BuildContext)

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

	// Regression Detection
	public.Get("/regression/tests", h.Regression.ListTests)
	public.Post("/regression/tests", h.Regression.CreateTest)
	public.Post("/regression/tests/:testId/run", h.Regression.RunTest)
	public.Get("/regression/tests/:testId/results/:resultId", h.Regression.GetResult)
	public.Post("/regression/gate", h.Regression.CheckGate)

	// Agent Regression Test Suite
	public.Post("/regression/golden-datasets", h.RegressionSuite.CreateGoldenDataset)
	public.Get("/regression/golden-datasets/:datasetId", h.RegressionSuite.GetGoldenDataset)
	public.Get("/regression/golden-datasets", h.RegressionSuite.ListGoldenDatasets)
	public.Post("/regression/run", h.RegressionSuite.RunRegression)
	public.Get("/regression/runs/:runId", h.RegressionSuite.GetRegressionRun)
	public.Get("/regression/runs", h.RegressionSuite.ListRegressionRuns)

	// Automated Regression Detection
	public.Post("/regression-detection/configs", h.RegressionDetection.CreateConfig)
	public.Get("/regression-detection/configs", h.RegressionDetection.ListConfigs)
	public.Get("/regression-detection/configs/:configId", h.RegressionDetection.GetConfig)
	public.Put("/regression-detection/configs/:configId", h.RegressionDetection.UpdateConfig)
	public.Delete("/regression-detection/configs/:configId", h.RegressionDetection.DeleteConfig)
	public.Post("/regression-detection/configs/:configId/run", h.RegressionDetection.RunDetection)
	public.Get("/regression-detection/detections", h.RegressionDetection.ListDetections)
	public.Post("/regression-detection/detections/:detectionId/acknowledge", h.RegressionDetection.AcknowledgeDetection)
	public.Post("/regression-detection/detections/:detectionId/resolve", h.RegressionDetection.ResolveDetection)
	public.Get("/regression-detection/dashboard", h.RegressionDetection.GetDashboard)

	// Predictive Agent Health
	public.Get("/health/analyze", h.Prediction.AnalyzeHealth)
	public.Get("/health/predictions", h.Prediction.GetPredictions)
	public.Get("/health/trends/:metricName", h.Prediction.GetTrend)

	// Semantic Search
	public.Post("/search", h.SemanticSearch.Search)
	public.Get("/search/suggestions", h.SemanticSearch.GetSuggestions)

	// Semantic Trace Search
	public.Post("/semantic-search", h.SemanticTraceSearch.Search)
	public.Get("/semantic-search/clusters", h.SemanticTraceSearch.GetClusters)
	public.Get("/semantic-search/anomaly-patterns", h.SemanticTraceSearch.GetAnomalyPatterns)
	public.Get("/semantic-search/dashboard", h.SemanticTraceSearch.GetDashboard)

	// Trace-to-Ticket Pipeline
	public.Get("/tickets", h.Tickets.ListTickets)
	public.Post("/tickets", h.Tickets.CreateTicket)
	public.Post("/tickets/preview", h.Tickets.PreviewTicket)
	public.Get("/tickets/integrations", h.Tickets.GetIntegrations)
	public.Post("/tickets/integrations", h.Tickets.ConfigureIntegration)

	// Chaos Testing
	public.Get("/chaos/experiments", h.Chaos.List)
	public.Post("/chaos/experiments", h.Chaos.Create)
	public.Get("/chaos/experiments/:experimentId", h.Chaos.Get)
	public.Post("/chaos/experiments/:experimentId/run", h.Chaos.Run)
	public.Get("/chaos/scorecard/:agentName", h.Chaos.GetScorecard)

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
}
