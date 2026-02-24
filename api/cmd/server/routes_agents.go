package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerAgentsRoutes registers agent management, graph, and comparison routes
func registerAgentsRoutes(public fiber.Router, h *Handlers) {
	// Agent Graph (multi-agent visualization)
	public.Get("/traces/:traceId/graph", h.AgentGraph.BuildGraph)
	public.Post("/agent-graph/compare", h.AgentGraph.CompareGraphs)

	// Agent Rollback & Versioning
	public.Get("/agent-versions", h.AgentVersion.ListVersions)
	public.Post("/agent-versions", h.AgentVersion.CreateVersion)
	public.Get("/agent-versions/active", h.AgentVersion.GetActive)
	public.Get("/agent-versions/:versionId", h.AgentVersion.GetVersion)
	public.Post("/agent-versions/:versionId/rollback", h.AgentVersion.Rollback)
	public.Post("/agent-versions/diff", h.AgentVersion.DiffVersions)

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

	// Multi-Agent Collaboration Graph
	public.Get("/multi-agent/sessions", h.MultiAgentGraph.ListSessions)
	public.Post("/multi-agent/analyze", h.MultiAgentGraph.AnalyzeSession)
	public.Get("/multi-agent/sessions/:sessionId", h.MultiAgentGraph.GetSession)

	// Agent Comparison Dashboard
	public.Post("/agent-comparison/profiles", h.AgentComparison.CreateProfile)
	public.Get("/agent-comparison/profiles", h.AgentComparison.ListProfiles)
	public.Get("/agent-comparison/profiles/:profileId", h.AgentComparison.GetProfile)
	public.Post("/agent-comparison/compare", h.AgentComparison.RunComparison)
	public.Get("/agent-comparison/comparisons", h.AgentComparison.ListComparisons)
	public.Get("/agent-comparison/comparisons/:comparisonId", h.AgentComparison.GetComparison)
	public.Get("/agent-comparison/trends", h.AgentComparison.GetTrends)
	public.Get("/agent-comparison/summary", h.AgentComparison.GetDashboardSummary)

	// Autonomy Gradient
	public.Get("/autonomy/dashboard", h.Autonomy.GetDashboard)
	public.Get("/autonomy/:agentName", h.Autonomy.GetConfig)
	public.Post("/autonomy", h.Autonomy.SetAutonomy)
	public.Get("/autonomy/:agentName/trust", h.Autonomy.GetTrustEvolution)

	// Agent Handoff Protocol
	public.Post("/handoffs", h.Handoff.Initiate)
	public.Post("/handoffs/:handoffId/accept", h.Handoff.Accept)
	public.Post("/handoffs/:handoffId/complete", h.Handoff.Complete)
	public.Get("/handoffs/chain/:traceId", h.Handoff.GetChain)
	public.Get("/handoffs/stats", h.Handoff.GetStats)

	// Agent Memory & Context
	public.Post("/memory/analyze", h.AgentMemory.AnalyzeMemory)
	public.Get("/memory/traces/:traceId/snapshots/:stepIndex", h.AgentMemory.GetSnapshot)
	public.Get("/memory/optimizations", h.AgentMemory.GetOptimizations)

	// Agent Reasoning Explorer
	public.Get("/traces/:traceId/reasoning", h.Reasoning.GetReasoningTree)
	public.Get("/traces/:traceId/reasoning/:nodeId", h.Reasoning.GetNode)

	// Observability Copilot
	public.Post("/copilot/ask", h.Copilot.Ask)
	public.Get("/copilot/suggestions", h.Copilot.GetSuggestions)
	public.Get("/copilot/insights", h.Copilot.GetInsights)

	// Knowledge Graph
	public.Get("/knowledge-graph", h.KnowledgeGraph.Build)
	public.Post("/knowledge-graph/query", h.KnowledgeGraph.Query)
	public.Get("/knowledge-graph/stats", h.KnowledgeGraph.GetStats)

	// Agent Knowledge Graph (next-gen)
	public.Get("/agent-knowledge-graph", h.AgentKnowledgeGraph.BuildGraph)
	public.Post("/agent-knowledge-graph/query", h.AgentKnowledgeGraph.QueryGraph)
	public.Get("/agent-knowledge-graph/evolution", h.AgentKnowledgeGraph.GetEvolution)
	public.Get("/agent-knowledge-graph/stats", h.AgentKnowledgeGraph.GetStats)

	// SLOs
	public.Get("/slos", h.SLO.List)
	public.Post("/slos", h.SLO.Create)
	public.Get("/slos/:sloId/status", h.SLO.GetStatus)
	public.Get("/slos/report", h.SLO.GetReport)
	public.Get("/slos/:sloId/history", h.SLO.GetHistory)

	// Intent Verification
	public.Post("/intents", h.Intent.Declare)
	public.Post("/intents/:intentId/verify", h.Intent.Verify)
	public.Get("/intents/:intentId", h.Intent.Get)
	public.Get("/intents/stats", h.Intent.GetStats)
}
