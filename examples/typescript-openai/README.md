# AgentTrace + OpenAI Example (TypeScript)

This example demonstrates integrating AgentTrace with OpenAI's API, showcasing:

- **Streaming responses** - Real-time token streaming with proper trace recording
- **Function calling** - Tool use with `get_weather` and `search_web` functions
- **Multi-turn conversations** - Agentic loop with tool result handling
- **Error handling** - Proper error capture and trace metadata

## Prerequisites

1. AgentTrace running locally or remotely:
   ```bash
   docker compose up -d
   ```

2. Environment variables:
   ```bash
   export AGENTTRACE_API_KEY="your-agenttrace-api-key"
   export AGENTTRACE_HOST="http://localhost:8080"
   export OPENAI_API_KEY="your-openai-api-key"
   ```

## Running the Example

```bash
npm install
npm start
```

Or with ts-node directly:
```bash
npx ts-node main.ts
```

## What It Does

1. Creates a trace for the entire agent session
2. Sends a user query: "What's the weather like in San Francisco and New York today?"
3. The LLM decides to call the `get_weather` tool for both cities
4. Tool calls are recorded as spans with input/output
5. Results are fed back to the LLM for a final response
6. All generations and spans are visible in the AgentTrace UI

## Trace Structure

```
openai-agent-with-tools (trace)
├── llm-call-1 (generation) - Initial request, returns tool calls
├── tool-get_weather (span) - San Francisco weather
├── tool-get_weather (span) - New York weather
└── llm-call-2 (generation) - Final response with weather info
```

## Key Patterns Demonstrated

### Streaming with Trace Recording

```typescript
const stream = await openai.chat.completions.create({
  model: "gpt-4o-mini",
  messages,
  tools,
  stream: true,
});

let content = "";
for await (const chunk of stream) {
  content += chunk.choices[0]?.delta?.content || "";
}

generation.end({
  output: { content },
  usage: { inputTokens: 10, outputTokens: 50, totalTokens: 60 },
});
```

### Tool Execution Spans

```typescript
const toolSpan = trace.span({
  name: `tool-${toolCall.function.name}`,
  input: { arguments: toolCall.function.arguments },
});

try {
  const result = executeTool(toolCall);
  toolSpan.end({ output: { result } });
} catch (error) {
  toolSpan.end({
    output: { error: error.message },
    level: "error"
  });
}
```

### Error Handling

```typescript
try {
  // Agent logic
} catch (error) {
  trace.update({
    output: { error: error.message },
    level: "error",
    statusMessage: error.message,
  });
}
```

## Viewing Traces

1. Open http://localhost:3000 in your browser
2. Navigate to **Traces**
3. Find the trace named "openai-agent-with-tools"
4. Expand to see all LLM calls and tool executions
5. Click any generation to see tokens, latency, and costs

## Extending This Example

- Add more tools (calculator, file operations, etc.)
- Implement retry logic for failed tool calls
- Add user ID and session ID for conversation tracking
- Score the final response with custom evaluators
