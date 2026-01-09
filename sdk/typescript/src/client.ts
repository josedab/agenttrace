/**
 * AgentTrace TypeScript SDK - Main Client
 */

import { HttpTransport } from "./transport/http";
import { BatchQueue } from "./transport/batch";
import { setClient, setCurrentTrace } from "./context";
import { CheckpointClient, type CheckpointOptions, type CheckpointInfo } from "./checkpoint";
import { GitClient, type GitLinkOptions, type GitLinkInfo } from "./git";
import { FileOperationClient, type FileOperationOptions, type FileOperationInfo } from "./fileops";
import { TerminalClient, type TerminalCommandOptions, type TerminalCommandInfo, type RunCommandOptions, type RunCommandResult } from "./terminal";

export interface AgentTraceConfig {
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

export interface TraceOptions {
  name: string;
  id?: string;
  userId?: string;
  sessionId?: string;
  metadata?: Record<string, unknown>;
  tags?: string[];
  input?: unknown;
  public?: boolean;
}

export interface SpanOptions {
  name: string;
  id?: string;
  parentObservationId?: string;
  metadata?: Record<string, unknown>;
  input?: unknown;
  level?: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
}

export interface GenerationOptions {
  name: string;
  id?: string;
  parentObservationId?: string;
  model?: string;
  modelParameters?: Record<string, unknown>;
  input?: unknown;
  metadata?: Record<string, unknown>;
  level?: "DEBUG" | "DEFAULT" | "WARNING" | "ERROR";
}

export interface ScoreOptions {
  traceId: string;
  name: string;
  value: number | boolean | string;
  observationId?: string;
  dataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
  comment?: string;
}

export interface UsageDetails {
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
}

/**
 * Main AgentTrace client for tracing and observability.
 *
 * @example
 * ```typescript
 * const client = new AgentTrace({
 *   apiKey: "your-api-key",
 *   host: "https://api.agenttrace.io"
 * });
 *
 * const trace = client.trace({ name: "my-trace" });
 * const generation = trace.generation({
 *   name: "llm-call",
 *   model: "gpt-4",
 *   input: { query: "Hello" }
 * });
 * generation.end({ output: "Hi there!" });
 * trace.end();
 *
 * client.flush();
 * ```
 */
export class AgentTrace {
  public readonly apiKey: string;
  public readonly host: string;
  public readonly publicKey?: string;
  public readonly projectId?: string;
  public readonly enabled: boolean;

  private _transport: HttpTransport;
  private _batchQueue: BatchQueue;

  constructor(config: AgentTraceConfig) {
    this.apiKey = config.apiKey;
    this.host = (config.host || "https://api.agenttrace.io").replace(/\/$/, "");
    this.publicKey = config.publicKey;
    this.projectId = config.projectId;
    this.enabled = config.enabled ?? true;

    this._transport = new HttpTransport({
      host: this.host,
      apiKey: this.apiKey,
      timeout: config.timeout ?? 10000,
      maxRetries: config.maxRetries ?? 3,
    });

    this._batchQueue = new BatchQueue({
      transport: this._transport,
      flushAt: config.flushAt ?? 20,
      flushInterval: config.flushInterval ?? 5000,
    });

    // Set as global client
    setClient(this);

    // Flush on process exit
    if (typeof process !== "undefined") {
      process.on("beforeExit", () => this.flush());
      process.on("SIGINT", () => {
        this.shutdown();
        process.exit(0);
      });
      process.on("SIGTERM", () => {
        this.shutdown();
        process.exit(0);
      });
    }
  }

  /**
   * Create a new trace.
   */
  trace(options: TraceOptions): Trace {
    const trace = new Trace({
      client: this,
      id: options.id || generateId(),
      name: options.name,
      userId: options.userId,
      sessionId: options.sessionId,
      metadata: options.metadata || {},
      tags: options.tags || [],
      input: options.input,
      public: options.public ?? false,
    });

    setCurrentTrace(trace);
    return trace;
  }

  /**
   * Submit a score for a trace or observation.
   */
  score(options: ScoreOptions): void {
    if (!this.enabled) return;

    this._batchQueue.add({
      type: "score-create",
      body: {
        id: generateId(),
        traceId: options.traceId,
        observationId: options.observationId,
        name: options.name,
        value: options.value,
        dataType: options.dataType || "NUMERIC",
        comment: options.comment,
        source: "API",
      },
    });
  }

  /**
   * Flush all pending events to the server.
   */
  flush(): Promise<void> {
    return this._batchQueue.flush();
  }

  /**
   * Shutdown the client and flush remaining events.
   */
  async shutdown(): Promise<void> {
    await this.flush();
    this._batchQueue.stop();
  }

