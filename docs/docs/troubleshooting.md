---
sidebar_position: 50
---

# Troubleshooting

Common issues and solutions when using AgentTrace.

## Connection Issues

### Traces Not Appearing in the UI

**Symptoms**: You're sending traces but they don't show up in the dashboard.

**Solutions**:

1. **Verify AgentTrace is running**:
   ```bash
   curl http://localhost:8080/health
   # Should return: {"status":"ok"}
   ```

2. **Check your API key**:
   ```bash
   curl -X GET "http://localhost:8080/api/public/traces" \
     -H "Authorization: Bearer $AGENTTRACE_API_KEY"
   # Should return traces, not 401 Unauthorized
   ```

3. **Ensure you're flushing data**:
   ```python
   # Python - Always flush before exiting
   client.flush()
   client.shutdown()
   ```
   ```typescript
   // TypeScript - Always await flush
   await client.flush();
   await client.shutdown();
   ```

4. **Check the correct host**:
   ```bash
   # Verify AGENTTRACE_HOST is set correctly
   echo $AGENTTRACE_HOST
   # Should be http://localhost:8080 for local, or your production URL
   ```

5. **Look at SDK logs**:
   ```python
   # Python - Enable debug logging
   import logging
   logging.basicConfig(level=logging.DEBUG)
   ```

### Connection Refused Errors

**Symptoms**: `ConnectionRefusedError` or `ECONNREFUSED`

**Solutions**:

1. **Check if services are running**:
   ```bash
   docker compose ps
   # All services should show "Up"
   ```

2. **Check port bindings**:
   ```bash
   # API should be on 8080
   lsof -i :8080

   # Web UI should be on 3000
   lsof -i :3000
   ```

3. **Restart services**:
   ```bash
   docker compose down
   docker compose up -d
   ```

### SSL/TLS Certificate Errors

**Symptoms**: Certificate verification failed

**Solutions**:

1. **For local development** (not recommended for production):
   ```python
   # Python
   client = AgentTrace(
       api_key="...",
       host="https://localhost:8080",
       verify_ssl=False  # Only for local dev
   )
   ```

2. **For production**: Ensure your SSL certificates are valid and properly configured.

## Authentication Issues

### 401 Unauthorized

**Symptoms**: API returns `{"error": {"code": "unauthorized"}}`

**Solutions**:

1. **Check API key format**:
   ```bash
   # Keys should start with "sk-at-"
   echo $AGENTTRACE_API_KEY
   ```

2. **Verify key is active**: Check in Settings > API Keys in the UI

3. **Check header format**:
   ```bash
   # Correct
   -H "Authorization: Bearer sk-at-your-key"

   # Wrong
   -H "Authorization: sk-at-your-key"
   -H "X-API-Key: sk-at-your-key"
   ```

### 403 Forbidden

**Symptoms**: API returns `{"error": {"code": "forbidden"}}`

**Solutions**:

1. **Check API key scopes**: Ensure the key has required permissions
2. **Verify project access**: The key must belong to the correct project

## SDK Issues

### Python SDK

#### Import Errors

**Symptom**: `ModuleNotFoundError: No module named 'agenttrace'`

**Solution**:
```bash
pip install agenttrace
# or
pip install -e ".[dev]"  # for development
```

#### Async Context Issues

**Symptom**: `RuntimeError: This event loop is already running`

**Solution**:
```python
# Use nest_asyncio for Jupyter notebooks
import nest_asyncio
nest_asyncio.apply()

# Or use the sync client
from agenttrace import AgentTrace
client = AgentTrace(api_key="...", async_mode=False)
```

#### Decorator Not Working

**Symptom**: `@observe()` decorator doesn't create traces

**Solution**:
```python
# Initialize client BEFORE using decorators
from agenttrace import AgentTrace

# This must come first
client = AgentTrace(api_key="...")

# Now decorators will work
@observe()
def my_function():
    pass
```

### TypeScript SDK

#### Type Errors

**Symptom**: TypeScript compilation errors

**Solution**:
```bash
# Ensure types are installed
npm install @types/node

# Check tsconfig.json has correct settings
{
  "compilerOptions": {
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

#### Promise Not Awaited

**Symptom**: Traces are incomplete or missing

**Solution**:
```typescript
// Always await async operations
await client.flush();
await client.shutdown();

// Or use finally block
try {
  // Your code
} finally {
  await client.shutdown();
}
```

### Go SDK

#### Context Cancellation

**Symptom**: Traces are cut off or incomplete

**Solution**:
```go
// Don't cancel context before flushing
ctx := context.Background()
defer func() {
    // Use a fresh context for shutdown
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    client.Shutdown(shutdownCtx)
}()
```

#### Missing Traces

**Symptom**: Some traces don't appear

**Solution**:
```go
// Always defer client shutdown
defer client.Shutdown(ctx)

// Or explicitly flush
if err := client.Flush(ctx); err != nil {
    log.Printf("Failed to flush: %v", err)
}
```

## Integration Issues

### OpenAI Integration

**Symptom**: OpenAI calls not being traced

**Solution**:
```python
from agenttrace import AgentTrace
from agenttrace.integrations.openai import OpenAIInstrumentation

# Initialize AgentTrace FIRST
client = AgentTrace(api_key="...")

# Enable instrumentation BEFORE creating OpenAI client
OpenAIInstrumentation.enable()

# Now create OpenAI client
from openai import OpenAI
openai_client = OpenAI()
```

### LangChain Integration

**Symptom**: Chain executions not appearing as traces

**Solution**:
```python
from agenttrace import AgentTrace
from agenttrace.integrations import LangChainInstrumentation

