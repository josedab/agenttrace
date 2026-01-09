/**
 * AgentTrace TypeScript SDK - Observability for AI Coding Agents
 *
 * @example
 * ```typescript
 * import { AgentTrace, observe, startGeneration } from "agenttrace";
 *
 * const client = new AgentTrace({ apiKey: "your-api-key" });
 *
 * // Using observe wrapper
 * const myFunction = observe(async (query: string) => {
 *   const response = await llm.generate(query);
 *   return response;
 * });
 *
 * // Using startGeneration for manual control
 * const gen = startGeneration({
 *   name: "chat",
 *   model: "gpt-4",
 *   input: messages
 * });
 *
 * const response = await openai.chat.completions.create({ messages });
 * gen.end({
 *   output: response.choices[0].message.content,
 *   usage: {
 *     inputTokens: response.usage.prompt_tokens,
 *     outputTokens: response.usage.completion_tokens
 *   }
 * });
 *
 * await client.flush();
 * ```
 */

// Main client
export {
  AgentTrace,
  Trace,
  Span,
  Generation,
  type AgentTraceConfig,
  type TraceOptions,
  type SpanOptions,
  type GenerationOptions,
  type ScoreOptions,
  type UsageDetails,
} from "./client";

// Context
export {
  getClient,
  setClient,
  getCurrentTrace,
  setCurrentTrace,
  getCurrentObservation,
  setCurrentObservation,
  runWithTrace,
  runWithObservation,
  runWithClient,
  TraceContext,
  ObservationContext,
} from "./context";

// Observe wrapper
export { observe, type ObserveOptions } from "./observe";

// Generation helpers
export {
  startGeneration,
  withGeneration,
  withGenerationSync,
  GenerationContext,
  type GenerationContextOptions,
} from "./generation";

// Prompt management
export {
  Prompt,
  PromptVersion,
  getPrompt,
  type PromptVersionData,
  type ChatMessage,
  type GetPromptOptions,
} from "./prompt";

// Transport (for advanced use)
export { HttpTransport, type HttpTransportConfig, type BatchEvent } from "./transport/http";
export { BatchQueue, type BatchQueueConfig } from "./transport/batch";

// Agent-specific features - Checkpoints
export {
  CheckpointClient,
  withCheckpoint,
  type CheckpointType,
  type CheckpointOptions,
  type CheckpointInfo,
  type GitInfo,
} from "./checkpoint";

// Agent-specific features - Git Links
export {
  GitClient,
  type GitLinkType,
  type GitLinkOptions,
  type GitLinkInfo,
} from "./git";

// Agent-specific features - File Operations
export {
  FileOperationClient,
  withFileOp,
  type FileOperationType,
  type FileOperationOptions,
  type FileOperationInfo,
} from "./fileops";

// Agent-specific features - Terminal Commands
export {
  TerminalClient,
  runCommand,
  type TerminalCommandOptions,
  type TerminalCommandInfo,
  type RunCommandOptions,
  type RunCommandResult,
} from "./terminal";

// Version
export const VERSION = "0.1.0";
