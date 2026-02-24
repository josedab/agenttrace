package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerTracesRoutes registers trace, session, and observation routes
func registerTracesRoutes(public fiber.Router, h *Handlers) {
	// Ingestion endpoints
	public.Post("/ingestion", h.Ingestion.BatchIngestion)
	public.Post("/traces", h.Ingestion.CreateTrace)
	public.Post("/spans", h.Ingestion.CreateSpan)
	public.Post("/generations", h.Ingestion.CreateGeneration)
	public.Post("/events", h.Ingestion.CreateEvent)

	// OTLP-compatible ingestion
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

	// Real-time events (SSE)
	public.Get("/events", h.Events.StreamEvents)

	// Metrics
	public.Get("/metrics/project", h.Traces.GetMetrics)

	// Distributed Tracing
	public.Get("/distributed/traces/:traceId", h.DistributedTrace.GetTrace)
	public.Get("/distributed/service-map", h.DistributedTrace.GetServiceMap)
	public.Post("/distributed/correlate", h.DistributedTrace.Correlate)

	// Multi-Modal Traces
	public.Post("/multimodal/attachments", h.MultiModal.Register)
	public.Get("/multimodal/traces/:traceId", h.MultiModal.GetTraceAttachments)
	public.Get("/multimodal/attachments/:attachmentId", h.MultiModal.GetAttachment)
	public.Get("/multimodal/traces/:traceId/summary", h.MultiModal.GetSummary)
	public.Get("/multimodal/attachments", h.MultiModal.List)

	// Real-time streaming
	public.Get("/streams", h.Streaming.GetActiveStreams)
	public.Get("/traces/:traceId/stream", h.Streaming.StreamTrace)
	public.Get("/traces/:traceId/live-metrics", h.Streaming.GetLiveMetrics)
	public.Get("/traces/:traceId/activities", h.Streaming.GetRecentActivities)
	public.Post("/traces/:traceId/intervene", h.Streaming.RequestIntervention)
	public.Get("/traces/:traceId/interventions", h.Streaming.GetPendingInterventions)
	public.Post("/traces/:traceId/interventions/:interventionId/ack", h.Streaming.AcknowledgeIntervention)

	// IDE Inline Trace Viewer
	public.Get("/ide/file-mapping", h.IDETraceView.GetFileMapping)
	public.Post("/ide/batch-mappings", h.IDETraceView.GetBatchMappings)
	public.Get("/ide/trace-context/:traceId", h.IDETraceView.GetTraceContext)
}