client = AgentTrace(api_key="...")
LangChainInstrumentation.enable()

# Use callback handler for explicit control
from agenttrace.integrations.langchain import AgentTraceCallbackHandler
handler = AgentTraceCallbackHandler()

chain.invoke(input, config={"callbacks": [handler]})
```

### Anthropic Integration

**Symptom**: Claude API calls not traced

**Solution**:
```python
from agenttrace import AgentTrace
from agenttrace.integrations.anthropic import AnthropicInstrumentation

client = AgentTrace(api_key="...")
AnthropicInstrumentation.enable()

# Create Anthropic client after enabling instrumentation
from anthropic import Anthropic
anthropic_client = Anthropic()
```

## Data Issues

### Missing Token Counts

**Symptom**: Token counts show as 0 or null

**Solutions**:

1. **Provide usage data explicitly**:
   ```python
   generation.end(
       output=response,
       usage={
           "input_tokens": 100,
           "output_tokens": 50,
           "total_tokens": 150
       }
   )
   ```

2. **Use auto-instrumentation** which captures usage automatically

### Incorrect Costs

**Symptom**: Costs don't match expected values

**Solutions**:

1. **Check model name**: Ensure the model name matches our pricing database
   ```python
   # Correct
   generation = trace.generation(model="gpt-4")

   # May not be recognized
   generation = trace.generation(model="my-custom-gpt4")
   ```

2. **Costs recalculate in background**: Wait a few seconds and refresh

### Large Payloads Truncated

**Symptom**: Input/output data appears cut off

**Solution**:
```python
# Summarize large data before logging
import json

large_response = call_api()
summary = {
    "length": len(large_response),
    "preview": large_response[:1000],
    "keys": list(large_response.keys()) if isinstance(large_response, dict) else None
}
span.end(output=summary)
```

## Performance Issues

### High Latency

**Symptom**: Tracing adds noticeable latency

**Solutions**:

1. **Enable batching**:
   ```python
   client = AgentTrace(
       api_key="...",
       flush_interval=5.0,  # Batch for 5 seconds
       max_batch_size=100   # Or until 100 events
   )
   ```

2. **Use async mode**:
   ```python
   client = AgentTrace(api_key="...", async_mode=True)
   ```

3. **Sample traces** for high-volume applications:
   ```python
   import random

   if random.random() < 0.1:  # 10% sampling
       with client.trace(name="my-trace"):
           # Your code
   ```

### Memory Usage

**Symptom**: Application memory grows over time

**Solutions**:

1. **Flush regularly**:
   ```python
   # In long-running applications
   if trace_count % 100 == 0:
       client.flush()
   ```

2. **Avoid storing trace references**:
   ```python
   # Bad - holds references
   traces = []
   for i in range(10000):
       traces.append(client.trace(name=f"trace-{i}"))

   # Good - let traces be garbage collected
   for i in range(10000):
       with client.trace(name=f"trace-{i}"):
           # Your code
   ```

## Docker/Deployment Issues

### ClickHouse Connection Failed

**Symptom**: API logs show ClickHouse connection errors

**Solutions**:

1. **Check ClickHouse health**:
   ```bash
   curl http://localhost:8123/ping
   # Should return "Ok."
   ```

2. **Verify credentials**:
   ```bash
   # Check environment variables
   echo $CLICKHOUSE_USER
   echo $CLICKHOUSE_PASSWORD
   ```

3. **Check network connectivity**:
   ```bash
   docker compose exec api ping clickhouse
   ```

### PostgreSQL Connection Issues

**Symptom**: "connection refused" to PostgreSQL

**Solutions**:

1. **Wait for PostgreSQL to be ready**:
   ```bash
   # PostgreSQL takes a few seconds to start
   docker compose logs postgres
   # Look for "database system is ready to accept connections"
   ```

2. **Check credentials match**:
   ```yaml
   # In docker-compose.yml
   postgres:
     environment:
       POSTGRES_USER: agenttrace
       POSTGRES_PASSWORD: agenttrace  # Must match API config
   ```

### Migrations Not Applied

**Symptom**: API errors about missing tables

**Solution**:
```bash
# Run migrations manually
cd api
make migrate-pg-up
make migrate-ch-up
```

## UI Issues

### Dashboard Not Loading

**Symptom**: Blank page or loading spinner

**Solutions**:

1. **Check browser console** for JavaScript errors

2. **Clear browser cache**:
   ```
   Ctrl+Shift+Delete (Windows/Linux)
   Cmd+Shift+Delete (Mac)
   ```

3. **Check API connectivity**:
   ```bash
   curl http://localhost:8080/health
   ```

### Graphs Not Rendering

**Symptom**: Charts show "No data" despite having traces

**Solutions**:

1. **Check date range filter**: Ensure it covers your trace timestamps
2. **Wait for aggregation**: Analytics may take a few seconds to update
3. **Check project selection**: Ensure you're viewing the correct project

## Getting Help

If you're still stuck:

1. **Check logs**:
   ```bash
   # API logs
   docker compose logs api

   # Worker logs
   docker compose logs worker

   # All logs
   docker compose logs -f
   ```

2. **Enable debug mode**:
   ```bash
   export AGENTTRACE_DEBUG=true
   ```

3. **Search existing issues**: [GitHub Issues](https://github.com/agenttrace/agenttrace/issues)

4. **Join the community**: [Discord](https://discord.gg/agenttrace)

5. **File a bug report** with:
   - AgentTrace version
   - SDK version
   - Reproduction steps
   - Error messages and logs
