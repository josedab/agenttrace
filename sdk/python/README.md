# AgentTrace Python SDK

[![PyPI](https://img.shields.io/pypi/v/agenttrace)](https://pypi.org/project/agenttrace/)
[![Python](https://img.shields.io/pypi/pyversions/agenttrace)](https://pypi.org/project/agenttrace/)

Python SDK for [AgentTrace](https://github.com/agenttrace/agenttrace) — observability for AI coding agents.

## Installation

```bash
pip install agenttrace
```

## Quick Start

```python
from agenttrace import AgentTrace, observe

client = AgentTrace(
    api_key="your-api-key",
    host="http://localhost:8080"
)

@observe()
def my_llm_function(query: str) -> str:
    # Your LLM call here
    return "response"

result = my_llm_function("Hello, world!")
client.flush()
```

## Features

- **Decorator-based tracing** — `@observe()` for automatic function tracing
- **Manual instrumentation** — Fine-grained control with `trace()`, `span()`, `generation()`
- **Prompt management** — Version-controlled prompts with `Prompt.get()`
- **Auto-instrumentation** — OpenAI and Anthropic integrations
- **Async support** — Full async/await compatibility
- **Type hints** — Complete type annotations (PEP 561 compliant)

## Documentation

See the [full documentation](https://docs.agenttrace.io/sdks/python).

## API Reference

### `AgentTrace`

Main client class.

| Method | Description |
|--------|-------------|
| `trace(name, **kwargs)` | Create a new trace |
| `score(trace_id, name, value)` | Submit a score for a trace or observation |
| `flush()` | Flush all pending events to the server |
| `shutdown()` | Flush remaining events and stop the client |

### `@observe()`

Decorator for automatic function tracing.

### `Trace`

| Method | Description |
|--------|-------------|
| `generation(name, **kwargs)` | Create a generation (LLM call) observation |
| `span(name, **kwargs)` | Create a span observation |
| `score(name, value)` | Add a score to this trace |
| `update(**kwargs)` | Update trace properties |
| `end(output=None)` | End the trace |
| `checkpoint(name, **kwargs)` | Create a checkpoint |
| `git_link(**kwargs)` | Link a git commit to this trace |
| `file_op(operation, file_path)` | Track a file operation |
| `terminal_cmd(command)` | Track a terminal command |

### `Generation` / `Span`

| Method | Description |
|--------|-------------|
| `end(output=None, **kwargs)` | End the observation |

### `generation()` context manager

| Function | Description |
|----------|-------------|
| `generation(name, **kwargs)` | Sync context manager for LLM tracking |
| `ageneration(name, **kwargs)` | Async context manager for LLM tracking |
| `start_generation(name, **kwargs)` | Start a generation without a context manager |

## License

MIT
