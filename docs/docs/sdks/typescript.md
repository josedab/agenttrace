---
sidebar_position: 3
---

# TypeScript SDK

The AgentTrace TypeScript SDK provides full observability for Node.js and browser-based AI agents.

## Installation

```bash
npm install agenttrace
# or
yarn add agenttrace
# or
pnpm add agenttrace
```

**Requirements**: Node.js 18+ or modern browsers with ES2020 support

## Quick Start

```typescript
import { AgentTrace, observe } from "agenttrace";

// Initialize the client
const client = new AgentTrace({
  apiKey: "sk-at-...",
  host: "https://api.agenttrace.io"
});

// Use the observe() wrapper for automatic tracing
const myAgentTask = observe(async (prompt: string) => {
  const response = await callLLM(prompt);
  return response;
});

// Call your function - it's automatically traced!
const result = await myAgentTask("Write a function to sort a list");

// Ensure events are sent before exiting
await client.flush();
```

## Configuration

### Environment Variables

```bash
export AGENTTRACE_API_KEY="sk-at-..."
export AGENTTRACE_PROJECT_ID="your-project-id"  # optional
export AGENTTRACE_HOST="https://api.agenttrace.io"  # for self-hosted
```

### Programmatic Configuration

```typescript
import { AgentTrace } from "agenttrace";

const client = new AgentTrace({
  apiKey: "sk-at-...",           // Required (or use env var)
  host: "https://api.agenttrace.io",  // Optional (default: cloud)
  projectId: "your-project-id",  // Optional
  publicKey: "pk-...",           // Optional: for public traces
  enabled: true,                 // Optional: disable tracing entirely
  flushAt: 20,                   // Optional: events per batch
  flushInterval: 5000,           // Optional: ms between flushes
  maxRetries: 3,                 // Optional: retry attempts
  timeout: 10000,                // Optional: request timeout in ms
});
```

## Core Concepts

AgentTrace uses a hierarchical model for tracing:

```
Trace (top-level)
├── Span (grouping/step)
│   ├── Generation (LLM call)
│   └── Span (nested step)
├── Generation (LLM call)
└── Span (another step)
```

### Traces

A trace represents a complete execution path:

```typescript
const trace = client.trace({
  name: "code-review-agent",
  userId: "user-123",
  sessionId: "session-456",
  metadata: { version: "1.0" },
  tags: ["production", "code-review"],
  input: { code: "function add(a, b) { return a + b; }" }
});

// Do work...

trace.end({ output: { result: "Code looks good!" } });
```

### Spans

Spans represent non-LLM operations within a trace:

```typescript
const trace = client.trace({ name: "process-document" });

const parseSpan = trace.span({
  name: "parse-document",
  input: { documentId: "doc-123" }
});

// Do parsing work...
const parsed = await parseDocument();

parseSpan.end({ output: parsed });
trace.end();
```

### Generations

Generations track LLM calls with model-specific metadata:

```typescript
const trace = client.trace({ name: "chat-completion" });

const generation = trace.generation({
  name: "gpt-4-call",
  model: "gpt-4",
  modelParameters: { temperature: 0.7, maxTokens: 1000 },
  input: [{ role: "user", content: "Hello!" }]
});

// Make LLM call
const response = await openai.chat.completions.create({
  model: "gpt-4",
  messages: [{ role: "user", content: "Hello!" }]
});

generation.end({
  output: response.choices[0].message.content,
  usage: {
    inputTokens: response.usage.prompt_tokens,
    outputTokens: response.usage.completion_tokens
  }
});

trace.end();
```

## The `observe()` Wrapper

The simplest way to add tracing is with the `observe()` function:

```typescript
import { observe } from "agenttrace";

// Basic usage - automatically creates spans
const processData = observe(async (data: Record<string, unknown>) => {
  const result = await transform(data);
  return result;
});

// With custom name
const namedFunction = observe(
  async (x: number) => x * 2,
  { name: "double-value" }
);

// As a generation (LLM call)
const llmCall = observe(
  async (messages: Message[]) => {
    return await openai.chat.completions.create({
      model: "gpt-4",
      messages
    });
  },
  {
    name: "chat-completion",
    asType: "generation",
    model: "gpt-4",
    captureInput: true,
    captureOutput: true
  }
);
```

