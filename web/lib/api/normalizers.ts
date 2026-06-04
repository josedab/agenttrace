// Response DTOs and normalizers that map API payloads onto public contracts.
import { createSearchParams, fetchWithAuth, parseJsonRecord, parseJsonValue } from './transport';
import type {
  APIKey,
  Checkpoint,
  CheckpointFile,
  CompiledPrompt,
  CreateScoreInput,
  DashboardMetrics,
  Dataset,
  DatasetItem,
  DatasetRun,
  DatasetRunResult,
  Evaluator,
  EvaluatorTemplate,
  Observation,
  Prompt,
  PromptVersion,
  Score,
  Session,
  Trace,
  TraceStats,
  UpdateScoreInput,
} from './contracts';

export function dateRangeParams(dateRange?: string): { from?: string; to?: string } {
  if (!dateRange) return {};

  const match = /^(\d+)([hd])$/.exec(dateRange);
  if (!match) return {};

  const amount = Number(match[1]);
  const milliseconds = amount * (match[2] === 'h' ? 60 * 60 * 1000 : 24 * 60 * 60 * 1000);
  const to = new Date();
  const from = new Date(to.getTime() - milliseconds);
  return { from: from.toISOString(), to: to.toISOString() };
}

export interface TraceMetricsResponse {
  totalCount: number;
  avgDuration: number;
  totalCost: number;
  totalTokens: number;
  errorCount: number;
  errorRate: number;
}

export function toTraceStats(metrics: TraceMetricsResponse): TraceStats {
  return {
    totalTraces: metrics.totalCount,
    averageLatency: metrics.avgDuration,
    totalCost: metrics.totalCost,
    totalTokens: metrics.totalTokens,
    errorCount: metrics.errorCount,
    errorRate: metrics.errorRate,
  };
}

export function toDashboardMetrics(metrics: TraceMetricsResponse): DashboardMetrics {
  return {
    totalTraces: metrics.totalCount,
    totalCost: metrics.totalCost,
    avgLatency: metrics.avgDuration,
    activeSessions: 0,
    totalTokens: metrics.totalTokens,
    tracesChange: 0,
    costChange: 0,
    latencyChange: 0,
    sessionsChange: 0,
    tokensChange: 0,
    traceVolume: [],
    costBreakdown: [],
    latencyPercentiles: [],
    recentTraces: [],
  };
}

export function fetchTraceMetrics(dateRange?: string, projectId?: string) {
  const endpoint = projectId
    ? `/api/v1/projects/${projectId}/metrics`
    : `/api/public/metrics/project?${createSearchParams(dateRangeParams(dateRange))}`;
  return fetchWithAuth<TraceMetricsResponse>(endpoint);
}

