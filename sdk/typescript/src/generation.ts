/**
 * Generation helpers for tracking LLM calls.
 */

import {
  getClient,
  getCurrentTrace,
  getCurrentObservation,
  setCurrentObservation,
} from "./context";
import type { Generation, UsageDetails } from "./client";

export interface GenerationContextOptions {
  /** Name of the generation */
  name: string;
  /** Model name/identifier */
  model?: string;
  /** Model parameters (temperature, max_tokens, etc.) */
  modelParameters?: Record<string, unknown>;
  /** Input prompt or messages */
  input?: unknown;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
  /** Log level */
  level?: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
}

/**
 * Context object for managing a generation's lifecycle.
 */
export class GenerationContext {
  private _generation: Generation | null;
  private _previousObservation: ReturnType<typeof getCurrentObservation>;
  private _ended: boolean = false;

  constructor(
    generation: Generation | null,
    previousObservation: ReturnType<typeof getCurrentObservation>
  ) {
    this._generation = generation;
    this._previousObservation = previousObservation;
  }

  /**
   * Get the generation ID.
   */
  get id(): string {
    return this._generation?.id ?? "";
  }

  /**
   * Get the trace ID.
   */
  get traceId(): string {
    return this._generation?.traceId ?? "";
  }

  /**
   * Update the generation with additional data.
   */
  update(updates: {
    output?: unknown;
    usage?: UsageDetails;
    model?: string;
    metadata?: Record<string, unknown>;
  }): this {
    if (this._generation && !this._ended) {
      this._generation.update(updates);
    }
    return this;
  }

  /**
   * End the generation.
   */
  end(options?: { output?: unknown; usage?: UsageDetails; model?: string }): void {
    if (this._generation && !this._ended) {
      this._generation.end(options);
      this._ended = true;
      setCurrentObservation(this._previousObservation);
    }
  }

  /**
   * Check if the generation has ended.
   */
  get ended(): boolean {
    return this._ended;
  }
}

/**
 * Start a generation for tracking an LLM call.
 *
 * This is useful when you need more control over when to end the generation.
 * Remember to call .end() when done!
 *
 * @example
 * ```typescript
 * const gen = startGeneration({
 *   name: "chat-completion",
 *   model: "gpt-4",
 *   input: { messages }
 * });
 *
 * try {
 *   const response = await openai.chat.completions.create({ messages });
 *   gen.end({
 *     output: response.choices[0].message.content,
 *     usage: {
 *       inputTokens: response.usage.prompt_tokens,
 *       outputTokens: response.usage.completion_tokens,
 *     }
 *   });
 * } catch (error) {
 *   gen.end({ output: { error: String(error) } });
 *   throw error;
 * }
 * ```
 */
export function startGeneration(options: GenerationContextOptions): GenerationContext {
  const client = getClient();
  if (client === null || !client.enabled) {
    return new GenerationContext(null, null);
  }

  let trace = getCurrentTrace();
  const parentObservation = getCurrentObservation();

  if (trace === null) {
    trace = client.trace({ name: options.name, input: options.input });
  }

  const generation = trace.generation({
    name: options.name,
    model: options.model,
    modelParameters: options.modelParameters,
    input: options.input,
    metadata: options.metadata,
    parentObservationId: parentObservation?.id,
    level: options.level,
  });

  setCurrentObservation(generation);

  return new GenerationContext(generation, parentObservation);
}

/**
 * Run a function as a generation with automatic start/end.
 *
 * @example
 * ```typescript
 * const response = await withGeneration(
 *   { name: "chat", model: "gpt-4", input: messages },
 *   async (gen) => {
 *     const response = await openai.chat.completions.create({ messages });
 *     gen.update({
 *       output: response.choices[0].message.content,
 *       usage: { inputTokens: response.usage.prompt_tokens, outputTokens: response.usage.completion_tokens }
 *     });
 *     return response;
 *   }
 * );
 * ```
 */
export async function withGeneration<T>(
  options: GenerationContextOptions,
  fn: (gen: GenerationContext) => Promise<T>
): Promise<T> {
  const gen = startGeneration(options);

  try {
    const result = await fn(gen);
    if (!gen.ended) {
      gen.end();
    }
    return result;
  } catch (error) {
    if (!gen.ended) {
      gen.end({ output: { error: String(error) } });
    }
    throw error;
  }
}

/**
 * Synchronous version of withGeneration.
 */
export function withGenerationSync<T>(
  options: GenerationContextOptions,
  fn: (gen: GenerationContext) => T
): T {
  const gen = startGeneration(options);

  try {
    const result = fn(gen);
    if (!gen.ended) {
      gen.end();
    }
    return result;
  } catch (error) {
    if (!gen.ended) {
      gen.end({ output: { error: String(error) } });
    }
    throw error;
  }
}