### Options

```typescript
interface ObserveOptions {
  /** Custom name for the observation (defaults to function name) */
  name?: string;
  /** Type of observation: "span" or "generation" */
  asType?: "span" | "generation";
  /** Whether to capture function arguments (default: true) */
  captureInput?: boolean;
  /** Whether to capture return value (default: true) */
  captureOutput?: boolean;
  /** Model name for generations */
  model?: string;
  /** Model parameters for generations */
  modelParameters?: Record<string, unknown>;
}
```

## Context Propagation

The SDK uses `AsyncLocalStorage` for automatic parent-child relationships:

```typescript
import { getCurrentTrace, runWithTrace } from "agenttrace";

// Nested observations are automatically linked
const parentFunction = observe(async () => {
  // childFunction will be nested under parentFunction
  await childFunction();
});

const childFunction = observe(async () => {
  // This becomes a child observation
  return "result";
});

// Manual context access
const trace = client.trace({ name: "manual-context" });

await runWithTrace(trace, async () => {
  // getCurrentTrace() returns the trace here
  const currentTrace = getCurrentTrace();
  console.log(currentTrace?.id);
});
```

### Context Functions

```typescript
import {
  getClient,
  getCurrentTrace,
  getCurrentObservation,
  runWithTrace,
  runWithObservation,
  runWithClient
} from "agenttrace";

// Get current context
const client = getClient();
const trace = getCurrentTrace();
const observation = getCurrentObservation();

// Run with specific context
const result = await runWithTrace(trace, async () => {
  return await doWork();
});
```

## Nested Traces

Create hierarchical traces for complex workflows:

```typescript
const trace = client.trace({ name: "document-processor" });

// First step
const extractSpan = trace.span({ name: "extract-text" });
const text = await extractText(document);
extractSpan.end({ output: { text } });

// Nested operations
const analyzeSpan = trace.span({ name: "analyze-content" });

const sentimentGen = trace.generation({
  name: "sentiment-analysis",
  model: "gpt-4",
  parentObservationId: analyzeSpan.id,
  input: { text }
});
const sentiment = await analyzeSentiment(text);
sentimentGen.end({ output: sentiment });

const summaryGen = trace.generation({
  name: "summarize",
  model: "gpt-4",
  parentObservationId: analyzeSpan.id,
  input: { text }
});
const summary = await summarize(text);
summaryGen.end({ output: summary });

analyzeSpan.end({ output: { sentiment, summary } });
trace.end({ output: { text, sentiment, summary } });
```

## Agent Features

### Checkpoints

Create state snapshots during agent execution:

```typescript
const trace = client.trace({ name: "code-editor-agent" });

// Checkpoint before making changes
const beforeCp = trace.checkpoint({
  name: "before-edit",
  type: "manual",
  description: "State before code modification",
  files: ["src/main.ts", "src/utils.ts"],
  includeGitInfo: true
});

// Make changes
await editFiles();

// Checkpoint after changes
const afterCp = trace.checkpoint({
  name: "after-edit",
  type: "milestone",
  files: ["src/main.ts", "src/utils.ts"]
});

trace.end();
```

#### Checkpoint Types

- `manual` - User-initiated checkpoint
- `auto` - Automatically created checkpoint
- `tool_call` - Checkpoint before/after tool execution
- `error` - Checkpoint on error for debugging
- `milestone` - Significant progress point
- `restore` - Checkpoint for state restoration

### Git Linking

Associate traces with git commits:

```typescript
const trace = client.trace({ name: "feature-implementation" });

// Auto-detect git info (default behavior)
trace.gitLink();

// With explicit values
trace.gitLink({
  type: "commit",
  commitSha: "abc123def456",
  branch: "feature/new-api",
  repoUrl: "https://github.com/org/repo",
  commitMessage: "Add new API endpoint",
  filesChanged: ["src/api.ts", "src/routes.ts"]
});

trace.end();
```

#### Git Link Types

