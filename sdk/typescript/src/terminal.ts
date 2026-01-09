/**
 * Terminal command tracking for AgentTrace.
 *
 * Track terminal/shell commands executed during agent sessions for
 * full visibility and debugging capabilities.
 */

import { spawn, SpawnOptions } from "child_process";
import { cwd } from "process";
import type { AgentTrace } from "./client";

export interface TerminalCommandOptions {
  command: string;
  args?: string[];
  observationId?: string;
  workingDirectory?: string;
  shell?: string;
  envVars?: Record<string, string>;
  exitCode?: number;
  stdout?: string;
  stderr?: string;
  stdoutTruncated?: boolean;
  stderrTruncated?: boolean;
  timedOut?: boolean;
  killed?: boolean;
  maxMemoryBytes?: number;
  cpuTimeMs?: number;
  toolName?: string;
  reason?: string;
  startedAt?: Date;
  completedAt?: Date;
  success?: boolean;
}

export interface TerminalCommandInfo {
  id: string;
  traceId: string;
  observationId?: string;
  command: string;
  args: string[];
  workingDirectory: string;
  exitCode: number;
  stdout: string;
  stderr: string;
  success: boolean;
  durationMs: number;
  startedAt: Date;
  completedAt: Date;
}

export interface RunCommandOptions {
  args?: string[];
  observationId?: string;
  workingDirectory?: string;
  env?: Record<string, string>;
  timeout?: number;
  shell?: boolean | string;
  toolName?: string;
  reason?: string;
  maxOutputBytes?: number;
}

export interface RunCommandResult {
  info: TerminalCommandInfo;
  exitCode: number;
  stdout: string;
  stderr: string;
}

/**
 * Client for tracking terminal commands.
 *
 * @example
 * ```typescript
 * const client = new AgentTrace({ apiKey: "..." });
 * const trace = client.trace({ name: "my-agent" });
 *
 * // Track a command manually
 * trace.terminalCmd({
 *   command: "npm",
 *   args: ["test"],
 *   exitCode: 0,
 *   stdout: "All tests passed"
 * });
 *
 * // Or run and track automatically
 * const result = await trace.runCmd("npm", ["test"]);
 * ```
 */
export class TerminalClient {
  private _client: AgentTrace;

  constructor(client: AgentTrace) {
    this._client = client;
  }

  /**
   * Track a terminal command.
   */
  track(traceId: string, options: TerminalCommandOptions): TerminalCommandInfo {
    const cmdId = generateId();
    const now = new Date();

    const startedAt = options.startedAt || now;
    const completedAt = options.completedAt || now;
    const durationMs = completedAt.getTime() - startedAt.getTime();

    const workingDirectory = options.workingDirectory || cwd();
    const args = options.args || [];
    const stdout = options.stdout || "";
    const stderr = options.stderr || "";
    const exitCode = options.exitCode ?? 0;
    const success = options.success ?? (exitCode === 0);

    // Convert env vars to JSON string
    const envVarsStr = options.envVars ? JSON.stringify(options.envVars) : undefined;

    // Send to API
    if (this._client.enabled) {
      this._client._addEvent({
        type: "terminal-command-create",
        body: {
          id: cmdId,
          traceId,
          observationId: options.observationId,
          command: options.command,
          args,
          workingDirectory,
          shell: options.shell,
          envVars: envVarsStr,
          startedAt: startedAt.toISOString(),
          completedAt: completedAt.toISOString(),
          durationMs,
          exitCode,
          stdout,
          stderr,
          stdoutTruncated: options.stdoutTruncated || false,
          stderrTruncated: options.stderrTruncated || false,
          success,
          timedOut: options.timedOut || false,
          killed: options.killed || false,
          maxMemoryBytes: options.maxMemoryBytes,
          cpuTimeMs: options.cpuTimeMs,
          toolName: options.toolName,
          reason: options.reason,
        },
      });
    }

    return {
      id: cmdId,
      traceId,
      observationId: options.observationId,
      command: options.command,
      args,
      workingDirectory,
      exitCode,
      stdout,
      stderr,
      success,
      durationMs,
      startedAt,
      completedAt,
    };
  }

  /**
   * Run a command and track it.
   */
  async run(traceId: string, command: string, options: RunCommandOptions = {}): Promise<RunCommandResult> {
    const startedAt = new Date();
    const workingDirectory = options.workingDirectory || cwd();
    const maxOutputBytes = options.maxOutputBytes || 100000;

    return new Promise((resolve) => {
      const args = options.args || [];
      let stdout = "";
      let stderr = "";
      let timedOut = false;
      let killed = false;

      const spawnOptions: SpawnOptions = {
        cwd: workingDirectory,
        env: options.env ? { ...process.env, ...options.env } : process.env,
        shell: options.shell,
      };

      const proc = spawn(command, args, spawnOptions);

      let timeoutId: NodeJS.Timeout | undefined;
      if (options.timeout) {
        timeoutId = setTimeout(() => {
          timedOut = true;
          killed = true;
          proc.kill("SIGTERM");
        }, options.timeout);
      }

      proc.stdout?.on("data", (data: Buffer) => {
        if (stdout.length < maxOutputBytes) {
          stdout += data.toString().slice(0, maxOutputBytes - stdout.length);
        }
      });

      proc.stderr?.on("data", (data: Buffer) => {
        if (stderr.length < maxOutputBytes) {
          stderr += data.toString().slice(0, maxOutputBytes - stderr.length);
        }
      });

      proc.on("close", (code) => {
        if (timeoutId) clearTimeout(timeoutId);

        const completedAt = new Date();
        const exitCode = code ?? -1;

        const stdoutTruncated = stdout.length >= maxOutputBytes;
        const stderrTruncated = stderr.length >= maxOutputBytes;

        const info = this.track(traceId, {
          command,
          args,
          observationId: options.observationId,
          workingDirectory,
          exitCode,
          stdout,
          stderr,
          stdoutTruncated,
          stderrTruncated,
          timedOut,
          killed,
          toolName: options.toolName,
          reason: options.reason,
          startedAt,
          completedAt,
        });

        resolve({
          info,
          exitCode,
          stdout,
          stderr,
        });
      });

      proc.on("error", (error) => {
        if (timeoutId) clearTimeout(timeoutId);

        const completedAt = new Date();

        const info = this.track(traceId, {
          command,
          args,
          observationId: options.observationId,
          workingDirectory,
          exitCode: -1,
          stdout,
          stderr: error.message,
          toolName: options.toolName,
          reason: options.reason,
          startedAt,
          completedAt,
          success: false,
        });

        resolve({
          info,
          exitCode: -1,
          stdout,
          stderr: error.message,
        });
      });
    });
  }
}

/**
 * Run a command with the global client.
 */
export async function runCommand(
  command: string,
  options: RunCommandOptions & { client?: AgentTrace; traceId?: string } = {}
): Promise<RunCommandResult> {
  const { getClient, getCurrentTrace } = await import("./context");

  const client = options.client || getClient();
  if (!client) {
    throw new Error("No AgentTrace client available. Initialize one first.");
  }

  let traceId = options.traceId;
  if (!traceId) {
    const trace = getCurrentTrace();
    if (!trace) {
      throw new Error("No active trace. Create a trace first.");
    }
    traceId = trace.id;
  }

  const termClient = new TerminalClient(client);
  return termClient.run(traceId, command, options);
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
