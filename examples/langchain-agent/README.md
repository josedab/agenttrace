# LangChain Agent with AgentTrace

This example demonstrates how to trace a LangChain ReAct agent with AgentTrace.

## Features Demonstrated

- Automatic tracing of LLM calls
- Chain execution tracing
- Tool usage tracking
- Agent decision tracing

## Setup

1. Install dependencies:

```bash
pip install -r requirements.txt
```

2. Set environment variables:

```bash
export AGENTTRACE_API_KEY="your-agenttrace-api-key"
export AGENTTRACE_HOST="http://localhost:8080"  # Or your AgentTrace instance
export OPENAI_API_KEY="your-openai-api-key"
```

3. Run the example:

```bash
python main.py
```

## What This Example Does

1. Creates a ReAct agent with tools for:
   - Wikipedia search
   - Calculator
   - Current time

2. Asks the agent a question that requires using multiple tools

3. All agent activity is automatically traced to AgentTrace

## Viewing Traces

After running the example:

1. Open AgentTrace UI (http://localhost:3000 by default)
2. Navigate to Traces
3. Find the trace named "langchain-react-agent"
4. Explore the nested spans showing:
   - The overall agent execution
   - Each LLM call with inputs/outputs
   - Each tool invocation
   - Token usage and costs
