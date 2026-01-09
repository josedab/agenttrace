# OpenTelemetry Integration

AgentTrace provides full OpenTelemetry Protocol (OTLP) support, enabling you to:
- **Export traces** from AgentTrace to external observability backends (Jaeger, Zipkin, Grafana Tempo, Datadog, Honeycomb, New Relic)
- **Receive traces** from OpenTelemetry-instrumented applications

This makes AgentTrace compatible with the industry-standard observability ecosystem.

## Overview

```
┌─────────────────┐     OTLP      ┌─────────────────┐     OTLP      ┌─────────────────┐
│  Your App with  │ ────────────► │   AgentTrace    │ ────────────► │  External       │
│  OTel SDK       │               │   (Receiver)    │               │  Backend        │
└─────────────────┘               └─────────────────┘               └─────────────────┘
                                         │
                                         ▼
                                  ┌─────────────────┐
                                  │  AgentTrace     │
                                  │  Web UI         │
                                  └─────────────────┘
```

## OTLP Receiver (Ingesting Traces)

AgentTrace can receive traces from any OpenTelemetry-instrumented application via HTTP or gRPC.

### Configuration

Enable the OTLP receiver in your AgentTrace configuration:

```yaml
otel:
  receiver_enabled: true
  receiver_http_port: 4318
  receiver_http_path: "/v1/traces"
  receiver_grpc_port: 4317
```

Or via environment variables:

```bash
AGENTTRACE_OTEL_RECEIVER_ENABLED=true
AGENTTRACE_OTEL_RECEIVER_HTTP_PORT=4318
AGENTTRACE_OTEL_RECEIVER_HTTP_PATH=/v1/traces
AGENTTRACE_OTEL_RECEIVER_GRPC_PORT=4317
```

### Sending Traces to AgentTrace

#### Python with OpenTelemetry

```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Configure the OTLP exporter to send to AgentTrace
exporter = OTLPSpanExporter(
    endpoint="http://localhost:4318/v1/traces",
    headers={
        "Authorization": "Bearer your-api-key",
        "X-Project-ID": "your-project-id"
    }
)

# Set up the tracer provider
provider = TracerProvider()
provider.add_span_processor(BatchSpanProcessor(exporter))
trace.set_tracer_provider(provider)

# Create spans
tracer = trace.get_tracer(__name__)
with tracer.start_as_current_span("my-operation") as span:
    span.set_attribute("gen_ai.system", "openai")
    span.set_attribute("gen_ai.request.model", "gpt-4")
    # Your code here
```

#### TypeScript/Node.js with OpenTelemetry

```typescript
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { trace } from '@opentelemetry/api';

// Configure the OTLP exporter
const exporter = new OTLPTraceExporter({
  url: 'http://localhost:4318/v1/traces',
  headers: {
    'Authorization': 'Bearer your-api-key',
    'X-Project-ID': 'your-project-id'
  }
});

// Set up the tracer provider
const provider = new NodeTracerProvider();
provider.addSpanProcessor(new BatchSpanProcessor(exporter));
provider.register();

// Create spans
const tracer = trace.getTracer('my-service');
const span = tracer.startSpan('my-operation');
span.setAttribute('gen_ai.system', 'anthropic');
span.setAttribute('gen_ai.request.model', 'claude-3-opus');
// Your code here
span.end();
```

#### Go with OpenTelemetry

```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
    ctx := context.Background()

    // Create OTLP exporter
    exporter, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint("localhost:4318"),
        otlptracehttp.WithHeaders(map[string]string{
            "Authorization": "Bearer your-api-key",
            "X-Project-ID":  "your-project-id",
        }),
    )
    if err != nil {
        panic(err)
    }

    // Create tracer provider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
    )
    otel.SetTracerProvider(tp)

    // Create spans
    tracer := otel.Tracer("my-service")
    ctx, span := tracer.Start(ctx, "my-operation")
    span.SetAttributes(
        attribute.String("gen_ai.system", "openai"),
        attribute.String("gen_ai.request.model", "gpt-4"),
    )
    // Your code here
    span.End()
}
```

