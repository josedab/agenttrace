/**
 * Checkpoint functionality for AgentTrace.
 *
 * Checkpoints allow you to create snapshots of code state during agent execution,
 * enabling restoration and debugging of agent sessions.
 */

import { execSync } from "child_process";
import { statSync, readFileSync, existsSync } from "fs";
import { createHash } from "crypto";
import type { AgentTrace } from "./client";

export type CheckpointType = "manual" | "auto" | "tool_call" | "error" | "milestone" | "restore";

export interface CheckpointOptions {
  name: string;
  type?: CheckpointType;
  observationId?: string;
  description?: string;
  files?: string[];
  includeGitInfo?: boolean;
}

export interface CheckpointInfo {
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

export interface GitInfo {
  commitSha?: string;
  branch?: string;
  repoUrl?: string;
}

/**
 * Client for creating and managing checkpoints.
 *
 * @example
 * ```typescript
 * const client = new AgentTrace({ apiKey: "..." });
 * const trace = client.trace({ name: "my-agent" });
 *
 * // Create a checkpoint
 * const cp = trace.checkpoint({
 *   name: "before-edit",
 *   type: "manual",
 *   files: ["src/main.ts", "src/utils.ts"]
 * });
 * ```
 */
export class CheckpointClient {
  private _client: AgentTrace;

  constructor(client: AgentTrace) {
    this._client = client;
  }

  /**
   * Create a new checkpoint.
   */
  create(traceId: string, options: CheckpointOptions): CheckpointInfo {
    const checkpointId = generateId();
    const now = new Date();
    const type = options.type || "manual";

    // Gather git info if requested
    let gitInfo: GitInfo = {};
    if (options.includeGitInfo !== false) {
      gitInfo = this.getGitInfo();
    }

    // Calculate file info
    const filesChanged = options.files || [];
    const totalFiles = filesChanged.length;
    let totalSizeBytes = 0;

    const filesSnapshot: Record<string, { size: number; hash: string }> = {};
    for (const filePath of filesChanged) {
      if (existsSync(filePath)) {
        try {
          const stat = statSync(filePath);
          totalSizeBytes += stat.size;

          const content = readFileSync(filePath);
          const hash = createHash("sha256").update(content).digest("hex");
          filesSnapshot[filePath] = { size: stat.size, hash };
        } catch {
          // Ignore errors
        }
      }
    }

    // Send to API
    if (this._client.enabled) {
      this._client._addEvent({
        type: "checkpoint-create",
        body: {
          id: checkpointId,
          traceId,
          observationId: options.observationId,
          name: options.name,
          description: options.description,
          type,
          gitCommitSha: gitInfo.commitSha,
          gitBranch: gitInfo.branch,
          gitRepoUrl: gitInfo.repoUrl,
          filesSnapshot,
          filesChanged,
          totalFiles,
          totalSizeBytes,
          timestamp: now.toISOString(),
        },
      });
    }

    return {
      id: checkpointId,
      name: options.name,
      type,
      traceId,
      observationId: options.observationId,
      gitCommitSha: gitInfo.commitSha,
      gitBranch: gitInfo.branch,
      filesChanged,
      totalFiles,
      totalSizeBytes,
      createdAt: now,
    };
  }

  /**
   * Get current git repository info.
   */
  getGitInfo(): GitInfo {
    const result: GitInfo = {};

    try {
      result.commitSha = execSync("git rev-parse HEAD", { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Not in a git repo or git not available
    }

    try {
      result.branch = execSync("git rev-parse --abbrev-ref HEAD", { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Ignore
    }

    try {
      result.repoUrl = execSync("git config --get remote.origin.url", { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Ignore
    }

    return result;
  }
}

/**
 * Create a checkpoint scope for automatic tracking.
 */
export async function withCheckpoint<T>(
  client: AgentTrace,
  traceId: string,
  options: CheckpointOptions,
  fn: (checkpoint: CheckpointInfo) => Promise<T>
): Promise<T> {
  const cpClient = new CheckpointClient(client);
  const cp = cpClient.create(traceId, options);
  return fn(cp);
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
