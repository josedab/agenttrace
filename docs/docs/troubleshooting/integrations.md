---
sidebar_position: 4
title: Integration Issues
---

# Integration Issues

## OpenAI Integration

**Symptom**: OpenAI calls not being traced

**Solution**:
```python
from agenttrace import AgentTrace
from agenttrace.integrations.openai import OpenAIInstrumentation

# Initialize AgentTrace FIRST
client = AgentTrace(api_key="...")

# Enable instrumentation BEFORE creating OpenAI client
OpenAIInstrumentation.enable()

# Now create OpenAI client
from openai import OpenAI
openai_client = OpenAI()
```

## LangChain Integration

**Symptom**: Chain executions not appearing as traces

**Solution**:
```python
from agenttrace import AgentTrace
from agenttrace.integrations import LangChainInstrumentation

client = AgentTrace(api_key="...")
LangChainInstrumentation.enable()

# Use callback handler for explicit control
from agenttrace.integrations.langchain import AgentTraceCallbackHandler
handler = AgentTraceCallbackHandler()

chain.invoke(input, config={"callbacks": [handler]})
```

## Anthropic Integration

**Symptom**: Claude API calls not traced

**Solution**:
```python
from agenttrace import AgentTrace
from agenttrace.integrations.anthropic import AnthropicInstrumentation

client = AgentTrace(api_key="...")
AnthropicInstrumentation.enable()

# Create Anthropic client after enabling instrumentation
from anthropic import Anthropic
anthropic_client = Anthropic()
```
