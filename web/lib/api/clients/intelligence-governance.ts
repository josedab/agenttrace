// intelligenceGovernance API client.
import { createSearchParams, fetchWithAuth } from '../transport';
import type {
  BudgetPlan,
  BudgetPlanInput,
  CostOptimizerAnalysis,
  CostRecommendation,
  CreateEnrichmentRuleInput,
  EnrichmentRule,
  EnrichmentRuleTestResult,
  ForecastPoint,
  GuardrailRule,
  GuardrailViolationStats,
  SimulationChange,
  SimulationResult,
  UpdateEnrichmentRuleInput,
} from '../contracts';

export const intelligenceGovernanceApi = {
  // Anomaly Detection
  anomaly: {
    getDashboard: () => fetchWithAuth<unknown>('/api/public/anomaly/dashboard'),
    list: () => fetchWithAuth<unknown>('/api/public/rca/anomalies'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/rca/anomalies/${id}`),
    detect: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rca/anomalies', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    acknowledge: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/rca/anomalies/${id}/acknowledge`, { method: 'POST' }),
    getRootCause: (anomalyId: string) =>
      fetchWithAuth<unknown>(`/api/public/anomaly/anomalies/${anomalyId}/root-cause`),
    listChannels: () => fetchWithAuth<unknown>('/api/public/rca/alert-channels'),
    createChannel: (data: { name: string; type: string; config: unknown }) =>
      fetchWithAuth<unknown>('/api/public/rca/alert-channels', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    testChannel: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/rca/alert-channels/${id}/test`, { method: 'POST' }),
    listCorrelationRules: () => fetchWithAuth<unknown>('/api/public/rca/correlation-rules'),
    createCorrelationRule: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rca/correlation-rules', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listInvestigations: () => fetchWithAuth<unknown>('/api/public/rca/investigations'),
    createInvestigation: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rca/investigations', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    updateInvestigation: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/rca/investigations/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
  },
  rca: {
    analyze: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rca/analyze', { method: 'POST', body: JSON.stringify(data) }),
    listReports: () => fetchWithAuth<unknown>('/api/public/rca/reports'),
    getReport: (id: string) => fetchWithAuth<unknown>(`/api/public/rca/reports/${id}`),
  },
  // Diff Intelligence
  diffAnalysis: {
    list: (params?: { limit?: number; offset?: number }) =>
      fetchWithAuth<{ analyses: unknown[]; totalCount: number }>(
        `/api/public/diff-analysis?${createSearchParams(params)}`
      ),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/diff-analysis/${id}`),
    analyze: (data: { traceId: string; fileChanges: unknown[] }) =>
      fetchWithAuth<unknown>('/api/public/diff-analysis', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    trend: (days = 30) => fetchWithAuth<unknown>(`/api/public/diff-analysis/trend?days=${days}`),
    getForTrace: (traceId: string) =>
      fetchWithAuth<{ analyses: unknown[] }>(`/api/public/traces/${traceId}/diff-analysis`),
  },
  // AI Debugger
  aiDebugger: {
    debug: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/debug', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getHistory: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/traces/${traceId}/debug-history`),
    getContext: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/traces/${traceId}/debug-context`),
  },
  costForecast: {
    getForecast: (params: { period: string }) =>
      fetchWithAuth<ForecastPoint[]>(`/api/public/cost-forecast?${createSearchParams(params)}`),
    simulate: (data: { changes: SimulationChange[] }) =>
      fetchWithAuth<SimulationResult>('/api/public/cost-forecast/simulate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    createBudget: (data: BudgetPlanInput) =>
      fetchWithAuth<BudgetPlan>('/api/public/cost-forecast/budget-plan', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Cost Optimizer
  costOptimizer: {
    analyze: (params: { dateRange?: string } = {}) =>
      fetchWithAuth<CostOptimizerAnalysis>(
        `/api/public/cost-optimizer/analyze?${createSearchParams(params)}`
      ),
    getRecommendations: () =>
      fetchWithAuth<{ recommendations: unknown[] }>('/api/public/cost-optimizer/recommendations'),
    applyRecommendation: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/cost-optimizer/recommendations/${id}/apply`, {
        method: 'POST',
      }),
    apply: (id: string) =>
      fetchWithAuth<CostRecommendation>(`/api/public/cost-optimizer/recommendations/${id}/apply`, {
        method: 'POST',
      }),
    dismiss: (id: string) =>
      fetchWithAuth<CostRecommendation>(
        `/api/public/cost-optimizer/recommendations/${id}/dismiss`,
        { method: 'POST' }
      ),
    getForecast: () => fetchWithAuth<unknown>('/api/public/cost-optimizer/forecast'),
    generateReport: (period?: { startDate: string; endDate: string }) =>
      fetchWithAuth<unknown>('/api/public/cost-optimizer/report', {
        method: 'POST',
        body: JSON.stringify(period || {}),
      }),
    configureAutopilot: (config: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-optimizer/autopilot', {
        method: 'POST',
        body: JSON.stringify(config),
      }),
    getAutopilotReport: (dateRange?: string) =>
      fetchWithAuth<unknown>(
        `/api/public/cost-optimizer/autopilot/report?dateRange=${dateRange || '30d'}`
      ),
  },
  // Cost Guardrails
  costGuardrails: {
    getDashboard: () => fetchWithAuth<unknown>('/api/public/cost-guardrails/dashboard'),
    createPolicy: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-guardrails/policies', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listPolicies: () => fetchWithAuth<unknown>('/api/public/cost-guardrails/policies'),
    checkBudget: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-guardrails/check', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getForecast: () => fetchWithAuth<unknown>('/api/public/cost-guardrails/forecast'),
    listViolations: () => fetchWithAuth<unknown>('/api/public/cost-guardrails/violations'),
  },
  // Cost Alerting
  costAlerting: {
    createRule: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-alerts/rules', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listRules: () => fetchWithAuth<unknown>('/api/public/cost-alerts/rules'),
    deleteRule: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/cost-alerts/rules/${id}`, { method: 'DELETE' }),
    listAlerts: () => fetchWithAuth<unknown>('/api/public/cost-alerts'),
    acknowledgeAlert: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/cost-alerts/${id}/acknowledge`, { method: 'POST' }),
    getCircuitBreaker: () => fetchWithAuth<unknown>('/api/public/cost-alerts/circuit-breaker'),
    updateCircuitBreaker: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-alerts/circuit-breaker', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    checkCost: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-alerts/check', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Cost Attribution
  costAttribution: {
    attribute: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost-attribution', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getReport: () => fetchWithAuth<unknown>('/api/public/cost-attribution/report'),
    list: () => fetchWithAuth<unknown>('/api/public/cost-attribution'),
    getBreakdown: (data: { period: string; groupBy: string }) =>
      fetchWithAuth<unknown>(
        `/api/public/cost-attribution/report?period=${encodeURIComponent(data.period)}&groupBy=${encodeURIComponent(data.groupBy)}`
      ),
  },
  // Cost Autopilot (v12)
  costAutopilot: {
    getHotspots: (days: number) => fetchWithAuth<unknown>(`/api/public/cost/hotspots?days=${days}`),
    getRules: () => fetchWithAuth<unknown>('/api/public/cost/autopilot/rules'),
    createRule: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cost/autopilot/rules', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getPredictions: (days: number, budget: number) =>
      fetchWithAuth<unknown>(`/api/public/cost/predictions?days=${days}&budget=${budget}`),
    getDashboard: () => fetchWithAuth<unknown>('/api/public/cost/autopilot/dashboard'),
  },
  predictions: {
    predict: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/predictions/cost', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    list: () => fetchWithAuth<unknown>('/api/public/predictions'),
    requestApproval: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/predictions/${id}/approve`, { method: 'POST' }),
    decideApproval: (id: string, decision: unknown) =>
      fetchWithAuth<unknown>(`/api/public/approvals/${id}/decide`, {
        method: 'POST',
        body: JSON.stringify(decision),
      }),
  },
  // Guardrails
  guardrails: {
    list: () => fetchWithAuth<GuardrailRule[]>('/api/public/guardrails'),
    getViolationStats: () =>
      fetchWithAuth<GuardrailViolationStats>('/api/public/guardrails/violations/stats'),
    listRules: () => fetchWithAuth<unknown>('/api/public/guardrails'),
    createRule: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/guardrails', { method: 'POST', body: JSON.stringify(data) }),
    getTemplates: () => fetchWithAuth<{ templates: unknown[] }>('/api/public/guardrails/templates'),
    createPlaybook: (data: { name: string; template?: string; enforceMode?: string }) =>
      fetchWithAuth<unknown>('/api/public/guardrails/playbooks', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listViolations: () => fetchWithAuth<unknown>('/api/public/guardrails/violations'),
    createPolicy: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/guardrails/policies', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listPolicies: () => fetchWithAuth<{ policies: unknown[] }>('/api/public/guardrails/policies'),
    evaluatePipeline: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/guardrails/evaluate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getDashboardStats: () => fetchWithAuth<unknown>('/api/public/guardrails/dashboard'),
    getAuditTrail: (policyId: string) =>
      fetchWithAuth<{ auditTrail: unknown[] }>(`/api/public/guardrails/policies/${policyId}/audit`),
  },
  enrichment: {
    listRules: () => fetchWithAuth<EnrichmentRule[]>('/api/public/enrichment/rules'),
    createRule: (data: CreateEnrichmentRuleInput) =>
      fetchWithAuth<EnrichmentRule>('/api/public/enrichment/rules', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    updateRule: (id: string, data: UpdateEnrichmentRuleInput) =>
      fetchWithAuth<EnrichmentRule>(`/api/public/enrichment/rules/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    testRule: (id: string) =>
      fetchWithAuth<EnrichmentRuleTestResult>('/api/public/enrichment/test', {
        method: 'POST',
        body: JSON.stringify({ ruleId: id }),
      }),
  },
  // Compliance Reports
  complianceReports: {
    list: () => fetchWithAuth<unknown>('/api/public/compliance-reports'),
    generate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/compliance-reports', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/compliance-reports/${id}`),
    getTemplates: () => fetchWithAuth<unknown>('/api/public/compliance-reports/templates'),
  },
  // Compliance Monitor
  complianceMonitor: {
    listPolicies: () => fetchWithAuth<unknown>('/api/public/compliance-monitor/policies'),
    createPolicy: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/compliance-monitor/policies', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    evaluate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/compliance-monitor/evaluate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getScore: (fw: string) => fetchWithAuth<unknown>(`/api/public/compliance-monitor/score/${fw}`),
    configure: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/compliance-monitor/configure', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Security Scanner
  securityScanner: {
    scan: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/security/scan', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    createPolicy: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/security/policies', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listPolicies: () => fetchWithAuth<unknown>('/api/public/security/policies'),
    getDashboard: () => fetchWithAuth<unknown>('/api/public/security/dashboard'),
    acknowledgeFinding: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/security/findings/${id}/acknowledge`, {
        method: 'POST',
      }),
  },
  privacy: {
    scan: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/privacy/scan', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getConfig: () => fetchWithAuth<unknown>('/api/public/privacy/config'),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/privacy/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    requestDeletion: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/privacy/deletion-requests', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listDeletionRequests: () => fetchWithAuth<unknown>('/api/public/privacy/deletion-requests'),
  },
  // Chaos Testing
  chaos: {
    list: () => fetchWithAuth<unknown>('/api/public/chaos/experiments'),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/chaos/experiments', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/chaos/experiments/${id}`),
    run: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/chaos/experiments/${id}/run`, { method: 'POST' }),
    getScorecard: (agentName: string) =>
      fetchWithAuth<unknown>(`/api/public/chaos/scorecard/${agentName}`),
  },
  // Autonomy Gradient
  autonomy: {
    getDashboard: () => fetchWithAuth<unknown>('/api/public/autonomy/dashboard'),
    getConfig: (agent: string) => fetchWithAuth<unknown>(`/api/public/autonomy/${agent}`),
    set: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/autonomy', { method: 'POST', body: JSON.stringify(data) }),
    getTrust: (agent: string) => fetchWithAuth<unknown>(`/api/public/autonomy/${agent}/trust`),
  },
  // Intent Verification
  intents: {
    declare: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/intents', { method: 'POST', body: JSON.stringify(data) }),
    verify: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/intents/${id}/verify`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/intents/${id}`),
    getStats: () => fetchWithAuth<unknown>('/api/public/intents/stats'),
  },
};
