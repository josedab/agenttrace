import { DataQuery, DataSourceJsonData } from '@grafana/data';

/**
 * Query types supported by the AgentTrace datasource
 */
export enum QueryType {
  Traces = 'traces',
  Observations = 'observations',
  Scores = 'scores',
  Costs = 'costs',
  Metrics = 'metrics',
  Sessions = 'sessions',
}

/**
 * Aggregation functions for metrics
 */
export enum AggregationType {
  Count = 'count',
  Sum = 'sum',
  Avg = 'avg',
  Min = 'min',
  Max = 'max',
  P50 = 'p50',
  P95 = 'p95',
  P99 = 'p99',
}

/**
 * Group by dimensions
 */
export enum GroupByDimension {
  None = 'none',
  Model = 'model',
  Name = 'name',
  UserId = 'userId',
  SessionId = 'sessionId',
  Release = 'release',
  Version = 'version',
  Tag = 'tag',
}

/**
 * Query interface for AgentTrace datasource
 */
export interface AgentTraceQuery extends DataQuery {
  queryType: QueryType;

  // Filtering
  name?: string;
  userId?: string;
  sessionId?: string;
  release?: string;
  version?: string;
  tags?: string[];
  model?: string;

  // Aggregation
  aggregation?: AggregationType;
  groupBy?: GroupByDimension;

  // Metrics-specific
  metricField?: string;

  // Additional options
  includeMetadata?: boolean;
  limit?: number;
}

/**
 * Default query values
 */
export const defaultQuery: Partial<AgentTraceQuery> = {
  queryType: QueryType.Metrics,
  aggregation: AggregationType.Count,
  groupBy: GroupByDimension.None,
  limit: 1000,
};

/**
 * Datasource configuration options (stored securely)
 */
export interface AgentTraceDataSourceOptions extends DataSourceJsonData {
  url?: string;
  projectId?: string;
  defaultTimeRange?: string;
}

/**
 * Secure configuration options (API key, etc.)
 */
export interface AgentTraceSecureJsonData {
  apiKey?: string;
}

/**
 * API response types
 */
export interface TraceResponse {
  id: string;
  name: string;
  startTime: string;
  endTime?: string;
  latency?: number;
  totalCost?: number;
  totalTokens?: number;
  userId?: string;
  sessionId?: string;
  metadata?: Record<string, unknown>;
  tags?: string[];
}

export interface ObservationResponse {
  id: string;
  traceId: string;
  type: 'SPAN' | 'GENERATION' | 'EVENT';
  name: string;
  startTime: string;
  endTime?: string;
  model?: string;
  promptTokens?: number;
  completionTokens?: number;
  totalTokens?: number;
  calculatedCost?: number;
}

export interface ScoreResponse {
  id: string;
  traceId: string;
  observationId?: string;
  name: string;
  value?: number;
  stringValue?: string;
  source: string;
  createdAt: string;
}

export interface MetricsResponse {
  timestamp: string;
  value: number;
  dimensions?: Record<string, string>;
}

export interface CostBreakdown {
  date: string;
  totalCost: number;
  traceCount: number;
  modelCosts?: Array<{
    model: string;
    cost: number;
    count: number;
  }>;
}
