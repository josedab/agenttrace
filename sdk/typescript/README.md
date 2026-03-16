# AgentTrace TypeScript SDK

[![npm](https://img.shields.io/npm/v/agenttrace)](https://www.npmjs.com/package/agenttrace)
[![Node](https://img.shields.io/node/v/agenttrace)](https://www.npmjs.com/package/agenttrace)

TypeScript SDK for [AgentTrace](https://github.com/agenttrace/agenttrace) — observability for AI coding agents.

## Installation

```bash
npm install agenttrace
# or
yarn add agenttrace
# or
pnpm add agenttrace
```

**Requirements**: Node.js 18+

### Browser Bundle

A lightweight browser-compatible build is available:

```typescript
import { AgentTrace } from 'agenttrace/browser';
```

## Quick Start

```typescript
import { AgentTrace, observe } from 'agenttrace';

const client = new AgentTrace({
  apiKey: 'your-api-key',
  host: 'http://localhost:8080',
});

const myLLMFunction = observe(async (query: string) => {
  // Your LLM call here
  return 'response';
});

const result = await myLLMFunction('Hello, world!');
await client.flush();
```

## Features

- **Function wrapper tracing** — `observe()` for automatic function tracing
- **Manual instrumentation** — Fine-grained control with `trace()`, `generation()`, `span()`
- **Prompt management** — Version-controlled prompts with `getPrompt()`
- **Checkpoints** — Create and restore agent checkpoints
- **Git linking** — Correlate traces with git commits
- **File & terminal tracking** — Track file operations and terminal commands
- **Async support** — Full async/await compatibility
- **Browser support** — Lightweight browser bundle available
- **Auto-batching** — Automatic batching and flushing of events

## Configuration

### Environment Variables

```bash
export AGENTTRACE_API_KEY="sk-at-..."
export AGENTTRACE_PROJECT_ID="your-project-id"
export AGENTTRACE_API_URL="http://localhost:8080"
export AGENTTRACE_DEBUG="true"
```

### Programmatic

```typescript
import { AgentTrace } from 'agenttrace';

const client = new AgentTrace({
  apiKey: 'sk-at-...',
  projectId: 'your-project-id',
  apiUrl: 'https://api.agenttrace.io',
  debug: true,
  flushInterval: 5000,  // ms, default 5000
  batchSize: 100,       // default 100
});
```

## Usage

### Traces and Observations

```typescript
const trace = client.trace({ name: 'code-review' });

// Generation (LLM call)
const gen = trace.generation({
  name: 'analyze-code',
  model: 'gpt-4',
  input: [{ role: 'user', content: 'Review this code...' }],
});
gen.end({
  output: 'The code looks good...',
  usage: { inputTokens: 150, outputTokens: 200 },
});

// Span (generic operation)
const span = trace.span({ name: 'process-result' });
span.end({ output: { processed: true } });

trace.end();
```

### Prompt Management

```typescript
import { Prompt } from 'agenttrace';

// Fetch a versioned prompt
const prompt = await Prompt.get({ name: 'code-review', label: 'production' });

// Compile with variables
const compiled = prompt.compile({
  language: 'TypeScript',
  code: 'function hello() { console.log("world"); }',
});
```

### Scores

```typescript
await client.score({
  traceId: 'trace-123',
  name: 'quality',
  value: 0.95,
  comment: 'Excellent output',
});
```

### Checkpoints

```typescript
import { checkpoint } from 'agenttrace';

// Create a checkpoint
await checkpoint({
  traceId: 'trace-123',
  name: 'before-refactor',
  data: { files: ['src/index.ts'] },
});
```

### Git Linking

```typescript
import { gitLink } from 'agenttrace';

// Link a git commit to a trace
await gitLink({
  traceId: 'trace-123',
  commitSha: 'abc123',
  repository: 'my-org/my-repo',
  branch: 'main',
});
```

### OpenAI Integration

```typescript
import { AgentTrace } from 'agenttrace';
import OpenAI from 'openai';

const client = new AgentTrace({ apiKey: 'sk-at-...' });
const openai = new OpenAI();

const trace = client.trace({ name: 'chat' });
const gen = trace.generation({ name: 'openai-call', model: 'gpt-4' });

const response = await openai.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }],
});

gen.end({
  output: response.choices[0].message,
  usage: {
    inputTokens: response.usage?.prompt_tokens,
    outputTokens: response.usage?.completion_tokens,
  },
});
trace.end();
await client.flush();
```

## API Reference

### `AgentTrace`

Main client class.

| Method | Description |
|--------|-------------|
| `trace(options)` | Create a new trace |
| `flush()` | Flush all pending events |
| `score(options)` | Create a score |
| `getPrompt(options)` | Fetch a prompt by name |
| `shutdown()` | Flush and close the client |

### `observe(fn, options?)`

Wraps a function with automatic tracing.

### `Trace`

| Method | Description |
|--------|-------------|
| `generation(options)` | Create a generation observation |
| `span(options)` | Create a span observation |
| `score(options)` | Score this trace |
| `end(options?)` | End the trace |

### `Generation` / `Span`

| Method | Description |
|--------|-------------|
| `end(options?)` | End the observation |
| `update(options)` | Update observation metadata |

## Development

```bash
# Install dependencies
npm install

# Build
npm run build

# Run tests
npm test

# Lint
npm run lint

# Type check
npm run typecheck
```

## Documentation

See the [full documentation](https://docs.agenttrace.io/sdks/typescript).

## License

MIT
