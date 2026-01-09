import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput, FieldSet } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { AgentTraceDataSourceOptions, AgentTraceSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<AgentTraceDataSourceOptions, AgentTraceSecureJsonData> {}

/**
 * ConfigEditor component for AgentTrace datasource configuration
 */
export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      url: event.target.value,
    });
  };

  const onProjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        projectId: event.target.value,
      },
    });
  };

  const onApiKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        apiKey: event.target.value,
      },
    });
  };

  const onResetApiKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...secureJsonData,
        apiKey: '',
      },
    });
  };

  return (
    <>
      <FieldSet label="Connection">
        <InlineField
          label="URL"
          labelWidth={14}
          tooltip="The URL of your AgentTrace API server"
        >
          <Input
            width={40}
            value={options.url || ''}
            onChange={onUrlChange}
            placeholder="http://localhost:8080"
          />
        </InlineField>
      </FieldSet>

      <FieldSet label="Authentication">
        <InlineField
          label="API Key"
          labelWidth={14}
          tooltip="Your AgentTrace API key (starts with sk-at-)"
        >
          <SecretInput
            width={40}
            isConfigured={secureJsonFields?.apiKey ?? false}
            value={secureJsonData?.apiKey || ''}
            placeholder="sk-at-your-api-key"
            onReset={onResetApiKey}
            onChange={onApiKeyChange}
          />
        </InlineField>

        <InlineField
          label="Project ID"
          labelWidth={14}
          tooltip="Optional: Specify a project ID to filter data"
        >
          <Input
            width={40}
            value={jsonData.projectId || ''}
            onChange={onProjectIdChange}
            placeholder="(optional)"
          />
        </InlineField>
      </FieldSet>

      <FieldSet label="Help">
        <p style={{ marginLeft: '14px', color: '#666' }}>
          To get your API key:
          <ol style={{ marginTop: '8px', paddingLeft: '20px' }}>
            <li>Log in to your AgentTrace dashboard</li>
            <li>Go to Settings &gt; API Keys</li>
            <li>Create a new API key with read permissions</li>
          </ol>
        </p>
        <p style={{ marginLeft: '14px', marginTop: '16px' }}>
          <a
            href="https://docs.agenttrace.io/integrations/grafana"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: '#6E9FFF' }}
          >
            View documentation
          </a>
        </p>
      </FieldSet>
    </>
  );
}
