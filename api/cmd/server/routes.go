package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerRoutes registers all HTTP routes by delegating to domain-specific
// registration functions defined in routes_*.go files.
func registerRoutes(app *fiber.App, deps *Dependencies) {
	h := deps.Handlers

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

	// Register domain-specific public routes
	registerTracesRoutes(public, h)
	registerScoresRoutes(public, h)
	registerPromptsRoutes(public, h)
	registerCostRoutes(public, h)
	registerComplianceRoutes(public, h)
	registerAgentsRoutes(public, h)
	registerAnalysisRoutes(public, h)
	registerCollaborationRoutes(public, h)
	registerInfraRoutes(public, h)
	registerV10Routes(public, h)

	// Internal API routes (JWT auth)
	internal := app.Group("/api/v1")
	internal.Use(deps.AuthMiddleware.RequireJWT())
	internal.Use(deps.RateLimitMiddleware.UserRateLimit(100))
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

	// WebSocket routes for real-time trace streaming
	app.Use("/ws/streaming", h.StreamingWS.UpgradeCheck())
	app.Get("/ws/streaming/:traceId", h.StreamingWS.HandleWebSocket())

	// Billing webhook (no auth — Stripe sends directly with signature)
	app.Post("/api/billing/webhook", h.Billing.HandleWebhook)

	// Embed widget script (no auth — token-based access)
	app.Get("/api/public/embed/widget.js", h.Embed.GetWidget)

	// Cloud Sandbox (no auth — demo environment)
	app.Post("/api/public/cloud-sandbox", h.Sandbox.CreateCloudSandbox)
	app.Get("/api/public/cloud-sandbox/:sessionId", h.Sandbox.GetCloudSandbox)
	app.Post("/api/public/cloud-sandbox/:sessionId/extend", h.Sandbox.ExtendCloudSandbox)
}
