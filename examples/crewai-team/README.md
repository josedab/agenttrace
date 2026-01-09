# CrewAI Team with AgentTrace

This example demonstrates how to trace a CrewAI multi-agent team with AgentTrace.

## Features Demonstrated

- Multi-agent collaboration tracing
- Task execution tracking
- Agent communication patterns
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

1. Creates a CrewAI team with three agents:
   - **Researcher**: Finds information about topics
   - **Writer**: Creates content based on research
   - **Editor**: Reviews and improves the content

2. Assigns tasks that flow between agents

3. All agent activity is traced to AgentTrace

## Viewing Traces

After running the example:

1. Open AgentTrace UI (http://localhost:3000 by default)
2. Navigate to Traces
3. Find the trace named "crewai-content-team"
4. Explore the nested spans showing:
   - Overall crew execution
   - Each agent's task execution
   - LLM calls per agent
   - Task delegation patterns
