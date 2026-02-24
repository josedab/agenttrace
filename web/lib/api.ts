import { GraphQLClient } from "graphql-request";

export const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

/**
 * Fetch wrapper with error handling
 */
async function fetchWithAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_URL}${endpoint}`;

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: "Unknown error" }));
    throw new ApiError(response.status, error.message || response.statusText);
  }

  return response.json();
}

/**
 * Custom API error class
 */
export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Create authenticated API client
 */
export function createApiClient(token?: string) {
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return {
    get: <T>(endpoint: string) =>
      fetchWithAuth<T>(endpoint, { method: "GET", headers }),

    post: <T>(endpoint: string, data?: unknown) =>
      fetchWithAuth<T>(endpoint, {
        method: "POST",
        headers,
        body: data ? JSON.stringify(data) : undefined,
      }),

    put: <T>(endpoint: string, data?: unknown) =>
      fetchWithAuth<T>(endpoint, {
        method: "PUT",
        headers,
        body: data ? JSON.stringify(data) : undefined,
      }),

    patch: <T>(endpoint: string, data?: unknown) =>
      fetchWithAuth<T>(endpoint, {
        method: "PATCH",
        headers,
        body: data ? JSON.stringify(data) : undefined,
      }),

    delete: <T>(endpoint: string) =>
      fetchWithAuth<T>(endpoint, { method: "DELETE", headers }),
  };
}

/**
 * Create GraphQL client
 */
export function createGraphQLClient(token?: string) {
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return new GraphQLClient(`${API_URL}/graphql`, { headers });
}

/**
 * API endpoints
 */
export const api = {
  // Auth
  auth: {
    login: (data: { email: string; password: string }) =>
      fetchWithAuth<{ token: string; user: User }>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    register: (data: { email: string; password: string; name: string }) =>
      fetchWithAuth<{ token: string; user: User }>("/api/auth/register", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    refresh: (refreshToken: string) =>
      fetchWithAuth<{ token: string }>("/api/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refreshToken }),
      }),
  },

  // Traces
  traces: {
    list: (projectId: string, params?: TraceListParams) =>
      fetchWithAuth<{ traces: Trace[]; nextCursor?: string }>(
        `/api/public/traces?${new URLSearchParams(params as Record<string, string>)}`,
        {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }
      ),

    get: (projectId: string, id: string) =>
      fetchWithAuth<Trace>(`/api/public/traces/${id}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    getObservations: (projectId: string, traceId: string) =>
      fetchWithAuth<Observation[]>(`/api/public/traces/${traceId}/observations`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),
  },

  // Sessions
  sessions: {
    list: (projectId: string, params?: SessionListParams) =>
      fetchWithAuth<{ sessions: Session[]; nextCursor?: string }>(
        `/api/public/sessions?${new URLSearchParams(params as Record<string, string>)}`,
        {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }
      ),

    get: (projectId: string, id: string) =>
      fetchWithAuth<Session>(`/api/public/sessions/${id}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),
  },

  // Scores
  scores: {
    list: (projectId: string, params?: ScoreListParams) =>
      fetchWithAuth<{ scores: Score[]; nextCursor?: string }>(
        `/api/public/scores?${new URLSearchParams(params as Record<string, string>)}`,
        {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }
      ),

    create: (projectId: string, data: CreateScoreInput) =>
      fetchWithAuth<Score>("/api/public/scores", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    update: (projectId: string, id: string, data: UpdateScoreInput) =>
      fetchWithAuth<Score>(`/api/public/scores/${id}`, {
        method: "PUT",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    delete: (projectId: string, id: string) =>
      fetchWithAuth<void>(`/api/public/scores/${id}`, {
        method: "DELETE",
        headers: { "X-Project-ID": projectId },
      }),
  },

  // Prompts
  prompts: {
    list: (projectId: string) =>
      fetchWithAuth<Prompt[]>("/api/public/prompts", {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    get: (projectId: string, name: string, version?: number, label?: string) => {
      const params = new URLSearchParams();
      if (version) params.set("version", version.toString());
      if (label) params.set("label", label);
      return fetchWithAuth<Prompt>(`/api/public/prompts/${name}?${params}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      });
    },

    create: (projectId: string, data: CreatePromptInput) =>
      fetchWithAuth<Prompt>("/api/public/prompts", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    update: (projectId: string, name: string, data: UpdatePromptInput) =>
      fetchWithAuth<Prompt>(`/api/public/prompts/${name}`, {
        method: "PUT",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    delete: (projectId: string, name: string) =>
      fetchWithAuth<void>(`/api/public/prompts/${name}`, {
        method: "DELETE",
        headers: { "X-Project-ID": projectId },
      }),

    compile: (projectId: string, name: string, variables: Record<string, string>) =>
      fetchWithAuth<{ prompt: string }>(`/api/public/prompts/${name}/compile`, {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify({ variables }),
      }),
  },

  // Datasets
  datasets: {
    list: (projectId: string) =>
      fetchWithAuth<Dataset[]>("/api/public/datasets", {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    get: (projectId: string, id: string) =>
      fetchWithAuth<Dataset>(`/api/public/datasets/${id}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateDatasetInput) =>
      fetchWithAuth<Dataset>("/api/public/datasets", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    update: (projectId: string, id: string, data: UpdateDatasetInput) =>
      fetchWithAuth<Dataset>(`/api/public/datasets/${id}`, {
        method: "PUT",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    delete: (projectId: string, id: string) =>
      fetchWithAuth<void>(`/api/public/datasets/${id}`, {
        method: "DELETE",
        headers: { "X-Project-ID": projectId },
      }),

    items: {
      list: (projectId: string, datasetId: string) =>
        fetchWithAuth<DatasetItem[]>(`/api/public/datasets/${datasetId}/items`, {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }),

      create: (projectId: string, datasetId: string, data: CreateDatasetItemInput) =>
        fetchWithAuth<DatasetItem>(`/api/public/datasets/${datasetId}/items`, {
          method: "POST",
          headers: { "X-Project-ID": projectId },
          body: JSON.stringify(data),
        }),
    },

    runs: {
      list: (projectId: string, datasetId: string) =>
        fetchWithAuth<DatasetRun[]>(`/api/public/datasets/${datasetId}/runs`, {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }),

      create: (projectId: string, datasetId: string, data: CreateDatasetRunInput) =>
        fetchWithAuth<DatasetRun>(`/api/public/datasets/${datasetId}/runs`, {
          method: "POST",
          headers: { "X-Project-ID": projectId },
          body: JSON.stringify(data),
        }),
    },
  },

  // Evaluators
  evaluators: {
    list: (projectId: string) =>
      fetchWithAuth<Evaluator[]>("/api/public/evaluators", {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    get: (projectId: string, id: string) =>
      fetchWithAuth<Evaluator>(`/api/public/evaluators/${id}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateEvaluatorInput) =>
      fetchWithAuth<Evaluator>("/api/public/evaluators", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    update: (projectId: string, id: string, data: UpdateEvaluatorInput) =>
      fetchWithAuth<Evaluator>(`/api/public/evaluators/${id}`, {
        method: "PUT",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    delete: (projectId: string, id: string) =>
      fetchWithAuth<void>(`/api/public/evaluators/${id}`, {
        method: "DELETE",
        headers: { "X-Project-ID": projectId },
      }),

    templates: (projectId: string) =>
      fetchWithAuth<EvaluatorTemplate[]>("/api/public/evaluator-templates", {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),
  },

  // Metrics
  metrics: {
    get: (projectId: string, params: MetricsParams) =>
      fetchWithAuth<Metrics>(`/api/public/metrics/project?${new URLSearchParams(params as Record<string, string>)}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    dailyCosts: (projectId: string, params: DailyCostsParams) =>
      fetchWithAuth<DailyCost[]>(`/api/v1/projects/${projectId}/daily-costs?${new URLSearchParams(params as Record<string, string>)}`, {
        method: "GET",
      }),
  },

  // Organizations
  organizations: {
    list: () => fetchWithAuth<Organization[]>("/api/v1/organizations"),

    get: (id: string) => fetchWithAuth<Organization>(`/api/v1/organizations/${id}`),

    create: (data: { name: string }) =>
      fetchWithAuth<Organization>("/api/v1/organizations", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  },

  // Projects
  projects: {
    list: (orgId?: string) => {
      const params = orgId ? `?organizationId=${orgId}` : "";
      return fetchWithAuth<Project[]>(`/api/v1/projects${params}`);
    },

    get: (id: string) => fetchWithAuth<Project>(`/api/v1/projects/${id}`),

    create: (data: CreateProjectInput) =>
      fetchWithAuth<Project>("/api/v1/projects", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    update: (id: string, data: UpdateProjectInput) =>
      fetchWithAuth<Project>(`/api/v1/projects/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
  },

  // API Keys
  apiKeys: {
    list: (projectId: string) =>
      fetchWithAuth<APIKey[]>(`/api/v1/projects/${projectId}/api-keys`),

    create: (projectId: string, data: CreateAPIKeyInput) =>
      fetchWithAuth<APIKeyWithSecret>(`/api/v1/projects/${projectId}/api-keys`, {
        method: "POST",
        body: JSON.stringify(data),
      }),

    delete: (id: string) =>
      fetchWithAuth<void>(`/api/v1/api-keys/${id}`, { method: "DELETE" }),
  },

  // SSO
  sso: {
    get: (organizationId: string) =>
      fetchWithAuth<SSOConfiguration>(`/api/v1/organizations/${organizationId}/sso`),

    list: (organizationId: string) =>
      fetchWithAuth<SSOConfiguration[]>(`/api/v1/organizations/${organizationId}/sso/configurations`),

    create: (organizationId: string, data: CreateSSOInput) =>
      fetchWithAuth<SSOConfiguration>(`/api/v1/organizations/${organizationId}/sso`, {
        method: "POST",
        body: JSON.stringify(data),
      }),

    update: (organizationId: string, data: UpdateSSOInput) =>
      fetchWithAuth<SSOConfiguration>(`/api/v1/organizations/${organizationId}/sso`, {
        method: "PATCH",
        body: JSON.stringify(data),
      }),

    delete: (organizationId: string) =>
      fetchWithAuth<void>(`/api/v1/organizations/${organizationId}/sso`, {
        method: "DELETE",
      }),

    test: (organizationId: string) =>
      fetchWithAuth<{ success: boolean; message: string }>(
        `/api/v1/organizations/${organizationId}/sso/test`,
        { method: "POST" }
      ),
  },

  // Audit Logs
  auditLogs: {
    list: (organizationId: string, params?: AuditLogListParams) => {
      const searchParams = new URLSearchParams();
      if (params?.userId) searchParams.set("userId", params.userId);
      if (params?.action) searchParams.set("action", params.action);
      if (params?.resourceType) searchParams.set("resourceType", params.resourceType);
      if (params?.startDate) searchParams.set("startDate", params.startDate);
      if (params?.endDate) searchParams.set("endDate", params.endDate);
      if (params?.cursor) searchParams.set("cursor", params.cursor);
      if (params?.limit) searchParams.set("limit", params.limit.toString());
      return fetchWithAuth<{ logs: AuditLog[]; nextCursor?: string }>(
        `/api/v1/organizations/${organizationId}/audit-logs?${searchParams}`
      );
    },

    get: (organizationId: string, logId: string) =>
      fetchWithAuth<AuditLog>(`/api/v1/organizations/${organizationId}/audit-logs/${logId}`),

    summary: (organizationId: string, params?: { startDate?: string; endDate?: string }) => {
      const searchParams = new URLSearchParams();
      if (params?.startDate) searchParams.set("startDate", params.startDate);
      if (params?.endDate) searchParams.set("endDate", params.endDate);
      return fetchWithAuth<AuditSummary>(
        `/api/v1/organizations/${organizationId}/audit-logs/summary?${searchParams}`
      );
    },

    exportJobs: (organizationId: string) =>
      fetchWithAuth<AuditExportJob[]>(`/api/v1/organizations/${organizationId}/audit-logs/exports`),

    createExport: (organizationId: string, params: { startDate: string; endDate: string; format?: "json" | "csv" }) =>
      fetchWithAuth<AuditExportJob>(`/api/v1/organizations/${organizationId}/audit-logs/exports`, {
        method: "POST",
        body: JSON.stringify(params),
      }),

    downloadExport: (organizationId: string, jobId: string) =>
      fetchWithAuth<unknown>(`/api/v1/organizations/${organizationId}/audit-logs/exports/${jobId}/download`),
  },

  // Checkpoints
  checkpoints: {
    list: (projectId: string, params?: CheckpointListParams) => {
      const searchParams = new URLSearchParams();
      if (params?.traceId) searchParams.set("traceId", params.traceId);
      if (params?.type) searchParams.set("type", params.type);
      if (params?.cursor) searchParams.set("cursor", params.cursor);
      if (params?.limit) searchParams.set("limit", params.limit.toString());
      return fetchWithAuth<{ checkpoints: Checkpoint[]; nextCursor?: string }>(
        `/api/v1/checkpoints?${searchParams}`,
        {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }
      );
    },

    get: (projectId: string, checkpointId: string) =>
      fetchWithAuth<Checkpoint>(`/api/v1/checkpoints/${checkpointId}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    listByTrace: (projectId: string, traceId: string) =>
      fetchWithAuth<Checkpoint[]>(`/api/v1/checkpoints?traceId=${traceId}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateCheckpointInput) =>
      fetchWithAuth<Checkpoint>("/api/v1/checkpoints", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    restore: (projectId: string, checkpointId: string) =>
      fetchWithAuth<{ success: boolean }>(`/api/v1/checkpoints/${checkpointId}/restore`, {
        method: "POST",
        headers: { "X-Project-ID": projectId },
      }),
  },

  // Git Links
  gitLinks: {
    list: (projectId: string, params?: { traceId?: string; cursor?: string; limit?: number }) => {
      const searchParams = new URLSearchParams();
      if (params?.traceId) searchParams.set("traceId", params.traceId);
      if (params?.cursor) searchParams.set("cursor", params.cursor);
      if (params?.limit) searchParams.set("limit", params.limit.toString());
      return fetchWithAuth<{ gitLinks: GitLink[]; nextCursor?: string }>(
        `/api/v1/git-links?${searchParams}`,
        {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }
      );
    },

    timeline: (projectId: string, traceId: string) =>
      fetchWithAuth<GitTimelineEntry[]>(`/api/v1/git-links/timeline?traceId=${traceId}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateGitLinkInput) =>
      fetchWithAuth<GitLink>("/api/v1/git-links", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),
  },

  // File Operations
  fileOperations: {
    list: (projectId: string, traceId: string) =>
      fetchWithAuth<FileOperation[]>(`/api/v1/file-operations?traceId=${traceId}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateFileOperationInput) =>
      fetchWithAuth<FileOperation>("/api/v1/file-operations", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),
  },

  // Terminal Commands
  terminalCommands: {
    list: (projectId: string, traceId: string) =>
      fetchWithAuth<TerminalCommand[]>(`/api/v1/terminal-commands?traceId=${traceId}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateTerminalCommandInput) =>
      fetchWithAuth<TerminalCommand>("/api/v1/terminal-commands", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),
  },

  // CI Runs
  ciRuns: {
    list: (projectId: string, params?: { cursor?: string; limit?: number }) => {
      const searchParams = new URLSearchParams();
      if (params?.cursor) searchParams.set("cursor", params.cursor);
      if (params?.limit) searchParams.set("limit", params.limit.toString());
      return fetchWithAuth<{ ciRuns: CIRun[]; nextCursor?: string }>(
        `/api/v1/ci-runs?${searchParams}`,
        {
          method: "GET",
          headers: { "X-Project-ID": projectId },
        }
      );
    },

    get: (projectId: string, runId: string) =>
      fetchWithAuth<CIRun>(`/api/v1/ci-runs/${runId}`, {
        method: "GET",
        headers: { "X-Project-ID": projectId },
      }),

    create: (projectId: string, data: CreateCIRunInput) =>
      fetchWithAuth<CIRun>("/api/v1/ci-runs", {
        method: "POST",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),

    update: (projectId: string, runId: string, data: UpdateCIRunInput) =>
      fetchWithAuth<CIRun>(`/api/v1/ci-runs/${runId}`, {
        method: "PATCH",
        headers: { "X-Project-ID": projectId },
        body: JSON.stringify(data),
      }),
  },

  // Streaming
  streaming: {
    getActiveStreams: () =>
      fetchWithAuth<{ streams: any[]; count: number }>("/api/public/streams"),
    getLiveMetrics: (traceId: string) =>
      fetchWithAuth<any>(`/api/public/traces/${traceId}/live-metrics`),
    getActivities: (traceId: string, limit = 100) =>
      fetchWithAuth<{ activities: any[]; count: number }>(`/api/public/traces/${traceId}/activities?limit=${limit}`),
    requestIntervention: (traceId: string, data: { action: string; message?: string }) =>
      fetchWithAuth<any>(`/api/public/traces/${traceId}/intervene`, { method: "POST", body: JSON.stringify(data) }),
  },

  // Diff Intelligence
  diffAnalysis: {
    list: (params?: { limit?: number; offset?: number }) =>
      fetchWithAuth<{ analyses: any[]; totalCount: number }>(`/api/public/diff-analysis?${new URLSearchParams(params as any)}`),
    get: (id: string) =>
      fetchWithAuth<any>(`/api/public/diff-analysis/${id}`),
    analyze: (data: { traceId: string; fileChanges: any[] }) =>
      fetchWithAuth<any>("/api/public/diff-analysis", { method: "POST", body: JSON.stringify(data) }),
    trend: (days = 30) =>
      fetchWithAuth<any>(`/api/public/diff-analysis/trend?days=${days}`),
    getForTrace: (traceId: string) =>
      fetchWithAuth<{ analyses: any[] }>(`/api/public/traces/${traceId}/diff-analysis`),
  },

  // Anomaly Detection
  anomaly: {
    getDashboard: () =>
      fetchWithAuth<any>("/api/public/anomaly/dashboard"),
    getRootCause: (anomalyId: string) =>
      fetchWithAuth<any>(`/api/public/anomaly/anomalies/${anomalyId}/root-cause`),
    createChannel: (data: { name: string; type: string; config: any }) =>
      fetchWithAuth<any>("/api/public/anomaly/channels", { method: "POST", body: JSON.stringify(data) }),
  },

  // Cost Optimizer
  costOptimizer: {
    analyze: () =>
      fetchWithAuth<any>("/api/public/cost-optimizer/analyze"),
    getRecommendations: () =>
      fetchWithAuth<{ recommendations: any[] }>("/api/public/cost-optimizer/recommendations"),
    applyRecommendation: (id: string) =>
      fetchWithAuth<any>(`/api/public/cost-optimizer/recommendations/${id}/apply`, { method: "POST" }),
    getForecast: () =>
      fetchWithAuth<any>("/api/public/cost-optimizer/forecast"),
    generateReport: (period?: { startDate: string; endDate: string }) =>
      fetchWithAuth<any>("/api/public/cost-optimizer/report", { method: "POST", body: JSON.stringify(period || {}) }),
    configureAutopilot: (config: any) =>
      fetchWithAuth<any>("/api/public/cost-optimizer/autopilot", { method: "POST", body: JSON.stringify(config) }),
  },

  // Guardrails
  guardrails: {
    listRules: () =>
      fetchWithAuth<any>("/api/public/guardrails"),
    createRule: (data: any) =>
      fetchWithAuth<any>("/api/public/guardrails", { method: "POST", body: JSON.stringify(data) }),
    getTemplates: () =>
      fetchWithAuth<{ templates: any[] }>("/api/public/guardrails/templates"),
    createPlaybook: (data: { name: string; template?: string; enforceMode?: string }) =>
      fetchWithAuth<any>("/api/public/guardrails/playbooks", { method: "POST", body: JSON.stringify(data) }),
    listViolations: () =>
      fetchWithAuth<any>("/api/public/guardrails/violations"),
  },

  // Benchmarks
  benchmarks: {
    list: () =>
      fetchWithAuth<any>("/api/public/benchmarks"),
    get: (id: string) =>
      fetchWithAuth<any>(`/api/public/benchmarks/${id}`),
    create: (data: any) =>
      fetchWithAuth<any>("/api/public/benchmarks", { method: "POST", body: JSON.stringify(data) }),
    submit: (benchmarkId: string, data: any) =>
      fetchWithAuth<any>(`/api/public/benchmarks/${benchmarkId}/submit`, { method: "POST", body: JSON.stringify(data) }),
    getLeaderboard: (benchmarkId: string) =>
      fetchWithAuth<any>(`/api/public/benchmarks/${benchmarkId}/leaderboard`),
    compare: (benchmarkId: string, data: { submissionIdA: string; submissionIdB: string }) =>
      fetchWithAuth<any>(`/api/public/benchmarks/${benchmarkId}/compare`, { method: "POST", body: JSON.stringify(data) }),
    getStats: (benchmarkId: string) =>
      fetchWithAuth<any>(`/api/public/benchmarks/${benchmarkId}/stats`),
  },

  // Federation
  federation: {
    listPeers: () =>
      fetchWithAuth<{ peers: any[]; count: number }>("/api/public/federation/peers"),
    addPeer: (data: { name: string; url: string; apiKey?: string }) =>
      fetchWithAuth<any>("/api/public/federation/peers", { method: "POST", body: JSON.stringify(data) }),
    removePeer: (peerId: string) =>
      fetchWithAuth<any>(`/api/public/federation/peers/${peerId}`, { method: "DELETE" }),
    query: (data: { query: string; peerIds?: string[] }) =>
      fetchWithAuth<any>("/api/public/federation/query", { method: "POST", body: JSON.stringify(data) }),
    listDestinations: () =>
      fetchWithAuth<{ destinations: any[]; count: number }>("/api/public/federation/destinations"),
    createDestination: (data: any) =>
      fetchWithAuth<any>("/api/public/federation/destinations", { method: "POST", body: JSON.stringify(data) }),
  },

  // Skill Profiles
  skillProfiles: {
    list: () =>
      fetchWithAuth<any>("/api/public/skill-profiles"),
    get: (agentName: string) =>
      fetchWithAuth<any>(`/api/public/skill-profiles/${agentName}`),
    compare: (agents: string[]) =>
      fetchWithAuth<any>(`/api/public/skill-profiles/compare?agents=${agents.join(",")}`),
  },

  // Prompt Lab
  promptLab: {
    listExperiments: () =>
      fetchWithAuth<any>("/api/public/prompt-lab/experiments"),
    createExperiment: (data: any) =>
      fetchWithAuth<any>("/api/public/prompt-lab/experiments", { method: "POST", body: JSON.stringify(data) }),
    getExperiment: (id: string) =>
      fetchWithAuth<any>(`/api/public/prompt-lab/experiments/${id}`),
    startExperiment: (id: string) =>
      fetchWithAuth<any>(`/api/public/prompt-lab/experiments/${id}/start`, { method: "POST" }),
    completeExperiment: (id: string) =>
      fetchWithAuth<any>(`/api/public/prompt-lab/experiments/${id}/complete`, { method: "POST" }),
    getSuggestions: (promptName?: string) =>
      fetchWithAuth<any>(`/api/public/prompt-lab/suggestions${promptName ? `?promptName=${promptName}` : ""}`),
  },

  // Sandbox
  sandbox: {
    listPending: () =>
      fetchWithAuth<any>("/api/public/sandbox/reviews/pending"),
    submitReview: (data: any) =>
      fetchWithAuth<any>("/api/public/sandbox/reviews", { method: "POST", body: JSON.stringify(data) }),
    getReview: (id: string) =>
      fetchWithAuth<any>(`/api/public/sandbox/reviews/${id}`),
    decide: (id: string, decision: any) =>
      fetchWithAuth<any>(`/api/public/sandbox/reviews/${id}/decide`, { method: "POST", body: JSON.stringify(decision) }),
    listPolicies: () =>
      fetchWithAuth<any>("/api/public/sandbox/policies"),
    createPolicy: (data: any) =>
      fetchWithAuth<any>("/api/public/sandbox/policies", { method: "POST", body: JSON.stringify(data) }),
    getStats: () =>
      fetchWithAuth<any>("/api/public/sandbox/stats"),
  },

  // Team Intelligence
  team: {
    getDashboard: () =>
      fetchWithAuth<any>("/api/public/team/dashboard"),
    calculateROI: (hourlyRate?: number) =>
      fetchWithAuth<any>(`/api/public/team/roi${hourlyRate ? `?hourlyRate=${hourlyRate}` : ""}`),
  },

  // Semantic Search
  search: {
    query: (data: any) =>
      fetchWithAuth<any>("/api/public/search", { method: "POST", body: JSON.stringify(data) }),
    suggestions: (prefix: string) =>
      fetchWithAuth<any>(`/api/public/search/suggestions?prefix=${prefix}`),
  },

  // Training Pipeline
  training: {
    listDatasets: () =>
      fetchWithAuth<any>("/api/public/training/datasets"),
    createDataset: (data: any) =>
      fetchWithAuth<any>("/api/public/training/datasets", { method: "POST", body: JSON.stringify(data) }),
    exportDataset: (id: string) =>
      fetchWithAuth<any>(`/api/public/training/datasets/${id}/export`, { method: "POST" }),
    detectFailures: () =>
      fetchWithAuth<any>("/api/public/training/failure-patterns"),
  },

  // RBAC
  rbac: {
    getPermissions: (role?: string) =>
      fetchWithAuth<any>(`/api/public/rbac/permissions${role ? `?role=${role}` : ""}`),
    assignRole: (data: any) =>
      fetchWithAuth<any>("/api/public/rbac/roles", { method: "POST", body: JSON.stringify(data) }),
    checkPermission: (data: any) =>
      fetchWithAuth<any>("/api/public/rbac/check", { method: "POST", body: JSON.stringify(data) }),
    getSSOConfig: () =>
      fetchWithAuth<any>("/api/public/rbac/sso"),
    configureSSO: (data: any) =>
      fetchWithAuth<any>("/api/public/rbac/sso", { method: "POST", body: JSON.stringify(data) }),
  },

  // Webhook Rules
  webhookRules: {
    list: () =>
      fetchWithAuth<any>("/api/public/webhook-rules"),
    create: (data: any) =>
      fetchWithAuth<any>("/api/public/webhook-rules", { method: "POST", body: JSON.stringify(data) }),
    delete: (id: string) =>
      fetchWithAuth<any>(`/api/public/webhook-rules/${id}`, { method: "DELETE" }),
    getTemplates: () =>
      fetchWithAuth<any>("/api/public/webhook-rules/templates"),
    test: (id: string) =>
      fetchWithAuth<any>(`/api/public/webhook-rules/${id}/test`, { method: "POST" }),
  },

  // Marketplace
  marketplace: {
    search: (params?: any) =>
      fetchWithAuth<any>(`/api/public/marketplace${params ? `?${new URLSearchParams(params)}` : ""}`),
    featured: () =>
      fetchWithAuth<any>("/api/public/marketplace/featured"),
    get: (id: string) =>
      fetchWithAuth<any>(`/api/public/marketplace/${id}`),
    publish: (data: any) =>
      fetchWithAuth<any>("/api/public/marketplace", { method: "POST", body: JSON.stringify(data) }),
    install: (id: string) =>
      fetchWithAuth<any>(`/api/public/marketplace/${id}/install`, { method: "POST" }),
    rate: (id: string, data: any) =>
      fetchWithAuth<any>(`/api/public/marketplace/${id}/rate`, { method: "POST", body: JSON.stringify(data) }),
  },

  // Compliance Reports
  complianceReports: {
    list: () =>
      fetchWithAuth<any>("/api/public/compliance-reports"),
    generate: (data: any) =>
      fetchWithAuth<any>("/api/public/compliance-reports", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) =>
      fetchWithAuth<any>(`/api/public/compliance-reports/${id}`),
    getTemplates: () =>
      fetchWithAuth<any>("/api/public/compliance-reports/templates"),
  },
  orchestration: {
    listSessions: () => fetchWithAuth<any>("/api/public/orchestration/sessions"),
    createSession: (data: any) => fetchWithAuth<any>("/api/public/orchestration/sessions", { method: "POST", body: JSON.stringify(data) }),
    getSession: (id: string) => fetchWithAuth<any>(`/api/public/orchestration/sessions/${id}`),
    executeCommand: (id: string, cmd: any) => fetchWithAuth<any>(`/api/public/orchestration/sessions/${id}/command`, { method: "POST", body: JSON.stringify(cmd) }),
    addBreakpoint: (id: string, bp: any) => fetchWithAuth<any>(`/api/public/orchestration/sessions/${id}/breakpoints`, { method: "POST", body: JSON.stringify(bp) }),
  },
  rca: {
    analyze: (data: any) => fetchWithAuth<any>("/api/public/rca/analyze", { method: "POST", body: JSON.stringify(data) }),
    listReports: () => fetchWithAuth<any>("/api/public/rca/reports"),
    getReport: (id: string) => fetchWithAuth<any>(`/api/public/rca/reports/${id}`),
  },
  agentVersions: {
    list: (agentName?: string) => fetchWithAuth<any>(`/api/public/agent-versions${agentName ? `?agentName=${agentName}` : ""}`),
    create: (data: any) => fetchWithAuth<any>("/api/public/agent-versions", { method: "POST", body: JSON.stringify(data) }),
    getActive: (agentName?: string) => fetchWithAuth<any>(`/api/public/agent-versions/active${agentName ? `?agentName=${agentName}` : ""}`),
    get: (id: string) => fetchWithAuth<any>(`/api/public/agent-versions/${id}`),
    rollback: (id: string) => fetchWithAuth<any>(`/api/public/agent-versions/${id}/rollback`, { method: "POST" }),
    diff: (data: { versionIdA: string; versionIdB: string }) => fetchWithAuth<any>("/api/public/agent-versions/diff", { method: "POST", body: JSON.stringify(data) }),
  },
  predictions: {
    predict: (data: any) => fetchWithAuth<any>("/api/public/predictions/cost", { method: "POST", body: JSON.stringify(data) }),
    list: () => fetchWithAuth<any>("/api/public/predictions"),
    requestApproval: (id: string) => fetchWithAuth<any>(`/api/public/predictions/${id}/approve`, { method: "POST" }),
    decideApproval: (id: string, decision: any) => fetchWithAuth<any>(`/api/public/approvals/${id}/decide`, { method: "POST", body: JSON.stringify(decision) }),
  },
  embed: {
    getConfig: () => fetchWithAuth<any>("/api/public/embed/config"),
    createConfig: (data: any) => fetchWithAuth<any>("/api/public/embed/config", { method: "POST", body: JSON.stringify(data) }),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/embed/config", { method: "PUT", body: JSON.stringify(data) }),
    generateToken: () => fetchWithAuth<any>("/api/public/embed/token", { method: "POST" }),
  },
  agentBuilder: {
    generate: (data: any) => fetchWithAuth<any>("/api/public/agent-builder/generate", { method: "POST", body: JSON.stringify(data) }),
    list: () => fetchWithAuth<any>("/api/public/agent-builder/blueprints"),
    get: (id: string) => fetchWithAuth<any>(`/api/public/agent-builder/blueprints/${id}`),
    deploy: (id: string) => fetchWithAuth<any>(`/api/public/agent-builder/blueprints/${id}/deploy`, { method: "POST" }),
  },
  fleet: {
    getDashboard: () => fetchWithAuth<any>("/api/public/fleet/dashboard"),
    listAgents: () => fetchWithAuth<any>("/api/public/fleet/agents"),
    listPolicies: () => fetchWithAuth<any>("/api/public/fleet/policies"),
    createPolicy: (data: any) => fetchWithAuth<any>("/api/public/fleet/policies", { method: "POST", body: JSON.stringify(data) }),
    bulkUpdate: (data: any) => fetchWithAuth<any>("/api/public/fleet/bulk-update", { method: "POST", body: JSON.stringify(data) }),
    getScaling: () => fetchWithAuth<any>("/api/public/fleet/scaling"),
  },
  privacy: {
    scan: (data: any) => fetchWithAuth<any>("/api/public/privacy/scan", { method: "POST", body: JSON.stringify(data) }),
    getConfig: () => fetchWithAuth<any>("/api/public/privacy/config"),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/privacy/config", { method: "PUT", body: JSON.stringify(data) }),
    requestDeletion: (data: any) => fetchWithAuth<any>("/api/public/privacy/deletion-requests", { method: "POST", body: JSON.stringify(data) }),
    listDeletionRequests: () => fetchWithAuth<any>("/api/public/privacy/deletion-requests"),
  },
  mobile: {
    registerDevice: (data: any) => fetchWithAuth<any>("/api/public/mobile/devices", { method: "POST", body: JSON.stringify(data) }),
    getDashboard: () => fetchWithAuth<any>("/api/public/mobile/dashboard"),
    listNotifications: () => fetchWithAuth<any>("/api/public/mobile/notifications"),
  },
  plugins: {
    list: () => fetchWithAuth<any>("/api/public/plugins"),
    install: (data: any) => fetchWithAuth<any>("/api/public/plugins", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<any>(`/api/public/plugins/${id}`),
    activate: (id: string) => fetchWithAuth<any>(`/api/public/plugins/${id}/activate`, { method: "POST" }),
    disable: (id: string) => fetchWithAuth<any>(`/api/public/plugins/${id}/disable`, { method: "POST" }),
    execute: (id: string, data: any) => fetchWithAuth<any>(`/api/public/plugins/${id}/execute`, { method: "POST", body: JSON.stringify(data) }),
    uninstall: (id: string) => fetchWithAuth<any>(`/api/public/plugins/${id}`, { method: "DELETE" }),
  },

  // Memory
  memory: {
    analyze: (data: any) => fetchWithAuth<any>("/api/public/memory/analyze", { method: "POST", body: JSON.stringify(data) }),
    getSnapshot: (traceId: string, step: number) => fetchWithAuth<any>(`/api/public/memory/traces/${traceId}/snapshots/${step}`),
    getOptimizations: () => fetchWithAuth<any>("/api/public/memory/optimizations"),
  },

  // Distributed Traces
  distributedTraces: {
    get: (traceId: string) => fetchWithAuth<any>(`/api/public/distributed/traces/${traceId}`),
    getServiceMap: () => fetchWithAuth<any>("/api/public/distributed/service-map"),
    correlate: (data: any) => fetchWithAuth<any>("/api/public/distributed/correlate", { method: "POST", body: JSON.stringify(data) }),
  },

  // Prompt Cache
  promptCache: {
    analyze: () => fetchWithAuth<any>("/api/public/prompt-cache/analyze"),
    getConfig: () => fetchWithAuth<any>("/api/public/prompt-cache/config"),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/prompt-cache/config", { method: "PUT", body: JSON.stringify(data) }),
    getStats: () => fetchWithAuth<any>("/api/public/prompt-cache/stats"),
    invalidate: () => fetchWithAuth<any>("/api/public/prompt-cache/invalidate", { method: "POST" }),
  },

  // Chaos Testing
  chaos: {
    list: () => fetchWithAuth<any>("/api/public/chaos/experiments"),
    create: (data: any) => fetchWithAuth<any>("/api/public/chaos/experiments", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<any>(`/api/public/chaos/experiments/${id}`),
    run: (id: string) => fetchWithAuth<any>(`/api/public/chaos/experiments/${id}/run`, { method: "POST" }),
    getScorecard: (agentName: string) => fetchWithAuth<any>(`/api/public/chaos/scorecard/${agentName}`),
  },

  // Custom Metrics
  customMetrics: {
    list: () => fetchWithAuth<any>("/api/public/custom-metrics"),
    create: (data: any) => fetchWithAuth<any>("/api/public/custom-metrics", { method: "POST", body: JSON.stringify(data) }),
    getValues: (id: string) => fetchWithAuth<any>(`/api/public/custom-metrics/${id}/values`),
    listDashboards: () => fetchWithAuth<any>("/api/public/custom-metrics/dashboards"),
    createDashboard: (data: any) => fetchWithAuth<any>("/api/public/custom-metrics/dashboards", { method: "POST", body: JSON.stringify(data) }),
    listAlerts: () => fetchWithAuth<any>("/api/public/custom-metrics/alerts"),
    createAlert: (data: any) => fetchWithAuth<any>("/api/public/custom-metrics/alerts", { method: "POST", body: JSON.stringify(data) }),
  },

  // Handoffs
  handoffs: {
    initiate: (data: any) => fetchWithAuth<any>("/api/public/handoffs", { method: "POST", body: JSON.stringify(data) }),
    accept: (id: string) => fetchWithAuth<any>(`/api/public/handoffs/${id}/accept`, { method: "POST" }),
    complete: (id: string) => fetchWithAuth<any>(`/api/public/handoffs/${id}/complete`, { method: "POST" }),
    getChain: (traceId: string) => fetchWithAuth<any>(`/api/public/handoffs/chain/${traceId}`),
    getStats: () => fetchWithAuth<any>("/api/public/handoffs/stats"),
  },

  // Annotations
  annotations: {
    list: (traceId: string) => fetchWithAuth<any>(`/api/public/annotations/traces/${traceId}`),
    create: (data: any) => fetchWithAuth<any>("/api/public/annotations", { method: "POST", body: JSON.stringify(data) }),
    reply: (id: string, data: any) => fetchWithAuth<any>(`/api/public/annotations/${id}/reply`, { method: "POST", body: JSON.stringify(data) }),
    resolve: (id: string) => fetchWithAuth<any>(`/api/public/annotations/${id}/resolve`, { method: "POST" }),
    getPresence: (traceId: string) => fetchWithAuth<any>(`/api/public/annotations/presence/${traceId}`),
  },

  // Carbon
  carbon: {
    getFootprint: () => fetchWithAuth<any>("/api/public/carbon/footprint"),
    getConfig: () => fetchWithAuth<any>("/api/public/carbon/config"),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/carbon/config", { method: "PUT", body: JSON.stringify(data) }),
    getSuggestions: () => fetchWithAuth<any>("/api/public/carbon/suggestions"),
  },

  // Synthetic Data
  syntheticData: {
    generate: (data: any) => fetchWithAuth<any>("/api/public/synthetic-data/generate", { method: "POST", body: JSON.stringify(data) }),
    list: () => fetchWithAuth<any>("/api/public/synthetic-data/datasets"),
    get: (id: string) => fetchWithAuth<any>(`/api/public/synthetic-data/datasets/${id}`),
    getStats: () => fetchWithAuth<any>("/api/public/synthetic-data/stats"),
  },

  // SLOs
  slos: {
    list: () => fetchWithAuth<any>("/api/public/slos"),
    create: (data: any) => fetchWithAuth<any>("/api/public/slos", { method: "POST", body: JSON.stringify(data) }),
    getStatus: (id: string) => fetchWithAuth<any>(`/api/public/slos/${id}/status`),
    getReport: () => fetchWithAuth<any>("/api/public/slos/report"),
    getHistory: (id: string) => fetchWithAuth<any>(`/api/public/slos/${id}/history`),
  },

  // Autonomy Gradient
  autonomy: {
    getDashboard: () => fetchWithAuth<any>("/api/public/autonomy/dashboard"),
    getConfig: (agent: string) => fetchWithAuth<any>(`/api/public/autonomy/${agent}`),
    set: (data: any) => fetchWithAuth<any>("/api/public/autonomy", { method: "POST", body: JSON.stringify(data) }),
    getTrust: (agent: string) => fetchWithAuth<any>(`/api/public/autonomy/${agent}/trust`),
  },

  // Cross-Org Benchmarks
  crossOrg: {
    submit: (data: any) => fetchWithAuth<any>("/api/public/cross-org/submit", { method: "POST", body: JSON.stringify(data) }),
    getReport: () => fetchWithAuth<any>("/api/public/cross-org/report"),
    getIndustry: (cat: string) => fetchWithAuth<any>(`/api/public/cross-org/industry/${cat}`),
  },

  // Intent Verification
  intents: {
    declare: (data: any) => fetchWithAuth<any>("/api/public/intents", { method: "POST", body: JSON.stringify(data) }),
    verify: (id: string, data: any) => fetchWithAuth<any>(`/api/public/intents/${id}/verify`, { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<any>(`/api/public/intents/${id}`),
    getStats: () => fetchWithAuth<any>("/api/public/intents/stats"),
  },

  // Cost Attribution
  costAttribution: {
    attribute: (data: any) => fetchWithAuth<any>("/api/public/cost-attribution", { method: "POST", body: JSON.stringify(data) }),
    getReport: () => fetchWithAuth<any>("/api/public/cost-attribution/report"),
    list: () => fetchWithAuth<any>("/api/public/cost-attribution"),
  },

  // Knowledge Graph
  knowledgeGraph: {
    build: () => fetchWithAuth<any>("/api/public/knowledge-graph"),
    query: (data: any) => fetchWithAuth<any>("/api/public/knowledge-graph/query", { method: "POST", body: JSON.stringify(data) }),
    getStats: () => fetchWithAuth<any>("/api/public/knowledge-graph/stats"),
  },

  // Compliance Monitor
  complianceMonitor: {
    listPolicies: () => fetchWithAuth<any>("/api/public/compliance-monitor/policies"),
    createPolicy: (data: any) => fetchWithAuth<any>("/api/public/compliance-monitor/policies", { method: "POST", body: JSON.stringify(data) }),
    evaluate: (data: any) => fetchWithAuth<any>("/api/public/compliance-monitor/evaluate", { method: "POST", body: JSON.stringify(data) }),
    getScore: (fw: string) => fetchWithAuth<any>(`/api/public/compliance-monitor/score/${fw}`),
    configure: (data: any) => fetchWithAuth<any>("/api/public/compliance-monitor/configure", { method: "POST", body: JSON.stringify(data) }),
  },

  // Multi-Modal Traces
  multimodal: {
    register: (data: any) => fetchWithAuth<any>("/api/public/multimodal/attachments", { method: "POST", body: JSON.stringify(data) }),
    getTraceAttachments: (traceId: string) => fetchWithAuth<any>(`/api/public/multimodal/traces/${traceId}`),
    getAttachment: (id: string) => fetchWithAuth<any>(`/api/public/multimodal/attachments/${id}`),
    getSummary: (traceId: string) => fetchWithAuth<any>(`/api/public/multimodal/traces/${traceId}/summary`),
    list: () => fetchWithAuth<any>("/api/public/multimodal/attachments"),
  },

  // Collaboration Patterns
  collabPatterns: {
    list: () => fetchWithAuth<any>("/api/public/collab-patterns"),
    get: (id: string) => fetchWithAuth<any>(`/api/public/collab-patterns/${id}`),
    deploy: (id: string, data: any) => fetchWithAuth<any>(`/api/public/collab-patterns/${id}/deploy`, { method: "POST", body: JSON.stringify(data) }),
    getDeployments: () => fetchWithAuth<any>("/api/public/collab-patterns/deployments"),
    getAnalytics: (id: string) => fetchWithAuth<any>(`/api/public/collab-patterns/${id}/analytics`),
  },

  // Federated Learning
  federatedLearning: {
    listRings: () => fetchWithAuth<any>("/api/public/federated/rings"),
    joinRing: (data: any) => fetchWithAuth<any>("/api/public/federated/rings/join", { method: "POST", body: JSON.stringify(data) }),
    getInsights: (ringId: string) => fetchWithAuth<any>(`/api/public/federated/rings/${ringId}/insights`),
    getConfig: () => fetchWithAuth<any>("/api/public/federated/config"),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/federated/config", { method: "PUT", body: JSON.stringify(data) }),
  },

  // Observability Copilot
  copilot: {
    ask: (data: any) => fetchWithAuth<any>("/api/public/copilot/ask", { method: "POST", body: JSON.stringify(data) }),
    getSuggestions: () => fetchWithAuth<any>("/api/public/copilot/suggestions"),
    getInsights: () => fetchWithAuth<any>("/api/public/copilot/insights"),
  },

  // Replay Sessions
  replaySessions: {
    list: () => fetchWithAuth<any>("/api/public/replay-sessions"),
    create: (data: any) => fetchWithAuth<any>("/api/public/replay-sessions", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<any>(`/api/public/replay-sessions/${id}`),
    getTimeline: (id: string) => fetchWithAuth<any>(`/api/public/replay-sessions/${id}/timeline`),
    branch: (id: string, data: any) => fetchWithAuth<any>(`/api/public/replay-sessions/${id}/branch`, { method: "POST", body: JSON.stringify(data) }),
    getPlayback: (id: string) => fetchWithAuth<any>(`/api/public/replay-sessions/${id}/playback`),
    share: (id: string) => fetchWithAuth<any>(`/api/public/replay-sessions/${id}/share`, { method: "POST" }),
  },

  // Cost Guardrails
  costGuardrails: {
    getDashboard: () => fetchWithAuth<any>("/api/public/cost-guardrails/dashboard"),
    createPolicy: (data: any) => fetchWithAuth<any>("/api/public/cost-guardrails/policies", { method: "POST", body: JSON.stringify(data) }),
    listPolicies: () => fetchWithAuth<any>("/api/public/cost-guardrails/policies"),
    checkBudget: (data: any) => fetchWithAuth<any>("/api/public/cost-guardrails/check", { method: "POST", body: JSON.stringify(data) }),
    getForecast: () => fetchWithAuth<any>("/api/public/cost-guardrails/forecast"),
    listViolations: () => fetchWithAuth<any>("/api/public/cost-guardrails/violations"),
  },

  // Multi-Agent
  multiAgent: {
    listSessions: () => fetchWithAuth<any>("/api/public/multi-agent/sessions"),
    analyze: (data: any) => fetchWithAuth<any>("/api/public/multi-agent/analyze", { method: "POST", body: JSON.stringify(data) }),
    getSession: (id: string) => fetchWithAuth<any>(`/api/public/multi-agent/sessions/${id}`),
  },

  // Prompt CI
  promptCI: {
    listBaselines: () => fetchWithAuth<any>("/api/public/prompt-ci/baselines"),
    createBaseline: (data: any) => fetchWithAuth<any>("/api/public/prompt-ci/baselines", { method: "POST", body: JSON.stringify(data) }),
    getBaseline: (id: string) => fetchWithAuth<any>(`/api/public/prompt-ci/baselines/${id}`),
    runComparison: (data: any) => fetchWithAuth<any>("/api/public/prompt-ci/compare", { method: "POST", body: JSON.stringify(data) }),
    listRuns: () => fetchWithAuth<any>("/api/public/prompt-ci/runs"),
  },

  // Agent Benchmarks
  agentBenchmarks: {
    listSuites: () => fetchWithAuth<any>("/api/public/agent-benchmarks/suites"),
    createSuite: (data: any) => fetchWithAuth<any>("/api/public/agent-benchmarks/suites", { method: "POST", body: JSON.stringify(data) }),
    getSuite: (id: string) => fetchWithAuth<any>(`/api/public/agent-benchmarks/suites/${id}`),
    runBenchmark: (data: any) => fetchWithAuth<any>("/api/public/agent-benchmarks/run", { method: "POST", body: JSON.stringify(data) }),
    getLeaderboard: (id: string) => fetchWithAuth<any>(`/api/public/agent-benchmarks/suites/${id}/leaderboard`),
  },

  // Semantic Search
  semanticSearch: {
    search: (data: any) => fetchWithAuth<any>("/api/public/semantic-search/query", { method: "POST", body: JSON.stringify(data) }),
    getClusters: () => fetchWithAuth<any>("/api/public/semantic-search/clusters"),
    getAnomalyPatterns: () => fetchWithAuth<any>("/api/public/semantic-search/anomaly-patterns"),
    getDashboard: () => fetchWithAuth<any>("/api/public/semantic-search/dashboard"),
  },

  // Agent Knowledge Graph
  agentKnowledgeGraph: {
    build: (params?: any) => fetchWithAuth<any>("/api/public/agent-knowledge-graph/build", { method: "POST", body: params ? JSON.stringify(params) : undefined }),
    query: (data: any) => fetchWithAuth<any>("/api/public/agent-knowledge-graph/query", { method: "POST", body: JSON.stringify(data) }),
    getEvolution: () => fetchWithAuth<any>("/api/public/agent-knowledge-graph/evolution"),
    getStats: () => fetchWithAuth<any>("/api/public/agent-knowledge-graph/stats"),
  },

  // IDE Trace View
  ideTraceView: {
    getFileMapping: (filePath: string) => fetchWithAuth<any>(`/api/public/ide-trace-view/file-mapping?filePath=${encodeURIComponent(filePath)}`),
    batchMappings: (data: any) => fetchWithAuth<any>("/api/public/ide-trace-view/batch-mappings", { method: "POST", body: JSON.stringify(data) }),
    getTraceContext: (traceId: string) => fetchWithAuth<any>(`/api/public/ide-trace-view/trace-context/${traceId}`),
  },

  // Federated Aggregation
  federatedAggregation: {
    getDashboard: () => fetchWithAuth<any>("/api/public/federated-aggregation/dashboard"),
    registerInstance: (data: any) => fetchWithAuth<any>("/api/public/federated-aggregation/instances", { method: "POST", body: JSON.stringify(data) }),
    listInstances: () => fetchWithAuth<any>("/api/public/federated-aggregation/instances"),
    submitMetrics: (data: any) => fetchWithAuth<any>("/api/public/federated-aggregation/metrics", { method: "POST", body: JSON.stringify(data) }),
    getBenchmarks: (metricType?: string) => fetchWithAuth<any>(`/api/public/federated-aggregation/benchmarks${metricType ? `?metricType=${encodeURIComponent(metricType)}` : ""}`),
    getInsights: () => fetchWithAuth<any>("/api/public/federated-aggregation/insights"),
  },

  // Workflow Simulator
  workflowSimulator: {
    list: () => fetchWithAuth<any>("/api/public/workflow-simulator/workflows"),
    create: (data: any) => fetchWithAuth<any>("/api/public/workflow-simulator/workflows", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<any>(`/api/public/workflow-simulator/workflows/${id}`),
    update: (id: string, data: any) => fetchWithAuth<any>(`/api/public/workflow-simulator/workflows/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth<any>(`/api/public/workflow-simulator/workflows/${id}`, { method: "DELETE" }),
    validate: (data: any) => fetchWithAuth<any>("/api/public/workflow-simulator/validate", { method: "POST", body: JSON.stringify(data) }),
    simulate: (data: any) => fetchWithAuth<any>("/api/public/workflow-simulator/simulate", { method: "POST", body: JSON.stringify(data) }),
    getSimulation: (id: string) => fetchWithAuth<any>(`/api/public/workflow-simulator/simulations/${id}`),
    listSimulations: (workflowId: string) => fetchWithAuth<any>(`/api/public/workflow-simulator/workflows/${workflowId}/simulations`),
  },

  // Auto-Discovery
  autoDiscovery: {
    scan: () => fetchWithAuth<any>("/api/public/auto-discovery/scan", { method: "POST" }),
    getFramework: (id: string) => fetchWithAuth<any>(`/api/public/auto-discovery/frameworks/${id}`),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/auto-discovery/config", { method: "PUT", body: JSON.stringify(data) }),
    toggleInstrumentation: (id: string, enabled: boolean) => fetchWithAuth<any>(`/api/public/auto-discovery/frameworks/${id}/instrumentation`, { method: "PUT", body: JSON.stringify({ enabled }) }),
  },

  // Cloud Onboarding
  cloudOnboarding: {
    get: () => fetchWithAuth<any>("/api/public/cloud-onboarding"),
    completeStep: (data: any) => fetchWithAuth<any>("/api/public/cloud-onboarding/steps/complete", { method: "POST", body: JSON.stringify(data) }),
    generateQuickstart: (data: any) => fetchWithAuth<any>("/api/public/cloud-onboarding/quickstart", { method: "POST", body: JSON.stringify(data) }),
    getUsage: () => fetchWithAuth<any>("/api/public/cloud-onboarding/usage"),
    checkQuota: (data: any) => fetchWithAuth<any>("/api/public/cloud-onboarding/quota/check", { method: "POST", body: JSON.stringify(data) }),
  },

  // AI Debugger
  aiDebugger: {
    debug: (data: any) => fetchWithAuth<any>("/api/public/ai-debugger/debug", { method: "POST", body: JSON.stringify(data) }),
    getHistory: (traceId: string) => fetchWithAuth<any>(`/api/public/ai-debugger/traces/${traceId}/history`),
    getContext: (traceId: string) => fetchWithAuth<any>(`/api/public/ai-debugger/traces/${traceId}/context`),
  },

  // Prompt Optimization
  promptOptimization: {
    start: (data: any) => fetchWithAuth<any>("/api/public/prompt-optimization/start", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<any>(`/api/public/prompt-optimization/${id}`),
    list: () => fetchWithAuth<any>("/api/public/prompt-optimization"),
    getConfig: () => fetchWithAuth<any>("/api/public/prompt-optimization/config"),
    updateConfig: (data: any) => fetchWithAuth<any>("/api/public/prompt-optimization/config", { method: "PUT", body: JSON.stringify(data) }),
    approveVariant: (id: string) => fetchWithAuth<any>(`/api/public/prompt-optimization/variants/${id}/approve`, { method: "POST" }),
    rejectVariant: (id: string) => fetchWithAuth<any>(`/api/public/prompt-optimization/variants/${id}/reject`, { method: "POST" }),
  },

  // Cost Alerting
  costAlerting: {
    createRule: (data: any) => fetchWithAuth<any>("/api/public/cost-alerting/rules", { method: "POST", body: JSON.stringify(data) }),
    listRules: () => fetchWithAuth<any>("/api/public/cost-alerting/rules"),
    deleteRule: (id: string) => fetchWithAuth<any>(`/api/public/cost-alerting/rules/${id}`, { method: "DELETE" }),
    listAlerts: () => fetchWithAuth<any>("/api/public/cost-alerting/alerts"),
    acknowledgeAlert: (id: string) => fetchWithAuth<any>(`/api/public/cost-alerting/alerts/${id}/acknowledge`, { method: "POST" }),
    getCircuitBreaker: () => fetchWithAuth<any>("/api/public/cost-alerting/circuit-breaker"),
    updateCircuitBreaker: (data: any) => fetchWithAuth<any>("/api/public/cost-alerting/circuit-breaker", { method: "PUT", body: JSON.stringify(data) }),
    checkCost: (data: any) => fetchWithAuth<any>("/api/public/cost-alerting/check", { method: "POST", body: JSON.stringify(data) }),
  },

  // Regression Suite
  regressionSuite: {
    createDataset: (data: any) => fetchWithAuth<any>("/api/public/regression-suite/datasets", { method: "POST", body: JSON.stringify(data) }),
    getDataset: (id: string) => fetchWithAuth<any>(`/api/public/regression-suite/datasets/${id}`),
    listDatasets: () => fetchWithAuth<any>("/api/public/regression-suite/datasets"),
    runRegression: (data: any) => fetchWithAuth<any>("/api/public/regression-suite/run", { method: "POST", body: JSON.stringify(data) }),
    getRun: (id: string) => fetchWithAuth<any>(`/api/public/regression-suite/runs/${id}`),
    listRuns: () => fetchWithAuth<any>("/api/public/regression-suite/runs"),
  },

  // Collaboration Hub
  collabHub: {
    createQueue: (data: any) => fetchWithAuth<any>("/api/public/collab-hub/queues", { method: "POST", body: JSON.stringify(data) }),
    listQueues: () => fetchWithAuth<any>("/api/public/collab-hub/queues"),
    assignReview: (data: any) => fetchWithAuth<any>("/api/public/collab-hub/reviews/assign", { method: "POST", body: JSON.stringify(data) }),
    completeReview: (id: string, data: any) => fetchWithAuth<any>(`/api/public/collab-hub/reviews/${id}/complete`, { method: "POST", body: JSON.stringify(data) }),
    createStandard: (data: any) => fetchWithAuth<any>("/api/public/collab-hub/standards", { method: "POST", body: JSON.stringify(data) }),
    listStandards: () => fetchWithAuth<any>("/api/public/collab-hub/standards"),
    getActivityFeed: () => fetchWithAuth<any>("/api/public/collab-hub/activity"),
  },

  // OpenTelemetry Compatibility
  otelCompat: {
    createDestination: (data: any) => fetchWithAuth<any>("/api/public/otel-compat/destinations", { method: "POST", body: JSON.stringify(data) }),
    listDestinations: () => fetchWithAuth<any>("/api/public/otel-compat/destinations"),
    deleteDestination: (id: string) => fetchWithAuth<any>(`/api/public/otel-compat/destinations/${id}`, { method: "DELETE" }),
    getMappings: () => fetchWithAuth<any>("/api/public/otel-compat/mappings"),
    getDashboard: () => fetchWithAuth<any>("/api/public/otel-compat/dashboard"),
    generateCollectorConfig: () => fetchWithAuth<any>("/api/public/otel-compat/collector-config", { method: "POST" }),
  },

  // Security Scanner
  securityScanner: {
    scan: (data: any) => fetchWithAuth<any>("/api/public/security-scanner/scan", { method: "POST", body: JSON.stringify(data) }),
    createPolicy: (data: any) => fetchWithAuth<any>("/api/public/security-scanner/policies", { method: "POST", body: JSON.stringify(data) }),
    listPolicies: () => fetchWithAuth<any>("/api/public/security-scanner/policies"),
    getDashboard: () => fetchWithAuth<any>("/api/public/security-scanner/dashboard"),
    acknowledgeFinding: (id: string) => fetchWithAuth<any>(`/api/public/security-scanner/findings/${id}/acknowledge`, { method: "POST" }),
  },
};
export interface User {
  id: string;
  email: string;
  name: string | null;
  image: string | null;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  createdAt: string;
}

export interface Project {
  id: string;
  organizationId: string;
  name: string;
  slug: string;
  description: string | null;
  settings: Record<string, unknown> | null;
  retentionDays: number;
  rateLimitPerMin: number;
  createdAt: string;
}

export interface Trace {
  id: string;
  projectId: string;
  name: string;
  timestamp: string;
  input: Record<string, unknown> | null;
  output: Record<string, unknown> | null;
  metadata: Record<string, unknown> | null;
  tags: string[];
  userId: string | null;
  sessionId: string | null;
  release: string | null;
  version: string | null;
  level: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
  statusMessage: string | null;
  public: boolean;
  latency: number;
  totalCost: number;
  observations?: Observation[];
  scores?: Score[];
}

export interface Observation {
  id: string;
  traceId: string;
  projectId: string;
  parentObservationId: string | null;
  name: string;
  type: "SPAN" | "GENERATION" | "EVENT";
  startTime: string;
  endTime: string | null;
  input: Record<string, unknown> | null;
  output: Record<string, unknown> | null;
  metadata: Record<string, unknown> | null;
  level: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
  statusMessage: string | null;
  version: string | null;
  model: string | null;
  modelParameters: Record<string, unknown> | null;
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  cost: number;
  children?: Observation[];
  scores?: Score[];
}

export interface Session {
  id: string;
  projectId: string;
  createdAt: string;
  traceCount: number;
  totalDuration: number | null;
  totalCost: number | null;
  traces?: Trace[];
}

export interface Score {
  id: string;
  traceId: string;
  observationId: string | null;
  projectId: string;
  name: string;
  value: number | null;
  stringValue: string | null;
  dataType: "NUMERIC" | "CATEGORICAL" | "BOOLEAN";
  source: "API" | "ANNOTATION" | "EVAL";
  comment: string | null;
  timestamp: string;
}

export interface Prompt {
  id: string;
  projectId: string;
  name: string;
  type: "TEXT" | "CHAT";
  isActive: boolean;
  version: PromptVersion;
  versions: PromptVersion[];
  labels: string[];
  createdAt: string;
}

export interface PromptVersion {
  id: string;
  promptId: string;
  version: number;
  prompt: string | null;
  messages: PromptMessage[] | null;
  config: Record<string, unknown> | null;
  labels: string[];
  variables: string[];
  createdAt: string;
}

export interface PromptMessage {
  role: string;
  content: string;
}

export interface Dataset {
  id: string;
  projectId: string;
  name: string;
  description: string | null;
  metadata: Record<string, unknown> | null;
  itemCount: number;
  runCount: number;
  createdAt: string;
}

export interface DatasetItem {
  id: string;
  datasetId: string;
  input: Record<string, unknown>;
  expectedOutput: Record<string, unknown> | null;
  metadata: Record<string, unknown> | null;
  sourceTraceId: string | null;
  sourceObservationId: string | null;
  status: "ACTIVE" | "ARCHIVED";
  createdAt: string;
}

export interface DatasetRun {
  id: string;
  datasetId: string;
  name: string;
  description: string | null;
  metadata: Record<string, unknown> | null;
  itemCount: number;
  createdAt: string;
}

export interface Evaluator {
  id: string;
  projectId: string;
  name: string;
  description: string | null;
  type: "LLM_AS_JUDGE" | "RULE_BASED" | "HUMAN";
  scoreName: string;
  scoreDataType: "NUMERIC" | "CATEGORICAL" | "BOOLEAN";
  scoreCategories: string[] | null;
  promptTemplate: string | null;
  variables: string[] | null;
  config: Record<string, unknown> | null;
  targetFilter: Record<string, unknown> | null;
  samplingRate: number;
  enabled: boolean;
  evalCount: number;
  createdAt: string;
}

export interface EvaluatorTemplate {
  id: string;
  name: string;
  description: string;
  type: "LLM_AS_JUDGE" | "RULE_BASED";
  promptTemplate: string;
  variables: string[];
  scoreDataType: "NUMERIC" | "CATEGORICAL" | "BOOLEAN";
  scoreCategories: string[] | null;
}

export interface APIKey {
  id: string;
  projectId: string;
  name: string;
  displayKey: string;
  scopes: string[] | null;
  expiresAt: string | null;
  lastUsedAt: string | null;
  createdAt: string;
}

export interface APIKeyWithSecret extends APIKey {
  key: string;
}

export interface Metrics {
  traceCount: number;
  observationCount: number;
  totalCost: number;
  totalTokens: number;
  avgLatency: number | null;
  p50Latency: number | null;
  p95Latency: number | null;
  p99Latency: number | null;
  modelUsage: ModelUsage[];
}

export interface ModelUsage {
  model: string;
  count: number;
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  cost: number;
}

export interface DailyCost {
  date: string;
  totalCost: number;
  traceCount: number;
  modelCosts: ModelCost[];
}

export interface ModelCost {
  model: string;
  cost: number;
  count: number;
}

// Input types
export interface TraceListParams {
  limit?: string;
  cursor?: string;
  userId?: string;
  sessionId?: string;
  name?: string;
  tags?: string;
  fromTimestamp?: string;
  toTimestamp?: string;
}

export interface SessionListParams {
  limit?: string;
  cursor?: string;
  fromTimestamp?: string;
  toTimestamp?: string;
}

export interface ScoreListParams {
  limit?: string;
  cursor?: string;
  traceId?: string;
  observationId?: string;
  name?: string;
  source?: string;
}

export interface CreateScoreInput {
  traceId: string;
  observationId?: string;
  name: string;
  value?: number;
  stringValue?: string;
  dataType?: "NUMERIC" | "CATEGORICAL" | "BOOLEAN";
  source?: "API" | "ANNOTATION" | "EVAL";
  comment?: string;
}

export interface UpdateScoreInput {
  value?: number;
  stringValue?: string;
  comment?: string;
}

export interface CreatePromptInput {
  name: string;
  type?: "TEXT" | "CHAT";
  prompt?: string;
  messages?: PromptMessage[];
  config?: Record<string, unknown>;
  labels?: string[];
}

export interface UpdatePromptInput {
  prompt?: string;
  messages?: PromptMessage[];
  config?: Record<string, unknown>;
  labels?: string[];
}

export interface CreateDatasetInput {
  name: string;
  description?: string;
  metadata?: Record<string, unknown>;
}

export interface UpdateDatasetInput {
  name?: string;
  description?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateDatasetItemInput {
  input: Record<string, unknown>;
  expectedOutput?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  sourceTraceId?: string;
  sourceObservationId?: string;
}

export interface CreateDatasetRunInput {
  name: string;
  description?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateEvaluatorInput {
  name: string;
  description?: string;
  type?: "LLM_AS_JUDGE" | "RULE_BASED" | "HUMAN";
  scoreName: string;
  scoreDataType?: "NUMERIC" | "CATEGORICAL" | "BOOLEAN";
  scoreCategories?: string[];
  promptTemplate?: string;
  variables?: string[];
  config?: Record<string, unknown>;
  targetFilter?: Record<string, unknown>;
  samplingRate?: number;
  enabled?: boolean;
  templateId?: string;
}

export interface UpdateEvaluatorInput {
  name?: string;
  description?: string;
  promptTemplate?: string;
  variables?: string[];
  scoreCategories?: string[];
  config?: Record<string, unknown>;
  targetFilter?: Record<string, unknown>;
  samplingRate?: number;
  enabled?: boolean;
}

export interface CreateProjectInput {
  organizationId: string;
  name: string;
  description?: string;
  settings?: Record<string, unknown>;
  retentionDays?: number;
  rateLimitPerMin?: number;
}

export interface UpdateProjectInput {
  name?: string;
  description?: string;
  settings?: Record<string, unknown>;
  retentionDays?: number;
  rateLimitPerMin?: number;
}

export interface CreateAPIKeyInput {
  name: string;
  scopes?: string[];
  expiresAt?: string;
}

export interface MetricsParams {
  fromTimestamp: string;
  toTimestamp: string;
  userId?: string;
  sessionId?: string;
  name?: string;
  tags?: string;
}

export interface DailyCostsParams {
  fromDate: string;
  toDate: string;
  groupBy?: string;
}

// SSO Types
export interface SSOConfiguration {
  id: string;
  organizationId: string;
  provider: "saml" | "oidc";
  enabled: boolean;
  issuer?: string;
  ssoUrl?: string;
  certificate?: string;
  clientId?: string;
  discoveryUrl?: string;
  allowedDomains?: string[];
  defaultRole?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateSSOInput {
  provider: "saml" | "oidc";
  enabled?: boolean;
  issuer?: string;
  ssoUrl?: string;
  certificate?: string;
  clientId?: string;
  clientSecret?: string;
  discoveryUrl?: string;
  allowedDomains?: string[];
  defaultRole?: string;
}

export interface UpdateSSOInput {
  enabled?: boolean;
  issuer?: string;
  ssoUrl?: string;
  certificate?: string;
  clientId?: string;
  clientSecret?: string;
  discoveryUrl?: string;
  allowedDomains?: string[];
  defaultRole?: string;
}

// Audit Log Types
export interface AuditLog {
  id: string;
  organizationId: string;
  userId: string;
  userName: string;
  userEmail: string;
  action: string;
  resourceType: string;
  resourceId: string;
  metadata: Record<string, unknown>;
  ipAddress: string;
  userAgent: string;
  timestamp: string;
}

export interface AuditSummary {
  totalEvents: number;
  eventsTrend: number;
  activeUsers: number;
  usersTrend: number;
  apiKeyEvents: number;
  apiKeyTrend: number;
  securityEvents: number;
  securityTrend: number;
}

export interface AuditExportJob {
  id: string;
  organizationId: string;
  status: "pending" | "processing" | "completed" | "failed";
  format: "json" | "csv";
  startDate: string;
  endDate: string;
  downloadUrl?: string;
  createdAt: string;
  completedAt?: string;
}

export interface AuditLogListParams {
  userId?: string;
  action?: string;
  resourceType?: string;
  startDate?: string;
  endDate?: string;
  cursor?: string;
  limit?: number;
}

// Checkpoint Types
export interface Checkpoint {
  id: string;
  projectId: string;
  traceId: string;
  traceName: string;
  type: "auto" | "manual";
  description: string | null;
  files: CheckpointFile[];
  metadata: Record<string, unknown> | null;
  createdAt: string;
}

export interface CheckpointFile {
  path: string;
  content: string;
  hash: string;
  size: number;
}

export interface CheckpointListParams {
  traceId?: string;
  type?: "auto" | "manual";
  cursor?: string;
  limit?: number;
}

export interface CreateCheckpointInput {
  traceId: string;
  type?: "auto" | "manual";
  description?: string;
  files: {
    path: string;
    content: string;
    hash: string;
  }[];
  metadata?: Record<string, unknown>;
}

// Git Link Types
export interface GitLink {
  id: string;
  projectId: string;
  traceId: string;
  commitSha: string;
  commitMessage: string;
  branch: string;
  repository: string;
  author: string;
  authorEmail: string;
  timestamp: string;
  createdAt: string;
}

export interface GitTimelineEntry {
  commitSha: string;
  commitMessage: string;
  author: string;
  timestamp: string;
  traceCount: number;
}

export interface CreateGitLinkInput {
  traceId: string;
  commitSha: string;
  commitMessage?: string;
  branch?: string;
  repository?: string;
  author?: string;
  authorEmail?: string;
  timestamp?: string;
}

// File Operation Types
export interface FileOperation {
  id: string;
  projectId: string;
  traceId: string;
  observationId?: string;
  operation: "read" | "write" | "delete" | "rename" | "create";
  path: string;
  oldPath?: string;
  size?: number;
  linesChanged?: number;
  timestamp: string;
}

export interface CreateFileOperationInput {
  traceId: string;
  observationId?: string;
  operation: "read" | "write" | "delete" | "rename" | "create";
  path: string;
  oldPath?: string;
  size?: number;
  linesChanged?: number;
}

// Terminal Command Types
export interface TerminalCommand {
  id: string;
  projectId: string;
  traceId: string;
  observationId?: string;
  command: string;
  args?: string[];
  exitCode?: number;
  stdout?: string;
  stderr?: string;
  durationMs?: number;
  workingDirectory?: string;
  timestamp: string;
}

export interface CreateTerminalCommandInput {
  traceId: string;
  observationId?: string;
  command: string;
  args?: string[];
  exitCode?: number;
  stdout?: string;
  stderr?: string;
  durationMs?: number;
  workingDirectory?: string;
}

// CI Run Types
export interface CIRun {
  id: string;
  projectId: string;
  provider: "github_actions" | "gitlab_ci" | "jenkins" | "circleci" | "other";
  runId: string;
  runUrl?: string;
  workflowName?: string;
  jobName?: string;
  status: "pending" | "running" | "completed" | "failed" | "cancelled";
  branch?: string;
  commitSha?: string;
  triggeredBy?: string;
  startedAt: string;
  completedAt?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
}

export interface CreateCIRunInput {
  provider: "github_actions" | "gitlab_ci" | "jenkins" | "circleci" | "other";
  runId: string;
  runUrl?: string;
  workflowName?: string;
  jobName?: string;
  status?: "pending" | "running" | "completed" | "failed" | "cancelled";
  branch?: string;
  commitSha?: string;
  triggeredBy?: string;
  metadata?: Record<string, unknown>;
}

export interface UpdateCIRunInput {
  status?: "pending" | "running" | "completed" | "failed" | "cancelled";
  completedAt?: string;
  metadata?: Record<string, unknown>;
}
