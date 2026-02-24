package main

import (
	"github.com/gofiber/fiber/v2"
)

// registerComplianceRoutes registers compliance, privacy, and security routes
func registerComplianceRoutes(public fiber.Router, h *Handlers) {
	// EU AI Act Compliance
	public.Get("/compliance/assess", h.Compliance.AssessProject)
	public.Get("/compliance/status", h.Compliance.GetStatus)
	public.Get("/compliance/audit-trail", h.Compliance.GetAuditTrail)
	public.Post("/compliance/assessments", h.Compliance.CreateAssessment)
	public.Get("/compliance/assessments/:id", h.Compliance.GetAssessment)
	public.Post("/compliance/reports", h.Compliance.GenerateReport)

	// Compliance Export
	public.Post("/compliance/exports", h.ComplianceExport.StartExport)
	public.Get("/compliance/exports", h.ComplianceExport.ListExports)
	public.Get("/compliance/exports/:id", h.ComplianceExport.GetExport)
	public.Get("/compliance/templates", h.ComplianceExport.GetTemplates)

	// Compliance Reports
	public.Get("/compliance-reports", h.ComplianceReport.List)
	public.Post("/compliance-reports", h.ComplianceReport.Generate)
	public.Get("/compliance-reports/templates", h.ComplianceReport.GetTemplates)
	public.Get("/compliance-reports/:reportId", h.ComplianceReport.Get)

	// Compliance Monitoring
	public.Get("/compliance-monitor/policies", h.ComplianceMonitor.ListPolicies)
	public.Post("/compliance-monitor/policies", h.ComplianceMonitor.CreatePolicy)
	public.Post("/compliance-monitor/evaluate", h.ComplianceMonitor.Evaluate)
	public.Get("/compliance-monitor/score/:framework", h.ComplianceMonitor.GetScore)
	public.Post("/compliance-monitor/configure", h.ComplianceMonitor.Configure)

	// Privacy
	public.Post("/privacy/scan", h.Privacy.ScanPII)
	public.Get("/privacy/config", h.Privacy.GetConfig)
	public.Put("/privacy/config", h.Privacy.UpdateConfig)
	public.Post("/privacy/deletion-requests", h.Privacy.RequestDeletion)
	public.Get("/privacy/deletion-requests", h.Privacy.ListDeletionRequests)

	// RBAC & SSO
	public.Get("/rbac/permissions", h.RBAC.GetPermissions)
	public.Post("/rbac/roles", h.RBAC.AssignRole)
	public.Post("/rbac/check", h.RBAC.CheckPermission)
	public.Get("/rbac/sso", h.RBAC.GetSSOConfig)
	public.Post("/rbac/sso", h.RBAC.ConfigureSSO)
	public.Post("/rbac/api-key-scope", h.RBAC.ScopeAPIKey)

	// Agent Security Scanner
	public.Post("/security/scan", h.SecurityScanner.ScanTrace)
	public.Post("/security/policies", h.SecurityScanner.CreateSecurityPolicy)
	public.Get("/security/policies", h.SecurityScanner.ListSecurityPolicies)
	public.Get("/security/dashboard", h.SecurityScanner.GetSecurityDashboard)
	public.Post("/security/findings/:findingId/acknowledge", h.SecurityScanner.AcknowledgeSecurityFinding)

	// Guardrails
	public.Get("/guardrails", h.Guardrails.ListRules)
	public.Post("/guardrails", h.Guardrails.CreateRule)
	public.Put("/guardrails/:ruleId", h.Guardrails.UpdateRule)
	public.Delete("/guardrails/:ruleId", h.Guardrails.DeleteRule)
	public.Get("/guardrails/violations", h.Guardrails.ListViolations)
	public.Get("/guardrails/violations/stats", h.Guardrails.GetViolationStats)
	public.Get("/guardrails/templates", h.Guardrails.GetPlaybookTemplates)
	public.Post("/guardrails/playbooks", h.Guardrails.CreatePlaybook)
}
