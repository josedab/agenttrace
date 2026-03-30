---
sidebar_position: 3
title: SDK Issues
---

# SDK Issues

## Python SDK

### Import Errors

**Symptom**: `ModuleNotFoundError: No module named 'agenttrace'`

**Solution**:
```bash
pip install agenttrace
# or
pip install -e ".[dev]"  # for development
```

### Async Context Issues

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

### Decorator Not Working

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

## TypeScript SDK

### Type Errors

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

### Promise Not Awaited

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

## Go SDK

### Context Cancellation

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

### Missing Traces

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