### LLM Semantic Conventions

AgentTrace recognizes OpenTelemetry semantic conventions for LLM observability:

| Attribute | Description | Example |
|-----------|-------------|---------|
| `gen_ai.system` | The LLM provider | `openai`, `anthropic`, `cohere` |
| `gen_ai.request.model` | Requested model name | `gpt-4`, `claude-3-opus` |
| `gen_ai.response.model` | Actual model used | `gpt-4-0613` |
| `gen_ai.request.max_tokens` | Max tokens requested | `1000` |
| `gen_ai.request.temperature` | Temperature setting | `0.7` |
| `gen_ai.request.top_p` | Top-p setting | `0.9` |
| `gen_ai.usage.input_tokens` | Input token count | `150` |
| `gen_ai.usage.output_tokens` | Output token count | `500` |
| `gen_ai.response.finish_reasons` | Completion reason | `stop`, `length` |

AgentTrace also adds custom attributes:

| Attribute | Description |
|-----------|-------------|
| `agenttrace.trace.id` | AgentTrace trace ID |
| `agenttrace.span.id` | AgentTrace span ID |
| `agenttrace.project.id` | Project ID |
| `agenttrace.trace.name` | Trace name |
| `agenttrace.span.type` | Span type (generation, span, event) |
| `agenttrace.cost` | Computed cost in USD |
| `agenttrace.latency_ms` | Latency in milliseconds |

## OTLP Exporter (Sending Traces)

Export traces from AgentTrace to external observability backends.

### Creating an Exporter

#### Via API

```bash
curl -X POST "https://api.agenttrace.io/v1/projects/{projectId}/otel/exporters" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Jaeger",
    "type": "http",
    "endpoint": "https://jaeger.example.com:4318/v1/traces",
    "enabled": true,
    "headers": {
      "Authorization": "Bearer jaeger-api-key"
    },
    "batchConfig": {
      "maxBatchSize": 512,
      "maxQueueSize": 2048,
      "batchTimeoutMs": 5000,
      "exportTimeoutMs": 30000
    },
    "retryConfig": {
      "enabled": true,
      "initialIntervalMs": 1000,
      "maxIntervalMs": 30000,
      "maxElapsedTimeMs": 300000,
      "multiplier": 1.5
    },
    "samplingRate": 1.0,
    "resourceAttributes": {
      "deployment.environment": "production",
      "service.namespace": "ai-agents"
    }
  }'
```

#### Via Web UI

1. Navigate to **Project Settings** > **Integrations** > **OpenTelemetry**
2. Click **Add Exporter**
3. Configure the exporter settings
4. Test the connection
5. Enable the exporter

### Supported Backends

AgentTrace includes pre-configured templates for popular backends:

#### Jaeger

```json
{
  "name": "Jaeger",
  "type": "http",
  "endpoint": "http://jaeger:4318/v1/traces",
  "compression": "gzip"
}
```

#### Grafana Tempo

```json
{
  "name": "Grafana Tempo",
  "type": "http",
  "endpoint": "http://tempo:4318/v1/traces",
  "compression": "gzip"
}
```

#### Datadog

```json
{
  "name": "Datadog",
  "type": "http",
  "endpoint": "https://trace.agent.datadoghq.com/v1/traces",
  "headers": {
    "DD-API-KEY": "${DD_API_KEY}"
  }
}
```

#### Honeycomb

```json
{
  "name": "Honeycomb",
  "type": "http",
  "endpoint": "https://api.honeycomb.io/v1/traces",
  "headers": {
    "x-honeycomb-team": "${HONEYCOMB_API_KEY}",
    "x-honeycomb-dataset": "agenttrace"
  }
}
```

#### New Relic

```json
{
  "name": "New Relic",
  "type": "http",
  "endpoint": "https://otlp.nr-data.net:4318/v1/traces",
  "headers": {
    "api-key": "${NEW_RELIC_LICENSE_KEY}"
  }
}
```

