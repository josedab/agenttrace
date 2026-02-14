---
sidebar_position: 4
title: Terminal Commands
description: Log terminal commands executed by agents, including exit codes and stdout/stderr output.
---

# Terminal Commands

Terminal command tracking logs every shell command an agent executes, capturing the command, arguments, exit code, and output. This provides full execution visibility for debugging and auditing.

## Overview

Each logged command includes the working directory, exit code, stdout/stderr (automatically truncated), and execution duration. You can review the full command history in the trace detail view on the dashboard.

## CLI Auto-Capture

The CLI wrapper automatically captures subprocess commands when tracing agent execution:

```bash
agenttrace wrap --name "build-task" -- python agent.py

# Disable output capture if needed
agenttrace wrap --capture-stdout=false --capture-stderr=false -- python agent.py
```

Output is automatically truncated to 10KB to prevent excessive storage. Both stdout and stderr are captured by default.

## SDK Usage

### Python

```python
from agenttrace import AgentTrace

at = AgentTrace()

with at.trace("build-project") as trace:
    # Log a completed command
    trace.terminal_cmd(
        command="npm",
        args=["install"],
        exit_code=0,
        stdout="added 150 packages...",
        working_directory="/project"
    )

    # Run and track a command with timeout
    result = trace.run_cmd(
        "pytest",
        args=["tests/", "-v"],
        timeout=120.0,
        max_output_bytes=50000
    )
    print(f"Tests passed: {result.exit_code == 0}")
```

### TypeScript

```typescript
import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({ apiKey: 'at-your-api-key' });
const trace = client.trace({ name: 'build-project' });

// Log a completed command
trace.terminalCmd({
  command: 'npm',
  args: ['test'],
  exitCode: 0,
  stdout: 'All tests passed',
  stderr: '',
  workingDirectory: '/project'
});

// Run and track a command
const result = await trace.runCmd('npm', {
  args: ['test'],
  timeout: 60000,
  maxOutputBytes: 10000
});
console.log(`Exit code: ${result.exitCode}`);

trace.end();
```

## Terminal Command Data

Each command captures:

| Field | Description |
|-------|-------------|
| `command` | Command name |
| `args` | Command arguments |
| `workingDirectory` | Execution directory |
| `exitCode` | Process exit code |
| `stdout` | Standard output (truncated) |
| `stderr` | Standard error (truncated) |
| `stdoutTruncated` | Whether stdout was truncated |
| `stderrTruncated` | Whether stderr was truncated |
| `timedOut` | Whether command timed out |
| `killed` | Whether process was killed |
| `durationMs` | Execution duration |

## Viewing Command History

In the dashboard, navigate to a trace and open the **Commands** tab to see:

- Each command with its exit code (color-coded: green for 0, red for non-zero)
- Expandable stdout/stderr output
- Execution duration and timeout status
- Working directory context

Failed commands are highlighted, making it easy to spot build or test failures during an agent run.

## Best Practices

1. **Set timeouts** — prevent runaway commands from blocking agent execution
2. **Limit output size** — use `max_output_bytes` to truncate large outputs and avoid storage bloat
3. **Track failures** — log failed commands with stderr for post-run debugging
4. **Include context** — add `tool_name` and `reason` for richer audit trails
