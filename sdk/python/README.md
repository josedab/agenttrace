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

## License

MIT
