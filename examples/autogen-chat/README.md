# AutoGen Chat with AgentTrace

This example demonstrates how to trace an AutoGen multi-agent conversation with AgentTrace.

## Features Demonstrated

- Multi-agent conversation tracing
- Message exchange tracking
- Code execution monitoring
- LLM call tracing per agent

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

1. Creates an AutoGen conversation between:
   - **User Proxy**: Represents the human user
   - **Assistant**: An AI assistant that can write and execute code
   - **Critic**: Reviews code and suggests improvements

2. Initiates a conversation about creating a data visualization

3. All agent interactions are traced to AgentTrace

## Viewing Traces

After running the example:

1. Open AgentTrace UI (http://localhost:3000 by default)
2. Navigate to Traces
3. Find the trace named "autogen-coding-chat"
4. Explore the nested spans showing:
   - Overall conversation flow
   - Message exchanges between agents
   - Code generation and execution
   - LLM calls with token usage