export interface ScoreResponse {
  id: string;
  projectId: string;
  traceId: string;
  observationId?: string;
  name: string;
  source: Score['source'];
  dataType: Score['dataType'];
  value?: number;
  stringValue?: string;
  comment?: string;
  configId?: string;
  authorUserId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ScoreDataResponse {
  data: ScoreResponse[];
}

export function normalizeScore(score: ScoreResponse): Score {
  const value =
    score.dataType === 'BOOLEAN'
      ? score.value === undefined
        ? null
        : score.value !== 0
      : score.dataType === 'CATEGORICAL'
        ? (score.stringValue ?? null)
        : (score.value ?? null);

  return {
    id: score.id,
    projectId: score.projectId,
    traceId: score.traceId,
    observationId: score.observationId ?? null,
    name: score.name,
    source: score.source,
    dataType: score.dataType,
    value,
    stringValue: score.stringValue ?? null,
    comment: score.comment ?? null,
    configId: score.configId ?? null,
    authorUserId: score.authorUserId ?? null,
    createdAt: score.createdAt,
    updatedAt: score.updatedAt,
  };
}

export interface TraceResponse {
  id: string;
  projectId?: string;
  name: string;
  userId?: string;
  sessionId?: string;
  release?: string;
  version?: string;
  tags?: string[];
  metadata?: unknown;
  public?: boolean;
  startTime: string;
  endTime?: string;
  durationMs?: number;
  input?: unknown;
  output?: unknown;
  level: Trace['level'];
  statusMessage?: string;
  totalCost?: number;
  totalTokens?: number;
  inputTokens?: number;
  outputTokens?: number;
  observations?: ObservationResponse[];
  scores?: ScoreResponse[];
}

export interface ObservationResponse {
  id: string;
  traceId: string;
  projectId?: string;
  parentObservationId?: string;
  name?: string;
  type: Observation['type'];
  startTime: string;
  endTime?: string;
  durationMs?: number;
  input?: unknown;
  output?: unknown;
  metadata?: unknown;
  level: Observation['level'];
  statusMessage?: string;
  version?: string;
  model?: string;
  modelParameters?: unknown;
  usageDetails?: {
    inputTokens?: number;
    outputTokens?: number;
    totalTokens?: number;
  };
  costDetails?: {
    inputCost?: number;
    outputCost?: number;
    totalCost?: number;
  };
  children?: ObservationResponse[];
  scores?: ScoreResponse[];
}

export interface ObservationTreeResponse {
  observation: ObservationResponse;
  children?: ObservationTreeResponse[];
}

export interface SessionResponse {
  id: string;
  projectId: string;
  userId?: string;
  bookmarked: boolean;
  public: boolean;
  createdAt: string;
  updatedAt: string;
  traceCount: number;
  totalCost: number;
  totalTokens: number;
  firstTraceTime: string;
  lastTraceTime: string;
  traces?: TraceResponse[];
}

export function normalizeObservation(observation: ObservationResponse): Observation {
  const inputTokens = observation.usageDetails?.inputTokens ?? 0;
  const outputTokens = observation.usageDetails?.outputTokens ?? 0;
  const totalTokens = observation.usageDetails?.totalTokens ?? inputTokens + outputTokens;
  const inputCost = observation.costDetails?.inputCost ?? 0;
  const outputCost = observation.costDetails?.outputCost ?? 0;
  const totalCost = observation.costDetails?.totalCost ?? inputCost + outputCost;

  return {
    id: observation.id,
    traceId: observation.traceId,
    projectId: observation.projectId,
    parentObservationId: observation.parentObservationId ?? null,
    name: observation.name ?? null,
    type: observation.type,
    startTime: observation.startTime,
    endTime: observation.endTime ?? null,
    latency: observation.durationMs ?? null,
    input: parseJsonValue(observation.input),
    output: parseJsonValue(observation.output),
    metadata: parseJsonRecord(observation.metadata),
    level: observation.level,
    statusMessage: observation.statusMessage ?? null,
    version: observation.version ?? null,
    model: observation.model ?? null,
    modelParameters: parseJsonRecord(observation.modelParameters),
    promptTokens: inputTokens,
    completionTokens: outputTokens,
    totalTokens,
    totalCost,
    inputCost,
    outputCost,
    usage: { promptTokens: inputTokens, completionTokens: outputTokens, totalTokens },
    cost: { input: inputCost, output: outputCost, total: totalCost },
    children: observation.children?.map(normalizeObservation),
    scores: observation.scores?.map(normalizeScore),
  };
}

export function flattenObservationTree(tree: ObservationTreeResponse): Observation[] {
  const observation = normalizeObservation(tree.observation);
  observation.children = undefined;
  return [observation, ...(tree.children ?? []).flatMap(flattenObservationTree)];
}

export function normalizeSession(session: SessionResponse): Session {
  return {
    id: session.id,
    projectId: session.projectId,
    userId: session.userId ?? null,
    bookmarked: session.bookmarked,
    public: session.public,
    createdAt: session.createdAt,
    updatedAt: session.updatedAt,
    traceCount: session.traceCount,
    totalCost: session.totalCost,
    totalTokens: session.totalTokens,
    firstTraceTime: session.firstTraceTime,
    lastTraceTime: session.lastTraceTime,
    traces: session.traces?.map(normalizeTrace),
  };
}

export function normalizeTrace(trace: TraceResponse): Trace {
  const inputTokens = trace.inputTokens ?? 0;
  const outputTokens = trace.outputTokens ?? 0;
  const totalTokens = trace.totalTokens ?? inputTokens + outputTokens;

  return {
    id: trace.id,
    projectId: trace.projectId,
    name: trace.name,
    startTime: trace.startTime,
    endTime: trace.endTime ?? null,
    input: parseJsonValue(trace.input),
    output: parseJsonValue(trace.output),
    metadata: parseJsonRecord(trace.metadata),
    tags: trace.tags ?? [],
    userId: trace.userId ?? null,
    sessionId: trace.sessionId ?? null,
    release: trace.release ?? null,
    version: trace.version ?? null,
    level: trace.level,
    statusMessage: trace.statusMessage ?? null,
    public: trace.public ?? false,
    latency: trace.durationMs ?? null,
    totalCost: trace.totalCost ?? null,
    usage: { promptTokens: inputTokens, completionTokens: outputTokens, totalTokens },
    observations: trace.observations?.map(normalizeObservation),
    scores: trace.scores?.map(normalizeScore),
  };
}

export interface PromptVersionResponse {
  id: string;
  promptId?: string;
  version: number;
  content: string;
  config?: unknown;
  labels?: string[];
  variables?: string[];
  createdAt: string;
  commitMessage?: string;
}

export interface PromptResponse {
  id: string;
  projectId?: string;
  name: string;
  type: 'text' | 'chat';
  description?: string;
  tags?: string[];
  latestVersion?: PromptVersionResponse;
  versions?: PromptVersionResponse[];
  createdAt: string;
  updatedAt: string;
}

export function normalizePromptVersion(version: PromptVersionResponse): PromptVersion {
  return {
    id: version.id,
    promptId: version.promptId,
    version: version.version,
    prompt: version.content,
    messages: null,
    config: parseJsonRecord(version.config),
    labels: version.labels ?? [],
    variables: version.variables ?? [],
    createdAt: version.createdAt,
    commitMessage: version.commitMessage,
  };
}

export function normalizePrompt(prompt: PromptResponse): Prompt {
  const rawVersions = prompt.versions?.length
    ? prompt.versions
    : prompt.latestVersion
      ? [prompt.latestVersion]
      : [];
  const versions = rawVersions.map(normalizePromptVersion);
  const labels: Record<string, number> = {};
  for (const version of versions) {
    for (const label of version.labels ?? []) {
      labels[label] = version.version;
    }
  }

  return {
    id: prompt.id,
    projectId: prompt.projectId,
    name: prompt.name,
    description: prompt.description ?? null,
    type: prompt.type === 'chat' ? 'CHAT' : 'TEXT',
    tags: prompt.tags ?? [],
    activeVersion: prompt.latestVersion
      ? normalizePromptVersion(prompt.latestVersion)
      : (versions.at(-1) ?? null),
    versions,
    labels,
    versionCount: versions.length,
    createdAt: prompt.createdAt,
    updatedAt: prompt.updatedAt,
  };
}

export interface CompiledPromptResponse {
  prompt: PromptResponse;
  version: number;
  compiled: string;
  variables: Record<string, string>;
}

export function normalizeCompiledPrompt(response: CompiledPromptResponse): CompiledPrompt {
  return {
    prompt: normalizePrompt(response.prompt),
    version: response.version,
    compiled: response.compiled,
    variables: response.variables,
  };
}

export interface DatasetResponse {
  id: string;
  projectId?: string;
  name: string;
  description?: string;
  metadata?: unknown;
  itemCount?: number;
  runCount?: number;
  createdAt: string;
  updatedAt: string;
}

export interface DatasetItemResponse {
  id: string;
  datasetId: string;
  input: unknown;
  expectedOutput?: unknown;
  metadata?: unknown;
  sourceTraceId?: string;
  sourceObservationId?: string;
  status: 'active' | 'archived';
  createdAt: string;
  updatedAt: string;
}

export interface DatasetRunItemResponse {
  id: string;
  datasetItemId: string;
  traceId: string;
  datasetItem?: DatasetItemResponse;
  trace?: TraceResponse;
  scores?: ScoreResponse[];
}

export interface DatasetRunResponse {
  id: string;
  datasetId: string;
  name: string;
  description?: string;
  metadata?: unknown;
  itemCount?: number;
  completedCount?: number;
  items?: DatasetRunItemResponse[];
  createdAt: string;
  updatedAt: string;
}

export function normalizeDataset(dataset: DatasetResponse): Dataset {
  return {
    id: dataset.id,
    projectId: dataset.projectId,
    name: dataset.name,
    description: dataset.description || null,
    metadata: parseJsonRecord(dataset.metadata),
    itemCount: dataset.itemCount ?? 0,
    runCount: dataset.runCount ?? 0,
    createdAt: dataset.createdAt,
    updatedAt: dataset.updatedAt,
  };
}

export function normalizeDatasetItem(item: DatasetItemResponse): DatasetItem {
  return {
    id: item.id,
    datasetId: item.datasetId,
    input: parseJsonValue(item.input),
    expectedOutput: item.expectedOutput === undefined ? null : parseJsonValue(item.expectedOutput),
    metadata: parseJsonRecord(item.metadata),
    sourceTraceId: item.sourceTraceId ?? null,
    sourceObservationId: item.sourceObservationId ?? null,
    status: item.status === 'archived' ? 'ARCHIVED' : 'ACTIVE',
    createdAt: item.createdAt,
    updatedAt: item.updatedAt,
  };
}

export function normalizeDatasetRun(run: DatasetRunResponse): DatasetRun {
  const results = (run.items ?? []).map((item): DatasetRunResult => {
    const numericScores = (item.scores ?? [])
      .map(normalizeScore)
      .map((score) => score.value)
      .filter((value): value is number => typeof value === 'number');
    const failed = item.trace?.level === 'ERROR';

    return {
      id: item.id,
      input: item.datasetItem ? parseJsonValue(item.datasetItem.input) : null,
      output: item.trace ? parseJsonValue(item.trace.output) : null,
      expectedOutput:
        item.datasetItem?.expectedOutput !== undefined
          ? parseJsonValue(item.datasetItem.expectedOutput)
          : null,
      score:
        numericScores.length > 0
          ? numericScores.reduce((total, score) => total + score, 0) / numericScores.length
          : null,
      status: failed ? 'FAILED' : item.trace ? 'COMPLETED' : 'PENDING',
    };
  });
  const scores = results
    .map((result) => result.score)
    .filter((score): score is number => score !== null);
  const itemCount = run.itemCount ?? results.length;
  const completedCount =
    run.completedCount ??
    results.filter((result) => result.status === 'COMPLETED' || result.status === 'FAILED').length;
  const failedCount = results.filter((result) => result.status === 'FAILED').length;
  const status: DatasetRun['status'] =
    itemCount === 0
      ? 'PENDING'
      : completedCount >= itemCount
        ? failedCount > 0
          ? 'FAILED'
          : 'COMPLETED'
        : completedCount > 0
          ? 'RUNNING'
          : 'PENDING';

  return {
    id: run.id,
    datasetId: run.datasetId,
    name: run.name,
    description: run.description || null,
    metadata: parseJsonRecord(run.metadata),
    status,
    itemCount,
    totalCount: itemCount,
    completedCount,
    failedCount,
    avgScore:
      scores.length > 0 ? scores.reduce((total, score) => total + score, 0) / scores.length : null,
    totalCost: run.items?.reduce((total, item) => total + (item.trace?.totalCost ?? 0), 0) ?? null,
    scores,
    results,
    createdAt: run.createdAt,
    updatedAt: run.updatedAt,
  };
}

export interface EvaluatorResponse {
  id: string;
  projectId?: string;
  name: string;
  description?: string;
  type: 'llm' | 'llm_as_judge' | 'rule' | 'custom';
  config?: unknown;
  promptTemplate?: string;
  variables?: string[];
  targetFilter?: unknown;
  samplingRate?: number;
  scoreName: string;
  scoreDataType: Evaluator['scoreDataType'];
  scoreCategories?: string[];
  enabled: boolean;
  evalCount?: number;
  avgScore?: number;
  lastEvalTime?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EvaluatorTemplateResponse {
  id: string;
  name: string;
  description?: string;
  promptTemplate: string;
  variables?: string[];
  scoreDataType: Evaluator['scoreDataType'];
  scoreCategories?: string[];
}

export function normalizeEvaluator(evaluator: EvaluatorResponse): Evaluator {
  return {
    id: evaluator.id,
    projectId: evaluator.projectId,
    name: evaluator.name,
    description: evaluator.description,
    type: evaluator.type === 'llm' || evaluator.type === 'llm_as_judge' ? 'LLM_AS_JUDGE' : 'CODE',
    status: evaluator.enabled ? 'ACTIVE' : 'INACTIVE',
    scoreName: evaluator.scoreName,
    scoreDataType: evaluator.scoreDataType,
    scoreCategories: evaluator.scoreCategories?.length ? evaluator.scoreCategories : null,
    promptTemplate: evaluator.promptTemplate || null,
    variables: evaluator.variables ?? [],
    config: parseJsonRecord(evaluator.config) ?? {},
    targetFilter: parseJsonRecord(evaluator.targetFilter),
    samplingRate: evaluator.samplingRate ?? 1,
    enabled: evaluator.enabled,
    evalCount: evaluator.evalCount ?? 0,
    totalRuns: evaluator.evalCount ?? 0,
    avgScore: evaluator.avgScore ?? null,
    lastRunAt: evaluator.lastEvalTime,
    createdAt: evaluator.createdAt,
    updatedAt: evaluator.updatedAt,
  };
}

export function normalizeEvaluatorTemplate(template: EvaluatorTemplateResponse): EvaluatorTemplate {
  return {
    id: template.id,
    name: template.name,
    description: template.description ?? '',
    type: 'LLM_AS_JUDGE',
    promptTemplate: template.promptTemplate,
    variables: template.variables ?? [],
    scoreDataType: template.scoreDataType,
    scoreCategories: template.scoreCategories?.length ? template.scoreCategories : null,
  };
}

export interface CheckpointResponse {
  id: string;
  projectId: string;
  traceId: string;
  observationId?: string;
  name: string;
  description?: string;
  type: Checkpoint['type'];
  gitCommitSha?: string;
  gitBranch?: string;
  gitRepoUrl?: string;
  filesSnapshot?: unknown;
  filesChanged?: string[];
  storagePath?: string;
  totalFiles?: number;
  totalSizeBytes?: number;
  restoredFrom?: string;
  restoredAt?: string;
  createdAt: string;
}

export function checkpointFiles(snapshot: unknown): CheckpointFile[] {
  const parsed = parseJsonValue(snapshot);
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    return [];
  }

