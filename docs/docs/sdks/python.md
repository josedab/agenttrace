---
sidebar_position: 2
---

# Python SDK

The AgentTrace Python SDK provides full observability for Python-based AI agents.

## Installation

```bash
pip install agenttrace
```

**Requirements**: Python 3.9+

## Quick Start

```python
from agenttrace import AgentTrace, observe

# Initialize the client
at = AgentTrace(api_key="sk-at-...")

# Use the @observe decorator
@observe()
def my_agent_task(prompt: str) -> str:
    response = call_llm(prompt)
    return response

# Call your function - it's automatically traced!
result = my_agent_task("Write a function to sort a list")
```

## Configuration

### Environment Variables

```bash
export AGENTTRACE_API_KEY="sk-at-..."
export AGENTTRACE_PROJECT_ID="your-project-id"  # optional
export AGENTTRACE_API_URL="https://api.agenttrace.io"  # for self-hosted
export AGENTTRACE_DEBUG="true"  # enable debug logging
```

### Programmatic Configuration

```python
from agenttrace import AgentTrace

at = AgentTrace(
    api_key="sk-at-...",  # Required (or use env var)
    project_id="your-project-id",  # Optional
    api_url="https://api.agenttrace.io",  # Optional (default: cloud)
    debug=True,  # Optional: enable debug logging
    flush_interval=5.0,  # Optional: seconds between flushes
    batch_size=100,  # Optional: max events per batch
    enabled=True,  # Optional: disable tracing entirely
)
```

## Tracing

### The `@observe` Decorator

The simplest way to add tracing is with the `@observe` decorator:

```python
from agenttrace import observe

@observe()
def simple_function(x: int) -> int:
    return x * 2

@observe(name="custom-name")
def named_function(x: int) -> int:
    return x * 2

@observe(capture_input=True, capture_output=True)
def full_capture(data: dict) -> dict:
    return {"result": data["value"] * 2}
```

### Async Functions

The decorator works with async functions too:

```python
@observe()
async def async_function(prompt: str) -> str:
    response = await async_llm_call(prompt)
    return response
```

### Manual Traces

For more control, use the context manager:

```python
with at.trace("my-task") as trace:
    # Set input
    trace.input = {"prompt": "Hello"}

    # Do work
    result = process()

    # Set output
    trace.output = result

    # Add metadata
    trace.metadata = {"version": "1.0"}

    # Add tags
    trace.tags = ["production", "important"]
```

### Nested Traces

Create hierarchical traces:

```python
with at.trace("parent-task") as parent:
    with parent.span("step-1") as step1:
        intermediate = process_step_1()
        step1.output = intermediate

    with parent.span("step-2") as step2:
        result = process_step_2(intermediate)
        step2.output = result

    parent.output = result
```

## Observations

### Generations (LLM Calls)

Track LLM calls with the `generation` context manager:

```python
with at.trace("chat") as trace:
    with trace.generation(
        name="gpt-call",
        model="gpt-4",
        model_parameters={"temperature": 0.7}
    ) as gen:
        messages = [{"role": "user", "content": "Hello"}]
        gen.input = messages

        response = openai.chat.completions.create(
            model="gpt-4",
            messages=messages
        )

        gen.output = response.choices[0].message.content
        gen.usage = {
            "inputTokens": response.usage.prompt_tokens,
            "outputTokens": response.usage.completion_tokens
        }
```

### Spans (Non-LLM Operations)

Track other operations:

```python
with at.trace("process") as trace:
    with trace.span(name="fetch-data") as span:
        data = fetch_from_database()
        span.output = {"records": len(data)}

    with trace.span(name="transform") as span:
        result = transform_data(data)
        span.output = {"transformed": len(result)}
```

### Events

Log discrete events:

```python
with at.trace("task") as trace:
    trace.event(name="checkpoint-reached", data={"step": 1})

    process_step_1()

    trace.event(name="checkpoint-reached", data={"step": 2})
```

## LLM Integrations

### OpenAI

```python
from agenttrace.integrations.openai import wrap_openai
import openai

# Wrap the OpenAI client
client = wrap_openai(openai.OpenAI())

# All calls are automatically traced
with at.trace("openai-chat") as trace:
    response = client.chat.completions.create(
        model="gpt-4",
        messages=[{"role": "user", "content": "Hello"}]
    )
```

### Anthropic

```python
from agenttrace.integrations.anthropic import wrap_anthropic
import anthropic

# Wrap the Anthropic client
client = wrap_anthropic(anthropic.Anthropic())

# All calls are automatically traced
with at.trace("anthropic-chat") as trace:
    response = client.messages.create(
        model="claude-3-sonnet-20240229",
        max_tokens=1024,
        messages=[{"role": "user", "content": "Hello"}]
    )
```

## Prompts

### Fetching Prompts

```python
# Get the latest version
prompt = at.get_prompt("code-review")

# Get a specific version
prompt = at.get_prompt("code-review", version=3)

# Get by label
prompt = at.get_prompt("code-review", label="production")
```

### Compiling Prompts

```python
prompt = at.get_prompt("code-review")

# Compile with variables
compiled = prompt.compile(
    language="Python",
    code="def hello(): print('world')"
)

# Use the compiled prompt
response = call_llm(compiled)
```

