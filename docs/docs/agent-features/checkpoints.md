---
sidebar_position: 1
title: Checkpoints
description: Capture file state snapshots during agent execution for debugging and recovery.
---

# Checkpoints

Checkpoints create snapshots of code state during agent execution, enabling you to debug failures and restore previous states.

## Overview

When an AI agent modifies files, checkpoints record file hashes, sizes, and git context at that point in time. This gives you a full timeline of how the codebase evolved during an agent run.

### Checkpoint Types

| Type | Description | Use Case |
|------|-------------|----------|
| `manual` | Explicitly created by code | Before major changes |
| `auto` | Automatically created on file changes | Periodic snapshots |
| `tool_call` | Before/after tool execution | Tool debugging |
| `error` | Created on errors | Error recovery |
| `milestone` | Key progress points | Task completion |

## CLI Usage

The `--checkpoints` flag (used with `--watch`) auto-creates checkpoints when files change:

```bash
agenttrace wrap --watch --checkpoints -- python agent.py

# Full example with git linking
agenttrace wrap \
  --name "refactor-task" \
  --watch \
  --checkpoints \
  --git \
  -- python agent.py
```

Checkpoints are created automatically each time a monitored file is modified, tracking the file state via SHA256 hashes.

## SDK Usage

### Python

```python
from agenttrace import AgentTrace

at = AgentTrace()

with at.trace("code-refactor") as trace:
    # Checkpoint before making changes
    trace.checkpoint(
        name="before-refactor",
        type="manual",
        files=["src/main.py", "src/utils.py"],
        description="State before refactoring utils module"
    )

    # ... perform refactoring ...

    # Milestone checkpoint after completion
    trace.checkpoint(
        name="after-refactor",
        type="milestone",
        files=["src/main.py", "src/utils.py"],
        description="Refactoring complete"
    )
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });
const trace = client.trace({ name: 'code-refactor' });

const cp = trace.checkpoint({
  name: 'before-refactor',
  type: 'manual',
  files: ['src/main.ts', 'src/utils.ts'],
  description: 'State before refactoring utils module'
});

console.log(`Checkpoint: ${cp.id}, Git SHA: ${cp.gitCommitSha}`);

trace.end();
```

## Checkpoint Data

Each checkpoint captures:

| Field | Description |
|-------|-------------|
| `gitCommitSha` | Current git commit SHA |
| `gitBranch` | Current branch name |
| `filesSnapshot` | Map of files with size and SHA256 hash |
| `filesChanged` | List of tracked file paths |
| `totalFiles` | Number of files tracked |
| `totalSizeBytes` | Total size of tracked files |

## Best Practices

1. **Create before major changes** — checkpoint before refactoring or editing multiple files
2. **Use meaningful names** — `before-auth-refactor` is more useful than `checkpoint-1`
3. **Track relevant files only** — include only files that might need restoration
4. **Use appropriate types** — `error` for error states, `milestone` for progress points
