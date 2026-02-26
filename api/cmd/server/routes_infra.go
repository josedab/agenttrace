package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerInfraRoutes registers infrastructure, federation, and integration routes
func registerInfraRoutes(public fiber.Router, h *Handlers) {
	// Webhooks
	public.Get("/webhooks", h.Webhook.ListWebhooks)
	public.Get("/webhooks/:id", h.Webhook.GetWebhook)
	public.Post("/webhooks", h.Webhook.CreateWebhook)
	public.Patch("/webhooks/:id", h.Webhook.UpdateWebhook)
	public.Delete("/webhooks/:id", h.Webhook.DeleteWebhook)
	public.Post("/webhooks/:id/test", h.Webhook.TestWebhook)
	public.Get("/webhooks/:id/deliveries", h.Webhook.ListWebhookDeliveries)

	// Webhook Orchestration
	public.Get("/webhook-rules", h.WebhookOrchestration.ListRules)
	public.Post("/webhook-rules", h.WebhookOrchestration.CreateRule)
	public.Delete("/webhook-rules/:ruleId", h.WebhookOrchestration.DeleteRule)
	public.Get("/webhook-rules/templates", h.WebhookOrchestration.GetTemplates)
	public.Get("/webhook-rules/deliveries", h.WebhookOrchestration.ListDeliveries)
	public.Post("/webhook-rules/:ruleId/test", h.WebhookOrchestration.TestRule)

	// Federation & OTLP Export
	public.Get("/federation/peers", h.Federation.ListPeers)
	public.Post("/federation/peers", h.Federation.AddPeer)
	public.Delete("/federation/peers/:peerId", h.Federation.RemovePeer)
	public.Post("/federation/query", h.Federation.FederatedQuery)
	public.Get("/federation/destinations", h.Federation.ListExportDestinations)
	public.Post("/federation/destinations", h.Federation.CreateExportDestination)

	// Federated Learning
	public.Get("/federated/rings", h.FederatedLearning.ListRings)
	public.Post("/federated/rings/join", h.FederatedLearning.JoinRing)
	public.Get("/federated/rings/:ringId/insights", h.FederatedLearning.GetInsights)
	public.Get("/federated/config", h.FederatedLearning.GetConfig)
	public.Put("/federated/config", h.FederatedLearning.UpdateConfig)

	// Federated Trace Aggregation
	public.Get("/federated-aggregation/dashboard", h.FederatedAggregation.GetDashboard)
	public.Get("/federated-aggregation/instances", h.FederatedAggregation.ListInstances)
	public.Post("/federated-aggregation/instances", h.FederatedAggregation.RegisterInstance)
	public.Post("/federated-aggregation/metrics", h.FederatedAggregation.SubmitMetrics)
	public.Get("/federated-aggregation/benchmarks", h.FederatedAggregation.GetBenchmarks)
	public.Get("/federated-aggregation/insights", h.FederatedAggregation.GetInsights)
	public.Post("/federated-aggregation/anonymized-benchmark", h.FederatedAggregation.SubmitAnonymizedBenchmark)
	public.Get("/federated-aggregation/baselines", h.FederatedAggregation.GetIndustryBaselines)
	public.Get("/federated-aggregation/mesh-status", h.FederatedAggregation.GetMeshStatus)

	// OpenTelemetry Native Compatibility
	public.Post("/otel/destinations", h.OTelCompat.CreateExportDestination)
	public.Get("/otel/destinations", h.OTelCompat.ListExportDestinations)
	public.Delete("/otel/destinations/:destinationId", h.OTelCompat.DeleteExportDestination)
	public.Get("/otel/mappings", h.OTelCompat.GetOTelMappings)
	public.Get("/otel/dashboard", h.OTelCompat.GetOTelDashboard)
	public.Post("/otel/collector-config", h.OTelCompat.GenerateCollectorConfig)

	// Plugin System
	public.Get("/plugins", h.Plugin.List)
	public.Post("/plugins", h.Plugin.Install)
	public.Get("/plugins/:pluginId", h.Plugin.Get)
	public.Post("/plugins/:pluginId/activate", h.Plugin.Activate)
	public.Post("/plugins/:pluginId/disable", h.Plugin.Disable)
	public.Post("/plugins/:pluginId/execute", h.Plugin.Execute)
	public.Delete("/plugins/:pluginId", h.Plugin.Uninstall)

	// Universal Agent Protocol Adapters
	public.Get("/adapters", h.Plugin.ListAdapters)
	public.Post("/adapters", h.Plugin.InstallAdapter)
	public.Post("/adapters/events", h.Plugin.IngestAdapterEvent)

	// Migration
	public.Get("/migrations", h.Migration.ListMigrations)
	public.Post("/migrations", h.Migration.StartMigration)
	public.Get("/migrations/:jobId", h.Migration.GetMigration)
	public.Post("/migrations/validate", h.Migration.ValidateSource)

	// Framework Auto-Instrumentation
	public.Get("/instrumentation/frameworks", h.Instrumentation.ListFrameworks)
	public.Get("/instrumentation/setup/:framework", h.Instrumentation.GetSetup)

	// Agent Marketplace
	public.Get("/marketplace", h.Marketplace.Search)
	public.Get("/marketplace/featured", h.Marketplace.Featured)
	public.Get("/marketplace/:packageId", h.Marketplace.Get)
	public.Post("/marketplace", h.Marketplace.Publish)
	public.Post("/marketplace/:packageId/install", h.Marketplace.Install)
	public.Post("/marketplace/:packageId/rate", h.Marketplace.Rate)

	// White-Label & Embedding
	public.Get("/embed/config", h.Embed.GetConfig)
	public.Post("/embed/config", h.Embed.CreateConfig)
	public.Put("/embed/config", h.Embed.UpdateConfig)
	public.Post("/embed/token", h.Embed.GenerateToken)

	// Mobile API
	public.Post("/mobile/devices", h.Mobile.RegisterDevice)
	public.Get("/mobile/dashboard", h.Mobile.GetDashboard)
	public.Get("/mobile/notifications", h.Mobile.ListNotifications)

	// Custom Metrics
	public.Get("/custom-metrics", h.CustomMetrics.ListMetrics)
	public.Post("/custom-metrics", h.CustomMetrics.CreateMetric)
	public.Get("/custom-metrics/:metricId/values", h.CustomMetrics.GetValues)
	public.Get("/custom-metrics/dashboards", h.CustomMetrics.ListDashboards)
	public.Post("/custom-metrics/dashboards", h.CustomMetrics.CreateDashboard)
	public.Get("/custom-metrics/alerts", h.CustomMetrics.ListAlerts)
	public.Post("/custom-metrics/alerts", h.CustomMetrics.CreateAlert)
}
