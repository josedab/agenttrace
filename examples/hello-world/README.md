# Hello World Examples

The simplest possible AgentTrace examples to help you get started quickly.

## Overview

These examples demonstrate the basics of AgentTrace tracing in each supported language:
- Creating a trace
- Adding spans (child operations)
- Recording an LLM generation
- Viewing traces in the UI

## Prerequisites

1. **AgentTrace running locally**:
   ```bash
   cd /path/to/agenttrace
   docker compose up -d
   ```

2. **API Key**: Create one at http://localhost:3000 in Settings > API Keys

## Quick Start

### Python

```bash
cd python
pip install agenttrace
export AGENTTRACE_API_KEY="your-api-key"
python main.py
```

### TypeScript/Node.js

```bash
cd typescript
npm install
export AGENTTRACE_API_KEY="your-api-key"
npx ts-node main.ts
```

### Go

```bash
cd go
go mod tidy
export AGENTTRACE_API_KEY="your-api-key"
go run main.go
```

## What You'll Learn

Each example shows:

1. **Initializing the client** - Connect to AgentTrace with your API key
2. **Creating a trace** - The root container for an operation
3. **Adding spans** - Track individual steps within a trace
4. **Recording generations** - Log LLM calls with model info and tokens
5. **Adding metadata** - Attach custom key-value pairs
6. **Proper cleanup** - Flush data and shut down gracefully

## Viewing Your Traces

1. Open http://localhost:3000
2. Navigate to **Traces**
3. Find your "hello-world" trace
4. Click to explore the trace hierarchy

You'll see:
- The parent trace
- Child spans for each step
- A simulated LLM generation with token counts
- Timing information and latency

## Next Steps

After running these examples, try:

1. **Add real LLM calls** - Replace the simulated generation with actual API calls
2. **Use decorators** (Python) - Simplify tracing with `@observe()`
3. **Explore integrations** - Auto-instrument OpenAI, Anthropic, LangChain
4. **Add agent features** - Track git commits, file operations, terminal commands

See the other examples in this directory for more advanced use cases.
