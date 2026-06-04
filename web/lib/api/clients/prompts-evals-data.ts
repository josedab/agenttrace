// promptsEvalsData API client.
import { ApiError, createSearchParams, fetchWithAuth, unsupportedApiFeature } from '../transport';
import type { QueryParams } from '../transport';
import {
  CompiledPromptResponse,
  DatasetItemResponse,
  DatasetResponse,
  DatasetRunResponse,
  EvaluatorResponse,
  EvaluatorTemplateResponse,
  PromptResponse,
  PromptVersionResponse,
  normalizeCompiledPrompt,
  normalizeDataset,
  normalizeDatasetItem,
  normalizeDatasetRun,
  normalizeEvaluator,
  normalizeEvaluatorTemplate,
  normalizePrompt,
  normalizePromptVersion,
} from '../normalizers';
import type {
  AnnotationQueue,
  AnnotationQueueItem,
  AnnotationScoreInput,
  CreateDatasetInput,
  CreateDatasetItemInput,
  CreateDatasetRunInput,
  CreateEvaluatorInput,
  CreateKnowledgeBaseEntryInput,
  CreatePromptInput,
  CreatePromptVersionInput,
  DataListResponse,
  Evaluator,
  EvaluatorRun,
  EvaluatorStats,
  KnowledgeBaseEntry,
  KnowledgeBaseSearchParams,
  PaginationParams,
  PromptListParams,
  RunPromptInput,
  UpdateDatasetInput,
  UpdateDatasetItemInput,
  UpdateEvaluatorInput,
  UpdatePromptInput,
} from '../contracts';

