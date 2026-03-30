---
sidebar_position: 2
title: Authentication Issues
---

# Authentication Issues

## 401 Unauthorized

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

## 403 Forbidden

**Symptoms**: API returns `{"error": {"code": "forbidden"}}`

**Solutions**:

1. **Check API key scopes**: Ensure the key has required permissions
2. **Verify project access**: The key must belong to the correct project
