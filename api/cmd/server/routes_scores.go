package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerScoresRoutes registers score and evaluation routes
func registerScoresRoutes(public fiber.Router, h *Handlers) {
	// Scores
	public.Get("/scores", h.Scores.ListScores)
	public.Get("/scores/stats", h.Scores.GetScoreStats)
	public.Get("/scores/:scoreId", h.Scores.GetScore)
	public.Post("/scores", h.Scores.CreateScore)
	public.Post("/scores/batch", h.Scores.BatchCreateScores)
	public.Put("/scores/:scoreId", h.Scores.UpdateScore)
	public.Get("/traces/:traceId/scores", h.Scores.GetTraceScores)

	// Evaluators
	public.Get("/evaluators", h.Evaluators.ListEvaluators)
	public.Get("/evaluators/:id", h.Evaluators.GetEvaluator)
	public.Post("/evaluators", h.Evaluators.CreateEvaluator)
	public.Put("/evaluators/:id", h.Evaluators.UpdateEvaluator)
	public.Delete("/evaluators/:id", h.Evaluators.DeleteEvaluator)
	public.Get("/evaluator-templates", h.Evaluators.ListTemplates)

	// Agent Performance Scorecards
	public.Get("/scorecards", h.Scorecard.ListScorecards)
	public.Post("/scorecards", h.Scorecard.Generate)
	public.Get("/scorecards/config", h.Scorecard.GetConfig)
	public.Post("/scorecards/config", h.Scorecard.ConfigureAuto)
	public.Get("/scorecards/:id", h.Scorecard.GetScorecard)

	// Benchmarks
	public.Get("/benchmarks", h.Benchmarks.ListBenchmarks)
	public.Get("/benchmarks/:benchmarkId", h.Benchmarks.GetBenchmark)
	public.Post("/benchmarks", h.Benchmarks.CreateBenchmark)
	public.Post("/benchmarks/:benchmarkId/submit", h.Benchmarks.Submit)
	public.Get("/benchmarks/:benchmarkId/leaderboard", h.Benchmarks.GetLeaderboard)
	public.Post("/benchmarks/:benchmarkId/compare", h.Benchmarks.CompareSubmissions)
	public.Get("/benchmarks/:benchmarkId/stats", h.Benchmarks.GetStats)

	// Agent Performance Benchmarks (next-gen)
	public.Get("/agent-benchmarks/suites", h.AgentBenchmark.ListSuites)
	public.Post("/agent-benchmarks/suites", h.AgentBenchmark.CreateSuite)
	public.Get("/agent-benchmarks/suites/:suiteId", h.AgentBenchmark.GetSuite)
	public.Post("/agent-benchmarks/run", h.AgentBenchmark.RunBenchmark)
	public.Get("/agent-benchmarks/suites/:suiteId/leaderboard", h.AgentBenchmark.GetLeaderboard)

	// Code Quality Scoring
	public.Post("/code-quality/configs", h.CodeQuality.CreateConfig)
	public.Get("/code-quality/configs/:configId", h.CodeQuality.GetConfig)
	public.Post("/code-quality/analyze", h.CodeQuality.AnalyzeTrace)
	public.Get("/code-quality/reports", h.CodeQuality.ListReports)
	public.Get("/code-quality/reports/:reportId", h.CodeQuality.GetReport)
	public.Get("/code-quality/dashboard", h.CodeQuality.GetDashboard)
}
