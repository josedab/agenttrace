// observability API client.
import { createSearchParams, fetchWithAuth, unsupportedApiFeature } from '../transport';
import {
  CheckpointResponse,
  ObservationTreeResponse,
  ScoreDataResponse,
  ScoreResponse,
  SessionResponse,
  TraceMetricsResponse,
  TraceResponse,
  dateRangeParams,
  fetchTraceMetrics,
  flattenObservationTree,
  normalizeCheckpoint,
  normalizeScore,
  normalizeSession,
  normalizeTrace,
  scoreRequestBody,
  toDashboardMetrics,
  toTraceStats,
} from '../normalizers';
import type {
  AgentGraph,
  AnalyticsGroupParams,
  AnalyticsOverview,
  CIRun,
  CheckpointListParams,
  CostAnalytics,
  CostTrendPoint,
  CreateCIRunInput,
  CreateCheckpointInput,
  CreateFileOperationInput,
  CreateGitLinkInput,
  CreateScoreInput,
  CreateTerminalCommandInput,
  FileOperation,
  GitLink,
  GitTimelineEntry,
  LatencyAnalytics,
  LatencyPercentilePoint,
  MetricsParams,
  ModelDistributionPoint,
  ReplayStepState,
  ReplayTimeline,
  ScoreListParams,
  ScoreStats,
  Session,
  SessionListParams,
  TerminalCommand,
  TraceFilterParams,
  TraceListParams,
  TraceVolumePoint,
  UpdateCIRunInput,
  UpdateScoreInput,
  UpdateTraceInput,
  UsageAnalytics,
} from '../contracts';