  return Object.entries(parsed).map(([path, value]) => {
    const details =
      value && typeof value === 'object' && !Array.isArray(value)
        ? Object.fromEntries(Object.entries(value))
        : {};
    return {
      path,
      content: '',
      hash: typeof details.hash === 'string' ? details.hash : '',
      size: typeof details.size === 'number' ? details.size : 0,
    };
  });
}

export function normalizeCheckpoint(checkpoint: CheckpointResponse): Checkpoint {
  const files = checkpointFiles(checkpoint.filesSnapshot);
  return {
    id: checkpoint.id,
    projectId: checkpoint.projectId,
    traceId: checkpoint.traceId,
    observationId: checkpoint.observationId,
    name: checkpoint.name,
    type: checkpoint.type,
    description: checkpoint.description || null,
    gitCommitSha: checkpoint.gitCommitSha,
    gitBranch: checkpoint.gitBranch,
    gitRepoUrl: checkpoint.gitRepoUrl,
    filesSnapshot:
      typeof checkpoint.filesSnapshot === 'string'
        ? checkpoint.filesSnapshot
        : JSON.stringify(checkpoint.filesSnapshot ?? {}),
    filesChanged: checkpoint.filesChanged ?? [],
    storagePath: checkpoint.storagePath,
    totalFiles: checkpoint.totalFiles ?? files.length,
    totalSizeBytes:
      checkpoint.totalSizeBytes ?? files.reduce((totalSize, file) => totalSize + file.size, 0),
    restoredFrom: checkpoint.restoredFrom,
    restoredAt: checkpoint.restoredAt,
    files,
    fileCount: checkpoint.totalFiles ?? files.length,
    totalSize:
      checkpoint.totalSizeBytes ?? files.reduce((totalSize, file) => totalSize + file.size, 0),
    createdAt: checkpoint.createdAt,
  };
}

export function scoreRequestBody(data: CreateScoreInput | UpdateScoreInput) {
  const value =
    typeof data.value === 'boolean'
      ? Number(data.value)
      : typeof data.value === 'number'
        ? data.value
        : undefined;
  const stringValue = typeof data.value === 'string' ? data.value : data.stringValue;

  return {
    ...data,
    value,
    stringValue,
  };
}

export interface APIKeyResponse {
  id: string;
  name: string;
  publicKey: string;
  secretKeyPreview?: string;
  scopes: string[];
  expiresAt?: string;
  lastUsedAt?: string;
  createdAt: string;
}

export interface APIKeyCreateResponse extends APIKeyResponse {
  secretKey: string;
}

export function normalizeAPIKey(response: APIKeyResponse): APIKey {
  return {
    id: response.id,
    name: response.name,
    keyPrefix: response.publicKey,
    displayKey: response.secretKeyPreview,
    scopes: response.scopes,
    expiresAt: response.expiresAt,
    lastUsedAt: response.lastUsedAt,
    createdAt: response.createdAt,
  };
}
