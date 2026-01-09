/**
 * Git link functionality for AgentTrace.
 *
 * Git links allow you to associate traces and observations with git commits,
 * branches, and repositories for better traceability and debugging.
 */

import { execSync } from "child_process";
import type { AgentTrace } from "./client";

export type GitLinkType = "start" | "commit" | "restore" | "branch" | "diff";

export interface GitLinkOptions {
  type?: GitLinkType;
  observationId?: string;
  commitSha?: string;
  branch?: string;
  repoUrl?: string;
  commitMessage?: string;
  filesChanged?: string[];
  autoDetect?: boolean;
}

export interface GitLinkInfo {
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

/**
 * Client for creating and managing git links.
 *
 * @example
 * ```typescript
 * const client = new AgentTrace({ apiKey: "..." });
 * const trace = client.trace({ name: "my-agent" });
 *
 * // Auto-link current git state
 * trace.gitLink();
 *
 * // Or with explicit info
 * trace.gitLink({
 *   commitSha: "abc123",
 *   branch: "main",
 *   type: "commit"
 * });
 * ```
 */
export class GitClient {
  private _client: AgentTrace;

  constructor(client: AgentTrace) {
    this._client = client;
  }

  /**
   * Create a git link for a trace.
   */
  link(traceId: string, options: GitLinkOptions = {}): GitLinkInfo {
    const linkId = generateId();
    const now = new Date();
    const type = options.type || "commit";
    const autoDetect = options.autoDetect !== false;

    let commitSha = options.commitSha;
    let branch = options.branch;
    let repoUrl = options.repoUrl;
    let commitMessage = options.commitMessage;
    let authorName: string | undefined;
    let authorEmail: string | undefined;
    let filesChanged = options.filesChanged;

    // Auto-detect git info
    if (autoDetect) {
      const gitInfo = this.getGitInfo();
      if (!commitSha) commitSha = gitInfo.commitSha;
      if (!branch) branch = gitInfo.branch;
      if (!repoUrl) repoUrl = gitInfo.repoUrl;
      if (!commitMessage) commitMessage = gitInfo.commitMessage;

      const authorInfo = this.getAuthorInfo();
      authorName = authorInfo.name;
      authorEmail = authorInfo.email;

      if (!filesChanged) {
        filesChanged = this.getChangedFiles();
      }
    }

    filesChanged = filesChanged || [];
    commitSha = commitSha || "";

    // Calculate diff stats
    const diffStats = this.getDiffStats();

    // Send to API
    if (this._client.enabled) {
      this._client._addEvent({
        type: "git-link-create",
        body: {
          id: linkId,
          traceId,
          observationId: options.observationId,
          linkType: type,
          commitSha,
          branch,
          repoUrl,
          commitMessage,
          authorName,
          authorEmail,
          filesChanged,
          additions: diffStats.additions,
          deletions: diffStats.deletions,
          timestamp: now.toISOString(),
        },
      });
    }

    return {
      id: linkId,
      traceId,
      observationId: options.observationId,
      type,
      commitSha,
      branch,
      repoUrl,
      commitMessage,
      authorName,
      authorEmail,
      filesChanged,
      createdAt: now,
    };
  }

  /**
   * Get current git repository info.
   */
  getGitInfo(): {
    commitSha?: string;
    branch?: string;
    repoUrl?: string;
    commitMessage?: string;
  } {
    const result: {
      commitSha?: string;
      branch?: string;
      repoUrl?: string;
      commitMessage?: string;
    } = {};

    try {
      result.commitSha = execSync("git rev-parse HEAD", { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Not in a git repo
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

    try {
      result.commitMessage = execSync('git log -1 --format=%s', { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Ignore
    }

    return result;
  }

  /**
   * Get author info for current commit.
   */
  getAuthorInfo(): { name?: string; email?: string } {
    const result: { name?: string; email?: string } = {};

    try {
      result.name = execSync('git log -1 --format=%an', { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Ignore
    }

    try {
      result.email = execSync('git log -1 --format=%ae', { encoding: "utf-8", timeout: 5000 }).trim();
    } catch {
      // Ignore
    }

    return result;
  }

  /**
   * Get list of changed files (staged and unstaged).
   */
  getChangedFiles(): string[] {
    const files: string[] = [];

    try {
      const staged = execSync("git diff --cached --name-only", { encoding: "utf-8", timeout: 5000 });
      files.push(...staged.trim().split("\n").filter(Boolean));
    } catch {
      // Ignore
    }

    try {
      const unstaged = execSync("git diff --name-only", { encoding: "utf-8", timeout: 5000 });
      for (const f of unstaged.trim().split("\n").filter(Boolean)) {
        if (!files.includes(f)) {
          files.push(f);
        }
      }
    } catch {
      // Ignore
    }

    try {
      const untracked = execSync("git ls-files --others --exclude-standard", { encoding: "utf-8", timeout: 5000 });
      for (const f of untracked.trim().split("\n").filter(Boolean)) {
        if (!files.includes(f)) {
          files.push(f);
        }
      }
    } catch {
      // Ignore
    }

    return files;
  }

  /**
   * Get diff statistics.
   */
  getDiffStats(): { additions: number; deletions: number } {
    const result = { additions: 0, deletions: 0 };

    try {
      const output = execSync("git diff --shortstat", { encoding: "utf-8", timeout: 5000 }).trim();
      // Parse: "2 files changed, 10 insertions(+), 5 deletions(-)"
      const insertMatch = output.match(/(\d+) insertion/);
      const deleteMatch = output.match(/(\d+) deletion/);
      if (insertMatch) result.additions = parseInt(insertMatch[1], 10);
      if (deleteMatch) result.deletions = parseInt(deleteMatch[1], 10);
    } catch {
      // Ignore
    }

    return result;
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