- `start` - Beginning of work session
- `commit` - Associated with a specific commit
- `restore` - Restored to a previous state
- `branch` - Branch creation/switch
- `diff` - Current diff state

### File Operations

Track file read/write operations:

```typescript
const trace = client.trace({ name: "refactor-task" });

// Track a file edit
trace.fileOp({
  operation: "update",
  filePath: "src/main.ts",
  contentBefore: oldContent,
  contentAfter: newContent,
  linesAdded: 15,
  linesRemoved: 8,
  toolName: "edit-file",
  reason: "Refactoring for clarity"
});

// Track file creation
trace.fileOp({
  operation: "create",
  filePath: "src/utils/helpers.ts",
  contentAfter: newFileContent
});

// Track file deletion
trace.fileOp({
  operation: "delete",
  filePath: "src/old-file.ts"
});

trace.end();
```

#### Operation Types

- `create` - New file created
- `read` - File read
- `update` - File modified
- `delete` - File deleted
- `rename` - File renamed
- `copy` - File copied
- `move` - File moved
- `chmod` - Permissions changed

### Terminal Commands

Track shell command execution:

```typescript
const trace = client.trace({ name: "build-task" });

// Manual tracking
trace.terminalCmd({
  command: "npm",
  args: ["test"],
  exitCode: 0,
  stdout: "All tests passed (42 tests)",
  stderr: "",
  workingDirectory: "/project",
  toolName: "run-tests",
  reason: "Verify changes before commit"
});

// Run and track automatically
const result = await trace.runCmd("npm", {
  args: ["run", "build"],
  workingDirectory: "/project",
  timeout: 60000,
  maxOutputBytes: 100000
});

console.log(result.exitCode); // 0
console.log(result.stdout);   // Build output

trace.end();
```

## Prompts

### Fetching Prompts

```typescript
import { Prompt, getPrompt } from "agenttrace";

// Get latest version
const prompt = await Prompt.get({ name: "code-review" });

// Get specific version
const promptV2 = await Prompt.get({ name: "code-review", version: 2 });

// Get by label
const prodPrompt = await Prompt.get({
  name: "code-review",
  label: "production"
});

// With fallback
const safePrompt = await Prompt.get({
  name: "code-review",
  fallback: "Review the following code for issues:\n{{code}}"
});

// Convenience function
const prompt2 = await getPrompt({ name: "my-prompt" });
```

### Compiling Prompts

```typescript
const prompt = await Prompt.get({ name: "code-review" });

// Compile with variables
const compiled = prompt.compile({
  language: "TypeScript",
  code: "function add(a: number, b: number) { return a + b; }"
});

// Use the compiled prompt
const response = await callLLM(compiled);

// Get variable names
const variables = prompt.getVariables(); // ["language", "code"]
```

### Chat Prompts

```typescript
const prompt = await Prompt.get({ name: "chat-template" });

// Prompt content:
// system: You are a helpful {{role}}.
// user: {{question}}

const messages = prompt.compileChat({
  role: "coding assistant",
  question: "How do I sort an array?"
});

// Result:
// [
//   { role: "system", content: "You are a helpful coding assistant." },
//   { role: "user", content: "How do I sort an array?" }
// ]
```

### Cache Management

```typescript
import { Prompt } from "agenttrace";

// Set default cache TTL (default: 60000ms)
Prompt.setCacheTtl(120000);

// Clear all cached prompts
Prompt.clearCache();

// Invalidate specific prompt
Prompt.invalidate("code-review");
```

## Scoring

### Score a Trace

```typescript
// Score by trace ID
client.score({
  traceId: "trace-123",
  name: "quality",
  value: 0.95,
  dataType: "NUMERIC",
  comment: "Excellent response quality"
});

// Score within a trace
const trace = client.trace({ name: "task" });
// ... do work ...
trace.score("accuracy", 0.92, { comment: "High accuracy" });
trace.end();
```

### Score Types

```typescript
// Numeric score (0-1 range recommended)
client.score({
  traceId: "...",
  name: "quality",
  value: 0.95,
  dataType: "NUMERIC"
});

// Boolean score
client.score({
  traceId: "...",
  name: "correct",
  value: true,
  dataType: "BOOLEAN"
});

// Categorical score
client.score({
  traceId: "...",
  name: "rating",
  value: "excellent",
  dataType: "CATEGORICAL"
});
```

