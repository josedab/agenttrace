package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerCostRoutes registers cost optimization, budgets, and carbon routes
func registerCostRoutes(public fiber.Router, h *Handlers) {
	// Cost Optimization
	public.Get("/cost-optimizer/analyze", h.CostOptimizer.Analyze)
	public.Get("/cost-optimizer/recommendations", h.CostOptimizer.GetRecommendations)
	public.Post("/cost-optimizer/recommendations/:id/apply", h.CostOptimizer.ApplyRecommendation)
	public.Post("/cost-optimizer/recommendations/:id/dismiss", h.CostOptimizer.DismissRecommendation)
	public.Get("/cost-optimizer/forecast", h.CostOptimizer.GetForecast)
	public.Post("/cost-optimizer/report", h.CostOptimizer.GenerateReport)
	public.Post("/cost-optimizer/autopilot", h.CostOptimizer.ConfigureAutopilot)

	// Cost Budgets & Forecasting
	public.Get("/budgets", h.CostBudget.ListBudgets)
	public.Post("/budgets", h.CostBudget.CreateBudget)
	public.Get("/budgets/forecast", h.CostBudget.GetForecast)
	public.Post("/budgets/check", h.CostBudget.CheckBudget)
	public.Get("/budgets/:id", h.CostBudget.GetBudget)
	public.Put("/budgets/:id", h.CostBudget.UpdateBudget)
	public.Delete("/budgets/:id", h.CostBudget.DeleteBudget)

	// Predictive Cost Modeling
	public.Post("/predictions/cost", h.PredictiveCost.Predict)
	public.Get("/predictions", h.PredictiveCost.ListPredictions)
	public.Post("/predictions/:predictionId/approve", h.PredictiveCost.RequestApproval)
	public.Post("/approvals/:approvalId/decide", h.PredictiveCost.DecideApproval)

	// Cost Attribution
	public.Post("/cost-attribution", h.CostAttribution.Attribute)
	public.Get("/cost-attribution/report", h.CostAttribution.GetReport)
	public.Get("/cost-attribution", h.CostAttribution.List)

	// Intelligent Cost Guardrails
	public.Get("/cost-guardrails/dashboard", h.CostGuardrail.GetDashboard)
	public.Get("/cost-guardrails/policies", h.CostGuardrail.ListPolicies)
	public.Post("/cost-guardrails/policies", h.CostGuardrail.CreatePolicy)
	public.Post("/cost-guardrails/check", h.CostGuardrail.CheckBudget)
	public.Get("/cost-guardrails/forecast", h.CostGuardrail.GetForecast)
	public.Get("/cost-guardrails/violations", h.CostGuardrail.ListViolations)

	// Real-Time Cost Anomaly Alerting
	public.Post("/cost-alerts/rules", h.CostAlerting.CreateAlertRule)
	public.Get("/cost-alerts/rules", h.CostAlerting.ListAlertRules)
	public.Delete("/cost-alerts/rules/:ruleId", h.CostAlerting.DeleteAlertRule)
	public.Get("/cost-alerts", h.CostAlerting.ListCostAlerts)
	public.Post("/cost-alerts/:alertId/acknowledge", h.CostAlerting.AcknowledgeCostAlert)
	public.Get("/cost-alerts/circuit-breaker", h.CostAlerting.GetCircuitBreakerConfig)
	public.Put("/cost-alerts/circuit-breaker", h.CostAlerting.UpdateCircuitBreakerConfig)
	public.Post("/cost-alerts/check", h.CostAlerting.CheckCost)

	// Energy & Carbon
	public.Get("/carbon/footprint", h.Carbon.GetFootprint)
	public.Get("/carbon/config", h.Carbon.GetConfig)
	public.Put("/carbon/config", h.Carbon.UpdateConfig)
	public.Get("/carbon/suggestions", h.Carbon.GetSuggestions)
}
