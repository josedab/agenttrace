/**
 * AgentTrace Browser SDK - User Feedback and Client-Side Tracing
 *
 * This SDK is designed for browser usage, providing:
 * - User feedback collection
 * - Client-side score submission
 * - Public trace viewing
 *
 * @example
 * ```typescript
 * import { AgentTraceBrowser } from "agenttrace/browser";
 *
 * const client = new AgentTraceBrowser({
 *   publicKey: "your-public-key",
 *   host: "https://api.agenttrace.io"
 * });
 *
 * // Submit user feedback
 * await client.submitFeedback({
 *   traceId: "trace-123",
 *   name: "user-rating",
 *   value: 5,
 *   comment: "Very helpful response!"
 * });
 * ```
 */

export interface AgentTraceBrowserConfig {
  publicKey: string;
  host?: string;
  timeout?: number;
}

export interface FeedbackOptions {
  traceId: string;
  name: string;
  value: number | boolean | string;
  observationId?: string;
  dataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
  comment?: string;
}

export interface PublicTrace {
  id: string;
  name: string;
  input?: unknown;
  output?: unknown;
  metadata: Record<string, unknown>;
  startTime: string;
  endTime?: string;
}

/**
 * Browser SDK for AgentTrace - User Feedback and Public APIs
 */
export class AgentTraceBrowser {
  public readonly publicKey: string;
  public readonly host: string;
  private timeout: number;

  constructor(config: AgentTraceBrowserConfig) {
    this.publicKey = config.publicKey;
    this.host = (config.host || "https://api.agenttrace.io").replace(/\/$/, "");
    this.timeout = config.timeout ?? 10000;
  }

  private getHeaders(): Record<string, string> {
    return {
      "Content-Type": "application/json",
      "X-Langfuse-Public-Key": this.publicKey,
      "User-Agent": "agenttrace-browser/0.1.0",
    };
  }

  /**
   * Submit user feedback for a trace.
   *
   * @example
   * ```typescript
   * // Submit a rating
   * await client.submitFeedback({
   *   traceId: "trace-123",
   *   name: "rating",
   *   value: 5,
   *   dataType: "NUMERIC"
   * });
   *
   * // Submit a thumbs up/down
   * await client.submitFeedback({
   *   traceId: "trace-123",
   *   name: "helpful",
   *   value: true,
   *   dataType: "BOOLEAN"
   * });
   *
   * // Submit categorized feedback
   * await client.submitFeedback({
   *   traceId: "trace-123",
   *   name: "sentiment",
   *   value: "positive",
   *   dataType: "CATEGORICAL"
   * });
   * ```
   */
  async submitFeedback(options: FeedbackOptions): Promise<boolean> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.timeout);

      const response = await fetch(`${this.host}/api/public/scores`, {
        method: "POST",
        headers: this.getHeaders(),
        body: JSON.stringify({
          id: generateId(),
          traceId: options.traceId,
          observationId: options.observationId,
          name: options.name,
          value: options.value,
          dataType: options.dataType || inferDataType(options.value),
          comment: options.comment,
          source: "USER",
        }),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      return response.status === 200 || response.status === 201;
    } catch (error) {
      console.error("Failed to submit feedback:", error);
      return false;
    }
  }

  /**
   * Get a public trace by ID.
   *
   * Note: Only traces marked as public can be retrieved.
   */
  async getPublicTrace(traceId: string): Promise<PublicTrace | null> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.timeout);

      const response = await fetch(`${this.host}/api/public/traces/${traceId}`, {
        method: "GET",
        headers: this.getHeaders(),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (response.status === 200) {
        return (await response.json()) as PublicTrace;
      }

      return null;
    } catch (error) {
      console.error("Failed to get trace:", error);
      return null;
    }
  }
}

/**
 * Standalone feedback submission function.
 *
 * @example
 * ```typescript
 * import { submitFeedback } from "agenttrace/browser";
 *
 * await submitFeedback({
 *   publicKey: "your-public-key",
 *   traceId: "trace-123",
 *   name: "rating",
 *   value: 5
 * });
 * ```
 */
export async function submitFeedback(options: {
  publicKey: string;
  host?: string;
  traceId: string;
  name: string;
  value: number | boolean | string;
  observationId?: string;
  dataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
  comment?: string;
}): Promise<boolean> {
  const client = new AgentTraceBrowser({
    publicKey: options.publicKey,
    host: options.host,
  });

  return client.submitFeedback({
    traceId: options.traceId,
    name: options.name,
    value: options.value,
    observationId: options.observationId,
    dataType: options.dataType,
    comment: options.comment,
  });
}

/**
 * React hook for feedback (if React is available).
 * Returns a function to submit feedback with loading state.
 */
export function createFeedbackHandler(config: {
  publicKey: string;
  host?: string;
}) {
  const client = new AgentTraceBrowser(config);

  return async function submitFeedback(
    traceId: string,
    name: string,
    value: number | boolean | string,
    comment?: string
  ): Promise<boolean> {
    return client.submitFeedback({
      traceId,
      name,
      value,
      comment,
    });
  };
}

// Utility functions

function generateId(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

function inferDataType(value: unknown): "NUMERIC" | "BOOLEAN" | "CATEGORICAL" {
  if (typeof value === "boolean") return "BOOLEAN";
  if (typeof value === "number") return "NUMERIC";
  return "CATEGORICAL";
}

// Version
export const VERSION = "0.1.0";
