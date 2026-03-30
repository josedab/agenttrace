---
sidebar_position: 8
title: UI Issues
---

# UI Issues

## Dashboard Not Loading

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

## Graphs Not Rendering

**Symptom**: Charts show "No data" despite having traces

**Solutions**:

1. **Check date range filter**: Ensure it covers your trace timestamps
2. **Wait for aggregation**: Analytics may take a few seconds to update
3. **Check project selection**: Ensure you're viewing the correct project
