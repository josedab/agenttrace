package main

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	"github.com/agenttrace/agenttrace/api/internal/middleware"
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
	public.Use(deps.AuthMiddleware.RequireAuth())
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
	registerV11Routes(public, h)
	registerV12Routes(public, h)

	// Internal API routes (JWT auth)
	internal := app.Group("/api/v1")
	internal.Use(deps.AuthMiddleware.RequireJWT())
	internal.Use(deps.RateLimitMiddleware.UserRateLimit(deps.Config.RateLimit.UserMaxPerMinute))
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
		internal.Get(
			"/projects/:projectId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.Projects.GetProject,
		)
		internal.Post("/projects", h.Projects.CreateProject)
		internal.Put(
			"/projects/:projectId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.Projects.UpdateProject,
		)
		internal.Delete(
			"/projects/:projectId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.Projects.DeleteProject,
		)
		internal.Post(
			"/projects/:projectId/members",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.Projects.AddMember,
		)
		internal.Delete(
			"/projects/:projectId/members/:userId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.Projects.RemoveMember,
		)
		internal.Get(
			"/projects/:projectId/role",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.Projects.GetUserRole,
		)

		// API Keys
		internal.Get(
			"/projects/:projectId/api-keys",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.APIKeys.ListAPIKeys,
		)
		internal.Post(
			"/projects/:projectId/api-keys",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.APIKeys.CreateAPIKey,
		)
		internal.Delete(
			"/projects/:projectId/api-keys/:keyId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleAdmin),
			h.APIKeys.DeleteAPIKey,
		)

		// Dashboard metrics
		internal.Get(
			"/projects/:projectId/metrics",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.Traces.GetMetrics,
		)
		internal.Get(
			"/projects/:projectId/outcomes",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.Outcomes.GetOverview,
		)
		internal.Get(
			"/projects/:projectId/outcomes/digest",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.Outcomes.GetDigest,
		)
		internal.Post(
			"/projects/:projectId/outcomes/github-report",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.OutcomeDelivery.DeliverGitHub,
		)
		internal.Post(
			"/projects/:projectId/outcomes/digest/deliver",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.TeamDigest.Deliver,
		)
		internal.Get(
			"/projects/:projectId/traces/:traceId/replay-capabilities",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.ReplayPlan.AssessCapabilities,
		)
		internal.Post(
			"/projects/:projectId/traces/:traceId/replay-plans",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.ReplayPlan.CreatePlan,
		)
		internal.Get(
			"/projects/:projectId/replay-plans/:planId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.ReplayPlan.GetPlan,
		)
		internal.Post(
			"/projects/:projectId/replay-plans/:planId/execute",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.ReplayPlan.ExecutePlan,
		)
		internal.Post(
			"/projects/:projectId/replay-plans/:planId/retry",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.ReplayPlan.RetryPlan,
		)
		internal.Get(
			"/projects/:projectId/replay-plans/:planId/comparison",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.ReplayPlan.GetComparison,
		)
		internal.Get(
			"/projects/:projectId/eval-hub/packages",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.EvalHub.ListPackages,
		)
		internal.Get(
			"/projects/:projectId/eval-hub/packages/:packageId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.EvalHub.GetPackage,
		)
		internal.Post(
			"/projects/:projectId/eval-hub/packages",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.EvalHub.Publish,
		)
		internal.Post(
			"/projects/:projectId/eval-hub/packages/:packageId/fork",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.EvalHub.Fork,
		)
		internal.Post(
			"/projects/:projectId/eval-hub/packages/:packageId/runs",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.EvalHub.Run,
		)
		internal.Get(
			"/projects/:projectId/eval-hub/runs",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.EvalHub.ListRuns,
		)
		internal.Get(
			"/projects/:projectId/eval-hub/runs/:runId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleViewer),
			h.EvalHub.GetRun,
		)
		internal.Post(
			"/projects/:projectId/share-links",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.ShareLink.Create,
		)
		internal.Delete(
			"/projects/:projectId/share-links/:linkId",
			deps.AuthMiddleware.RequireProjectRole(domain.OrgRoleMember),
			h.ShareLink.Revoke,
		)
	}

	// Auth routes (no auth required, rate limited)
	auth := app.Group("/api/auth")
	auth.Use(deps.RateLimitMiddleware.AuthRateLimit(10, time.Minute))
	{
		auth.Post("/login", h.Auth.Login)
		auth.Post("/register", h.Auth.Register)
		auth.Post("/refresh", h.Auth.RefreshToken)
		auth.Post("/logout", h.Auth.Logout)
		auth.Post("/callback/:provider", h.Auth.OAuthCallback)
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

	// WebSocket routes for real-time collaboration (require authentication)
	app.Get(
		"/ws/collaboration/:traceId",
		deps.AuthMiddleware.RequireAuth(),
		deps.AuthMiddleware.RequireTraceAccess(deps.Repositories.Trace),
		h.CollaborationWS.UpgradeCheck(),
		h.CollaborationWS.HandleWebSocket(),
	)

	// WebSocket routes for real-time trace streaming (require authentication)
	app.Get(
		"/ws/streaming/:traceId",
		deps.AuthMiddleware.RequireAuth(),
		deps.AuthMiddleware.RequireTraceAccess(deps.Repositories.Trace),
		h.StreamingWS.UpgradeCheck(),
		h.StreamingWS.HandleWebSocket(),
	)

	// Billing webhook (no auth — Stripe sends directly with signature)
	app.Post("/api/billing/webhook", h.Billing.HandleWebhook)

	// Embed widget script (no auth — token-based access)
	app.Get("/api/public/embed/widget.js", h.Embed.GetWidget)

	shareLimiter := middleware.NewShareRateLimiter(60, time.Minute)
	app.Get("/api/share/:token", shareLimiter.Handler(), h.ShareLink.Resolve)

	// Cloud Sandbox (no auth — demo environment, rate limited)
	sandbox := app.Group("/api/public/cloud-sandbox")
	sandbox.Use(deps.RateLimitMiddleware.IPRateLimit(5, time.Hour))
	sandbox.Post("/", h.Sandbox.CreateCloudSandbox)
	sandbox.Get("/:sessionId", h.Sandbox.GetCloudSandbox)
	sandbox.Post("/:sessionId/extend", h.Sandbox.ExtendCloudSandbox)
}
