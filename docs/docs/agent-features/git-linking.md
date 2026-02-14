---
sidebar_position: 2
title: Git Linking
description: Link traces to git commits for trace-to-code correlation and commit-level traceability.
---

# Git Linking

Git linking correlates agent traces with git repository state, enabling commit-level traceability so you can answer "which agent run produced this commit?"

## Overview

When enabled, git links capture the branch, commit SHA, author info, and changed files at key points during agent execution. The AgentTrace dashboard displays this commit information alongside traces for easy navigation.

### Link Types

| Type | Description | Use Case |
|------|-------------|----------|
| `start` | Git state at trace start | Baseline tracking |
| `commit` | Link to a specific commit | Commit attribution |
| `branch` | Branch change detected | Context switching |
| `diff` | Uncommitted changes | Work in progress |

## CLI Usage

Use `--git` to auto-detect and link git state:

```bash
agenttrace wrap --git -- python agent.py

# Trace a git commit — the new commit is automatically linked
agenttrace wrap --git -- git commit -m "feat: add auth module"
```

The CLI automatically captures the current branch, commit SHA, repository root, and dirty/clean status. It also detects new commits and branch changes during execution.

## SDK Usage

### Python

```python
from agenttrace import AgentTrace

at = AgentTrace()

with at.trace("code-generation") as trace:
    # Auto-detect git state at start
    link = trace.git_link(type="start")
    print(f"Branch: {link.branch}, Commit: {link.commit_sha}")

    # ... agent generates code and commits ...

    # Link to the new commit
    commit_link = trace.git_link(
        type="commit",
        commit_message="feat: add user authentication"
    )
    print(f"New commit: {commit_link.commit_sha}")
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });
const trace = client.trace({ name: 'code-generation' });

// Auto-detect current git state
const link = trace.gitLink({ type: 'start' });
console.log(`Branch: ${link.branch}, Commit: ${link.commitSha}`);

// Link after committing
const commitLink = trace.gitLink({
  type: 'commit',
  commitMessage: 'feat: add user authentication'
});

trace.end();
```

## Auto-Detection

When auto-detect is enabled (the default), the SDK gathers:

- Current commit SHA via `git rev-parse HEAD`
- Current branch via `git rev-parse --abbrev-ref HEAD`
- Remote URL via `git config --get remote.origin.url`
- Changed files via `git diff --name-only`
- Author info via `git log -1 --format=%an/%ae`

You can also link manually by passing `commit_sha`, `branch`, and `repo_url` directly.

## Git Link Data

| Field | Description |
|-------|-------------|
| `commitSha` | Git commit SHA |
| `branch` | Branch name |
| `repoUrl` | Repository URL |
| `commitMessage` | Commit message |
| `authorName` | Commit author |
| `filesChanged` | List of changed files |
| `additions` | Lines added |
| `deletions` | Lines deleted |

## Best Practices

1. **Link at boundaries** — create links at trace start, after commits, and at branch changes
2. **Use auto-detect** — let the SDK gather git info automatically when possible
3. **Include commit context** — add commit messages and file lists for richer dashboard views
