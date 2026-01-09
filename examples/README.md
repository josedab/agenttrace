# AgentTrace Examples

This directory contains real-world examples demonstrating how to use AgentTrace with popular AI agent frameworks.

## Examples

### 0. Hello World (`hello-world/`)

The simplest possible examples to get started quickly. Available in Python, TypeScript, and Go. Demonstrates:
- Creating traces and spans
- Recording LLM generations
- Adding metadata and viewing traces

```bash
cd hello-world/python
pip install -r requirements.txt
python main.py
```

### 1. LangChain Agent (`langchain-agent/`)

A ReAct agent that uses tools to answer complex questions. Demonstrates:
- Automatic LLM call tracing
- Tool usage tracking
- Chain execution monitoring

```bash
cd langchain-agent
pip install -r requirements.txt
python main.py
```

### 2. CrewAI Team (`crewai-team/`)

A multi-agent team (Researcher, Writer, Editor) working together to create content. Demonstrates:
- Multi-agent collaboration
- Task delegation tracing
- Sequential process monitoring

```bash
cd crewai-team
pip install -r requirements.txt
python main.py
```

### 3. AutoGen Chat (`autogen-chat/`)

A conversational multi-agent system for code generation. Demonstrates:
- Agent-to-agent communication
- Code execution tracking
- Custom agent instrumentation

```bash
cd autogen-chat
pip install -r requirements.txt
python main.py
```

### 4. TypeScript + OpenAI (`typescript-openai/`)

Advanced TypeScript example with OpenAI integration. Demonstrates:
- Streaming responses with proper trace recording
- Function calling (tool use)
- Multi-turn agentic loops
- Error handling patterns

```bash
cd typescript-openai
npm install
npm start
```

### 5. Go Agent Workflow (`go-agent-workflow/`)

Multi-step agent workflow in Go. Demonstrates:
- Parallel document processing
- Nested spans and generations
- Checkpointing and scoring
- Error handling in concurrent code

```bash
cd go-agent-workflow
go mod tidy
go run main.go
```

## Prerequisites

All examples require:

1. **AgentTrace** running locally or accessible remotely:
   ```bash
   # Using Docker Compose from the root directory
   docker compose up -d
   ```

2. **Environment variables**:
   ```bash
   export AGENTTRACE_API_KEY="your-api-key"
   export AGENTTRACE_HOST="http://localhost:8080"
   export OPENAI_API_KEY="your-openai-key"
   ```

## Viewing Traces

After running any example:

1. Open the AgentTrace UI at http://localhost:3000
2. Navigate to the **Traces** section
3. Find your trace by name
4. Click to explore the full trace hierarchy

## What You'll See

Each example generates rich traces showing:

- **Spans**: High-level operations (agent runs, tool calls, retrievals)
- **Generations**: LLM calls with inputs, outputs, and token usage
- **Metadata**: Custom context like model names, agent types, and task details
- **Timing**: Latency for each operation with waterfall visualization
- **Costs**: Token usage and estimated costs per LLM call

## Creating Your Own Instrumentation

### Using Decorators

```python
from agenttrace import AgentTrace
from agenttrace.decorators import observe

client = AgentTrace(api_key="...")

@observe(name="my_function")
def my_function(input_data):
    # Your code here
    return result
```

### Using Context Managers

```python
from agenttrace import AgentTrace
from agenttrace.span import start_span

client = AgentTrace(api_key="...")

with client.trace(name="my-trace") as trace:
    with start_span(name="step_1") as span:
        result = do_something()
        span.end(output={"result": result})
```

### Using Framework Integrations

```python
from agenttrace import AgentTrace
from agenttrace.integrations import LangChainInstrumentation

client = AgentTrace(api_key="...")
LangChainInstrumentation.enable()

# All LangChain calls are now automatically traced!
```

## Troubleshooting

### Traces not appearing

1. Verify AgentTrace is running: `curl http://localhost:8080/health`
2. Check your API key is valid
3. Ensure `AGENTTRACE_HOST` is correct

### Missing LLM calls

1. Make sure you've enabled the integration before creating the LLM client
2. Pass the callback handler explicitly if auto-instrumentation doesn't work

### Import errors

1. Install the framework-specific dependencies
2. Check that you're using compatible versions (see requirements.txt)

## Contributing

We welcome additional examples! Please submit a PR with:

1. A new directory under `examples/`
2. A `README.md` explaining the example
3. A `main.py` with the example code
4. A `requirements.txt` with dependencies