  /** @internal */
  _addEvent(event: Record<string, unknown>): void {
    this._batchQueue.add(event);
  }
}

/**
 * Represents a trace in AgentTrace.
 */
export class Trace {
  public readonly id: string;
  public name: string;
  public userId?: string;
  public sessionId?: string;
  public metadata: Record<string, unknown>;
  public tags: string[];
  public input?: unknown;
  public output?: unknown;
  public public: boolean;
  public readonly startTime: Date;
  public endTime?: Date;

  private _client: AgentTrace;
  private _ended: boolean = false;

  constructor(options: {
    client: AgentTrace;
    id: string;
    name: string;
    userId?: string;
    sessionId?: string;
    metadata: Record<string, unknown>;
    tags: string[];
    input?: unknown;
    public: boolean;
  }) {
    this._client = options.client;
    this.id = options.id;
    this.name = options.name;
    this.userId = options.userId;
    this.sessionId = options.sessionId;
    this.metadata = options.metadata;
    this.tags = options.tags;
    this.input = options.input;
    this.public = options.public;
    this.startTime = new Date();

    this._sendCreate();
  }

  private _sendCreate(): void {
    if (!this._client.enabled) return;

    this._client._addEvent({
      type: "trace-create",
      body: {
        id: this.id,
        name: this.name,
        userId: this.userId,
        sessionId: this.sessionId,
        metadata: this.metadata,
        tags: this.tags,
        input: this.input,
        public: this.public,
        timestamp: this.startTime.toISOString(),
      },
    });
  }

  /**
   * Create a span within this trace.
   */
  span(options: SpanOptions): Span {
    return new Span({
      client: this._client,
      traceId: this.id,
      id: options.id || generateId(),
      name: options.name,
      parentObservationId: options.parentObservationId,
      metadata: options.metadata || {},
      input: options.input,
      level: options.level || "DEFAULT",
    });
  }

  /**
   * Create a generation (LLM call) within this trace.
   */
  generation(options: GenerationOptions): Generation {
    return new Generation({
      client: this._client,
      traceId: this.id,
      id: options.id || generateId(),
      name: options.name,
      parentObservationId: options.parentObservationId,
      model: options.model,
      modelParameters: options.modelParameters || {},
      input: options.input,
      metadata: options.metadata || {},
      level: options.level || "DEFAULT",
    });
  }

  /**
   * Update trace properties.
   */
  update(updates: Partial<{
    name: string;
    userId: string;
    sessionId: string;
    metadata: Record<string, unknown>;
    tags: string[];
    input: unknown;
    output: unknown;
    public: boolean;
  }>): this {
    if (updates.name !== undefined) this.name = updates.name;
    if (updates.userId !== undefined) this.userId = updates.userId;
    if (updates.sessionId !== undefined) this.sessionId = updates.sessionId;
    if (updates.metadata !== undefined) Object.assign(this.metadata, updates.metadata);
    if (updates.tags !== undefined) this.tags = updates.tags;
    if (updates.input !== undefined) this.input = updates.input;
    if (updates.output !== undefined) this.output = updates.output;
    if (updates.public !== undefined) this.public = updates.public;

    if (this._client.enabled) {
      this._client._addEvent({
        type: "trace-update",
        body: {
          id: this.id,
          name: this.name,
          userId: this.userId,
          sessionId: this.sessionId,
          metadata: this.metadata,
          tags: this.tags,
          input: this.input,
          output: this.output,
          public: this.public,
        },
      });
    }

    return this;
  }

  /**
   * End the trace.
   */
  end(options?: { output?: unknown }): void {
    if (this._ended) return;

    this._ended = true;
    this.endTime = new Date();
    if (options?.output !== undefined) {
      this.output = options.output;
    }

    this.update({ output: this.output });
    setCurrentTrace(null);
  }

  /**
   * Add a score to this trace.
   */
  score(name: string, value: number | boolean | string, options?: {
    dataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
    comment?: string;
  }): void {
    this._client.score({
      traceId: this.id,
      name,
      value,
      dataType: options?.dataType,
      comment: options?.comment,
    });
  }

  /**
   * Create a checkpoint for this trace.
   */
  checkpoint(options: CheckpointOptions): CheckpointInfo {
    const cpClient = new CheckpointClient(this._client);
    return cpClient.create(this.id, options);
  }

  /**
   * Create a git link for this trace.
   */
  gitLink(options: GitLinkOptions = {}): GitLinkInfo {
    const gitClient = new GitClient(this._client);
    return gitClient.link(this.id, options);
  }

  /**
   * Track a file operation for this trace.
   */
  fileOp(options: FileOperationOptions): FileOperationInfo {
    const fileClient = new FileOperationClient(this._client);
    return fileClient.track(this.id, options);
  }

  /**
   * Track a terminal command for this trace.
   */
  terminalCmd(options: TerminalCommandOptions): TerminalCommandInfo {
    const termClient = new TerminalClient(this._client);
    return termClient.track(this.id, options);
  }

