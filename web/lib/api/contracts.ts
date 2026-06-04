// Public API contracts (DTOs) exposed by the AgentTrace web client.
export interface PaginatedResponse<T> {
  data: T[];
  nextCursor?: string;
}

export interface DataListResponse<T> {
  data: T[];
  totalCount: number;
}

export interface ScoreListResponse {
  scores: Score[];
  totalCount: number;
  hasMore: boolean;
}

export interface PaginationParams {
  cursor?: string;
  offset?: number;
  limit?: number;
}

export interface TokenUsage {
  promptTokens: number | null;
  completionTokens: number | null;
  totalTokens: number | null;
}

export interface CostUsage {
  input?: number | null;
  output?: number | null;
  total: number | null;
}

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
  projectId?: string;
  name: string;
  timestamp?: string;
  startTime: string;
  endTime: string | null;
  input: unknown;
  output: unknown;
  metadata: Record<string, unknown> | null;
  tags: string[];
  userId: string | null;
  sessionId: string | null;
  release: string | null;
  version: string | null;
  level: 'DEBUG' | 'DEFAULT' | 'WARNING' | 'ERROR';
  statusMessage: string | null;
  public: boolean;
  latency: number | null;
  totalCost: number | null;
  usage: TokenUsage | null;
  observations?: Observation[];
  scores?: Score[];
}

export interface Observation {
  id: string;
  traceId: string;
  projectId?: string;
  parentObservationId: string | null;
  name: string | null;
  type: 'SPAN' | 'GENERATION' | 'EVENT';
  startTime: string;
  endTime: string | null;
  latency: number | null;
  input: unknown;
  output: unknown;
  metadata: Record<string, unknown> | null;
  level: 'DEBUG' | 'DEFAULT' | 'WARNING' | 'ERROR';
  statusMessage: string | null;
  version: string | null;
  model: string | null;
  modelParameters: Record<string, unknown> | null;
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  totalCost: number | null;
  inputCost: number | null;
  outputCost: number | null;
  usage: TokenUsage | null;
  cost: CostUsage | null;
  children?: Observation[];
  scores?: Score[];
}

export interface Session {
  id: string;
  projectId: string;
  userId: string | null;
  bookmarked: boolean;
  public: boolean;
  createdAt: string;
  updatedAt: string;
  traceCount: number;
  totalCost: number;
  totalTokens: number;
  firstTraceTime: string;
  lastTraceTime: string;
  traces?: Trace[];
}

export interface Score {
  id: string;
  traceId: string;
  observationId: string | null;
  projectId: string;
  name: string;
  value: number | boolean | string | null;
  stringValue: string | null;
  dataType: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN';
  source: 'API' | 'ANNOTATION' | 'EVAL';
  comment: string | null;
  configId?: string | null;
  authorUserId?: string | null;
  createdAt: string;
  updatedAt?: string;
}

