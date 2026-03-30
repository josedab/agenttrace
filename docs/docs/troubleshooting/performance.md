---
sidebar_position: 6
title: Performance Issues
---

# Performance Issues

## High Latency

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

## Memory Usage

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
