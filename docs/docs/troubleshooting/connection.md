---
sidebar_position: 1
title: Connection Issues
---

# Connection Issues

## Traces Not Appearing in the UI

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

## Connection Refused Errors

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

## SSL/TLS Certificate Errors

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
