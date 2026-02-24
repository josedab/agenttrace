package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerCollaborationRoutes registers collaboration and team routes
func registerCollaborationRoutes(public fiber.Router, h *Handlers) {
	// Collaboration
	public.Get("/collaboration/traces/:traceId/presence", h.Collaboration.GetPresence)
	public.Post("/collaboration/traces/:traceId/annotations", h.Collaboration.AddAnnotation)
	public.Get("/collaboration/traces/:traceId/annotations", h.Collaboration.ListAnnotations)
	public.Post("/collaboration/annotations/:annotationId/resolve", h.Collaboration.ResolveAnnotation)
	public.Post("/collaboration/sessions", h.Collaboration.CreateSharedSession)

	// Collaboration - Discussions
	public.Post("/collaboration/discussions", h.Collaboration.CreateDiscussion)
	public.Post("/collaboration/discussions/:threadId/messages", h.Collaboration.AddMessage)

	// Collaboration - Evaluation Queues
	public.Post("/collaboration/eval-queues", h.Collaboration.CreateEvalQueue)

	// Collaborative Annotations
	public.Get("/annotations/traces/:traceId", h.Annotation.List)
	public.Post("/annotations", h.Annotation.Create)
	public.Post("/annotations/:annotationId/reply", h.Annotation.Reply)
	public.Post("/annotations/:annotationId/resolve", h.Annotation.Resolve)
	public.Get("/annotations/presence/:traceId", h.Annotation.GetPresence)

	// Collaboration Patterns
	public.Get("/collab-patterns", h.CollabPattern.List)
	public.Get("/collab-patterns/:patternId", h.CollabPattern.Get)
	public.Post("/collab-patterns/:patternId/deploy", h.CollabPattern.Deploy)
	public.Get("/collab-patterns/deployments", h.CollabPattern.GetDeployments)
	public.Get("/collab-patterns/:patternId/analytics", h.CollabPattern.GetAnalytics)

	// Collaboration Hub
	public.Post("/collab/queues", h.CollabHub.CreateReviewQueue)
	public.Get("/collab/queues", h.CollabHub.ListReviewQueues)
	public.Post("/collab/reviews", h.CollabHub.AssignReview)
	public.Post("/collab/reviews/:assignmentId/complete", h.CollabHub.CompleteReview)
	public.Post("/collab/standards", h.CollabHub.CreateQualityStandard)
	public.Get("/collab/standards", h.CollabHub.ListQualityStandards)
	public.Get("/collab/activity", h.CollabHub.GetActivityFeed)

	// Team Intelligence
	public.Get("/team/dashboard", h.TeamIntelligence.GetDashboard)
	public.Get("/team/roi", h.TeamIntelligence.CalculateROI)

	// Cross-Organization Benchmarking
	public.Post("/cross-org/submit", h.CrossOrg.Submit)
	public.Get("/cross-org/report", h.CrossOrg.GetReport)
	public.Get("/cross-org/industry/:category", h.CrossOrg.GetIndustryStats)
}