### Linking Prompts to Traces

```python
prompt = at.get_prompt("code-review", label="production")

with at.trace("review") as trace:
    # Link the prompt to this trace
    trace.prompt = prompt

    compiled = prompt.compile(language="Python", code="...")
    response = call_llm(compiled)
```

## Scores

### Scoring Traces

```python
# Score by trace ID
at.score(
    trace_id="trace-123",
    name="quality",
    value=0.95,
    comment="Excellent response"
)

# Score within a trace
with at.trace("task") as trace:
    result = process()
    trace.score(name="quality", value=0.95)
```

### Score Types

```python
# Numeric score (0-1)
at.score(trace_id="...", name="quality", value=0.95)

# Boolean score
at.score(trace_id="...", name="correct", value=1)  # True
at.score(trace_id="...", name="correct", value=0)  # False

# Categorical score
at.score(trace_id="...", name="rating", value="good")
```

## Agent Features

### Checkpoints

Create state checkpoints:

```python
with at.trace("edit-task") as trace:
    # Checkpoint before making changes
    trace.checkpoint(
        name="before-edit",
        description="State before code modification",
        files=["main.py", "utils.py"]
    )

    # Make changes
    edit_files()

    # Checkpoint after changes
    trace.checkpoint(
        name="after-edit",
        description="State after code modification",
        files=["main.py", "utils.py"]
    )
```

### Git Linking

Link commits to traces:

```python
# Auto-detect git info
at.git_link(trace_id="trace-123")

# Or specify manually
at.git_link(
    trace_id="trace-123",
    commit_sha="abc123",
    branch="main",
    repo_url="https://github.com/org/repo"
)
```

### File Operations

Track file operations:

```python
with at.trace("refactor") as trace:
    trace.file_op(
        operation="edit",
        path="src/main.py",
        before="old content",
        after="new content"
    )
```

### Terminal Commands

Track terminal commands:

```python
with at.trace("build") as trace:
    trace.terminal_cmd(
        command="npm test",
        exit_code=0,
        stdout="All tests passed",
        stderr=""
    )
```

## Sessions

Group related traces into sessions:

```python
# Create a session
session = at.session(id="user-123-session")

# All traces in this context are grouped
with session.trace("task-1") as trace:
    process_task_1()

with session.trace("task-2") as trace:
    process_task_2()
```

## Error Handling

```python
with at.trace("risky-task") as trace:
    try:
        result = risky_operation()
        trace.output = result
    except Exception as e:
        trace.error = str(e)
        trace.level = "ERROR"
        raise
```

## Flushing

The SDK batches events and flushes periodically. You can force a flush:

```python
# Flush all pending events
at.flush()

# Or use shutdown for clean exit
at.shutdown()
```

## Context Propagation

The SDK uses context variables for automatic parent-child relationships:

```python
@observe()
def parent_function():
    # Child functions automatically become nested observations
    child_function()

@observe()
def child_function():
    # This will be nested under parent_function
    pass
```

## Advanced Configuration

### Custom HTTP Client

```python
import httpx

at = AgentTrace(
    api_key="...",
    http_client=httpx.Client(timeout=30.0)
)
```

### Disabled Mode

```python
# Disable tracing (e.g., in tests)
at = AgentTrace(enabled=False)

# Or via environment
# AGENTTRACE_ENABLED=false
```

### Sampling

```python
at = AgentTrace(
    api_key="...",
    sample_rate=0.1  # Only trace 10% of requests
)
```

## API Reference

### AgentTrace

```python
class AgentTrace:
    def __init__(
        self,
        api_key: str | None = None,
        project_id: str | None = None,
        api_url: str = "https://api.agenttrace.io",
        debug: bool = False,
        flush_interval: float = 5.0,
        batch_size: int = 100,
        enabled: bool = True,
        sample_rate: float = 1.0,
    ): ...

    def trace(self, name: str, **kwargs) -> TraceContext: ...
    def session(self, id: str) -> Session: ...
    def get_prompt(self, name: str, version: int | None = None, label: str | None = None) -> Prompt: ...
    def score(self, trace_id: str, name: str, value: float | int | str, comment: str | None = None): ...
    def git_link(self, trace_id: str, commit_sha: str | None = None, **kwargs): ...
    def flush(self): ...
    def shutdown(self): ...
```

### TraceContext

```python
class TraceContext:
    id: str
    input: Any
    output: Any
    metadata: dict
    tags: list[str]
    level: str
    error: str | None
    prompt: Prompt | None

    def generation(self, name: str, model: str, **kwargs) -> GenerationContext: ...
    def span(self, name: str, **kwargs) -> SpanContext: ...
    def event(self, name: str, data: dict | None = None): ...
    def score(self, name: str, value: float | int | str, comment: str | None = None): ...
    def checkpoint(self, name: str, files: list[str], **kwargs): ...
    def file_op(self, operation: str, path: str, **kwargs): ...
    def terminal_cmd(self, command: str, exit_code: int, **kwargs): ...
```

### Prompt

```python
class Prompt:
    name: str
    version: int
    content: str

    def compile(self, **variables) -> str: ...
```
