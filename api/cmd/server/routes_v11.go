package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerV11Routes registers v11 feature routes (trace diff, eval playground, etc.)
func registerV11Routes(public fiber.Router, h *Handlers) {
	// Trace Diff & Regression Bisect
	public.Post("/trace-diff", h.TraceDiff.DiffTraces)
	public.Get("/bisect/sessions", h.TraceDiff.ListBisectSessions)
	public.Post("/bisect/sessions", h.TraceDiff.StartBisect)
	public.Get("/bisect/sessions/:sessionId", h.TraceDiff.GetBisectSession)
	public.Post("/bisect/sessions/:sessionId/verdict", h.TraceDiff.SubmitBisectVerdict)
	public.Get("/bisect/sessions/:sessionId/result", h.TraceDiff.GetBisectResult)

	// Interactive Evaluation Playground
	public.Post("/eval-playground/sessions", h.EvalPlayground.CreateSession)
	public.Get("/eval-playground/sessions/:sessionId", h.EvalPlayground.GetSession)
	public.Post("/eval-playground/execute", h.EvalPlayground.Execute)
	public.Get("/eval-playground/templates", h.EvalPlayground.ListTemplates)
	public.Post("/eval-playground/share", h.EvalPlayground.ShareSession)
	public.Get("/eval-playground/shared/:shareToken", h.EvalPlayground.GetSharedSession)

	// Trace-Linked Code Impact Map
	public.Get("/traces/:traceId/code-impact", h.CodeImpact.GetCodeImpact)
	public.Get("/code-impact/summary", h.CodeImpact.GetProjectImpactSummary)
	public.Get("/code-impact/file-tree", h.CodeImpact.GetFileTree)

	// Real-Time Streaming Dashboard
	public.Get("/streaming-dashboard", h.StreamingDashboard.GetDashboard)
	public.Get("/streaming-dashboard/ws", h.StreamingDashboard.WebSocketHandler)
	public.Get("/streaming-dashboard/config", h.StreamingDashboard.GetConfig)
	public.Put("/streaming-dashboard/config", h.StreamingDashboard.UpdateConfig)

	// Evaluation Dataset Marketplace
	public.Get("/eval-marketplace/datasets", h.EvalMarketplace.ListDatasets)
	public.Get("/eval-marketplace/datasets/:datasetId", h.EvalMarketplace.GetDataset)
	public.Post("/eval-marketplace/datasets", h.EvalMarketplace.PublishDataset)
	public.Post("/eval-marketplace/datasets/:datasetId/import", h.EvalMarketplace.ImportDataset)
	public.Get("/eval-marketplace/categories", h.EvalMarketplace.ListCategories)
	public.Post("/eval-marketplace/datasets/:datasetId/rate", h.EvalMarketplace.RateDataset)

	// Session-Based Trace Journeys
	public.Get("/sessions/:sessionId/journey", h.SessionJourney.GetJourney)
	public.Get("/sessions/:sessionId/phases", h.SessionJourney.GetPhases)
	public.Get("/session-journeys/recent", h.SessionJourney.ListRecentJourneys)

	// Webhook Trace Enrichment Pipeline
	public.Get("/enrichment/rules", h.TraceEnrichment.ListRules)
	public.Post("/enrichment/rules", h.TraceEnrichment.CreateRule)
	public.Put("/enrichment/rules/:ruleId", h.TraceEnrichment.UpdateRule)
	public.Delete("/enrichment/rules/:ruleId", h.TraceEnrichment.DeleteRule)
	public.Get("/enrichment/sources", h.TraceEnrichment.ListSources)
	public.Post("/enrichment/test", h.TraceEnrichment.TestRule)

	// Cost Forecast & Budget Simulator
	public.Get("/cost-forecast", h.CostForecast.GetForecast)
	public.Post("/cost-forecast/simulate", h.CostForecast.Simulate)
	public.Get("/cost-forecast/history", h.CostForecast.GetHistory)
	public.Post("/cost-forecast/budget-plan", h.CostForecast.CreateBudgetPlan)

	// Trace Annotation Knowledge Base
	public.Get("/knowledge-base/entries", h.TraceKB.ListEntries)
	public.Post("/knowledge-base/entries", h.TraceKB.CreateEntry)
	public.Get("/knowledge-base/entries/:entryId", h.TraceKB.GetEntry)
	public.Get("/knowledge-base/search", h.TraceKB.Search)
	public.Get("/knowledge-base/suggestions", h.TraceKB.GetSuggestions)

	// OpenTelemetry-Native Trace Bridge
	public.Get("/otel-bridge/config", h.OTelBridge.GetConfig)
	public.Put("/otel-bridge/config", h.OTelBridge.UpdateConfig)
	public.Get("/otel-bridge/destinations", h.OTelBridge.ListDestinations)
	public.Post("/otel-bridge/destinations", h.OTelBridge.AddDestination)
	public.Delete("/otel-bridge/destinations/:destId", h.OTelBridge.RemoveDestination)
	public.Post("/otel-bridge/import", h.OTelBridge.ImportSpans)
	public.Get("/otel-bridge/stats", h.OTelBridge.GetStats)
}
