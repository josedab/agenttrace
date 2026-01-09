/**
 * observe() wrapper for automatic tracing.
 */

import {
  getClient,
  getCurrentTrace,
  getCurrentObservation,
  setCurrentObservation,
} from "./context";
import type { Span, Generation } from "./client";

export interface ObserveOptions {
  /** Custom name for the observation (defaults to function name) */
  name?: string;
  /** Type of observation ("span" or "generation") */
  asType?: "span" | "generation";
  /** Whether to capture function arguments */
  captureInput?: boolean;
  /** Whether to capture function return value */
  captureOutput?: boolean;
  /** Model name for generations */
  model?: string;
  /** Model parameters for generations */
  modelParameters?: Record<string, unknown>;
}

type AnyFunction = (...args: unknown[]) => unknown;
type AsyncFunction = (...args: unknown[]) => Promise<unknown>;

/**
 * Wrap a function with automatic tracing.
 *
 * @example
 * ```typescript
 * // Basic usage
 * const myFunction = observe(async (query: string) => {
 *   const response = await llm.generate(query);
 *   return response;
 * });
 *
 * // With options
 * const myLLMCall = observe(
 *   async (messages: Message[]) => {
 *     return await openai.chat.completions.create({ messages });
 *   },
 *   { name: "chat-completion", asType: "generation", model: "gpt-4" }
 * );
 * ```
 */
export function observe<T extends AnyFunction>(
  fn: T,
  options?: ObserveOptions
): T {
  const observationName = options?.name ?? fn.name ?? "anonymous";
  const observationType = options?.asType ?? "span";
  const captureInput = options?.captureInput ?? true;
  const captureOutput = options?.captureOutput ?? true;

  const isAsync = fn.constructor.name === "AsyncFunction";

  if (isAsync) {
    const wrappedAsync = async function (
      this: unknown,
      ...args: unknown[]
    ): Promise<unknown> {
      return traceAsync(
        fn as AsyncFunction,
        args,
        observationName,
        observationType,
        captureInput,
        captureOutput,
        options,
        this
      );
    };

    // Preserve function name and length
    Object.defineProperty(wrappedAsync, "name", { value: fn.name });
    Object.defineProperty(wrappedAsync, "length", { value: fn.length });

    return wrappedAsync as T;
  }

  const wrappedSync = function (this: unknown, ...args: unknown[]): unknown {
    return traceSync(
      fn,
      args,
      observationName,
      observationType,
      captureInput,
      captureOutput,
      options,
      this
    );
  };

  Object.defineProperty(wrappedSync, "name", { value: fn.name });
  Object.defineProperty(wrappedSync, "length", { value: fn.length });

  return wrappedSync as T;
}

function traceSync(
  fn: AnyFunction,
  args: unknown[],
  name: string,
  type: "span" | "generation",
  captureInput: boolean,
  captureOutput: boolean,
  options: ObserveOptions | undefined,
  context: unknown
): unknown {
  const client = getClient();
  if (client === null || !client.enabled) {
    return fn.apply(context, args);
  }

  let trace = getCurrentTrace();
  const parentObservation = getCurrentObservation();

  // Prepare input
  const inputData = captureInput ? captureArgs(fn, args) : undefined;

  let observation: Span | Generation | null = null;

  if (trace === null) {
    // Create a new trace
    trace = client.trace({ name, input: inputData });
  } else {
    const parentId = parentObservation?.id;

    if (type === "generation") {
      observation = trace.generation({
        name,
        input: inputData,
        parentObservationId: parentId,
        model: options?.model,
        modelParameters: options?.modelParameters,
      });
    } else {
      observation = trace.span({
        name,
        input: inputData,
        parentObservationId: parentId,
      });
    }

    // Set as current observation
    setCurrentObservation(observation);
  }

  try {
    const result = fn.apply(context, args);

    const output = captureOutput ? result : undefined;

    if (observation) {
      observation.end({ output });
    } else {
      trace.end({ output });
    }

    // Restore previous observation
    setCurrentObservation(parentObservation);

    return result;
  } catch (error) {
    const errorOutput = { error: String(error) };

    if (observation) {
      observation.end({ output: errorOutput });
    } else {
      trace.end({ output: errorOutput });
    }

    setCurrentObservation(parentObservation);
    throw error;
  }
}

async function traceAsync(
  fn: AsyncFunction,
  args: unknown[],
  name: string,
  type: "span" | "generation",
  captureInput: boolean,
  captureOutput: boolean,
  options: ObserveOptions | undefined,
  context: unknown
): Promise<unknown> {
  const client = getClient();
  if (client === null || !client.enabled) {
    return fn.apply(context, args);
  }

  let trace = getCurrentTrace();
  const parentObservation = getCurrentObservation();

  const inputData = captureInput ? captureArgs(fn, args) : undefined;

  let observation: Span | Generation | null = null;

  if (trace === null) {
    trace = client.trace({ name, input: inputData });
  } else {
    const parentId = parentObservation?.id;

    if (type === "generation") {
      observation = trace.generation({
        name,
        input: inputData,
        parentObservationId: parentId,
        model: options?.model,
        modelParameters: options?.modelParameters,
      });
    } else {
      observation = trace.span({
        name,
        input: inputData,
        parentObservationId: parentId,
      });
    }

    setCurrentObservation(observation);
  }

  try {
    const result = await fn.apply(context, args);

    const output = captureOutput ? result : undefined;

    if (observation) {
      observation.end({ output });
    } else {
      trace.end({ output });
    }

    setCurrentObservation(parentObservation);

    return result;
  } catch (error) {
    const errorOutput = { error: String(error) };

    if (observation) {
      observation.end({ output: errorOutput });
    } else {
      trace.end({ output: errorOutput });
    }

    setCurrentObservation(parentObservation);
    throw error;
  }
}

/**
 * Capture function arguments as a dictionary.
 */
function captureArgs(fn: AnyFunction, args: unknown[]): Record<string, unknown> {
  try {
    // Try to get parameter names from function source
    const fnStr = fn.toString();
    const paramMatch = fnStr.match(/\(([^)]*)\)/);

    if (paramMatch && paramMatch[1]) {
      const paramNames = paramMatch[1]
        .split(",")
        .map((p) => p.trim().split(/[=:]/)[0].trim())
        .filter((p) => p && p !== "this");

      const result: Record<string, unknown> = {};
      paramNames.forEach((name, i) => {
        if (i < args.length) {
          result[name] = serializeValue(args[i]);
        }
      });

      return result;
    }
  } catch {
    // Fall through to default
  }

  // Fallback: use positional args
  return {
    args: args.map(serializeValue),
  };
}

/**
 * Serialize a value for JSON encoding.
 */
function serializeValue(value: unknown): unknown {
  if (
    value === null ||
    value === undefined ||
    typeof value === "string" ||
    typeof value === "number" ||
    typeof value === "boolean"
  ) {
    return value;
  }

  if (Array.isArray(value)) {
    return value.map(serializeValue);
  }

  if (typeof value === "object") {
    const result: Record<string, unknown> = {};
    for (const [key, val] of Object.entries(value)) {
      result[key] = serializeValue(val);
    }
    return result;
  }

  try {
    return String(value);
  } catch {
    return "<unserializable>";
  }
}