export interface Prompt {
  id: string;
  projectId?: string;
  name: string;
  description?: string | null;
  type: 'TEXT' | 'CHAT';
  tags: string[];
  isActive?: boolean;
  activeVersion: PromptVersion | null;
  versions: PromptVersion[];
  labels: Record<string, number>;
  versionCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface PromptVersion {
  id: string;
  promptId?: string;
  version: number;
  prompt: string | null;
  messages: PromptMessage[] | null;
  config: Record<string, unknown> | null;
  labels?: string[];
  variables: string[];
  createdAt: string;
  author?: string;
  commitMessage?: string;
}

export interface PromptMessage {
  role: string;
  content: string;
}

export interface Dataset {
  id: string;
  projectId?: string;
  name: string;
  description: string | null;
  metadata: Record<string, unknown> | null;
  itemCount: number;
  runCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface DatasetItem {
  id: string;
  datasetId: string;
  input: unknown;
  expectedOutput: unknown;
  metadata: Record<string, unknown> | null;
  sourceTraceId: string | null;
  sourceObservationId: string | null;
  status: 'ACTIVE' | 'ARCHIVED';
  createdAt: string;
  updatedAt: string;
}

export interface DatasetRun {
  id: string;
  datasetId: string;
  name: string;
  description: string | null;
  metadata: Record<string, unknown> | null;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  itemCount: number;
  totalCount: number;
  completedCount: number;
  failedCount: number;
  avgScore: number | null;
  totalCost: number | null;
  scores?: number[];
  results?: DatasetRunResult[];
  createdAt: string;
  updatedAt: string;
}

export interface DatasetRunResult {
  id: string;
  input: unknown;
  output: unknown;
  expectedOutput?: unknown;
  score: number | null;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
}

export interface Evaluator {
  id: string;
  projectId?: string;
  name: string;
  description?: string;
  type: 'LLM_AS_JUDGE' | 'CODE' | 'HUMAN';
  status: 'ACTIVE' | 'INACTIVE';
  scoreName: string;
  scoreDataType: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN';
  scoreCategories: string[] | null;
  promptTemplate: string | null;
  variables: string[] | null;
  config: EvaluatorConfig;
  targetFilter: Record<string, unknown> | null;
  samplingRate: number;
  enabled: boolean;
  evalCount: number;
  totalRuns: number;
  avgScore: number | null;
  lastRunAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EvaluatorConfig {
  model?: string;
  prompt?: string;
  code?: string;
  categories?: string[];
  minValue?: number;
  maxValue?: number;
  [key: string]: unknown;
}

export interface EvaluatorStats {
  totalCount: number;
  avgScore: number | null;
  last24hCount: number;
  passRate?: number;
  scoreDistribution: Array<{ range: string; count: number }>;
}

export interface EvaluatorRun {
  id: string;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  startedAt: string;
  completedAt?: string;
  totalCount: number;
  completedCount: number;
  failedCount: number;
  avgScore?: number;
  error?: string;
}

export interface EvaluatorTemplate {
  id: string;
  name: string;
  description: string;
  type?: 'LLM_AS_JUDGE' | 'RULE_BASED';
  promptTemplate: string;
  variables: string[];
  scoreDataType: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN';
  scoreCategories: string[] | null;
}

export interface APIKey {
  id: string;
  projectId?: string;
  name: string;
  displayKey?: string;
  keyPrefix: string;
  scopes: string[];
  expiresAt?: string;
  lastUsedAt?: string;
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
  limit?: number;
  cursor?: string;
  offset?: number;
  userId?: string;
  sessionId?: string;
  name?: string;
  search?: string;
  level?: string;
  tags?: string[];
  startDate?: string | Date;
  endDate?: string | Date;
  minLatency?: number;
  maxLatency?: number;
  minCost?: number;
  maxCost?: number;
  fromTimestamp?: string;
  toTimestamp?: string;
}

export type TraceFilterParams = Omit<TraceListParams, 'cursor' | 'limit'>;

export interface UpdateTraceInput {
  name?: string;
  userId?: string;
  sessionId?: string;
  metadata?: Record<string, unknown>;
  tags?: string[];
  public?: boolean;
}

export interface SessionListParams {
  limit?: string;
  cursor?: string;
  fromTimestamp?: string;
  toTimestamp?: string;
}

export interface ScoreListParams {
  limit?: number;
  cursor?: string;
  offset?: number;
  traceId?: string;
  observationId?: string;
  name?: string;
  scoreName?: string;
  source?: '' | 'API' | 'ANNOTATION' | 'EVAL';
  dataType?: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN';
  fromTimestamp?: string;
  toTimestamp?: string;
  minScore?: number;
  maxScore?: number;
  evaluatorId?: string;
}

export interface CreateScoreInput {
  traceId: string;
  observationId?: string;
  name: string;
  value?: number | boolean | string;
  stringValue?: string;
  dataType?: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN';
  source?: 'API' | 'ANNOTATION' | 'EVAL';
  comment?: string;
}

export interface UpdateScoreInput {
  value?: number | boolean | string;
  stringValue?: string;
  comment?: string;
}

export interface CreatePromptInput {
  name: string;
  description?: string;
  type?: 'TEXT' | 'CHAT';
  tags?: string[];
  prompt?: string;
  messages?: PromptMessage[];
  config?: Record<string, unknown>;
  labels?: string[];
}

export interface UpdatePromptInput {
  name?: string;
  description?: string;
  type?: 'TEXT' | 'CHAT';
  tags?: string[];
}

export interface CreatePromptVersionInput {
  prompt?: string;
  messages?: PromptMessage[];
  config?: Record<string, unknown>;
  labels?: string[];
  commitMessage?: string;
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
  input: unknown;
  expectedOutput?: unknown;
  metadata?: Record<string, unknown>;
  sourceTraceId?: string;
  sourceObservationId?: string;
}

export interface UpdateDatasetItemInput {
  input?: unknown;
  expectedOutput?: unknown;
  metadata?: Record<string, unknown>;
  status?: 'ACTIVE' | 'ARCHIVED';
}

export interface CreateDatasetRunInput {
  name: string;
  description?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateEvaluatorInput {
  name: string;
  description?: string;
  type: 'LLM_AS_JUDGE' | 'CODE' | 'HUMAN';
  scoreName: string;
  scoreDataType?: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN';
  scoreCategories?: string[];
  promptTemplate?: string;
  variables?: string[];
  config?: Record<string, unknown>;
  targetFilter?: Record<string, unknown>;
  samplingRate?: number;
  enabled?: boolean;
  templateId?: string;
  template?: string;
}

export interface UpdateEvaluatorInput {
  name?: string;
  description?: string;
  status?: 'ACTIVE' | 'INACTIVE';
  promptTemplate?: string;
  variables?: string[];
  scoreCategories?: string[];
  config?: EvaluatorConfig;
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
  expiresIn?: string;
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

export interface TraceStats {
  totalTraces: number;
  errorRate?: number;
  averageLatency?: number;
  totalCost?: number;
  totalTokens?: number;
  errorCount?: number;
}

export interface ScoreStats {
  name: string;
  count: number;
  avgValue?: number | null;
  minValue?: number | null;
  maxValue?: number | null;
  medianValue?: number | null;
}

export interface ScoreDistribution {
  buckets: Array<{
    min: number;
    max: number;
    count: number;
  }>;
}

export interface AgentGraphNode {
  id: string;
  name: string;
  type: string;
  model?: string;
  tokensUsed: number;
  cost: number;
  durationMs: number;
  status: string;
}

export interface AgentGraphEdge {
  sourceId: string;
  targetId: string;
  label?: string;
  messageCount: number;
  tokenCount: number;
}

export interface AgentGraph {
  traceId: string;
  agents: AgentGraphNode[];
  edges: AgentGraphEdge[];
  totalCost: number;
  totalDurationMs: number;
}

export interface ReplayEvent {
  id: string;
  type: string;
  timestamp: string;
  durationMs?: number;
  title: string;
  description?: string;
  status: string;
  data: Record<string, unknown>;
  children?: ReplayEvent[];
}

export interface ReplayTimeline {
  traceId: string;
  traceName: string;
  startTime: string;
  endTime?: string;
  durationMs: number;
  events: ReplayEvent[];
  summary: {
    totalEvents: number;
    llmCalls: number;
    toolCalls: number;
    fileOperations: number;
    terminalCommands: number;
    checkpoints: number;
    errors: number;
    totalTokens: number;
    totalCost: number;
  };
}

export interface ReplayStepState {
  costSoFar: number;
  tokensSoFar: number;
  elapsedMs: number;
}

export interface PromptListParams {
  search?: string;
  label?: string;
  tags?: string[];
}

export interface CompiledPrompt {
  compiled: string;
  prompt: Prompt;
  version: number;
  variables: Record<string, string>;
}

export interface RunPromptInput {
  version?: number;
  variables: Record<string, string>;
}

export interface RunPromptResult {
  output: string;
}

export interface AnnotationQueue {
  id: string;
  name: string;
  description?: string;
  scoreName: string;
  scoreDataType: 'NUMERIC' | 'BOOLEAN' | 'CATEGORICAL';
  categories?: string[];
  pendingCount: number;
  completedCount: number;
  totalCount: number;
  createdAt: string;
}

export interface AnnotationQueueItem {
  id: string;
  traceId: string;
  input: unknown;
  output: unknown;
  expectedOutput?: unknown;
  metadata?: Record<string, unknown>;
}

export interface AnnotationScoreInput {
  score: number | boolean | string;
  comment?: string;
}

export interface UserProfile {
  id: string;
  name: string;
  email: string;
  bio?: string;
  avatar?: string;
}

export interface UpdateUserProfileInput {
  name: string;
  email: string;
  bio?: string;
}

export interface ProjectSettings {
  id: string;
  name: string;
  description?: string;
  defaultRetentionDays: number;
  publicDashboard: boolean;
  publicUrl?: string;
}

export type UpdateProjectSettingsInput = Omit<ProjectSettings, 'id'>;

export type TeamRole = 'OWNER' | 'ADMIN' | 'MEMBER' | 'VIEWER';

export interface TeamMember {
  id: string;
  userId: string;
  name: string;
  email: string;
  avatar?: string;
  role: TeamRole;
  joinedAt: string;
}

export interface TraceVolumePoint {
  date: string;
  traces: number;
  generations: number;
}

export interface CostBreakdownPoint {
  model: string;
  cost: number;
  tokens: number;
}

export interface LatencyPercentilePoint {
  date: string;
  p50: number;
  p95: number;
  p99: number;
}

export interface CostTrendPoint {
  date: string;
  cost: number;
  inputCost: number;
  outputCost: number;
}

export interface TokenUsagePoint {
  date: string;
  inputTokens: number;
  outputTokens: number;
}

export interface ModelDistributionPoint {
  name: string;
  count: number;
  percentage: number;
}

export interface DashboardMetrics {
  totalTraces: number;
  totalCost: number;
  avgLatency: number;
  activeSessions: number;
  totalTokens?: number;
  tracesChange: number;
  costChange: number;
  latencyChange: number;
  sessionsChange: number;
  tokensChange?: number;
  traceVolume: TraceVolumePoint[];
  costBreakdown: CostBreakdownPoint[];
  latencyPercentiles: LatencyPercentilePoint[];
  recentTraces: Trace[];
}

export interface AnalyticsOverview {
  totalTraces: number;
  totalCost: number;
  avgLatency: number;
  totalTokens: number;
  tracesChange: number;
  costChange: number;
  latencyChange: number;
  tokensChange: number;
  traceVolume: TraceVolumePoint[];
  costByModel: CostBreakdownPoint[];
  latencyPercentiles: LatencyPercentilePoint[];
}

export interface AnalyticsGroupParams {
  dateRange: string;
  groupBy: string;
}

export interface CostAnalyticsBreakdown {
  name: string;
  traces: number;
  inputTokens: number;
  outputTokens: number;
  totalCost: number;
  percentage: number;
}

export interface CostAnalytics {
  totalCost: number;
  costChange: number;
  inputCost: number;
  outputCost: number;
  avgCostPerTrace: number;
  costOverTime: CostTrendPoint[];
  costByGroup: CostBreakdownPoint[];
  breakdown: CostAnalyticsBreakdown[];
}

export interface LatencyBreakdown {
  name: string;
  requests: number;
  p50: number;
  p95: number;
  p99: number;
  avg: number;
}

export interface LatencyDistributionBucket {
  range: string;
  count: number;
  percentage: number;
}

export interface LatencyAnalytics {
  p50: number;
  p95: number;
  p99: number;
  avg: number;
  avgChange: number;
  totalRequests: number;
  latencyOverTime: LatencyPercentilePoint[];
  distribution: LatencyDistributionBucket[];
  breakdown: LatencyBreakdown[];
}

export interface TopUsageTrace {
  id: string;
  model: string;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  cost: number;
}

export interface UsageAnalytics {
  totalTraces: number;
  tracesChange: number;
  totalGenerations: number;
  inputTokens: number;
  outputTokens: number;
  volumeOverTime: TraceVolumePoint[];
  tokenUsageOverTime: TokenUsagePoint[];
  modelDistribution: ModelDistributionPoint[];
  topTraces: TopUsageTrace[];
}

export interface RecentTraceError {
  id: string;
  traceId: string;
  message: string;
  timestamp: string;
}

export interface ForecastPoint {
  date: string;
  predictedCost: number;
  lowerBound: number;
  upperBound: number;
}

export interface SimulationChange {
  fromModel: string;
  toModel: string;
  trafficPercent: number;
}

export interface SimulationResult {
  baselineCost: number;
  projectedCost: number;
  savings: number;
  qualityImpact: number;
}

export interface BudgetPlanInput {
  monthlyBudget: string;
  alertThreshold80: boolean;
  alertThreshold90: boolean;
  alertThreshold100: boolean;
  modelAllocations: string;
}

export interface BudgetPlan extends BudgetPlanInput {
  id: string;
}

export interface EnrichmentRule {
  id: string;
  name: string;
  triggerEvent: string;
  sourceType: string;
  condition: string;
  transform: string;
  priority: number;
  enabled: boolean;
  fireCount: number;
  lastFired: string | null;
}

export type CreateEnrichmentRuleInput = Pick<
  EnrichmentRule,
  'name' | 'triggerEvent' | 'sourceType' | 'condition' | 'transform' | 'priority'
>;

export interface UpdateEnrichmentRuleInput {
  enabled?: boolean;
  name?: string;
  triggerEvent?: string;
  sourceType?: string;
  condition?: string;
  transform?: string;
  priority?: number;
}

export interface EnrichmentRuleTestResult {
  success: boolean;
  output?: unknown;
  error?: string;
}

export interface KnowledgeBaseEntry {
  id: string;
  title: string;
  description: string;
  category: string;
  tags: string[];
  rootCause?: string;
  pattern?: string;
  fix?: string;
  createdAt: string;
}

export interface KnowledgeBaseSearchParams {
  query?: string;
  category?: string;
}

export type CreateKnowledgeBaseEntryInput = Omit<KnowledgeBaseEntry, 'id' | 'createdAt'>;

export interface OTelBridgeConfiguration {
  exportEnabled: boolean;
  importEnabled: boolean;
  samplingRate: number;
  resourceAttributes: Record<string, string>;
}

export interface OTelBridgeDestination {
  id: string;
  type: 'jaeger' | 'tempo' | 'datadog' | 'otlp' | 'zipkin';
  name: string;
  endpoint: string;
  enabled: boolean;
}

export type CreateOTelBridgeDestinationInput = Pick<
  OTelBridgeDestination,
  'type' | 'name' | 'endpoint'
>;

export interface OTelImportInput {
  correlateByTraceId: boolean;
  createMissingTraces: boolean;
}

export interface OTelImportResult {
  importedSpans: number;
  createdTraces: number;
}

export interface OTelBridgeStatsSection {
  totalSpans: number;
  successCount: number;
  errorCount: number;
  avgLatencyMs: number;
  last24hCount: number;
}

export interface OTelBridgeStats {
  exportStats: OTelBridgeStatsSection;
  importStats: OTelBridgeStatsSection;
}

export interface CostRecommendation {
  id: string;
  currentModel: string;
  recommendedModel: string;
  traceCount: number;
  estimatedSavingsPerMonth: number;
  qualityImpactEstimate: number;
  confidence: number;
  status: string;
}

export interface CostOptimizerAnalysis {
  projectId: string;
  totalCostPeriod: number;
  modelBreakdown: Array<{
    model: string;
    traceCount: number;
    totalCost: number;
    avgCostPerTrace: number;
  }>;
  recommendations: CostRecommendation[];
  potentialSavings: number;
}

export interface GuardrailRule {
  id: string;
  name: string;
  description: string;
  type: string;
  action: string;
  enabled: boolean;
  config: {
    maxCostPerTrace?: number;
    maxLatencyMs?: number;
    restrictedPaths?: string[];
    blockedPatterns?: string[];
  };
}

export interface GuardrailViolationStats {
  totalViolations: number;
  violationsByRule: Record<string, number>;
  violationsByAction: Record<string, number>;
  recentViolations: number;
}

// SSO Types
export interface SSOConfiguration {
  id: string;
  organizationId: string;
  provider: 'saml' | 'oidc';
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
  provider: 'saml' | 'oidc';
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
  status: 'pending' | 'processing' | 'completed' | 'failed';
  format: 'json' | 'csv';
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
  observationId?: string;
  name: string;
  type: 'auto' | 'manual' | 'pre_edit' | 'post_edit' | 'rollback';
  description: string | null;
  gitCommitSha?: string;
  gitBranch?: string;
  gitRepoUrl?: string;
  filesSnapshot: string;
  filesChanged: string[];
  storagePath?: string;
  totalFiles: number;
  totalSizeBytes: number;
  restoredFrom?: string;
  restoredAt?: string;
  files: CheckpointFile[];
  fileCount: number;
  totalSize: number;
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
  type?: Checkpoint['type'];
  offset?: number;
  limit?: number;
}

export interface CreateCheckpointInput {
  traceId: string;
  observationId?: string;
  name: string;
  type?: Checkpoint['type'];
  description?: string;
  gitCommitSha?: string;
  gitBranch?: string;
  gitRepoUrl?: string;
  filesSnapshot?: string;
  filesChanged?: string[];
  totalFiles?: number;
  totalSizeBytes?: number;
}

// Git Link Types
export interface GitLink {
  id: string;
  projectId: string;
  traceId: string;
  commitSha: string;
  parentSha?: string;
  branch?: string;
  tag?: string;
  repoUrl?: string;
  commitMessage?: string;
  commitAuthor?: string;
  commitAuthorEmail?: string;
  commitTimestamp: string;
  filesAdded?: string[];
  filesModified?: string[];
  filesDeleted?: string[];
  filesChangedCount: number;
  additions: number;
  deletions: number;
  linkType: 'current' | 'start' | 'end' | 'referenced';
  ciRunId?: string;
  createdAt: string;
}

export interface GitTimelineEntry {
  commitSha: string;
  commitMessage: string;
  commitAuthor: string;
  commitTime: string;
  branch: string;
  traceCount: number;
  traceIds: string[];
}

export interface CreateGitLinkInput {
  traceId: string;
  commitSha: string;
  parentSha?: string;
  branch?: string;
  tag?: string;
  repoUrl?: string;
  commitMessage?: string;
  commitAuthor?: string;
  commitAuthorEmail?: string;
  commitTimestamp?: string;
  filesAdded?: string[];
  filesModified?: string[];
  filesDeleted?: string[];
  additions?: number;
  deletions?: number;
  linkType?: GitLink['linkType'];
  ciRunId?: string;
}

// File Operation Types
export interface FileOperation {
  id: string;
  projectId: string;
  traceId: string;
  observationId?: string;
  operation: 'create' | 'read' | 'update' | 'delete' | 'rename' | 'move' | 'copy';
  filePath: string;
  newPath?: string;
  fileSize: number;
  fileMode?: string;
  contentHash?: string;
  mimeType?: string;
  linesAdded: number;
  linesRemoved: number;
  diffPreview?: string;
  contentBeforeHash?: string;
  contentAfterHash?: string;
  toolName?: string;
  reason?: string;
  startedAt: string;
  completedAt?: string;
  durationMs: number;
  success: boolean;
  errorMessage?: string;
}

export interface CreateFileOperationInput {
  traceId: string;
  observationId?: string;
  operation: FileOperation['operation'];
  filePath: string;
  newPath?: string;
  fileSize?: number;
  fileMode?: string;
  contentHash?: string;
  mimeType?: string;
  linesAdded?: number;
  linesRemoved?: number;
  diffPreview?: string;
  contentBeforeHash?: string;
  contentAfterHash?: string;
  toolName?: string;
  reason?: string;
  startedAt?: string;
  completedAt?: string;
  durationMs?: number;
  success?: boolean;
  errorMessage?: string;
}

// Terminal Command Types
export interface TerminalCommand {
  id: string;
  projectId: string;
  traceId: string;
  observationId?: string;
  command: string;
  args?: string[];
  workingDirectory?: string;
  shell?: string;
  envVars?: string;
  startedAt: string;
  completedAt?: string;
  durationMs: number;
  exitCode: number;
  stdout?: string;
  stderr?: string;
  stdoutTruncated: boolean;
  stderrTruncated: boolean;
  success: boolean;
  timedOut: boolean;
  killed: boolean;
  maxMemoryBytes: number;
  cpuTimeMs: number;
  toolName?: string;
  reason?: string;
}

export interface CreateTerminalCommandInput {
  traceId: string;
  observationId?: string;
  command: string;
  args?: string[];
  workingDirectory?: string;
  shell?: string;
  envVars?: string;
  startedAt?: string;
  completedAt?: string;
  durationMs?: number;
  exitCode?: number;
  stdout?: string;
  stderr?: string;
  stdoutTruncated?: boolean;
  stderrTruncated?: boolean;
  success?: boolean;
  timedOut?: boolean;
  killed?: boolean;
  maxMemoryBytes?: number;
  cpuTimeMs?: number;
  toolName?: string;
  reason?: string;
}

// CI Run Types
export interface CIRun {
  id: string;
  projectId: string;
  provider: CIProvider;
  providerRunId: string;
  providerRunUrl?: string;
  pipelineName?: string;
  jobName?: string;
  workflowName?: string;
  gitCommitSha?: string;
  gitBranch?: string;
  gitTag?: string;
  gitRepoUrl?: string;
  gitRef?: string;
  prNumber?: number;
  prTitle?: string;
  prSourceBranch?: string;
  prTargetBranch?: string;
  startedAt: string;
  completedAt?: string;
  durationMs: number;
  status: CIRunStatus;
  conclusion?: string;
  errorMessage?: string;
  traceIds?: string[];
  traceCount: number;
  totalCost: number;
  totalTokens: number;
  totalObservations: number;
  runnerName?: string;
  runnerOs?: string;
  runnerArch?: string;
  triggeredBy?: string;
  triggerEvent?: string;
  createdAt: string;
  updatedAt: string;
}

export type CIProvider =
  | 'github_actions'
  | 'gitlab_ci'
  | 'jenkins'
  | 'circleci'
  | 'azure_devops'
  | 'bitbucket'
  | 'other';

export type CIRunStatus = 'pending' | 'running' | 'success' | 'failure' | 'cancelled' | 'skipped';

export interface CreateCIRunInput {
  provider: CIProvider;
  providerRunId: string;
  providerRunUrl?: string;
  pipelineName?: string;
  jobName?: string;
  workflowName?: string;
  gitCommitSha?: string;
  gitBranch?: string;
  gitTag?: string;
  gitRepoUrl?: string;
  gitRef?: string;
  prNumber?: number;
  prTitle?: string;
  prSourceBranch?: string;
  prTargetBranch?: string;
  startedAt?: string;
  status?: CIRunStatus;
  conclusion?: string;
  errorMessage?: string;
  runnerName?: string;
  runnerOs?: string;
  runnerArch?: string;
  triggeredBy?: string;
  triggerEvent?: string;
}

export interface UpdateCIRunInput {
  status?: CIRunStatus;
  conclusion?: string;
  errorMessage?: string;
  completedAt?: string;
  traceIds?: string[];
  totalCost?: number;
  totalTokens?: number;
  totalObservations?: number;
}