export const observabilityApi = {
  // Traces
  traces: {
    list: (params: TraceListParams = {}) => {
      const {
        startDate,
        endDate,
        minLatency,
        maxLatency,
        minCost,
        maxCost,
        level,
        search,
        cursor,
        offset,
        ...filters
      } = params;
      const query = {
        ...filters,
        offset: offset ?? (cursor ? Number(cursor) : undefined),
        fromTimestamp:
          params.fromTimestamp ?? (startDate instanceof Date ? startDate.toISOString() : startDate),
        toTimestamp:
          params.toTimestamp ?? (endDate instanceof Date ? endDate.toISOString() : endDate),
      };

      const endpoint = search ? '/api/public/traces/search' : '/api/public/traces';
      return fetchWithAuth<{
        traces: TraceResponse[];
        totalCount: number;
        hasMore: boolean;
      }>(
        `${endpoint}?${createSearchParams({
          ...query,
          q: search,
        })}`
      ).then((response) => {
        const traces = response.traces
          .map(normalizeTrace)
          .filter((trace) => !level || trace.level === level)
          .filter(
            (trace) =>
              minLatency === undefined || (trace.latency !== null && trace.latency >= minLatency)
          )
          .filter(
            (trace) =>
              maxLatency === undefined || (trace.latency !== null && trace.latency <= maxLatency)
          )
          .filter(
            (trace) =>
              minCost === undefined || (trace.totalCost !== null && trace.totalCost >= minCost)
          )
          .filter(
            (trace) =>
              maxCost === undefined || (trace.totalCost !== null && trace.totalCost <= maxCost)
          );
        return { ...response, traces };
      });
    },

    get: (id: string) =>
      fetchWithAuth<TraceResponse>(`/api/public/traces/${id}`).then(normalizeTrace),

    getObservations: (traceId: string) =>
      fetchWithAuth<ObservationTreeResponse | null>(
        `/api/public/traces/${traceId}/observations`
      ).then((tree) => (tree ? flattenObservationTree(tree) : [])),

    count: (params: TraceFilterParams = {}) =>
      observabilityApi.traces.list({ ...params, limit: 1, offset: 0 }).then(({ totalCount }) => ({
        count: totalCount,
      })),

    sessions: (params: { limit?: number } = {}) =>
      fetchWithAuth<{ data: SessionResponse[]; totalCount: number; hasMore: boolean }>(
        `/api/public/sessions?${createSearchParams(params)}`
      ).then((response) => (response.data ?? []).map(normalizeSession)),

    stats: (params: { dateRange?: string } = {}) =>
      fetchTraceMetrics(params.dateRange).then(toTraceStats),

    update: (id: string, data: UpdateTraceInput) =>
      fetchWithAuth<TraceResponse>(`/api/public/traces/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }).then(normalizeTrace),

    delete: (id: string) =>
      fetchWithAuth<void>(`/api/public/traces/${id}`, {
        method: 'DELETE',
      }),

    getGraph: (traceId: string) => fetchWithAuth<AgentGraph>(`/api/public/traces/${traceId}/graph`),

    getReplay: (traceId: string) =>
      fetchWithAuth<ReplayTimeline>(`/api/public/traces/${traceId}/replay`),

    getReplayStep: (traceId: string, step: number) =>
      fetchWithAuth<ReplayStepState>(`/api/public/traces/${traceId}/debug/step/${step}`),
  },
  observations: {
    listByTrace: (traceId: string) =>
      fetchWithAuth<ObservationTreeResponse | null>(
        `/api/public/traces/${traceId}/observations`
      ).then((tree) => (tree ? flattenObservationTree(tree) : [])),
  },
  // Sessions
  sessions: {
    list: (projectId: string, params?: SessionListParams) =>
      fetchWithAuth<{ sessions: Session[]; nextCursor?: string }>(
        `/api/public/sessions?${createSearchParams(params)}`,
        {
          method: 'GET',
          headers: { 'X-Project-ID': projectId },
        }
      ),

    get: (projectId: string, id: string) =>
      fetchWithAuth<Session>(`/api/public/sessions/${id}`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }),
  },
  // Scores
  scores: {
    list: (params: ScoreListParams = {}) => {
      const { minScore, maxScore, evaluatorId, ...supportedParams } = params;
      return fetchWithAuth<{
        scores: ScoreResponse[];
        totalCount: number;
        hasMore: boolean;
      }>(
        `/api/public/scores?${createSearchParams({
          ...supportedParams,
          name: params.name ?? params.scoreName,
          scoreName: undefined,
          offset: params.offset ?? (params.cursor ? Number(params.cursor) : undefined),
          cursor: undefined,
        })}`
      ).then((response) => ({
        ...response,
        scores: response.scores
          .map(normalizeScore)
          .filter(
            (score) =>
              minScore === undefined || (typeof score.value === 'number' && score.value >= minScore)
          )
          .filter(
            (score) =>
              maxScore === undefined || (typeof score.value === 'number' && score.value <= maxScore)
          )
          .filter((score) => evaluatorId === undefined || score.configId === evaluatorId),
      }));
    },

    get: (id: string) =>
      fetchWithAuth<ScoreResponse>(`/api/public/scores/${id}`).then(normalizeScore),

    listByTrace: (traceId: string) =>
      fetchWithAuth<ScoreDataResponse>(`/api/public/traces/${traceId}/scores`).then((response) =>
        response.data.map(normalizeScore)
      ),

    listByObservation: (observationId: string) =>
      observabilityApi.scores.list({ observationId, limit: 100, offset: 0 }).then((response) => response.scores),

    stats: (params: { name: string }) =>
      fetchWithAuth<ScoreStats>(
        `/api/public/scores/stats?${createSearchParams({ name: params.name })}`
      ),

    distribution: async (params: { scoreName: string; dateRange?: string }) => {
      const range = dateRangeParams(params.dateRange);
      const response = await observabilityApi.scores.list({
        scoreName: params.scoreName,
        fromTimestamp: range.from,
        toTimestamp: range.to,
        limit: 100,
        offset: 0,
      });
      const values = response.scores
        .map((score) => score.value)
        .filter((value): value is number => typeof value === 'number');
      if (values.length === 0) return { buckets: [] };

      const min = Math.min(...values);
      const max = Math.max(...values);
      const bucketWidth = max === min ? 1 : (max - min) / 10;
      const buckets = Array.from({ length: max === min ? 1 : 10 }, (_, index) => {
        const bucketMin = min + index * bucketWidth;
        const bucketMax = index === 9 || max === min ? max : bucketMin + bucketWidth;
        return {
          min: bucketMin,
          max: bucketMax,
          count: values.filter(
            (value) =>
              value >= bucketMin &&
              (index === 9 || max === min ? value <= bucketMax : value < bucketMax)
          ).length,
        };
      });
      return { buckets };
    },

    names: () =>
      observabilityApi.scores
        .list({ limit: 100, offset: 0 })
        .then((response) => [...new Set(response.scores.map((score) => score.name))].sort()),

    create: (data: CreateScoreInput) =>
      fetchWithAuth<ScoreResponse>('/api/public/scores', {
        method: 'POST',
        body: JSON.stringify(scoreRequestBody(data)),
      }).then(normalizeScore),

    update: (id: string, data: UpdateScoreInput) =>
      fetchWithAuth<ScoreResponse>(`/api/public/scores/${id}`, {
        method: 'PUT',
        body: JSON.stringify(scoreRequestBody(data)),
      }).then(normalizeScore),

    delete: (_id: string) => unsupportedApiFeature<void>('Score deletion'),
  },
  // Metrics
  metrics: {
    get: (projectId: string, params: MetricsParams) =>
      fetchWithAuth<TraceMetricsResponse>(
        `/api/public/metrics/project?${createSearchParams({
          from: params.fromTimestamp,
          to: params.toTimestamp,
        })}`,
        {
          method: 'GET',
          headers: { 'X-Project-ID': projectId },
        }
      ).then((metrics) => ({
        traceCount: metrics.totalCount,
        observationCount: 0,
        totalCost: metrics.totalCost,
        totalTokens: metrics.totalTokens,
        avgLatency: metrics.avgDuration,
        p50Latency: null,
        p95Latency: null,
        p99Latency: null,
        modelUsage: [],
      })),

    getDashboard: () => fetchTraceMetrics().then(toDashboardMetrics),
  },
  analytics: {
    getDashboardMetrics: (params: { dateRange?: string } = {}) =>
      fetchTraceMetrics(params.dateRange).then(toDashboardMetrics),

    getOverview: () =>
      fetchTraceMetrics().then(
        (metrics): AnalyticsOverview => ({
          totalTraces: metrics.totalCount,
          totalCost: metrics.totalCost,
          avgLatency: metrics.avgDuration,
          totalTokens: metrics.totalTokens,
          tracesChange: 0,
          costChange: 0,
          latencyChange: 0,
          tokensChange: 0,
          traceVolume: [],
          costByModel: [],
          latencyPercentiles: [],
        })
      ),

    getCostAnalytics: (params: AnalyticsGroupParams) =>
      fetchTraceMetrics(params.dateRange).then(
        (metrics): CostAnalytics => ({
          totalCost: metrics.totalCost,
          costChange: 0,
          inputCost: 0,
          outputCost: 0,
          avgCostPerTrace: metrics.totalCount > 0 ? metrics.totalCost / metrics.totalCount : 0,
          costOverTime: [],
          costByGroup: [],
          breakdown: [],
        })
      ),

    getLatencyAnalytics: (params: AnalyticsGroupParams) =>
      fetchTraceMetrics(params.dateRange).then(
        (metrics): LatencyAnalytics => ({
          p50: metrics.avgDuration,
          p95: metrics.avgDuration,
          p99: metrics.avgDuration,
          avg: metrics.avgDuration,
          avgChange: 0,
          totalRequests: metrics.totalCount,
          latencyOverTime: [],
          distribution: [],
          breakdown: [],
        })
      ),

    getUsageAnalytics: (params: { dateRange: string }) =>
      fetchTraceMetrics(params.dateRange).then(
        (metrics): UsageAnalytics => ({
          totalTraces: metrics.totalCount,
          tracesChange: 0,
          totalGenerations: 0,
          inputTokens: metrics.totalTokens,
          outputTokens: 0,
          volumeOverTime: [],
          tokenUsageOverTime: [],
          modelDistribution: [],
          topTraces: [],
        })
      ),

    getTraceVolume: (_params: { dateRange?: string } = {}) =>
      Promise.resolve<TraceVolumePoint[]>([]),

    getCostOverTime: (_params: { dateRange?: string } = {}) =>
      Promise.resolve<CostTrendPoint[]>([]),

    getLatencyPercentiles: (_params: { dateRange?: string } = {}) =>
      Promise.resolve<LatencyPercentilePoint[]>([]),

    getModelUsage: (_params: { dateRange?: string } = {}) =>
      Promise.resolve<ModelDistributionPoint[]>([]),

    getTopTracesByTokens: (params: { limit?: number } = {}) =>
      observabilityApi.traces.list({ limit: params.limit ?? 10, offset: 0 }).then((response) =>
        response.traces.map((trace) => ({
          id: trace.id,
          model: '',
          inputTokens: trace.usage?.promptTokens ?? 0,
          outputTokens: trace.usage?.completionTokens ?? 0,
          totalTokens: trace.usage?.totalTokens ?? 0,
          cost: trace.totalCost ?? 0,
        }))
      ),

    getTopTracesByCost: (params: { limit?: number } = {}) =>
      observabilityApi.traces.list({ limit: params.limit ?? 10, offset: 0 }).then((response) =>
        response.traces
          .map((trace) => ({
            id: trace.id,
            model: '',
            inputTokens: trace.usage?.promptTokens ?? 0,
            outputTokens: trace.usage?.completionTokens ?? 0,
            totalTokens: trace.usage?.totalTokens ?? 0,
            cost: trace.totalCost ?? 0,
          }))
          .sort((left, right) => right.cost - left.cost)
      ),

    getRecentErrors: (params: { limit?: number } = {}) =>
      observabilityApi.traces.list({ level: 'ERROR', limit: params.limit ?? 10, offset: 0 }).then((response) =>
        response.traces.map((trace) => ({
          id: trace.id,
          traceId: trace.id,
          message: trace.statusMessage ?? 'Trace error',
          timestamp: trace.startTime,
        }))
      ),

    getProjectMetrics: (projectId: string) =>
      fetchTraceMetrics(undefined, projectId).then(toDashboardMetrics),
  },
  // Checkpoints
  checkpoints: {
    list: (projectId: string, params?: CheckpointListParams) => {
      const searchParams = new URLSearchParams();
      if (params?.traceId) searchParams.set('traceId', params.traceId);
      if (params?.type) searchParams.set('type', params.type);
      if (params?.offset !== undefined) searchParams.set('offset', params.offset.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      return fetchWithAuth<{
        data: CheckpointResponse[];
        totalCount: number;
        hasMore: boolean;
      }>(`/api/public/checkpoints?${searchParams}`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }).then((response) => ({
        checkpoints: response.data.map(normalizeCheckpoint),
        totalCount: response.totalCount,
        hasMore: response.hasMore,
      }));
    },

    get: (projectId: string, checkpointId: string) =>
      fetchWithAuth<CheckpointResponse>(`/api/public/checkpoints/${checkpointId}`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }).then(normalizeCheckpoint),

    listByTrace: (projectId: string, traceId: string) =>
      fetchWithAuth<{ data: CheckpointResponse[] }>(`/api/public/traces/${traceId}/checkpoints`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }).then((response) => response.data.map(normalizeCheckpoint)),

    create: (projectId: string, data: CreateCheckpointInput) =>
      fetchWithAuth<CheckpointResponse>('/api/public/checkpoints', {
        method: 'POST',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify(data),
      }).then(normalizeCheckpoint),

    restore: (projectId: string, checkpointId: string, traceId: string) =>
      fetchWithAuth<CheckpointResponse>(`/api/public/checkpoints/${checkpointId}/restore`, {
        method: 'POST',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify({ traceId }),
      }).then(normalizeCheckpoint),
  },
  // Git Links
  gitLinks: {
    list: (projectId: string, params?: { traceId?: string; offset?: number; limit?: number }) => {
      const searchParams = new URLSearchParams();
      if (params?.traceId) searchParams.set('traceId', params.traceId);
      if (params?.offset !== undefined) searchParams.set('offset', params.offset.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      return fetchWithAuth<{
        data: GitLink[];
        totalCount: number;
        hasMore: boolean;
      }>(`/api/public/git-links?${searchParams}`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }).then((response) => ({
        gitLinks: response.data,
        totalCount: response.totalCount,
        hasMore: response.hasMore,
      }));
    },

    timeline: (projectId: string, branch?: string) =>
      fetchWithAuth<{ commits: GitTimelineEntry[] }>(
        `/api/public/git-links/timeline?${createSearchParams({ branch })}`,
        {
          method: 'GET',
          headers: { 'X-Project-ID': projectId },
        }
      ).then((response) => response.commits),

    create: (projectId: string, data: CreateGitLinkInput) =>
      fetchWithAuth<GitLink>('/api/public/git-links', {
        method: 'POST',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify(data),
      }),
  },
  // File Operations
  fileOperations: {
    list: (projectId: string, traceId: string) =>
      fetchWithAuth<{ data: FileOperation[] }>(`/api/public/traces/${traceId}/file-operations`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }).then((response) => response.data),

    create: (projectId: string, data: CreateFileOperationInput) =>
      fetchWithAuth<FileOperation>('/api/public/file-operations', {
        method: 'POST',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify(data),
      }),
  },
  // Terminal Commands
  terminalCommands: {
    list: (projectId: string, traceId: string) =>
      fetchWithAuth<{ data: TerminalCommand[] }>(
        `/api/public/traces/${traceId}/terminal-commands`,
        {
          method: 'GET',
          headers: { 'X-Project-ID': projectId },
        }
      ).then((response) => response.data),

    create: (projectId: string, data: CreateTerminalCommandInput) =>
      fetchWithAuth<TerminalCommand>('/api/public/terminal-commands', {
        method: 'POST',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify(data),
      }),
  },
  // CI Runs
  ciRuns: {
    list: (projectId: string, params?: { offset?: number; limit?: number }) => {
      const searchParams = new URLSearchParams();
      if (params?.offset !== undefined) searchParams.set('offset', params.offset.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      return fetchWithAuth<{
        data: CIRun[];
        totalCount: number;
        hasMore: boolean;
      }>(`/api/public/ci-runs?${searchParams}`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }).then((response) => ({
        ciRuns: response.data,
        totalCount: response.totalCount,
        hasMore: response.hasMore,
      }));
    },

    get: (projectId: string, runId: string) =>
      fetchWithAuth<CIRun>(`/api/public/ci-runs/${runId}`, {
        method: 'GET',
        headers: { 'X-Project-ID': projectId },
      }),

    create: (projectId: string, data: CreateCIRunInput) =>
      fetchWithAuth<CIRun>('/api/public/ci-runs', {
        method: 'POST',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify(data),
      }),

    update: (projectId: string, runId: string, data: UpdateCIRunInput) =>
      fetchWithAuth<CIRun>(`/api/public/ci-runs/${runId}`, {
        method: 'PATCH',
        headers: { 'X-Project-ID': projectId },
        body: JSON.stringify(data),
      }),
  },
  // Streaming
  streaming: {
    getActiveStreams: () => fetchWithAuth<{ streams: unknown[]; count: number }>('/api/public/streams'),
    getLiveMetrics: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/traces/${traceId}/live-metrics`),
    getActivities: (traceId: string, limit = 100) =>
      fetchWithAuth<{ activities: unknown[]; count: number }>(
        `/api/public/traces/${traceId}/activities?limit=${limit}`
      ),
    requestIntervention: (traceId: string, data: { action: string; message?: string }) =>
      fetchWithAuth<unknown>(`/api/public/traces/${traceId}/intervene`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Distributed Traces
  distributedTraces: {
    get: (traceId: string) => fetchWithAuth<unknown>(`/api/public/distributed/traces/${traceId}`),
    getServiceMap: () => fetchWithAuth<unknown>('/api/public/distributed/service-map'),
    correlate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/distributed/correlate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Replay Sessions
  replaySessions: {
    list: () => fetchWithAuth<unknown>('/api/public/replay-sessions'),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/replay-sessions', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}`),
    getTimeline: (id: string) => fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/timeline`),
    branch: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/branch`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getPlayback: (id: string) => fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/playback`),
    share: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/share`, { method: 'POST' }),
    recordEvents: (id: string, events: unknown[]) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/events`, {
        method: 'POST',
        body: JSON.stringify(events),
      }),
    control: (id: string, cmd: unknown) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/control`, {
        method: 'POST',
        body: JSON.stringify(cmd),
      }),
    getFileState: (id: string, eventIndex: number) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/files?eventIndex=${eventIndex}`),
    complete: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/complete`, { method: 'POST' }),
    getUnifiedTimeline: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/unified-timeline`),
    getSnapshot: (id: string, eventIndex: number) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/snapshot?eventIndex=${eventIndex}`),
    addAnnotation: (id: string, data: { eventId: string; content: string }) =>
      fetchWithAuth<unknown>(`/api/public/replay-sessions/${id}/annotations`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // IDE Trace View
  ideTraceView: {
    getFileMapping: (filePath: string) =>
      fetchWithAuth<unknown>(`/api/public/ide/file-mapping?filePath=${encodeURIComponent(filePath)}`),
    batchMappings: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/ide/batch-mappings', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getTraceContext: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/ide/trace-context/${traceId}`),
  },
  // Custom Metrics
  customMetrics: {
    list: () => fetchWithAuth<unknown>('/api/public/custom-metrics'),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/custom-metrics', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getValues: (id: string) => fetchWithAuth<unknown>(`/api/public/custom-metrics/${id}/values`),
    listDashboards: () => fetchWithAuth<unknown>('/api/public/custom-metrics/dashboards'),
    createDashboard: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/custom-metrics/dashboards', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listAlerts: () => fetchWithAuth<unknown>('/api/public/custom-metrics/alerts'),
    createAlert: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/custom-metrics/alerts', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // SLOs
  slos: {
    list: () => fetchWithAuth<unknown>('/api/public/slos'),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/slos', { method: 'POST', body: JSON.stringify(data) }),
    getStatus: (id: string) => fetchWithAuth<unknown>(`/api/public/slos/${id}/status`),
    getReport: () => fetchWithAuth<unknown>('/api/public/slos/report'),
    getHistory: (id: string) => fetchWithAuth<unknown>(`/api/public/slos/${id}/history`),
  },
  // Carbon
  carbon: {
    getFootprint: () => fetchWithAuth<unknown>('/api/public/carbon/footprint'),
    getConfig: () => fetchWithAuth<unknown>('/api/public/carbon/config'),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/carbon/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    getSuggestions: () => fetchWithAuth<unknown>('/api/public/carbon/suggestions'),
  },
  // Memory
  memory: {
    analyze: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/memory/analyze', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getSnapshot: (traceId: string, step: number) =>
      fetchWithAuth<unknown>(`/api/public/memory/traces/${traceId}/snapshots/${step}`),
    getOptimizations: () => fetchWithAuth<unknown>('/api/public/memory/optimizations'),
  },
  // Handoffs
  handoffs: {
    initiate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/handoffs', { method: 'POST', body: JSON.stringify(data) }),
    accept: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/handoffs/${id}/accept`, { method: 'POST' }),
    complete: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/handoffs/${id}/complete`, { method: 'POST' }),
    getChain: (traceId: string) => fetchWithAuth<unknown>(`/api/public/handoffs/chain/${traceId}`),
    getStats: () => fetchWithAuth<unknown>('/api/public/handoffs/stats'),
  },
  // Auto-Discovery
  autoDiscovery: {
    scan: () => fetchWithAuth<unknown>('/api/public/discovery/scan', { method: 'POST' }),
    getFramework: (id: string) => fetchWithAuth<unknown>(`/api/public/discovery/frameworks/${id}`),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/discovery/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    toggleInstrumentation: (id: string, enabled: boolean) =>
      fetchWithAuth<unknown>(`/api/public/discovery/frameworks/${id}/toggle`, {
        method: 'POST',
        body: JSON.stringify({ enabled }),
      }),
  },
};