### Exporter Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | required | Exporter display name |
| `type` | string | `http` | Transport type (`http` or `grpc`) |
| `endpoint` | string | required | OTLP endpoint URL |
| `enabled` | boolean | `true` | Whether exporter is active |
| `headers` | object | `{}` | Custom headers (supports `${ENV_VAR}`) |
| `compression` | string | `none` | Compression (`gzip` or `none`) |
| `timeoutSeconds` | int | `30` | Request timeout |
| `insecure` | boolean | `false` | Skip TLS verification |
| `samplingRate` | float | `1.0` | Percentage of traces to export (0-1) |

#### Batch Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `maxBatchSize` | int | `512` | Max spans per batch |
| `maxQueueSize` | int | `2048` | Max spans in queue |
| `batchTimeoutMs` | int | `5000` | Max wait before sending |
| `exportTimeoutMs` | int | `30000` | Export request timeout |
| `scheduleDelayMs` | int | `1000` | Delay between batch checks |

#### Retry Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable retry on failure |
| `initialIntervalMs` | int | `1000` | Initial retry delay |
| `maxIntervalMs` | int | `30000` | Maximum retry delay |
| `maxElapsedTimeMs` | int | `300000` | Max total retry time |
| `multiplier` | float | `1.5` | Backoff multiplier |

### Filtering Traces

Control which traces are exported:

```json
{
  "traceNameFilter": "^prod-.*",
  "metadataFilters": {
    "environment": "production",
    "team": "ml-platform"
  },
  "samplingRate": 0.1
}
```

### TLS Configuration

For secure connections:

```json
{
  "tlsConfig": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key",
    "caFile": "/path/to/ca.crt",
    "serverName": "otlp.example.com"
  },
  "insecure": false
}
```

## API Reference

### Exporter Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/v1/projects/{id}/otel/exporters` | List all exporters |
| `POST` | `/v1/projects/{id}/otel/exporters` | Create exporter |
| `GET` | `/v1/projects/{id}/otel/exporters/{exporterId}` | Get exporter |
| `PUT` | `/v1/projects/{id}/otel/exporters/{exporterId}` | Update exporter |
| `DELETE` | `/v1/projects/{id}/otel/exporters/{exporterId}` | Delete exporter |
| `POST` | `/v1/projects/{id}/otel/exporters/{exporterId}/toggle` | Enable/disable |
| `POST` | `/v1/projects/{id}/otel/exporters/{exporterId}/test` | Test connection |
| `GET` | `/v1/projects/{id}/otel/exporters/{exporterId}/stats` | Get statistics |