  /**
   * Run a command and track it.
   */
  async runCmd(command: string, options: RunCommandOptions = {}): Promise<RunCommandResult> {
    const termClient = new TerminalClient(this._client);
    return termClient.run(this.id, command, options);
  }
}

/**
 * Represents a span within a trace.
 */
export class Span {
  public readonly id: string;
  public readonly traceId: string;
  public readonly name: string;
  public readonly parentObservationId?: string;
  public metadata: Record<string, unknown>;
  public input?: unknown;
  public output?: unknown;
  public readonly level: string;
  public readonly startTime: Date;
  public endTime?: Date;

  private _client: AgentTrace;
  private _ended: boolean = false;

  constructor(options: {
    client: AgentTrace;
    traceId: string;
    id: string;
    name: string;
    parentObservationId?: string;
    metadata: Record<string, unknown>;
    input?: unknown;
    level: string;
  }) {
    this._client = options.client;
    this.traceId = options.traceId;
    this.id = options.id;
    this.name = options.name;
    this.parentObservationId = options.parentObservationId;
    this.metadata = options.metadata;
    this.input = options.input;
    this.level = options.level;
    this.startTime = new Date();

    this._sendCreate();
  }

  private _sendCreate(): void {
    if (!this._client.enabled) return;

    this._client._addEvent({
      type: "span-create",
      body: {
        id: this.id,
        traceId: this.traceId,
        parentObservationId: this.parentObservationId,
        name: this.name,
        metadata: this.metadata,
        input: this.input,
        level: this.level,
        startTime: this.startTime.toISOString(),
      },
    });
  }

  /**
   * End the span.
   */
  end(options?: { output?: unknown }): void {
    if (this._ended) return;

    this._ended = true;
    this.endTime = new Date();
    if (options?.output !== undefined) {
      this.output = options.output;
    }

    if (this._client.enabled) {
      this._client._addEvent({
        type: "span-update",
        body: {
          id: this.id,
          output: this.output,
          endTime: this.endTime.toISOString(),
        },
      });
    }
  }
}

/**
 * Represents an LLM generation within a trace.
 */
export class Generation {
  public readonly id: string;
  public readonly traceId: string;
  public readonly name: string;
  public readonly parentObservationId?: string;
  public model?: string;
  public modelParameters: Record<string, unknown>;
  public input?: unknown;
  public output?: unknown;
  public metadata: Record<string, unknown>;
  public readonly level: string;
  public readonly startTime: Date;
  public endTime?: Date;
  public usage?: UsageDetails;

  private _client: AgentTrace;
  private _ended: boolean = false;

  constructor(options: {
    client: AgentTrace;
    traceId: string;
    id: string;
    name: string;
    parentObservationId?: string;
    model?: string;
    modelParameters: Record<string, unknown>;
    input?: unknown;
    metadata: Record<string, unknown>;
    level: string;
  }) {
    this._client = options.client;
    this.traceId = options.traceId;
    this.id = options.id;
    this.name = options.name;
    this.parentObservationId = options.parentObservationId;
    this.model = options.model;
    this.modelParameters = options.modelParameters;
    this.input = options.input;
    this.metadata = options.metadata;
    this.level = options.level;
    this.startTime = new Date();

    this._sendCreate();
  }

  private _sendCreate(): void {
    if (!this._client.enabled) return;

    this._client._addEvent({
      type: "generation-create",
      body: {
        id: this.id,
        traceId: this.traceId,
        parentObservationId: this.parentObservationId,
        name: this.name,
        model: this.model,
        modelParameters: this.modelParameters,
        input: this.input,
        metadata: this.metadata,
        level: this.level,
        startTime: this.startTime.toISOString(),
      },
    });
  }

  /**
   * Update the generation.
   */
  update(updates: Partial<{
    output: unknown;
    usage: UsageDetails;
    model: string;
    metadata: Record<string, unknown>;
  }>): this {
    if (updates.output !== undefined) this.output = updates.output;
    if (updates.usage !== undefined) this.usage = updates.usage;
    if (updates.model !== undefined) this.model = updates.model;
    if (updates.metadata !== undefined) Object.assign(this.metadata, updates.metadata);
    return this;
  }

  /**
   * End the generation.
   */
  end(options?: {
    output?: unknown;
    usage?: UsageDetails;
    model?: string;
  }): void {
    if (this._ended) return;

    this._ended = true;
    this.endTime = new Date();

    if (options?.output !== undefined) this.output = options.output;
    if (options?.usage !== undefined) this.usage = options.usage;
    if (options?.model !== undefined) this.model = options.model;

    if (this._client.enabled) {
      this._client._addEvent({
        type: "generation-update",
        body: {
          id: this.id,
          output: this.output,
          usage: this.usage,
          model: this.model,
          endTime: this.endTime.toISOString(),
        },
      });
    }
  }
}

/**
 * Generate a unique ID.
 */
function generateId(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  // Fallback for older environments
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
