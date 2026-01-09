import { GraphQLClient } from "graphql-request";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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
};

// Types
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
