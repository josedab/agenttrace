---
sidebar_position: 5
title: Data Issues
---

# Data Issues

## Missing Token Counts

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

## Incorrect Costs

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

## Large Payloads Truncated

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
