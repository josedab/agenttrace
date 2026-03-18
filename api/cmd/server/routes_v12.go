package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerV12Routes registers v12 feature routes (universal agent protocol adapter)
func registerV12Routes(public fiber.Router, h *Handlers) {
	// Universal Agent Protocol Adapter
	public.Get("/adapters/templates", h.Adapter.GetTemplates)
	public.Get("/adapters", h.Adapter.ListAdapters)
	public.Post("/adapters", h.Adapter.RegisterAdapter)
	public.Get("/adapters/:adapterId", h.Adapter.GetAdapter)
	public.Put("/adapters/:adapterId", h.Adapter.UpdateAdapter)
	public.Delete("/adapters/:adapterId", h.Adapter.DeleteAdapter)
	public.Post("/adapters/:adapterId/events", h.Adapter.IngestEvent)
	public.Post("/adapters/:adapterId/test", h.Adapter.TestAdapter)

	// Multi-Agent Topology Dashboard
	public.Get("/multi-agent-graph/sessions/:sessionId/analytics", h.MultiAgentGraph.GetTopologyAnalytics)
	public.Get("/multi-agent-graph/sessions/:sessionId/delegations", h.MultiAgentGraph.GetDelegationChains)

	// Agent Cost Autopilot
	public.Get("/cost/hotspots", h.CostOptimizer.GetHotspots)
	public.Post("/cost/autopilot/rules", h.CostOptimizer.CreateAutopilotRule)
	public.Get("/cost/autopilot/rules", h.CostOptimizer.ListAutopilotRules)
	public.Get("/cost/predictions", h.CostOptimizer.GetPredictions)
	public.Get("/cost/autopilot/dashboard", h.CostOptimizer.GetAutopilotDashboard)

	// Collaborative Trace Review
	public.Post("/reviews", h.Collaboration.CreateReview)
	public.Get("/reviews", h.Collaboration.ListReviews)
	public.Get("/reviews/:reviewId", h.Collaboration.GetReview)
	public.Post("/reviews/:reviewId/approve", h.Collaboration.ApproveReview)
	public.Post("/reviews/:reviewId/comments", h.Collaboration.AddReviewComment)
	public.Get("/reviews/:reviewId/comments", h.Collaboration.ListReviewComments)
	public.Post("/review-queues", h.Collaboration.CreateReviewQueue)
	public.Get("/review-queues", h.Collaboration.ListReviewQueues)
	public.Post("/notification-integrations", h.Collaboration.AddNotificationIntegration)
	public.Get("/notification-integrations", h.Collaboration.ListIntegrations)

	// Intelligent Alerting with RCA
	public.Post("/rca/anomalies", h.RCA.DetectAnomalies)
	public.Get("/rca/anomalies", h.RCA.ListAnomalies)
	public.Get("/rca/anomalies/:anomalyId", h.RCA.GetAnomaly)
	public.Post("/rca/anomalies/:anomalyId/acknowledge", h.RCA.AcknowledgeAnomaly)
	public.Post("/rca/alert-channels", h.RCA.CreateAlertChannel)
	public.Get("/rca/alert-channels", h.RCA.ListAlertChannels)
	public.Post("/rca/alert-channels/:channelId/test", h.RCA.TestAlertChannel)
	public.Post("/rca/correlation-rules", h.RCA.CreateCorrelationRule)
	public.Get("/rca/correlation-rules", h.RCA.ListCorrelationRules)
	public.Get("/rca/dashboard", h.RCA.GetAlertDashboardStats)
	public.Post("/rca/investigations", h.RCA.CreateInvestigation)
	public.Put("/rca/investigations/:investigationId", h.RCA.UpdateInvestigation)
	public.Get("/rca/investigations", h.RCA.ListInvestigations)

	// Prompt A/B Testing
	public.Post("/ab-tests", h.ABTesting.CreateTest)
	public.Get("/ab-tests", h.ABTesting.ListTests)
	public.Get("/ab-tests/:testId", h.ABTesting.GetTest)
	public.Post("/ab-tests/:testId/start", h.ABTesting.StartTest)
	public.Post("/ab-tests/:testId/pause", h.ABTesting.PauseTest)
	public.Post("/ab-tests/:testId/stop", h.ABTesting.StopTest)
	public.Post("/ab-tests/:testId/assign", h.ABTesting.AssignVariant)
	public.Post("/ab-tests/:testId/results", h.ABTesting.RecordResult)
	public.Get("/ab-tests/:testId/statistics", h.ABTesting.GetStatistics)
	public.Post("/ab-tests/:testId/select-winner", h.ABTesting.SelectWinner)
	public.Post("/ab-tests/:testId/rollout", h.ABTesting.StartGradualRollout)

	// Federated Trace Analytics
	public.Get("/federated/dashboard", h.FederatedAggregation.GetFederatedAnalyticsDashboard)
	public.Post("/federated/query", h.FederatedAggregation.RunFederatedQuery)

	// Self-Healing Agent Guardrails
	public.Post("/guardrails/policies", h.Guardrails.CreateSelfHealingPolicy)
	public.Get("/guardrails/policies", h.Guardrails.ListSelfHealingPolicies)
	public.Post("/guardrails/evaluate", h.Guardrails.EvaluatePipeline)
	public.Get("/guardrails/dashboard", h.Guardrails.GetDashboardStats)
	public.Get("/guardrails/policies/:policyId/audit", h.Guardrails.GetAuditTrail)
}