### Receiver Endpoint

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/traces` | Receive OTLP traces (HTTP) |

### Get Default Configuration

```bash
GET /v1/otel/config/defaults
```

Response:
```json
{
  "batchConfig": {
    "maxBatchSize": 512,
    "maxQueueSize": 2048,
    "batchTimeoutMs": 5000,
    "exportTimeoutMs": 30000,
    "scheduleDelayMs": 1000
  },
  "retryConfig": {
    "enabled": true,
    "initialIntervalMs": 1000,
    "maxIntervalMs": 30000,
    "maxElapsedTimeMs": 300000,
    "multiplier": 1.5
  }
}
```

### Get Supported Backends

```bash
GET /v1/otel/backends
```

Response:
```json
{
  "backends": [
    {
      "name": "Jaeger",
      "description": "Open-source distributed tracing",
      "defaultEndpoint": "http://jaeger:4318/v1/traces",
      "docsUrl": "https://www.jaegertracing.io/docs/"
    },
    {
      "name": "Grafana Tempo",
      "description": "Distributed tracing backend for Grafana",
      "defaultEndpoint": "http://tempo:4318/v1/traces",
      "docsUrl": "https://grafana.com/docs/tempo/"
    }
  ]
}
```

## Monitoring Exporters

### Exporter Statistics

View export metrics via the API:

```bash
GET /v1/projects/{id}/otel/exporters/{exporterId}/stats
```

Response:
```json
{
  "exporterId": "abc123",
  "exporterName": "Production Jaeger",
  "totalExported": 125000,
  "totalErrors": 12,
  "lastExportAt": "2024-01-15T10:30:00Z",
  "lastErrorAt": "2024-01-15T08:15:00Z",
  "avgLatencyMs": 45.2,
  "exportsLast24h": 8500,
  "errorsLast24h": 2
}
```

### Health Checks

AgentTrace automatically runs health checks on configured exporters every 5 minutes. Failed exporters are marked with `status: "error"` and can be investigated via logs.

## Best Practices

### 1. Use Sampling for High-Volume Workloads

```json
{
  "samplingRate": 0.1
}
```

Export only 10% of traces to reduce costs on external backends.

### 2. Add Resource Attributes

```json
{
  "resourceAttributes": {
    "service.name": "my-agent",
    "service.version": "1.0.0",
    "deployment.environment": "production"
  }
}
```

### 3. Enable Compression

```json
{
  "compression": "gzip"
}
```

Reduce bandwidth usage by 60-80%.

### 4. Configure Appropriate Timeouts

```json
{
  "timeoutSeconds": 30,
  "batchConfig": {
    "exportTimeoutMs": 30000
  }
}
```

### 5. Use Environment Variables for Secrets

```json
{
  "headers": {
    "Authorization": "Bearer ${API_KEY}"
  }
}
```

AgentTrace will expand `${VAR_NAME}` from environment variables.

## Troubleshooting

### Traces Not Appearing

1. **Check exporter status**: Ensure the exporter is enabled and healthy
2. **Verify endpoint**: Test the endpoint is reachable
3. **Check authentication**: Verify API keys and headers
4. **Review logs**: Check AgentTrace logs for export errors

### High Error Rate

1. **Check backend availability**: Ensure the target backend is running
2. **Increase timeouts**: The backend may be slow to respond
3. **Enable retry**: Configure retry settings for transient failures
4. **Reduce batch size**: Smaller batches may succeed more reliably

### Missing Attributes

1. **Check semantic conventions**: Use standard OTel attribute names
2. **Verify resource attributes**: Ensure they're configured on the exporter
3. **Check sampling**: Traces may be sampled out

## Examples

### Complete Python Example with LLM Tracing

```python
import openai
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource

# Configure resource
resource = Resource.create({
    "service.name": "my-llm-agent",
    "service.version": "1.0.0",
    "deployment.environment": "development"
})

# Configure exporter
exporter = OTLPSpanExporter(
    endpoint="http://localhost:4318/v1/traces",
    headers={
        "Authorization": "Bearer your-api-key",
        "X-Project-ID": "your-project-id"
    }
)

# Set up tracer
provider = TracerProvider(resource=resource)
provider.add_span_processor(BatchSpanProcessor(exporter))
trace.set_tracer_provider(provider)
tracer = trace.get_tracer(__name__)

def call_openai(prompt: str) -> str:
    with tracer.start_as_current_span("openai-completion") as span:
        # Set LLM attributes before the call
        span.set_attribute("gen_ai.system", "openai")
        span.set_attribute("gen_ai.request.model", "gpt-4")
        span.set_attribute("gen_ai.request.max_tokens", 1000)
        span.set_attribute("gen_ai.request.temperature", 0.7)

        # Make the API call
        response = openai.chat.completions.create(
            model="gpt-4",
            messages=[{"role": "user", "content": prompt}],
            max_tokens=1000,
            temperature=0.7
        )

        # Set response attributes
        span.set_attribute("gen_ai.response.model", response.model)
        span.set_attribute("gen_ai.usage.input_tokens", response.usage.prompt_tokens)
        span.set_attribute("gen_ai.usage.output_tokens", response.usage.completion_tokens)
        span.set_attribute("gen_ai.response.finish_reasons", response.choices[0].finish_reason)

        return response.choices[0].message.content

# Use the instrumented function
with tracer.start_as_current_span("agent-task") as parent:
    parent.set_attribute("task.type", "code-generation")
    result = call_openai("Write a Python function to calculate fibonacci numbers")
    print(result)
```

This trace will appear in both AgentTrace and any configured external backend with full LLM observability attributes.