### Score Observations

```typescript
const trace = client.trace({ name: "task" });
const generation = trace.generation({ name: "llm-call", model: "gpt-4" });

// ... LLM call ...

generation.end({ output: response });

// Score the specific generation
client.score({
  traceId: trace.id,
  observationId: generation.id,
  name: "relevance",
  value: 0.88
});
```

## Error Handling

```typescript
const trace = client.trace({ name: "risky-task" });

try {
  const result = await riskyOperation();
  trace.update({ output: result });
} catch (error) {
  trace.update({
    output: { error: String(error) },
    metadata: { errorType: error.name }
  });
  throw error;
} finally {
  trace.end();
}
```

### With observe()

Errors are automatically captured:

```typescript
const riskyFunction = observe(async () => {
  throw new Error("Something went wrong");
});

try {
  await riskyFunction();
} catch (error) {
  // Error is captured in the trace output
  console.error(error);
}
```

## Flushing and Shutdown

The SDK batches events for efficiency. Ensure events are sent:

```typescript
// Flush pending events
await client.flush();

// Shutdown with flush (recommended for process exit)
await client.shutdown();

// Auto-flush on process signals (automatic)
// - beforeExit
// - SIGINT
// - SIGTERM
```

### Graceful Shutdown Example

```typescript
const client = new AgentTrace({ apiKey: "..." });

// Handle shutdown gracefully
process.on("beforeExit", async () => {
  await client.shutdown();
});

// Or use try/finally
try {
  // Your application logic
  await runAgent();
} finally {
  await client.shutdown();
}
```

## Update Traces

Modify traces after creation:

```typescript
const trace = client.trace({ name: "my-task" });

// Update properties
trace.update({
  name: "updated-name",
  metadata: { step: 2 },
  tags: ["updated", "important"],
  output: { partial: "result" }
});

// Continue work...

trace.end({ output: { final: "result" } });
```

## Disabled Mode

Disable tracing for testing or specific environments:

```typescript
// Disable programmatically
const client = new AgentTrace({
  apiKey: "...",
  enabled: false
});

// Or via environment
// AGENTTRACE_ENABLED=false

// Traces still work but don't send data
const trace = client.trace({ name: "test" });
// ... operations work normally
trace.end();
```

## TypeScript Types Reference

### Configuration

```typescript
interface AgentTraceConfig {
  apiKey: string;
  host?: string;
  publicKey?: string;
  projectId?: string;
  enabled?: boolean;
  flushAt?: number;
  flushInterval?: number;
  maxRetries?: number;
  timeout?: number;
}
```

### Trace

```typescript
interface TraceOptions {
  name: string;
  id?: string;
  userId?: string;
  sessionId?: string;
  metadata?: Record<string, unknown>;
  tags?: string[];
  input?: unknown;
  public?: boolean;
}

class Trace {
  readonly id: string;
  name: string;
  userId?: string;
  sessionId?: string;
  metadata: Record<string, unknown>;
  tags: string[];
  input?: unknown;
  output?: unknown;
  readonly startTime: Date;
  endTime?: Date;

  span(options: SpanOptions): Span;
  generation(options: GenerationOptions): Generation;
  update(updates: Partial<TraceOptions>): this;
  end(options?: { output?: unknown }): void;
  score(name: string, value: number | boolean | string, options?: ScoreOptions): void;
  checkpoint(options: CheckpointOptions): CheckpointInfo;
  gitLink(options?: GitLinkOptions): GitLinkInfo;
  fileOp(options: FileOperationOptions): FileOperationInfo;
  terminalCmd(options: TerminalCommandOptions): TerminalCommandInfo;
  runCmd(command: string, options?: RunCommandOptions): Promise<RunCommandResult>;
}
```

### Generation

