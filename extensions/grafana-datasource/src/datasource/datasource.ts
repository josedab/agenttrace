import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  MutableDataFrame,
  FieldType,
  dateTime,
  MetricFindValue,
} from '@grafana/data';
import { getBackendSrv, getTemplateSrv, BackendSrvRequest } from '@grafana/runtime';

import {
  AgentTraceQuery,
  AgentTraceDataSourceOptions,
  AgentTraceSecureJsonData,
  QueryType,
  TraceResponse,
  ObservationResponse,
  ScoreResponse,
  MetricsResponse,
  CostBreakdown,
  defaultQuery,
} from '../types';

export class AgentTraceDataSource extends DataSourceApi<AgentTraceQuery, AgentTraceDataSourceOptions> {
  url: string;
  projectId?: string;

  constructor(instanceSettings: DataSourceInstanceSettings<AgentTraceDataSourceOptions>) {
    super(instanceSettings);
    this.url = instanceSettings.url || '';
    this.projectId = instanceSettings.jsonData.projectId;
  }

  /**
   * Main query method called by Grafana
   */
  async query(options: DataQueryRequest<AgentTraceQuery>): Promise<DataQueryResponse> {
    const { range } = options;
    const from = range!.from.toISOString();
    const to = range!.to.toISOString();

    const promises = options.targets
      .filter((target) => !target.hide)
      .map((target) => this.runQuery(target, from, to));

    const results = await Promise.all(promises);
    return { data: results.flat() };
  }

  /**
   * Run a single query
   */
  private async runQuery(query: AgentTraceQuery, from: string, to: string): Promise<MutableDataFrame[]> {
    const q = { ...defaultQuery, ...query };

    switch (q.queryType) {
      case QueryType.Traces:
        return this.queryTraces(q, from, to);
      case QueryType.Observations:
        return this.queryObservations(q, from, to);
      case QueryType.Scores:
        return this.queryScores(q, from, to);
      case QueryType.Costs:
        return this.queryCosts(q, from, to);
      case QueryType.Metrics:
      default:
        return this.queryMetrics(q, from, to);
    }
  }

  /**
   * Query traces
   */
  private async queryTraces(query: AgentTraceQuery, from: string, to: string): Promise<MutableDataFrame[]> {
    const params: Record<string, string> = {
      fromTimestamp: from,
      toTimestamp: to,
      limit: String(query.limit || 100),
    };

    if (query.name) params.name = this.templateReplace(query.name);
    if (query.userId) params.userId = this.templateReplace(query.userId);
    if (query.sessionId) params.sessionId = this.templateReplace(query.sessionId);
    if (query.release) params.release = this.templateReplace(query.release);
    if (query.version) params.version = this.templateReplace(query.version);

    const response = await this.doRequest<{ data: TraceResponse[] }>('/api/public/traces', params);
    const traces = response.data || [];

    const frame = new MutableDataFrame({
      refId: query.refId,
      fields: [
        { name: 'Time', type: FieldType.time },
        { name: 'ID', type: FieldType.string },
        { name: 'Name', type: FieldType.string },
        { name: 'Latency (ms)', type: FieldType.number },
        { name: 'Cost ($)', type: FieldType.number },
        { name: 'Tokens', type: FieldType.number },
        { name: 'User ID', type: FieldType.string },
        { name: 'Session ID', type: FieldType.string },
      ],
    });

    for (const trace of traces) {
      frame.appendRow([
        dateTime(trace.startTime).valueOf(),
        trace.id,
        trace.name,
        trace.latency,
        trace.totalCost,
        trace.totalTokens,
        trace.userId,
        trace.sessionId,
      ]);
    }

    return [frame];
  }

