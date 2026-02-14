---
sidebar_position: 3
title: File Operations
description: Track file create, read, write, and delete operations during agent runs.
---

# File Operations

File operation tracking records every file interaction during an agent run, giving you a clear picture of what the agent read, created, modified, or deleted.

## Overview

Each tracked operation includes the file path, operation type, content hash, line diff counts, and the tool that made the change. Operations appear in the trace timeline in the dashboard, making it easy to review agent behavior.

### Operation Types

| Type | Description |
|------|-------------|
| `create` | New file created |
| `read` | File read |
| `update` | File modified |
| `delete` | File deleted |
| `rename` | File renamed |
| `move` | File moved |

## CLI Auto-Capture

When using `--watch`, the CLI automatically captures file operations in the working directory:

```bash
agenttrace wrap --watch -- python agent.py

# Combined with checkpoints and git linking
agenttrace wrap --watch --checkpoints --git -- python agent.py
```

The file watcher monitors the working directory recursively, skipping hidden files and build artifacts (`.pyc`, `.swp`). Detected changes are logged to the trace as file operations.

## SDK Usage

### Python

```python
from agenttrace import AgentTrace

at = AgentTrace()

with at.trace("code-edit") as trace:
    # Track a file update
    trace.file_op(
        operation="update",
        file_path="src/main.py",
        content_before=old_content,
        content_after=new_content,
        tool_name="edit",
        reason="Adding error handling"
    )

    # Track file creation
    trace.file_op(
        operation="create",
        file_path="src/utils.py",
        content_after=new_file_content,
        tool_name="write"
    )
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });
const trace = client.trace({ name: 'code-edit' });

const op = trace.fileOp({
  operation: 'update',
  filePath: 'src/main.ts',
  contentBefore: oldContent,
  contentAfter: newContent,
  toolName: 'edit',
  reason: 'Adding error handling'
});

console.log(`Lines added: ${op.linesAdded}, removed: ${op.linesRemoved}`);

trace.end();
```

## File Operation Data

Each operation captures:

| Field | Description |
|-------|-------------|
| `filePath` | Path to the file |
| `newPath` | New path (for rename/move) |
| `fileSize` | File size in bytes |
| `contentHash` | SHA256 hash of content |
| `linesAdded` | Lines added (auto-calculated) |
| `linesRemoved` | Lines removed (auto-calculated) |
| `diffPreview` | Optional diff preview |
| `toolName` | Tool that made the change |
| `reason` | Why the change was made |
| `success` | Whether operation succeeded |
| `durationMs` | Operation duration |

## Viewing in the Dashboard

File operations appear in the **Trace Timeline** view. Each operation shows:

- The file path and operation type
- A diff preview for updates
- Lines added/removed counts
- The tool and reason, if provided

Use the file operations panel to understand which files an agent touched and in what order.

## Best Practices

1. **Track all modifications** — log every read, write, edit, and delete for a complete audit trail
2. **Include content hashes** — enables verification without storing full file content
3. **Add tool context** — include which tool made the change and why for easier debugging
4. **Calculate diffs** — auto-calculate lines added/removed for metrics and dashboards
