/**
 * Context propagation using AsyncLocalStorage.
 */

import { AsyncLocalStorage } from "async_hooks";
import type { AgentTrace, Trace, Span, Generation } from "./client";

// AsyncLocalStorage for context propagation
const clientStorage = new AsyncLocalStorage<AgentTrace | null>();
const traceStorage = new AsyncLocalStorage<Trace | null>();
const observationStorage = new AsyncLocalStorage<Span | Generation | null>();

// Fallback for non-async contexts
let globalClient: AgentTrace | null = null;
let globalTrace: Trace | null = null;
let globalObservation: Span | Generation | null = null;

/**
 * Set the global AgentTrace client.
 */
export function setClient(client: AgentTrace | null): void {
  globalClient = client;
}

/**
 * Get the global AgentTrace client.
 */
export function getClient(): AgentTrace | null {
  return clientStorage.getStore() ?? globalClient;
}

/**
 * Set the current trace in context.
 */
export function setCurrentTrace(trace: Trace | null): void {
  globalTrace = trace;
}

/**
 * Get the current trace from context.
 */
export function getCurrentTrace(): Trace | null {
  return traceStorage.getStore() ?? globalTrace;
}

/**
 * Set the current observation in context.
 */
export function setCurrentObservation(observation: Span | Generation | null): void {
  globalObservation = observation;
}

/**
 * Get the current observation from context.
 */
export function getCurrentObservation(): Span | Generation | null {
  return observationStorage.getStore() ?? globalObservation;
}

/**
 * Run a function within a trace context.
 *
 * @example
 * ```typescript
 * const result = await runWithTrace(trace, async () => {
 *   // Code here has access to the trace via getCurrentTrace()
 *   return await doSomething();
 * });
 * ```
 */
export function runWithTrace<T>(trace: Trace, fn: () => T): T {
  return traceStorage.run(trace, fn);
}

/**
 * Run a function within an observation context.
 *
 * @example
 * ```typescript
 * const result = await runWithObservation(span, async () => {
 *   // Code here has access to the observation via getCurrentObservation()
 *   return await doSomething();
 * });
 * ```
 */
export function runWithObservation<T>(observation: Span | Generation, fn: () => T): T {
  return observationStorage.run(observation, fn);
}

/**
 * Run a function within a client context.
 */
export function runWithClient<T>(client: AgentTrace, fn: () => T): T {
  return clientStorage.run(client, fn);
}

/**
 * Context wrapper class for managing trace context.
 */
export class TraceContext {
  private trace: Trace;
  private previousTrace: Trace | null = null;

  constructor(trace: Trace) {
    this.trace = trace;
  }

  /**
   * Enter the context.
   */
  enter(): Trace {
    this.previousTrace = globalTrace;
    globalTrace = this.trace;
    return this.trace;
  }

  /**
   * Exit the context.
   */
  exit(): void {
    globalTrace = this.previousTrace;
    this.previousTrace = null;
  }

  /**
   * Run a function within this trace context.
   */
  run<T>(fn: () => T): T {
    this.enter();
    try {
      return fn();
    } finally {
      this.exit();
    }
  }

  /**
   * Run an async function within this trace context.
   */
  async runAsync<T>(fn: () => Promise<T>): Promise<T> {
    return runWithTrace(this.trace, fn);
  }
}

/**
 * Context wrapper class for managing observation context.
 */
export class ObservationContext {
  private observation: Span | Generation;
  private previousObservation: Span | Generation | null = null;

  constructor(observation: Span | Generation) {
    this.observation = observation;
  }

  /**
   * Enter the context.
   */
  enter(): Span | Generation {
    this.previousObservation = globalObservation;
    globalObservation = this.observation;
    return this.observation;
  }

  /**
   * Exit the context.
   */
  exit(): void {
    globalObservation = this.previousObservation;
    this.previousObservation = null;
  }

  /**
   * Run a function within this observation context.
   */
  run<T>(fn: () => T): T {
    this.enter();
    try {
      return fn();
    } finally {
      this.exit();
    }
  }

  /**
   * Run an async function within this observation context.
   */
  async runAsync<T>(fn: () => Promise<T>): Promise<T> {
    return runWithObservation(this.observation, fn);
  }
}
