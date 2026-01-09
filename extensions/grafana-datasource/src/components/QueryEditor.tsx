import React from 'react';
import { InlineField, Select, Input, MultiSelect, InlineFieldRow } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { AgentTraceDataSource } from '../datasource/datasource';
import {
  AgentTraceDataSourceOptions,
  AgentTraceQuery,
  QueryType,
  AggregationType,
  GroupByDimension,
  defaultQuery,
} from '../types';

type Props = QueryEditorProps<AgentTraceDataSource, AgentTraceQuery, AgentTraceDataSourceOptions>;

const queryTypeOptions: Array<SelectableValue<QueryType>> = [
  { label: 'Metrics', value: QueryType.Metrics, description: 'Aggregated metrics over time' },
  { label: 'Traces', value: QueryType.Traces, description: 'Individual trace data' },
  { label: 'Observations', value: QueryType.Observations, description: 'Spans, generations, and events' },
  { label: 'Scores', value: QueryType.Scores, description: 'Evaluation scores' },
  { label: 'Costs', value: QueryType.Costs, description: 'Daily cost breakdown' },
];

const aggregationOptions: Array<SelectableValue<AggregationType>> = [
  { label: 'Count', value: AggregationType.Count },
  { label: 'Sum', value: AggregationType.Sum },
  { label: 'Average', value: AggregationType.Avg },
  { label: 'Min', value: AggregationType.Min },
  { label: 'Max', value: AggregationType.Max },
  { label: 'P50', value: AggregationType.P50 },
  { label: 'P95', value: AggregationType.P95 },
  { label: 'P99', value: AggregationType.P99 },
];

const groupByOptions: Array<SelectableValue<GroupByDimension>> = [
  { label: 'None', value: GroupByDimension.None },
  { label: 'Model', value: GroupByDimension.Model },
  { label: 'Name', value: GroupByDimension.Name },
  { label: 'User ID', value: GroupByDimension.UserId },
  { label: 'Session ID', value: GroupByDimension.SessionId },
  { label: 'Release', value: GroupByDimension.Release },
  { label: 'Version', value: GroupByDimension.Version },
  { label: 'Tag', value: GroupByDimension.Tag },
];

const metricFieldOptions: Array<SelectableValue<string>> = [
  { label: 'Trace Count', value: 'traceCount' },
  { label: 'Total Cost', value: 'totalCost' },
  { label: 'Total Tokens', value: 'totalTokens' },
  { label: 'Latency', value: 'latency' },
  { label: 'Prompt Tokens', value: 'promptTokens' },
  { label: 'Completion Tokens', value: 'completionTokens' },
  { label: 'Error Rate', value: 'errorRate' },
];

/**
 * QueryEditor component for building AgentTrace queries
 */
export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const q = { ...defaultQuery, ...query };

  const onQueryTypeChange = (value: SelectableValue<QueryType>) => {
    onChange({ ...q, queryType: value.value! });
    onRunQuery();
  };

  const onAggregationChange = (value: SelectableValue<AggregationType>) => {
    onChange({ ...q, aggregation: value.value! });
    onRunQuery();
  };

  const onGroupByChange = (value: SelectableValue<GroupByDimension>) => {
    onChange({ ...q, groupBy: value.value! });
    onRunQuery();
  };

  const onMetricFieldChange = (value: SelectableValue<string>) => {
    onChange({ ...q, metricField: value.value! });
    onRunQuery();
  };

  const onNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...q, name: event.target.value });
  };

  const onNameBlur = () => {
    onRunQuery();
  };

  const onUserIdChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...q, userId: event.target.value });
  };

  const onUserIdBlur = () => {
    onRunQuery();
  };

  const onSessionIdChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...q, sessionId: event.target.value });
  };

  const onSessionIdBlur = () => {
    onRunQuery();
  };

  const onModelChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...q, model: event.target.value });
  };

  const onModelBlur = () => {
    onRunQuery();
  };

  const onReleaseChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...q, release: event.target.value });
  };

  const onReleaseBlur = () => {
    onRunQuery();
  };

  const onLimitChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...q, limit: parseInt(event.target.value, 10) || 100 });
  };

  const onLimitBlur = () => {
    onRunQuery();
  };

  const showAggregationOptions = q.queryType === QueryType.Metrics || q.queryType === QueryType.Costs;
  const showMetricField = q.queryType === QueryType.Metrics;
  const showModelFilter = q.queryType === QueryType.Observations || q.queryType === QueryType.Metrics;

  return (
    <div>
      {/* Query Type Selection */}
      <InlineFieldRow>
        <InlineField label="Query Type" labelWidth={14}>
          <Select
            width={20}
            options={queryTypeOptions}
            value={queryTypeOptions.find((opt) => opt.value === q.queryType)}
            onChange={onQueryTypeChange}
          />
        </InlineField>

        {showAggregationOptions && (
          <>
            <InlineField label="Aggregation" labelWidth={12}>
              <Select
                width={16}
                options={aggregationOptions}
                value={aggregationOptions.find((opt) => opt.value === q.aggregation)}
                onChange={onAggregationChange}
              />
            </InlineField>

            <InlineField label="Group By" labelWidth={10}>
              <Select
                width={16}
                options={groupByOptions}
                value={groupByOptions.find((opt) => opt.value === q.groupBy)}
                onChange={onGroupByChange}
              />
            </InlineField>
          </>
        )}

        {showMetricField && (
          <InlineField label="Metric" labelWidth={10}>
            <Select
              width={20}
              options={metricFieldOptions}
              value={metricFieldOptions.find((opt) => opt.value === q.metricField)}
              onChange={onMetricFieldChange}
              placeholder="Select metric"
            />
          </InlineField>
        )}
      </InlineFieldRow>

      {/* Filters */}
      <InlineFieldRow>
        <InlineField label="Name" labelWidth={14} tooltip="Filter by trace or observation name">
          <Input
            width={20}
            value={q.name || ''}
            onChange={onNameChange}
            onBlur={onNameBlur}
            placeholder="(optional)"
          />
        </InlineField>

        <InlineField label="User ID" labelWidth={10} tooltip="Filter by user ID">
          <Input
            width={16}
            value={q.userId || ''}
            onChange={onUserIdChange}
            onBlur={onUserIdBlur}
            placeholder="(optional)"
          />
        </InlineField>

        <InlineField label="Session" labelWidth={10} tooltip="Filter by session ID">
          <Input
            width={16}
            value={q.sessionId || ''}
            onChange={onSessionIdChange}
            onBlur={onSessionIdBlur}
            placeholder="(optional)"
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        {showModelFilter && (
          <InlineField label="Model" labelWidth={14} tooltip="Filter by model name">
            <Input
              width={20}
              value={q.model || ''}
              onChange={onModelChange}
              onBlur={onModelBlur}
              placeholder="e.g., gpt-4"
            />
          </InlineField>
        )}

        <InlineField label="Release" labelWidth={10} tooltip="Filter by release version">
          <Input
            width={16}
            value={q.release || ''}
            onChange={onReleaseChange}
            onBlur={onReleaseBlur}
            placeholder="(optional)"
          />
        </InlineField>

        <InlineField label="Limit" labelWidth={10} tooltip="Maximum results to return">
          <Input
            width={10}
            type="number"
            value={q.limit || 100}
            onChange={onLimitChange}
            onBlur={onLimitBlur}
            min={1}
            max={10000}
          />
        </InlineField>
      </InlineFieldRow>
    </div>
  );
}
