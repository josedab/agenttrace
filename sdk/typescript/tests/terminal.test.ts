/**
 * Tests for the AgentTrace TypeScript SDK terminal module.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { spawn } from "child_process";
import { EventEmitter } from "events";

// Mock child_process
vi.mock("child_process", () => ({
  spawn: vi.fn(),
}));

// Mock process.cwd
vi.spyOn(process, "cwd").mockReturnValue("/mock/working/dir");

import { TerminalClient, TerminalCommandInfo, runCommand } from "../src/terminal";

describe("TerminalClient", () => {
  let mockClient: any;
  let terminalClient: TerminalClient;

  beforeEach(() => {
    mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };
    terminalClient = new TerminalClient(mockClient);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("track", () => {
    it("should track a command with minimal options", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "npm",
      });

      expect(cmd.command).toBe("npm");
      expect(cmd.traceId).toBe("trace-123");
      expect(cmd.id).toBeDefined();
      expect(cmd.args).toEqual([]);
      expect(cmd.workingDirectory).toBe("/mock/working/dir");
    });

    it("should track a command with args", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "npm",
        args: ["install", "--save", "lodash"],
      });

      expect(cmd.command).toBe("npm");
      expect(cmd.args).toEqual(["install", "--save", "lodash"]);
    });

    it("should track command with output", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "echo",
        args: ["hello"],
        stdout: "hello\n",
        stderr: "",
        exitCode: 0,
      });

      expect(cmd.stdout).toBe("hello\n");
      expect(cmd.stderr).toBe("");
      expect(cmd.exitCode).toBe(0);
      expect(cmd.success).toBe(true);
    });

    it("should track failed command", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "false",
        exitCode: 1,
        stderr: "Error occurred",
      });

      expect(cmd.exitCode).toBe(1);
      expect(cmd.success).toBe(false);
      expect(cmd.stderr).toBe("Error occurred");
    });

    it("should track with custom working directory", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "ls",
        workingDirectory: "/custom/path",
      });

      expect(cmd.workingDirectory).toBe("/custom/path");
    });

    it("should track with observation ID", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "test",
        observationId: "obs-456",
      });

      expect(cmd.observationId).toBe("obs-456");
    });

    it("should calculate duration", () => {
      const startedAt = new Date("2024-01-01T10:00:00Z");
      const completedAt = new Date("2024-01-01T10:00:05Z");

      const cmd = terminalClient.track("trace-123", {
        command: "long-command",
        startedAt,
        completedAt,
      });

      expect(cmd.durationMs).toBe(5000);
    });

    it("should send event to API when enabled", () => {
      terminalClient.track("trace-123", {
        command: "npm",
        args: ["test"],
        exitCode: 0,
        stdout: "All tests passed",
      });

      expect(mockClient._addEvent).toHaveBeenCalled();
      const eventArg = mockClient._addEvent.mock.calls[0][0];
      expect(eventArg.type).toBe("terminal-command-create");
      expect(eventArg.body.traceId).toBe("trace-123");
      expect(eventArg.body.command).toBe("npm");
      expect(eventArg.body.args).toEqual(["test"]);
    });

    it("should not send event when client is disabled", () => {
      mockClient.enabled = false;

      const cmd = terminalClient.track("trace-123", {
        command: "npm",
        args: ["test"],
      });

      expect(mockClient._addEvent).not.toHaveBeenCalled();
      expect(cmd.command).toBe("npm"); // Still returns info
    });

    it("should track timed out command", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "slow-command",
        timedOut: true,
        killed: true,
        exitCode: -1,
      });

      expect(mockClient._addEvent).toHaveBeenCalled();
      const eventArg = mockClient._addEvent.mock.calls[0][0];
      expect(eventArg.body.timedOut).toBe(true);
      expect(eventArg.body.killed).toBe(true);
    });

    it("should track truncated output", () => {
      const cmd = terminalClient.track("trace-123", {
        command: "verbose-command",
        stdout: "truncated...",
        stdoutTruncated: true,
        stderr: "errors...",
        stderrTruncated: true,
      });

      expect(mockClient._addEvent).toHaveBeenCalled();
      const eventArg = mockClient._addEvent.mock.calls[0][0];
      expect(eventArg.body.stdoutTruncated).toBe(true);
      expect(eventArg.body.stderrTruncated).toBe(true);
    });
  });

  describe("run", () => {
    it("should run a command and track it", async () => {
      // Create mock process
      const mockProc = new EventEmitter() as any;
      mockProc.stdout = new EventEmitter();
      mockProc.stderr = new EventEmitter();
      vi.mocked(spawn).mockReturnValue(mockProc);

      const runPromise = terminalClient.run("trace-123", "echo", {
        args: ["hello"],
      });

      // Simulate stdout
      mockProc.stdout.emit("data", Buffer.from("hello\n"));

      // Simulate process exit
      mockProc.emit("close", 0);

      const result = await runPromise;

      expect(result.exitCode).toBe(0);
      expect(result.stdout).toBe("hello\n");
      expect(result.info).toBeDefined();
      expect(result.info.command).toBe("echo");
    });

    it("should handle command with non-zero exit code", async () => {
      const mockProc = new EventEmitter() as any;
      mockProc.stdout = new EventEmitter();
      mockProc.stderr = new EventEmitter();
      vi.mocked(spawn).mockReturnValue(mockProc);

      const runPromise = terminalClient.run("trace-123", "false");

      mockProc.stderr.emit("data", Buffer.from("Error\n"));
      mockProc.emit("close", 1);

      const result = await runPromise;

      expect(result.exitCode).toBe(1);
      expect(result.stderr).toBe("Error\n");
      expect(result.info.success).toBe(false);
    });

    it("should handle command error", async () => {
      const mockProc = new EventEmitter() as any;
      mockProc.stdout = new EventEmitter();
      mockProc.stderr = new EventEmitter();
      vi.mocked(spawn).mockReturnValue(mockProc);

      const runPromise = terminalClient.run("trace-123", "nonexistent-command");

      mockProc.emit("error", new Error("Command not found"));

      const result = await runPromise;

      expect(result.exitCode).toBe(-1);
      expect(result.stderr).toBe("Command not found");
      expect(result.info.success).toBe(false);
    });

    it("should respect timeout", async () => {
      vi.useFakeTimers();

      const mockProc = new EventEmitter() as any;
      mockProc.stdout = new EventEmitter();
      mockProc.stderr = new EventEmitter();
      mockProc.kill = vi.fn();
      vi.mocked(spawn).mockReturnValue(mockProc);

      const runPromise = terminalClient.run("trace-123", "sleep", {
        args: ["10"],
        timeout: 1000,
      });

      // Fast-forward past timeout
      vi.advanceTimersByTime(1500);

      expect(mockProc.kill).toHaveBeenCalledWith("SIGTERM");

      // Clean up
      mockProc.emit("close", -1);

      vi.useRealTimers();
    });

    it("should truncate long output", async () => {
      const mockProc = new EventEmitter() as any;
      mockProc.stdout = new EventEmitter();
      mockProc.stderr = new EventEmitter();
      vi.mocked(spawn).mockReturnValue(mockProc);

      const runPromise = terminalClient.run("trace-123", "generate-output", {
        maxOutputBytes: 10,
      });

      // Send more data than maxOutputBytes
      mockProc.stdout.emit("data", Buffer.from("0123456789ABCDEF"));
      mockProc.emit("close", 0);

      const result = await runPromise;

      expect(result.stdout.length).toBeLessThanOrEqual(10);
    });
  });
});

describe("TerminalCommandInfo", () => {
  it("should have all expected fields", () => {
    const mockClient = {
      enabled: false,
      _addEvent: vi.fn(),
    };
    const client = new TerminalClient(mockClient);

    const cmd = client.track("trace-123", {
      command: "npm",
      args: ["test"],
      observationId: "obs-456",
      workingDirectory: "/path",
      exitCode: 0,
      stdout: "output",
      stderr: "errors",
    });

    expect(cmd).toHaveProperty("id");
    expect(cmd).toHaveProperty("traceId");
    expect(cmd).toHaveProperty("observationId");
    expect(cmd).toHaveProperty("command");
    expect(cmd).toHaveProperty("args");
    expect(cmd).toHaveProperty("workingDirectory");
    expect(cmd).toHaveProperty("exitCode");
    expect(cmd).toHaveProperty("stdout");
    expect(cmd).toHaveProperty("stderr");
    expect(cmd).toHaveProperty("success");
    expect(cmd).toHaveProperty("durationMs");
    expect(cmd).toHaveProperty("startedAt");
    expect(cmd).toHaveProperty("completedAt");
  });
});

describe("runCommand", () => {
  it("should throw error when no client available", async () => {
    // Mock getClient to return null
    vi.mock("../src/context", () => ({
      getClient: vi.fn().mockReturnValue(null),
      getCurrentTrace: vi.fn().mockReturnValue(null),
    }));

    await expect(runCommand("echo")).rejects.toThrow("No AgentTrace client available");
  });
});
