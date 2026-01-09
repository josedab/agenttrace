/**
 * File operations tracking for AgentTrace.
 *
 * Track file reads, writes, edits, and other operations performed during
 * agent execution for full visibility into agent actions.
 */

import { statSync, existsSync } from "fs";
import { createHash } from "crypto";
import type { AgentTrace } from "./client";

export type FileOperationType = "create" | "read" | "update" | "delete" | "rename" | "copy" | "move" | "chmod";

export interface FileOperationOptions {
  operation: FileOperationType;
  filePath: string;
  observationId?: string;
  newPath?: string;
  contentBefore?: string;
  contentAfter?: string;
  linesAdded?: number;
  linesRemoved?: number;
  diffPreview?: string;
  toolName?: string;
  reason?: string;
  startedAt?: Date;
  completedAt?: Date;
  success?: boolean;
  errorMessage?: string;
}

export interface FileOperationInfo {
  id: string;
  traceId: string;
  observationId?: string;
  operation: FileOperationType;
  filePath: string;
  newPath?: string;
  fileSize: number;
  contentHash?: string;
  linesAdded: number;
  linesRemoved: number;
  success: boolean;
  durationMs: number;
  startedAt: Date;
  completedAt: Date;
}

/**
 * Client for tracking file operations.
 *
 * @example
 * ```typescript
 * const client = new AgentTrace({ apiKey: "..." });
 * const trace = client.trace({ name: "my-agent" });
 *
 * // Track a file operation
 * trace.fileOp({
 *   operation: "update",
 *   filePath: "src/main.ts",
 *   linesAdded: 10,
 *   linesRemoved: 5
 * });
 * ```
 */
export class FileOperationClient {
  private _client: AgentTrace;

  constructor(client: AgentTrace) {
    this._client = client;
  }

  /**
   * Track a file operation.
   */
  track(traceId: string, options: FileOperationOptions): FileOperationInfo {
    const opId = generateId();
    const now = new Date();

    const startedAt = options.startedAt || now;
    const completedAt = options.completedAt || now;
    const durationMs = completedAt.getTime() - startedAt.getTime();

    // Calculate file info
    let fileSize = 0;
    let fileMode: string | undefined;
    let mimeType: string | undefined;
    let contentHash: string | undefined;
    let contentBeforeHash: string | undefined;
    let contentAfterHash: string | undefined;

    if (existsSync(options.filePath)) {
      try {
        const stat = statSync(options.filePath);
        fileSize = stat.size;
        fileMode = (stat.mode & 0o777).toString(8).padStart(3, "0");
      } catch {
        // Ignore
      }
    }

    if (options.contentBefore) {
      contentBeforeHash = createHash("sha256").update(options.contentBefore).digest("hex");
    }

    if (options.contentAfter) {
      contentAfterHash = createHash("sha256").update(options.contentAfter).digest("hex");
      contentHash = contentAfterHash;
    }

    // Auto-calculate lines changed
    let linesAdded = options.linesAdded ?? 0;
    let linesRemoved = options.linesRemoved ?? 0;

    if (options.linesAdded === undefined && options.contentBefore && options.contentAfter) {
      const beforeLines = new Set(options.contentBefore.split("\n"));
      const afterLines = options.contentAfter.split("\n");
      linesAdded = afterLines.filter(l => !beforeLines.has(l)).length;

      const afterSet = new Set(afterLines);
      linesRemoved = options.contentBefore.split("\n").filter(l => !afterSet.has(l)).length;
    }

    const success = options.success !== false;

    // Send to API
    if (this._client.enabled) {
      this._client._addEvent({
        type: "file-operation-create",
        body: {
          id: opId,
          traceId,
          observationId: options.observationId,
          operation: options.operation,
          filePath: options.filePath,
          newPath: options.newPath,
          fileSize,
          fileMode,
          contentHash,
          mimeType,
          linesAdded,
          linesRemoved,
          diffPreview: options.diffPreview,
          contentBeforeHash,
          contentAfterHash,
          toolName: options.toolName,
          reason: options.reason,
          startedAt: startedAt.toISOString(),
          completedAt: completedAt.toISOString(),
          durationMs,
          success,
          errorMessage: options.errorMessage,
        },
      });
    }

    return {
      id: opId,
      traceId,
      observationId: options.observationId,
      operation: options.operation,
      filePath: options.filePath,
      newPath: options.newPath,
      fileSize,
      contentHash,
      linesAdded,
      linesRemoved,
      success,
      durationMs,
      startedAt,
      completedAt,
    };
  }
}

/**
 * Run a function and track the file operation.
 */
export async function withFileOp<T>(
  client: AgentTrace,
  traceId: string,
  operation: FileOperationType,
  filePath: string,
  fn: (context: { contentBefore?: string; contentAfter?: string }) => Promise<T>,
  options: Omit<FileOperationOptions, "operation" | "filePath"> = {}
): Promise<T> {
  const startedAt = new Date();
  const context: { contentBefore?: string; contentAfter?: string; success: boolean; errorMessage?: string } = {
    success: true,
  };

  try {
    const result = await fn(context);
    return result;
  } catch (error) {
    context.success = false;
    context.errorMessage = error instanceof Error ? error.message : String(error);
    throw error;
  } finally {
    const completedAt = new Date();
    const opClient = new FileOperationClient(client);
    opClient.track(traceId, {
      operation,
      filePath,
      ...options,
      contentBefore: context.contentBefore,
      contentAfter: context.contentAfter,
      startedAt,
      completedAt,
      success: context.success,
      errorMessage: context.errorMessage,
    });
  }
}

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
