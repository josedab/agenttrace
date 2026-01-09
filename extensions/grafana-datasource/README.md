# AgentTrace Grafana Datasource

A Grafana datasource plugin for visualizing AgentTrace metrics, traces, costs, and performance data.

## Features

- **Metrics Visualization**: Track trace counts, costs, tokens, and latency over time
- **Trace Explorer**: Browse individual traces with filtering and search
- **Observation Analysis**: Analyze LLM calls, spans, and events
- **Cost Monitoring**: Daily and hourly cost breakdowns by model
- **Score Tracking**: Monitor evaluation scores and quality metrics
- **Alerting Support**: Create alerts based on cost, latency, or error thresholds
- **Variable Support**: Use template variables for dynamic dashboards

## Installation

### From Grafana Plugin Catalog

1. Go to **Configuration** > **Plugins**
2. Search for "AgentTrace"
3. Click **Install**

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/agenttrace/agenttrace/releases)
2. Extract to your Grafana plugins directory:
   ```bash
   unzip agenttrace-datasource-1.0.0.zip -d /var/lib/grafana/plugins/
   ```
3. Restart Grafana:
   ```bash
   sudo systemctl restart grafana-server
   ```

### Docker

Mount the plugin directory when running Grafana:

```yaml
services:
  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./extensions/grafana-datasource:/var/lib/grafana/plugins/agenttrace-datasource
    environment:
      - GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=agenttrace-datasource
```

## Configuration

1. Go to **Configuration** > **Data Sources**
2. Click **Add data source**
3. Search for "AgentTrace"
4. Configure the connection:
   - **URL**: Your AgentTrace API URL (e.g., `http://localhost:8080`)
   - **API Key**: Your AgentTrace API key (starts with `sk-at-`)
   - **Project ID** (optional): Filter to a specific project

## Query Types

### Metrics

Aggregated metrics over time with support for:
- **Aggregations**: Count, Sum, Avg, Min, Max, P50, P95, P99
- **Group By**: Model, Name, User ID, Session ID, Release, Version, Tag
- **Metric Fields**: Trace Count, Total Cost, Total Tokens, Latency, etc.

Example: Track average latency by model
```
Query Type: Metrics
Aggregation: Avg
Group By: Model
Metric: Latency
```

### Traces

Individual trace data including:
- Trace ID, Name, Start Time
- Latency, Cost, Token counts
- User ID, Session ID
- Metadata and tags

### Observations

LLM calls, spans, and events:
- Observation type (SPAN, GENERATION, EVENT)
- Model information
- Token usage (prompt/completion)
- Calculated costs

### Scores

Evaluation scores:
- Score name and value
- Source (API, Annotation, Eval)
- Associated trace/observation

### Costs

Daily cost breakdown:
- Total cost per day
- Trace count
- Model-level breakdown

## Template Variables

Use these queries in Grafana variables:

| Query | Description |
|-------|-------------|
| `models` | List of models used |
| `users` | List of user IDs |
| `sessions` | List of session IDs |
| `names` | List of trace names |
| `releases` | List of release versions |
| `tags` | List of tags |

## Example Dashboards

### Cost Dashboard

Track LLM spending:
- Total cost over time (Metrics: Sum, Field: totalCost)
- Cost by model (Metrics: Sum, Group By: Model)
- Daily cost breakdown (Costs query type)
- Top users by cost

### Performance Dashboard

Monitor latency and throughput:
- Request count over time
- P95 latency by trace name
- Error rate trending
- Tokens per request

### Quality Dashboard

Track evaluation metrics:
- Score distributions
- Score trends over time
- Pass/fail rates by evaluator

## Alerting

Create alerts on AgentTrace metrics:

1. Go to **Alerting** > **Alert rules**
2. Create a new rule using AgentTrace datasource
3. Example conditions:
   - Alert when daily cost exceeds $100
   - Alert when P95 latency > 5000ms
   - Alert when error rate > 5%

## Development

### Prerequisites

- Node.js 18+
- Yarn or npm
- Grafana 9.0+

### Setup

```bash
cd extensions/grafana-datasource
npm install
npm run dev
```

### Build

```bash
npm run build
```

### Test

```bash
npm run test
```

## Troubleshooting

### Plugin not loading

1. Check Grafana logs: `journalctl -u grafana-server -f`
2. Ensure plugin is in the correct directory
3. For unsigned plugins, add to `grafana.ini`:
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = agenttrace-datasource
   ```

### Connection errors

1. Verify AgentTrace API is running: `curl http://localhost:8080/health`
2. Check API key is valid and has read permissions
3. Ensure no firewall blocking the connection

### No data appearing

1. Check the time range matches your trace data
2. Verify filters aren't too restrictive
3. Test the connection with "Test" button in datasource config

## License

Apache License 2.0

## Links

- [AgentTrace Documentation](https://docs.agenttrace.io)
- [Grafana Integration Guide](https://docs.agenttrace.io/integrations/grafana)
- [GitHub Repository](https://github.com/agenttrace/agenttrace)
