# AgentTrace CLI

Command-line tool for wrapping any process with automatic AgentTrace tracing.

## Installation

```bash
go install github.com/agenttrace/agenttrace/sdk/cli@latest
```

Or download a pre-built binary from [Releases](https://github.com/agenttrace/agenttrace/releases).

## Quick Start

```bash
# Wrap any command for automatic tracing
agenttrace wrap --name "my-agent" -- python my_agent.py

# With git correlation
agenttrace wrap --name "coding-agent" --git -- npm run dev

# With file change tracking and checkpoints
agenttrace wrap --name "agent" --watch --checkpoints -- ./run.sh
```

## Features

- **Zero-code tracing** — Wrap any CLI tool without modifying its source
- **File watching** — Track file create/modify/delete during execution
- **Git linking** — Automatic commit correlation
- **Checkpoints** — Snapshot file state at intervals
- **Output capture** — Optionally capture stdout/stderr
- **Signal handling** — Clean shutdown on SIGINT/SIGTERM

## Flags

| Flag | Description |
|------|-------------|
| `--name` | Custom trace name |
| `--user-id` | Associate with a user |
| `--session-id` | Group in a session |
| `--tags` | Comma-separated tags |
| `--git` | Enable git commit linking |
| `--watch` | Watch for file changes |
| `--checkpoints` | Create checkpoints on file changes |
| `--capture-stdout` | Capture stdout output |
| `--capture-stderr` | Capture stderr output |

## Documentation

See the [full documentation](https://docs.agenttrace.io/sdks/cli).

## License

MIT
