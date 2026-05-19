package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerV10Routes registers Next-Gen v10 feature routes
func registerV10Routes(public fiber.Router, h *Handlers) {
	// LLM Gateway
	public.Get("/gateway/configs", h.Gateway.ListConfigs)
	public.Post("/gateway/configs", h.Gateway.CreateConfig)
	public.Get("/gateway/configs/:configId", h.Gateway.GetConfig)
	public.Put("/gateway/configs/:configId", h.Gateway.UpdateConfig)
	public.Delete("/gateway/configs/:configId", h.Gateway.DeleteConfig)
	public.Post("/gateway/proxy/:configId", h.Gateway.ProxyRequest)
	public.Post("/gateway/configs/:configId/rules", h.Gateway.AddRoutingRule)
	public.Get("/gateway/configs/:configId/rules", h.Gateway.ListRoutingRules)
	public.Delete("/gateway/configs/:configId/rules/:ruleId", h.Gateway.DeleteRoutingRule)
	public.Get("/gateway/stats", h.Gateway.GetStats)

	// Test Generation
	public.Get("/test-suites", h.TestGenerator.ListSuites)
	public.Post("/test-suites", h.TestGenerator.CreateSuite)
	public.Get("/test-suites/:suiteId", h.TestGenerator.GetSuite)
	public.Delete("/test-suites/:suiteId", h.TestGenerator.DeleteSuite)
	public.Post("/test-suites/generate", h.TestGenerator.GenerateFromTraces)
	public.Get("/test-suites/:suiteId/cases", h.TestGenerator.ListTestCases)
	public.Post("/test-suites/:suiteId/run", h.TestGenerator.RunSuite)
	public.Get("/test-suites/:suiteId/results", h.TestGenerator.GetResults)
	public.Post("/test-suites/:suiteId/snapshot", h.TestGenerator.CreateSnapshot)

	// Canary Deployment
	public.Get("/canary/deployments", h.CanaryDeployment.ListDeployments)
	public.Post("/canary/deployments", h.CanaryDeployment.CreateDeployment)
	public.Get("/canary/deployments/:deploymentId", h.CanaryDeployment.GetDeployment)
	public.Post("/canary/deployments/:deploymentId/promote", h.CanaryDeployment.Promote)
	public.Post("/canary/deployments/:deploymentId/rollback", h.CanaryDeployment.Rollback)
	public.Get("/canary/deployments/:deploymentId/metrics", h.CanaryDeployment.GetMetrics)
	public.Get("/canary/active-version", h.CanaryDeployment.GetActiveVersion)

	// Compliance Certification Export
	public.Get("/certifications", h.CertExport.ListCertifications)
	public.Post("/certifications/export", h.CertExport.ExportCertification)
	public.Get("/certifications/:certId", h.CertExport.GetCertification)
	public.Get("/certifications/:certId/download", h.CertExport.Download)
	public.Get("/certifications/frameworks", h.CertExport.ListFrameworks)

	// Data Warehouse Sync
	public.Get("/warehouse/connections", h.WarehouseSync.ListConnections)
	public.Post("/warehouse/connections", h.WarehouseSync.CreateConnection)
	public.Get("/warehouse/connections/:connId", h.WarehouseSync.GetConnection)
	public.Delete("/warehouse/connections/:connId", h.WarehouseSync.DeleteConnection)
	public.Post("/warehouse/connections/:connId/test", h.WarehouseSync.TestConnection)
	public.Post("/warehouse/connections/:connId/sync", h.WarehouseSync.TriggerSync)
	public.Get("/warehouse/connections/:connId/status", h.WarehouseSync.GetSyncStatus)
	public.Post("/warehouse/connections/:connId/schema", h.WarehouseSync.GetSchemaMapping)

	// Trace Review Workflow
	public.Get("/reviews", h.TraceReview.ListReviews)
	public.Post("/reviews", h.TraceReview.CreateReview)
	public.Get("/reviews/:reviewId", h.TraceReview.GetReview)
	public.Put("/reviews/:reviewId", h.TraceReview.UpdateReview)
	public.Post("/reviews/:reviewId/comments", h.TraceReview.AddComment)
	public.Post("/reviews/:reviewId/approve", h.TraceReview.Approve)
	public.Post("/reviews/:reviewId/reject", h.TraceReview.Reject)
	public.Get("/reviews/queue", h.TraceReview.GetQueue)

	// Runbook Engine (extends existing runbook)
	public.Get("/runbook-engine/runbooks", h.RunbookEngine.ListRunbooks)
	public.Post("/runbook-engine/runbooks", h.RunbookEngine.CreateRunbook)
	public.Get("/runbook-engine/runbooks/:runbookId", h.RunbookEngine.GetRunbook)
	public.Put("/runbook-engine/runbooks/:runbookId", h.RunbookEngine.UpdateRunbook)
	public.Post("/runbook-engine/runbooks/:runbookId/activate", h.RunbookEngine.Activate)
	public.Post("/runbook-engine/runbooks/:runbookId/test", h.RunbookEngine.TestRunbook)
	public.Get("/runbook-engine/executions", h.RunbookEngine.ListExecutions)
	public.Get("/runbook-engine/stats", h.RunbookEngine.GetStats)

	// Edge/Mobile Ingest
	public.Post("/edge/ingest", h.EdgeIngest.IngestBatch)
	public.Post("/edge/devices", h.EdgeIngest.RegisterDevice)
	public.Get("/edge/devices", h.EdgeIngest.ListDevices)
	public.Get("/edge/devices/:deviceId/status", h.EdgeIngest.GetDeviceStatus)
	public.Post("/edge/sync", h.EdgeIngest.SyncOfflineData)
	public.Get("/edge/stats", h.EdgeIngest.GetStats)

	// Prompt Impact Analysis
	public.Get("/prompt-impact/analyses", h.PromptImpact.ListAnalyses)
	public.Post("/prompt-impact/analyses", h.PromptImpact.CreateAnalysis)
	public.Get("/prompt-impact/analyses/:analysisId", h.PromptImpact.GetAnalysis)
	public.Get("/prompt-impact/analyses/:analysisId/report", h.PromptImpact.GetReport)
	public.Post("/prompt-impact/compare", h.PromptImpact.CompareVersions)
}