  /**
   * Query observations
   */
  private async queryObservations(query: AgentTraceQuery, from: string, to: string): Promise<MutableDataFrame[]> {
    const params: Record<string, string> = {
      fromTimestamp: from,
      toTimestamp: to,
      limit: String(query.limit || 100),
    };

    if (query.name) params.name = this.templateReplace(query.name);
    if (query.model) params.model = this.templateReplace(query.model);

    const response = await this.doRequest<{ data: ObservationResponse[] }>('/api/public/observations', params);
    const observations = response.data || [];

    const frame = new MutableDataFrame({
      refId: query.refId,
      fields: [
        { name: 'Time', type: FieldType.time },
        { name: 'ID', type: FieldType.string },
        { name: 'Trace ID', type: FieldType.string },
        { name: 'Type', type: FieldType.string },
        { name: 'Name', type: FieldType.string },
        { name: 'Model', type: FieldType.string },
        { name: 'Prompt Tokens', type: FieldType.number },
        { name: 'Completion Tokens', type: FieldType.number },
        { name: 'Cost ($)', type: FieldType.number },
      ],
    });

    for (const obs of observations) {
      frame.appendRow([
        dateTime(obs.startTime).valueOf(),
        obs.id,
        obs.traceId,
        obs.type,
        obs.name,
        obs.model,
        obs.promptTokens,
        obs.completionTokens,
        obs.calculatedCost,
      ]);
    }

    return [frame];
  }

  /**
   * Query scores
   */
  private async queryScores(query: AgentTraceQuery, from: string, to: string): Promise<MutableDataFrame[]> {
    const params: Record<string, string> = {
      fromTimestamp: from,
      toTimestamp: to,
      limit: String(query.limit || 100),
    };

    if (query.name) params.name = this.templateReplace(query.name);

    const response = await this.doRequest<{ data: ScoreResponse[] }>('/api/public/scores', params);
    const scores = response.data || [];

    const frame = new MutableDataFrame({
      refId: query.refId,
      fields: [
        { name: 'Time', type: FieldType.time },
        { name: 'ID', type: FieldType.string },
        { name: 'Trace ID', type: FieldType.string },
        { name: 'Name', type: FieldType.string },
        { name: 'Value', type: FieldType.number },
        { name: 'String Value', type: FieldType.string },
        { name: 'Source', type: FieldType.string },
      ],
    });

    for (const score of scores) {
      frame.appendRow([
        dateTime(score.createdAt).valueOf(),
        score.id,
        score.traceId,
        score.name,
        score.value,
        score.stringValue,
        score.source,
      ]);
    }

    return [frame];
  }

  /**
   * Query daily costs
   */
  private async queryCosts(query: AgentTraceQuery, from: string, to: string): Promise<MutableDataFrame[]> {
    const params: Record<string, string> = {
      fromDate: from.split('T')[0],
      toDate: to.split('T')[0],
    };

    if (query.groupBy && query.groupBy !== 'none') {
      params.groupBy = query.groupBy;
    }

    const response = await this.doRequest<{ data: CostBreakdown[] }>('/api/public/metrics/costs', params);
    const costs = response.data || [];

    const frame = new MutableDataFrame({
      refId: query.refId,
      fields: [
        { name: 'Time', type: FieldType.time },
        { name: 'Cost ($)', type: FieldType.number },
        { name: 'Trace Count', type: FieldType.number },
      ],
    });

    for (const cost of costs) {
      frame.appendRow([
        dateTime(cost.date).valueOf(),
        cost.totalCost,
        cost.traceCount,
      ]);
    }

    return [frame];
  }

  /**
   * Query aggregated metrics
   */
  private async queryMetrics(query: AgentTraceQuery, from: string, to: string): Promise<MutableDataFrame[]> {
    const params: Record<string, string> = {
      fromTimestamp: from,
      toTimestamp: to,
      aggregation: query.aggregation || 'count',
    };

    if (query.groupBy && query.groupBy !== 'none') {
      params.groupBy = query.groupBy;
    }
    if (query.name) params.name = this.templateReplace(query.name);
    if (query.userId) params.userId = this.templateReplace(query.userId);
    if (query.model) params.model = this.templateReplace(query.model);
    if (query.metricField) params.field = query.metricField;

    const response = await this.doRequest<{ data: MetricsResponse[] }>('/api/public/metrics', params);
    const metrics = response.data || [];

    const frame = new MutableDataFrame({
      refId: query.refId,
      fields: [
        { name: 'Time', type: FieldType.time },
        { name: 'Value', type: FieldType.number },
      ],
    });

    // Add dimension fields if grouped
    if (query.groupBy && query.groupBy !== 'none' && metrics.length > 0 && metrics[0].dimensions) {
      const dimensionKeys = Object.keys(metrics[0].dimensions);
      for (const key of dimensionKeys) {
        frame.addField({ name: key, type: FieldType.string });
      }
    }

    for (const metric of metrics) {
      const row: unknown[] = [dateTime(metric.timestamp).valueOf(), metric.value];
      if (metric.dimensions) {
        row.push(...Object.values(metric.dimensions));
      }
      frame.appendRow(row);
    }

    return [frame];
  }

