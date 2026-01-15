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
}