export const promptsEvalsDataApi = {
  // Prompts
  prompts: {
    list: (filters: PromptListParams = {}) =>
      fetchWithAuth<{
        prompts: PromptResponse[];
        totalCount: number;
        hasMore: boolean;
      }>(
        `/api/public/prompts?${createSearchParams({
          ...filters,
          name: filters.search,
          search: undefined,
        })}`
      ).then((response) => response.prompts.map(normalizePrompt)),

    get: (name: string, version?: number, label?: string) => {
      const params = new URLSearchParams();
      if (version) params.set('version', version.toString());
      if (label) params.set('label', label);
      return fetchWithAuth<PromptResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}?${params}`
      ).then(normalizePrompt);
    },

    getByName: (name: string) =>
      fetchWithAuth<PromptResponse>(`/api/public/prompts/${encodeURIComponent(name)}`).then(
        normalizePrompt
      ),

    listVersions: (name: string) =>
      fetchWithAuth<{ data: PromptVersionResponse[]; totalCount: number }>(
        `/api/public/prompts/${encodeURIComponent(name)}/versions`
      ).then((response) => response.data.map(normalizePromptVersion)),

    getVersion: (name: string, version: number) =>
      fetchWithAuth<PromptResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}?version=${version}`
      ).then((prompt) => {
        const promptVersion = prompt.latestVersion;
        if (!promptVersion) {
          throw new ApiError(404, `Prompt version ${version} not found`);
        }
        return normalizePromptVersion(promptVersion);
      }),

    create: (data: CreatePromptInput) =>
      fetchWithAuth<PromptResponse>('/api/public/prompts', {
        method: 'POST',
        body: JSON.stringify({
          ...data,
          type: data.type === 'CHAT' ? 'chat' : 'text',
          content: data.prompt ?? JSON.stringify(data.messages ?? []),
          prompt: undefined,
          messages: undefined,
        }),
      }).then(normalizePrompt),

    update: async (name: string, data: UpdatePromptInput) => {
      const existing = await fetchWithAuth<PromptResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}`
      );
      const response = await fetchWithAuth<PromptResponse>(`/api/public/prompts/${existing.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: data.name ?? existing.name,
          description: data.description ?? existing.description,
          type: data.type === undefined ? existing.type : data.type === 'CHAT' ? 'chat' : 'text',
          tags: data.tags ?? existing.tags,
        }),
      });
      return normalizePrompt(response);
    },

    createVersion: (name: string, data: CreatePromptVersionInput) =>
      fetchWithAuth<PromptVersionResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}/versions`,
        {
          method: 'POST',
          body: JSON.stringify({
            ...data,
            content: data.prompt ?? JSON.stringify(data.messages ?? []),
            prompt: undefined,
            messages: undefined,
          }),
        }
      ).then(normalizePromptVersion),

    setLabel: (name: string, labelOrVersion: string | number, versionOrLabel: number | string) => {
      const label = typeof labelOrVersion === 'string' ? labelOrVersion : versionOrLabel;
      const version = typeof labelOrVersion === 'number' ? labelOrVersion : versionOrLabel;

      return fetchWithAuth<{ message: string }>(
        `/api/public/prompts/${encodeURIComponent(name)}/labels`,
        {
          method: 'POST',
          body: JSON.stringify({
            label: String(label),
            version: Number(version),
          }),
        }
      );
    },

    removeLabel: async (name: string, label: string) => {
      const versions = await promptsEvalsDataApi.prompts.listVersions(name);
      const version = versions.find((candidate) => candidate.labels?.includes(label))?.version;
      if (version === undefined) {
        throw new ApiError(404, `Prompt label ${label} not found`);
      }
      return fetchWithAuth<void>(
        `/api/public/prompts/${encodeURIComponent(name)}/labels/${encodeURIComponent(label)}?version=${version}`,
        { method: 'DELETE' }
      );
    },

    delete: async (name: string) => {
      const prompt = await fetchWithAuth<PromptResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}`
      );
      return fetchWithAuth<void>(`/api/public/prompts/${prompt.id}`, {
        method: 'DELETE',
      });
    },

    compile: (name: string, version: number | undefined, variables: Record<string, string>) =>
      fetchWithAuth<CompiledPromptResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}/compile`,
        {
          method: 'POST',
          body: JSON.stringify({ version, variables }),
        }
      ).then(normalizeCompiledPrompt),

    run: (name: string, data: RunPromptInput) =>
      fetchWithAuth<CompiledPromptResponse>(
        `/api/public/prompts/${encodeURIComponent(name)}/compile`,
        {
          method: 'POST',
          body: JSON.stringify(data),
        }
      ).then((response) => ({ output: response.compiled })),
  },
  // Prompt Lab
  promptLab: {
    listExperiments: () => unsupportedApiFeature<never>('Prompt lab'),
    createExperiment: (_data: unknown) => unsupportedApiFeature<never>('Prompt lab'),
    getExperiment: (_id: string) => unsupportedApiFeature<never>('Prompt lab'),
    startExperiment: (_id: string) => unsupportedApiFeature<never>('Prompt lab'),
    completeExperiment: (_id: string) => unsupportedApiFeature<never>('Prompt lab'),
    getSuggestions: (_promptName?: string) => unsupportedApiFeature<never>('Prompt lab'),
  },
  // Prompt Cache
  promptCache: {
    analyze: () => fetchWithAuth<unknown>('/api/public/prompt-cache/analyze'),
    getConfig: () => fetchWithAuth<unknown>('/api/public/prompt-cache/config'),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-cache/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    getStats: () => fetchWithAuth<unknown>('/api/public/prompt-cache/stats'),
    invalidate: () => fetchWithAuth<unknown>('/api/public/prompt-cache/invalidate', { method: 'POST' }),
  },
  // Prompt CI
  promptCI: {
    listBaselines: () => fetchWithAuth<unknown>('/api/public/prompt-ci/baselines'),
    createBaseline: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-ci/baselines', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getBaseline: (id: string) => fetchWithAuth<unknown>(`/api/public/prompt-ci/baselines/${id}`),
    runComparison: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-ci/compare', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listRuns: () => fetchWithAuth<unknown>('/api/public/prompt-ci/runs'),
    createGateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-ci/gates', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listGateConfigs: () => fetchWithAuth<unknown>('/api/public/prompt-ci/gates'),
    updateGateConfig: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/prompt-ci/gates/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    evaluateGate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-ci/gates/evaluate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getRegressionHistory: () => fetchWithAuth<unknown>('/api/public/prompt-ci/history'),
    getDashboardStats: () => fetchWithAuth<unknown>('/api/public/prompt-ci/stats'),
    triggerCIWebhook: (_data: unknown) => unsupportedApiFeature<never>('Prompt CI webhooks'),
    generateCIConfig: (_provider: string) =>
      unsupportedApiFeature<never>('Prompt CI configuration generation'),
  },
  // Prompt Optimization
  promptOptimization: {
    start: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-optimization', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/prompt-optimization/${id}`),
    list: () => fetchWithAuth<unknown>('/api/public/prompt-optimization'),
    getConfig: () => fetchWithAuth<unknown>('/api/public/prompt-optimization/config'),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/prompt-optimization/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    approveVariant: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/prompt-optimization/variants/${id}/approve`, {
        method: 'POST',
      }),
    rejectVariant: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/prompt-optimization/variants/${id}/reject`, {
        method: 'POST',
      }),
  },
  // Datasets
  datasets: {
    list: () =>
      fetchWithAuth<{
        datasets: DatasetResponse[];
        totalCount: number;
        hasMore: boolean;
      }>('/api/public/datasets').then((response) => response.datasets.map(normalizeDataset)),

    get: (id: string) =>
      fetchWithAuth<DatasetResponse>(`/api/public/datasets/${id}`).then(normalizeDataset),

    create: (data: CreateDatasetInput) =>
      fetchWithAuth<DatasetResponse>('/api/public/datasets', {
        method: 'POST',
        body: JSON.stringify(data),
      }).then(normalizeDataset),

    update: (id: string, data: UpdateDatasetInput) =>
      fetchWithAuth<DatasetResponse>(`/api/public/datasets/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }).then(normalizeDataset),

    delete: (id: string) =>
      fetchWithAuth<void>(`/api/public/datasets/${id}`, {
        method: 'DELETE',
      }),

    listItems: (datasetId: string, params: PaginationParams = {}) =>
      fetchWithAuth<DataListResponse<DatasetItemResponse>>(
        `/api/public/datasets/${datasetId}/items?${createSearchParams({
          limit: params.limit,
          offset: params.offset ?? (params.cursor ? Number(params.cursor) : undefined),
        })}`
      ).then((response) => ({
        ...response,
        data: response.data.map(normalizeDatasetItem),
      })),

    getItems: (datasetId: string) =>
      fetchWithAuth<DataListResponse<DatasetItemResponse>>(
        `/api/public/datasets/${datasetId}/items`
      ).then((response) => response.data.map(normalizeDatasetItem)),

    addItem: (datasetId: string, data: CreateDatasetItemInput) =>
      fetchWithAuth<DatasetItemResponse>(`/api/public/datasets/${datasetId}/items`, {
        method: 'POST',
        body: JSON.stringify(data),
      }).then(normalizeDatasetItem),

    updateItem: (datasetId: string, itemId: string, data: UpdateDatasetItemInput) =>
      fetchWithAuth<DatasetItemResponse>(`/api/public/datasets/${datasetId}/items/${itemId}`, {
        method: 'PUT',
        body: JSON.stringify({
          ...data,
          status: data.status?.toLowerCase(),
        }),
      }).then(normalizeDatasetItem),

    deleteItem: (datasetId: string, itemId: string) =>
      fetchWithAuth<void>(`/api/public/datasets/${datasetId}/items/${itemId}`, {
        method: 'DELETE',
      }),

    listRuns: (datasetId: string) =>
      fetchWithAuth<DataListResponse<DatasetRunResponse>>(
        `/api/public/datasets/${datasetId}/runs`
      ).then((response) => response.data.map(normalizeDatasetRun)),

    getRuns: (datasetId: string) =>
      fetchWithAuth<DataListResponse<DatasetRunResponse>>(
        `/api/public/datasets/${datasetId}/runs`
      ).then((response) => response.data.map(normalizeDatasetRun)),

    getRun: (datasetId: string, runId: string) =>
      fetchWithAuth<DatasetRunResponse>(`/api/public/datasets/${datasetId}/runs/${runId}`).then(
        normalizeDatasetRun
      ),

    createRun: (datasetId: string, data: CreateDatasetRunInput) =>
      fetchWithAuth<DatasetRunResponse>(`/api/public/datasets/${datasetId}/runs`, {
        method: 'POST',
        body: JSON.stringify(data),
      }).then(normalizeDatasetRun),
  },
  // Synthetic Data
  syntheticData: {
    generate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/synthetic-data/generate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    list: () => fetchWithAuth<unknown>('/api/public/synthetic-data/datasets'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/synthetic-data/datasets/${id}`),
    getStats: () => fetchWithAuth<unknown>('/api/public/synthetic-data/stats'),
  },
  // Evaluators
  evaluators: {
    list: () =>
      fetchWithAuth<{
        evaluators: EvaluatorResponse[];
        totalCount: number;
        hasMore: boolean;
      }>('/api/public/evaluators').then((response) => response.evaluators.map(normalizeEvaluator)),

    get: (id: string) =>
      fetchWithAuth<EvaluatorResponse>(`/api/public/evaluators/${id}`).then(normalizeEvaluator),

    getStats: (_id: string) => unsupportedApiFeature<EvaluatorStats>('Evaluator statistics'),

    listRuns: (_id: string) => unsupportedApiFeature<EvaluatorRun[]>('Evaluator run history'),

    create: (data: CreateEvaluatorInput) => {
      if (data.type === 'HUMAN') {
        return unsupportedApiFeature<Evaluator>('Human evaluators');
      }
      return fetchWithAuth<EvaluatorResponse>('/api/public/evaluators', {
        method: 'POST',
        body: JSON.stringify({
          ...data,
          type: data.type === 'LLM_AS_JUDGE' ? 'llm_as_judge' : 'custom',
          templateId: data.templateId ?? data.template,
          template: undefined,
        }),
      }).then(normalizeEvaluator);
    },

    update: (id: string, data: UpdateEvaluatorInput) =>
      fetchWithAuth<EvaluatorResponse>(`/api/public/evaluators/${id}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: data.name,
          description: data.description,
          promptTemplate: data.promptTemplate,
          variables: data.variables,
          scoreCategories: data.scoreCategories,
          config: data.config,
          targetFilter: data.targetFilter,
          samplingRate: data.samplingRate,
          enabled: data.enabled ?? (data.status ? data.status === 'ACTIVE' : undefined),
        }),
      }).then(normalizeEvaluator),

    delete: (id: string) =>
      fetchWithAuth<void>(`/api/public/evaluators/${id}`, {
        method: 'DELETE',
      }),

    run: (_id: string) => unsupportedApiFeature<EvaluatorRun>('Evaluator execution'),

    templates: () =>
      fetchWithAuth<{ data: EvaluatorTemplateResponse[] }>('/api/public/evaluator-templates').then(
        (response) => response.data.map(normalizeEvaluatorTemplate)
      ),
  },
  annotationQueues: {
    list: () => unsupportedApiFeature<AnnotationQueue[]>('Annotation queues'),

    get: (_id: string) => unsupportedApiFeature<AnnotationQueue>('Annotation queues'),

    getItems: (_id: string) => unsupportedApiFeature<AnnotationQueueItem[]>('Annotation queues'),

    submitScore: (_queueId: string, _itemId: string, _data: AnnotationScoreInput) =>
      unsupportedApiFeature<AnnotationQueueItem>('Annotation queues'),

    skipItem: (_queueId: string, _itemId: string) =>
      unsupportedApiFeature<void>('Annotation queues'),
  },
  // Annotations
  annotations: {
    list: (traceId: string) => fetchWithAuth<unknown>(`/api/public/annotations/traces/${traceId}`),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/annotations', { method: 'POST', body: JSON.stringify(data) }),
    reply: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/annotations/${id}/reply`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    resolve: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/annotations/${id}/resolve`, { method: 'POST' }),
    getPresence: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/annotations/presence/${traceId}`),
  },
  // A/B Testing (v12)
  abTests: {
    list: () =>
      fetchWithAuth<{ tests: unknown[] }>('/api/public/ab-tests').then(
        (response) => response.tests
      ),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/ab-tests/${id}`),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/ab-tests', { method: 'POST', body: JSON.stringify(data) }),
    start: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${id}/start`, { method: 'POST' }),
    pause: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${id}/pause`, { method: 'POST' }),
    stop: (id: string) => fetchWithAuth<unknown>(`/api/public/ab-tests/${id}/stop`, { method: 'POST' }),
    assignVariant: (data: { testId: string; userId: string }) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${data.testId}/assign`, {
        method: 'POST',
        body: JSON.stringify({ assignmentKey: data.userId }),
      }),
    recordResult: (data: {
      testId: string;
      variantId: string;
      metrics: Record<string, number>;
      result?: Record<string, unknown>;
    }) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${data.testId}/results`, {
        method: 'POST',
        body: JSON.stringify({
          variantId: data.variantId,
          metrics: data.metrics,
          result: data.result,
        }),
      }),
    getStatistics: (testId: string) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${testId}/statistics`),
    selectWinner: (testId: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${testId}/select-winner`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    startRollout: (testId: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/ab-tests/${testId}/rollout`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Benchmarks
  benchmarks: {
    list: () => fetchWithAuth<unknown>('/api/public/benchmarks'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/benchmarks/${id}`),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/benchmarks', { method: 'POST', body: JSON.stringify(data) }),
    submit: (benchmarkId: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/benchmarks/${benchmarkId}/submit`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getLeaderboard: (benchmarkId: string) =>
      fetchWithAuth<unknown>(`/api/public/benchmarks/${benchmarkId}/leaderboard`),
    compare: (benchmarkId: string, data: { submissionIdA: string; submissionIdB: string }) =>
      fetchWithAuth<unknown>(`/api/public/benchmarks/${benchmarkId}/compare`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getStats: (benchmarkId: string) =>
      fetchWithAuth<unknown>(`/api/public/benchmarks/${benchmarkId}/stats`),
  },
  // Agent Benchmarks
  agentBenchmarks: {
    listSuites: () => fetchWithAuth<unknown>('/api/public/agent-benchmarks/suites'),
    createSuite: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/agent-benchmarks/suites', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getSuite: (id: string) => fetchWithAuth<unknown>(`/api/public/agent-benchmarks/suites/${id}`),
    runBenchmark: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/agent-benchmarks/run', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getLeaderboard: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/agent-benchmarks/suites/${id}/leaderboard`),
  },
  // Regression Suite
  regressionSuite: {
    createDataset: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/regression/golden-datasets', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getDataset: (id: string) => fetchWithAuth<unknown>(`/api/public/regression/golden-datasets/${id}`),
    listDatasets: () => fetchWithAuth<unknown>('/api/public/regression/golden-datasets'),
    runRegression: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/regression/run', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getRun: (id: string) => fetchWithAuth<unknown>(`/api/public/regression/runs/${id}`),
    listRuns: () => fetchWithAuth<unknown>('/api/public/regression/runs'),
  },
  // Semantic Search
  semanticSearch: {
    search: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/semantic-search', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getClusters: () => fetchWithAuth<unknown>('/api/public/semantic-search/clusters'),
    getAnomalyPatterns: () => fetchWithAuth<unknown>('/api/public/semantic-search/anomaly-patterns'),
    getDashboard: () => fetchWithAuth<unknown>('/api/public/semantic-search/dashboard'),
  },
  // Semantic Search
  search: {
    query: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/search', { method: 'POST', body: JSON.stringify(data) }),
    suggestions: (prefix: string) =>
      fetchWithAuth<unknown>(`/api/public/search/suggestions?prefix=${prefix}`),
  },
  knowledgeBase: {
    search: (params: KnowledgeBaseSearchParams) =>
      fetchWithAuth<KnowledgeBaseEntry[]>(
        `/api/public/knowledge-base/search?${createSearchParams(params)}`
      ),
    create: (data: CreateKnowledgeBaseEntryInput) =>
      fetchWithAuth<KnowledgeBaseEntry>('/api/public/knowledge-base/entries', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Knowledge Graph
  knowledgeGraph: {
    build: () => fetchWithAuth<unknown>('/api/public/knowledge-graph'),
    query: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/knowledge-graph/query', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getStats: () => fetchWithAuth<unknown>('/api/public/knowledge-graph/stats'),
  },
  // Agent Knowledge Graph
  agentKnowledgeGraph: {
    build: (params?: QueryParams) =>
      fetchWithAuth<unknown>(`/api/public/agent-knowledge-graph?${createSearchParams(params ?? {})}`),
    query: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/agent-knowledge-graph/query', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getEvolution: () => fetchWithAuth<unknown>('/api/public/agent-knowledge-graph/evolution'),
    getStats: () => fetchWithAuth<unknown>('/api/public/agent-knowledge-graph/stats'),
  },
  // Training Pipeline
  training: {
    listDatasets: () => fetchWithAuth<unknown>('/api/public/training/datasets'),
    createDataset: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/training/datasets', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    exportDataset: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/training/datasets/${id}/export`, { method: 'POST' }),
    detectFailures: () => fetchWithAuth<unknown>('/api/public/training/failure-patterns'),
  },
  embed: {
    getConfig: () => fetchWithAuth<unknown>('/api/public/embed/config'),
    createConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/embed/config', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/embed/config', { method: 'PUT', body: JSON.stringify(data) }),
    generateToken: () => fetchWithAuth<unknown>('/api/public/embed/token', { method: 'POST' }),
  },
  // Multi-Modal Traces
  multimodal: {
    register: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/multimodal/attachments', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getTraceAttachments: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/multimodal/traces/${traceId}`),
    getAttachment: (id: string) => fetchWithAuth<unknown>(`/api/public/multimodal/attachments/${id}`),
    getSummary: (traceId: string) =>
      fetchWithAuth<unknown>(`/api/public/multimodal/traces/${traceId}/summary`),
    list: () => fetchWithAuth<unknown>('/api/public/multimodal/attachments'),
  },
  // Natural Language Query
  nlQuery: {
    query: (data: { query: string; limit?: number }) =>
      fetchWithAuth<unknown>('/api/public/nl-query', { method: 'POST', body: JSON.stringify(data) }),
    getExamples: () => fetchWithAuth<unknown>('/api/public/nl-query/examples'),
    autocomplete: (partial: string) =>
      fetchWithAuth<unknown>('/api/public/nl-query/autocomplete', {
        method: 'POST',
        body: JSON.stringify({ partial }),
      }),
    getSchema: () => fetchWithAuth<unknown>('/api/public/nl-query/schema'),
  },
  // Marketplace
  marketplace: {
    search: (params?: QueryParams) =>
      fetchWithAuth<unknown>(
        `/api/public/marketplace${params ? `?${createSearchParams(params)}` : ''}`
      ),
    featured: () => fetchWithAuth<unknown>('/api/public/marketplace/featured'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/marketplace/${id}`),
    publish: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/marketplace', { method: 'POST', body: JSON.stringify(data) }),
    install: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/marketplace/${id}/install`, { method: 'POST' }),
    rate: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/marketplace/${id}/rate`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
};
