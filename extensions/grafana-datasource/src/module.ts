import { DataSourcePlugin } from '@grafana/data';
import { AgentTraceDataSource } from './datasource/datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { AgentTraceQuery, AgentTraceDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<AgentTraceDataSource, AgentTraceQuery, AgentTraceDataSourceOptions>(
  AgentTraceDataSource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