  /**
   * Test datasource connection
   */
  async testDatasource() {
    try {
      await this.doRequest('/health');
      return {
        status: 'success',
        message: 'Successfully connected to AgentTrace',
      };
    } catch (error) {
      return {
        status: 'error',
        message: `Failed to connect: ${error instanceof Error ? error.message : 'Unknown error'}`,
      };
    }
  }

  /**
   * Get variable options for templating
   */
  async metricFindQuery(query: string): Promise<MetricFindValue[]> {
    const queryLower = query.toLowerCase().trim();

    if (queryLower === 'models' || queryLower.startsWith('models(')) {
      return this.getModels();
    }
    if (queryLower === 'users' || queryLower.startsWith('users(')) {
      return this.getUsers();
    }
    if (queryLower === 'sessions' || queryLower.startsWith('sessions(')) {
      return this.getSessions();
    }
    if (queryLower === 'names' || queryLower.startsWith('names(')) {
      return this.getTraceNames();
    }
    if (queryLower === 'releases' || queryLower.startsWith('releases(')) {
      return this.getReleases();
    }
    if (queryLower === 'tags' || queryLower.startsWith('tags(')) {
      return this.getTags();
    }

    return [];
  }

  private async getModels(): Promise<MetricFindValue[]> {
    const response = await this.doRequest<{ data: string[] }>('/api/public/metrics/models');
    return (response.data || []).map((model) => ({ text: model, value: model }));
  }

  private async getUsers(): Promise<MetricFindValue[]> {
    const response = await this.doRequest<{ data: string[] }>('/api/public/metrics/users');
    return (response.data || []).map((user) => ({ text: user, value: user }));
  }

  private async getSessions(): Promise<MetricFindValue[]> {
    const response = await this.doRequest<{ data: string[] }>('/api/public/sessions');
    return (response.data || []).map((session: { id: string }) => ({
      text: session.id || session,
      value: session.id || session
    }));
  }

  private async getTraceNames(): Promise<MetricFindValue[]> {
    const response = await this.doRequest<{ data: string[] }>('/api/public/metrics/trace-names');
    return (response.data || []).map((name) => ({ text: name, value: name }));
  }

  private async getReleases(): Promise<MetricFindValue[]> {
    const response = await this.doRequest<{ data: string[] }>('/api/public/metrics/releases');
    return (response.data || []).map((release) => ({ text: release, value: release }));
  }

  private async getTags(): Promise<MetricFindValue[]> {
    const response = await this.doRequest<{ data: string[] }>('/api/public/metrics/tags');
    return (response.data || []).map((tag) => ({ text: tag, value: tag }));
  }

  /**
   * Make HTTP request to AgentTrace API
   */
  private async doRequest<T>(path: string, params?: Record<string, string>): Promise<T> {
    const options: BackendSrvRequest = {
      url: `${this.url}${path}`,
      method: 'GET',
      params,
    };

    const response = await getBackendSrv().datasourceRequest(options);
    return response.data;
  }

  /**
   * Replace template variables in a string
   */
  private templateReplace(value: string): string {
    return getTemplateSrv().replace(value);
  }

  /**
   * Get annotations for a time range
   */
  async annotationQuery(options: any): Promise<any[]> {
    const { annotation, range } = options;
    const from = range.from.toISOString();
    const to = range.to.toISOString();

    // Fetch traces that could be used as annotations
    const params: Record<string, string> = {
      fromTimestamp: from,
      toTimestamp: to,
      limit: '100',
    };

    if (annotation.name) {
      params.name = annotation.name;
    }
    if (annotation.tags) {
      params.tags = annotation.tags;
    }

    const response = await this.doRequest<{ data: TraceResponse[] }>('/api/public/traces', params);
    const traces = response.data || [];

    return traces.map((trace) => ({
      time: dateTime(trace.startTime).valueOf(),
      timeEnd: trace.endTime ? dateTime(trace.endTime).valueOf() : undefined,
      title: trace.name,
      text: `Cost: $${trace.totalCost?.toFixed(4) || 'N/A'}, Tokens: ${trace.totalTokens || 'N/A'}`,
      tags: trace.tags,
    }));
  }
}
