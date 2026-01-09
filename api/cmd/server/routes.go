package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerRoutes registers all HTTP routes
func registerRoutes(app *fiber.App, deps *Dependencies) {
	// Health check routes (no auth required)
	app.Get("/health", deps.HealthHandler.Health)
	app.Get("/healthz", deps.HealthHandler.Health)
	app.Get("/livez", deps.HealthHandler.Liveness)
	app.Get("/live", deps.HealthHandler.Liveness)
	app.Get("/readyz", deps.HealthHandler.Readiness)
	app.Get("/ready", deps.HealthHandler.Readiness)
	app.Get("/version", deps.HealthHandler.Version)

	// API Documentation routes (no auth required)
	deps.DocsHandler.RegisterRoutes(app)

	// Public API routes (API key auth)
	public := app.Group("/api/public")
	public.Use(deps.AuthMiddleware.RequireAPIKey())
	public.Use(deps.RateLimitMiddleware.Handler())
	{
		// Ingestion endpoints
		public.Post("/ingestion", deps.IngestionHandler.BatchIngestion)
		public.Post("/traces", deps.IngestionHandler.CreateTrace)
		public.Post("/spans", deps.IngestionHandler.CreateSpan)
		public.Post("/generations", deps.IngestionHandler.CreateGeneration)
		public.Post("/events", deps.IngestionHandler.CreateEvent)

		// OTLP-compatible ingestion (also uses BatchIngestion)
		public.Post("/v1/traces", deps.IngestionHandler.BatchIngestion)

		// Trace queries
		public.Get("/traces", deps.TracesHandler.ListTraces)
		public.Get("/traces/search", deps.TracesHandler.SearchTraces)
		public.Get("/traces/:id", deps.TracesHandler.GetTrace)
		public.Get("/traces/:id/observations", deps.TracesHandler.GetTraceObservations)
		public.Get("/traces/:id/stats", deps.TracesHandler.GetTraceStats)
		public.Delete("/traces/:id", deps.TracesHandler.DeleteTrace)

		// Sessions
		public.Get("/sessions", deps.TracesHandler.GetSessions)
		public.Get("/sessions/:id", deps.TracesHandler.GetSession)

		// Scores
		public.Get("/scores", deps.ScoresHandler.ListScores)
		public.Get("/scores/stats", deps.ScoresHandler.GetScoreStats)
		public.Get("/scores/:id", deps.ScoresHandler.GetScore)
		public.Post("/scores", deps.ScoresHandler.CreateScore)
		public.Post("/scores/batch", deps.ScoresHandler.BatchCreateScores)
		public.Put("/scores/:id", deps.ScoresHandler.UpdateScore)
		public.Get("/traces/:traceId/scores", deps.ScoresHandler.GetTraceScores)

		// Prompts
		public.Get("/prompts", deps.PromptsHandler.ListPrompts)
		public.Get("/prompts/:name", deps.PromptsHandler.GetPrompt)
		public.Post("/prompts", deps.PromptsHandler.CreatePrompt)
		public.Put("/prompts/:name", deps.PromptsHandler.UpdatePrompt)
		public.Delete("/prompts/:name", deps.PromptsHandler.DeletePrompt)
		public.Get("/prompts/:name/versions", deps.PromptsHandler.ListVersions)
		public.Post("/prompts/:name/labels", deps.PromptsHandler.SetLabel)
		public.Delete("/prompts/:name/labels/:label", deps.PromptsHandler.RemoveLabel)
		public.Post("/prompts/:name/compile", deps.PromptsHandler.CompilePrompt)

		// Datasets
		public.Get("/datasets", deps.DatasetsHandler.ListDatasets)
		public.Get("/datasets/:id", deps.DatasetsHandler.GetDataset)
		public.Post("/datasets", deps.DatasetsHandler.CreateDataset)
		public.Put("/datasets/:id", deps.DatasetsHandler.UpdateDataset)
		public.Delete("/datasets/:id", deps.DatasetsHandler.DeleteDataset)
		public.Get("/datasets/:id/items", deps.DatasetsHandler.ListItems)
		public.Post("/datasets/:id/items", deps.DatasetsHandler.CreateItem)
		public.Put("/datasets/:datasetId/items/:id", deps.DatasetsHandler.UpdateItem)
		public.Delete("/datasets/:datasetId/items/:id", deps.DatasetsHandler.DeleteItem)
		public.Get("/datasets/:id/runs", deps.DatasetsHandler.ListRuns)
		public.Post("/datasets/:id/runs", deps.DatasetsHandler.CreateRun)
		public.Get("/datasets/:datasetId/runs/:id", deps.DatasetsHandler.GetRun)
		public.Post("/datasets/:datasetId/runs/:id/items", deps.DatasetsHandler.AddRunItem)

		// Evaluators
		public.Get("/evaluators", deps.EvaluatorsHandler.ListEvaluators)
		public.Get("/evaluators/:id", deps.EvaluatorsHandler.GetEvaluator)
		public.Post("/evaluators", deps.EvaluatorsHandler.CreateEvaluator)
		public.Put("/evaluators/:id", deps.EvaluatorsHandler.UpdateEvaluator)
		public.Delete("/evaluators/:id", deps.EvaluatorsHandler.DeleteEvaluator)
		public.Get("/evaluator-templates", deps.EvaluatorsHandler.ListTemplates)

		// Metrics
		public.Get("/metrics/project", deps.TracesHandler.GetMetrics)

		// Real-time events (SSE)
		public.Get("/events", deps.EventsHandler.StreamEvents)

		// Checkpoints (agent-specific)
		public.Get("/checkpoints", deps.CheckpointsHandler.ListCheckpoints)
		public.Get("/checkpoints/:checkpointId", deps.CheckpointsHandler.GetCheckpoint)
		public.Post("/checkpoints", deps.CheckpointsHandler.CreateCheckpoint)
		public.Post("/checkpoints/:checkpointId/restore", deps.CheckpointsHandler.RestoreCheckpoint)
		public.Get("/traces/:traceId/checkpoints", deps.CheckpointsHandler.GetTraceCheckpoints)

		// Git Links (agent-specific)
		public.Get("/git-links", deps.GitLinksHandler.ListGitLinks)
		public.Get("/git-links/timeline", deps.GitLinksHandler.GetTimeline)
		public.Get("/git-links/commit/:commitSha", deps.GitLinksHandler.GetByCommit)
		public.Get("/git-links/:gitLinkId", deps.GitLinksHandler.GetGitLink)
		public.Post("/git-links", deps.GitLinksHandler.CreateGitLink)
		public.Get("/traces/:traceId/git-links", deps.GitLinksHandler.GetTraceGitLinks)

		// File Operations (agent-specific)
		public.Get("/file-operations", deps.FileOperationsHandler.ListFileOperations)
		public.Get("/file-operations/stats", deps.FileOperationsHandler.GetFileOperationStats)
		public.Post("/file-operations", deps.FileOperationsHandler.CreateFileOperation)
		public.Post("/file-operations/batch", deps.FileOperationsHandler.BatchCreateFileOperations)
		public.Get("/traces/:traceId/file-operations", deps.FileOperationsHandler.GetTraceFileOperations)

		// Terminal Commands (agent-specific)
		public.Get("/terminal-commands", deps.TerminalCommandsHandler.ListTerminalCommands)
		public.Get("/terminal-commands/stats", deps.TerminalCommandsHandler.GetTerminalCommandStats)
		public.Post("/terminal-commands", deps.TerminalCommandsHandler.CreateTerminalCommand)
		public.Post("/terminal-commands/batch", deps.TerminalCommandsHandler.BatchCreateTerminalCommands)
		public.Get("/traces/:traceId/terminal-commands", deps.TerminalCommandsHandler.GetTraceTerminalCommands)

		// CI Runs (agent-specific)
		public.Get("/ci-runs", deps.CIRunsHandler.ListCIRuns)
		public.Get("/ci-runs/stats", deps.CIRunsHandler.GetCIRunStats)
		public.Get("/ci-runs/provider/:providerRunId", deps.CIRunsHandler.GetCIRunByProviderID)
		public.Get("/ci-runs/:ciRunId", deps.CIRunsHandler.GetCIRun)
		public.Post("/ci-runs", deps.CIRunsHandler.CreateCIRun)
		public.Patch("/ci-runs/:ciRunId", deps.CIRunsHandler.UpdateCIRun)
		public.Post("/ci-runs/:ciRunId/traces", deps.CIRunsHandler.AddTraceToCIRun)
		public.Post("/ci-runs/:ciRunId/complete", deps.CIRunsHandler.CompleteCIRun)

		// Export endpoints
		public.Post("/export/data", deps.ExportHandler.ExportData)
		public.Post("/export/dataset", deps.ExportHandler.ExportDataset)

		// Import endpoints
		public.Post("/import/dataset", deps.ImportHandler.ImportDataset)
		public.Post("/import/dataset/csv", deps.ImportHandler.ImportDatasetCSV)
		public.Post("/import/dataset/openai-finetune", deps.ImportHandler.ImportOpenAIFinetune)
		public.Post("/import/prompt", deps.ImportHandler.ImportPrompt)
	}

	// Internal API routes (JWT auth)
	internal := app.Group("/api/v1")
	internal.Use(deps.AuthMiddleware.RequireJWT())
	{
		// Current user
		internal.Get("/me", deps.AuthHandler.GetCurrentUser)

		// Organizations
		internal.Get("/organizations", deps.OrganizationsHandler.ListOrganizations)
		internal.Get("/organizations/slug/:slug", deps.OrganizationsHandler.GetOrganizationBySlug)
		internal.Get("/organizations/:id", deps.OrganizationsHandler.GetOrganization)
		internal.Post("/organizations", deps.OrganizationsHandler.CreateOrganization)
		internal.Put("/organizations/:id", deps.OrganizationsHandler.UpdateOrganization)
		internal.Delete("/organizations/:id", deps.OrganizationsHandler.DeleteOrganization)
		internal.Get("/organizations/:orgId/members/:userId", deps.OrganizationsHandler.GetMember)

		// Projects
		internal.Get("/projects", deps.ProjectsHandler.ListProjects)
		internal.Get("/projects/:id", deps.ProjectsHandler.GetProject)
		internal.Post("/projects", deps.ProjectsHandler.CreateProject)
		internal.Put("/projects/:id", deps.ProjectsHandler.UpdateProject)
		internal.Delete("/projects/:id", deps.ProjectsHandler.DeleteProject)
		internal.Post("/projects/:id/members", deps.ProjectsHandler.AddMember)
		internal.Delete("/projects/:projectId/members/:userId", deps.ProjectsHandler.RemoveMember)
		internal.Get("/projects/:id/role", deps.ProjectsHandler.GetUserRole)

		// API Keys
		internal.Get("/projects/:id/api-keys", deps.APIKeysHandler.ListAPIKeys)
		internal.Post("/projects/:id/api-keys", deps.APIKeysHandler.CreateAPIKey)
		internal.Delete("/api-keys/:id", deps.APIKeysHandler.DeleteAPIKey)

		// Dashboard metrics
		internal.Get("/projects/:id/metrics", deps.TracesHandler.GetMetrics)
	}

	// Auth routes (no auth required)
	auth := app.Group("/api/auth")
	{
		auth.Post("/login", deps.AuthHandler.Login)
		auth.Post("/register", deps.AuthHandler.Register)
		auth.Post("/refresh", deps.AuthHandler.RefreshToken)
		auth.Post("/logout", deps.AuthHandler.Logout)
		auth.Get("/callback/:provider", deps.AuthHandler.OAuthCallback)
	}

	// User feedback endpoint (special auth - accepts both API key and user token)
	app.Post("/api/public/feedback", deps.AuthMiddleware.RequireAuth(), deps.ScoresHandler.SubmitFeedback)
}