```typescript
interface GenerationOptions {
  name: string;
  id?: string;
  parentObservationId?: string;
  model?: string;
  modelParameters?: Record<string, unknown>;
  input?: unknown;
  metadata?: Record<string, unknown>;
  level?: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
}

interface UsageDetails {
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
}

class Generation {
  readonly id: string;
  readonly traceId: string;
  readonly name: string;
  model?: string;
  modelParameters: Record<string, unknown>;
  input?: unknown;
  output?: unknown;
  usage?: UsageDetails;

  update(updates: Partial<{ output: unknown; usage: UsageDetails; model: string }>): this;
  end(options?: { output?: unknown; usage?: UsageDetails; model?: string }): void;
}
```

### Span

```typescript
interface SpanOptions {
  name: string;
  id?: string;
  parentObservationId?: string;
  metadata?: Record<string, unknown>;
  input?: unknown;
  level?: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
}

class Span {
  readonly id: string;
  readonly traceId: string;
  readonly name: string;
  input?: unknown;
  output?: unknown;

  end(options?: { output?: unknown }): void;
}
```

### Agent Features Types

```typescript
// Checkpoints
type CheckpointType = "manual" | "auto" | "tool_call" | "error" | "milestone" | "restore";

interface CheckpointOptions {
  name: string;
  type?: CheckpointType;
  observationId?: string;
  description?: string;
  files?: string[];
  includeGitInfo?: boolean;
}

interface CheckpointInfo {
  id: string;
  name: string;
  type: CheckpointType;
  traceId: string;
  observationId?: string;
  gitCommitSha?: string;
  gitBranch?: string;
  filesChanged: string[];
  totalFiles: number;
  totalSizeBytes: number;
  createdAt: Date;
}

// Git Links
type GitLinkType = "start" | "commit" | "restore" | "branch" | "diff";

interface GitLinkOptions {
  type?: GitLinkType;
  observationId?: string;
  commitSha?: string;
  branch?: string;
  repoUrl?: string;
  commitMessage?: string;
  filesChanged?: string[];
  autoDetect?: boolean;
}

interface GitLinkInfo {
  id: string;
  traceId: string;
  observationId?: string;
  type: GitLinkType;
  commitSha: string;
  branch?: string;
  repoUrl?: string;
  commitMessage?: string;
  authorName?: string;
  authorEmail?: string;
  filesChanged: string[];
  createdAt: Date;
}

// File Operations
type FileOperationType = "create" | "read" | "update" | "delete" | "rename" | "copy" | "move" | "chmod";

interface FileOperationOptions {
  operation: FileOperationType;
  filePath: string;
  observationId?: string;
  newPath?: string;
  contentBefore?: string;
  contentAfter?: string;
  linesAdded?: number;
  linesRemoved?: number;
  toolName?: string;
  reason?: string;
  success?: boolean;
  errorMessage?: string;
}

interface FileOperationInfo {
  id: string;
  traceId: string;
  operation: FileOperationType;
  filePath: string;
  fileSize: number;
  linesAdded: number;
  linesRemoved: number;
  success: boolean;
  durationMs: number;
}

// Terminal Commands
interface TerminalCommandOptions {
  command: string;
  args?: string[];
  observationId?: string;
  workingDirectory?: string;
  exitCode?: number;
  stdout?: string;
  stderr?: string;
  toolName?: string;
  reason?: string;
}

interface RunCommandOptions {
  args?: string[];
  workingDirectory?: string;
  env?: Record<string, string>;
  timeout?: number;
  shell?: boolean | string;
  maxOutputBytes?: number;
}

interface RunCommandResult {
  info: TerminalCommandInfo;
  exitCode: number;
  stdout: string;
  stderr: string;
}
```

### Prompts

```typescript
interface GetPromptOptions {
  name: string;
  version?: number;
  label?: string;
  fallback?: string;
  cacheTtl?: number;
}

class PromptVersion {
  readonly id: string;
  readonly version: number;
  readonly prompt: string;
  readonly config: Record<string, unknown>;
  readonly labels: string[];

  compile(variables: Record<string, unknown>): string;
  compileChat(variables: Record<string, unknown>): ChatMessage[];
  getVariables(): string[];
}

interface ChatMessage {
  role: "system" | "user" | "assistant" | "function";
  content: string;
}
```
