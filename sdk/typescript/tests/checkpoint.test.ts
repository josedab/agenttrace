/**
 * Tests for the AgentTrace TypeScript SDK checkpoint module.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { existsSync, statSync, readFileSync } from "fs";
import { execSync } from "child_process";

// Mock fs and child_process
vi.mock("fs", () => ({
  existsSync: vi.fn(),
  statSync: vi.fn(),
  readFileSync: vi.fn(),
}));

vi.mock("child_process", () => ({
  execSync: vi.fn(),
}));

vi.mock("crypto", () => ({
  createHash: vi.fn(() => ({
    update: vi.fn().mockReturnThis(),
    digest: vi.fn().mockReturnValue("abc123hash"),
  })),
}));

import { CheckpointClient, CheckpointInfo, CheckpointOptions, withCheckpoint } from "../src/checkpoint";

describe("CheckpointClient", () => {
  let mockClient: any;
  let checkpointClient: CheckpointClient;

  beforeEach(() => {
    mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };
    checkpointClient = new CheckpointClient(mockClient);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("create", () => {
    it("should create a checkpoint with minimal options", () => {
      const cp = checkpointClient.create("trace-123", {
        name: "test-checkpoint",
      });

      expect(cp.name).toBe("test-checkpoint");
      expect(cp.type).toBe("manual");
      expect(cp.traceId).toBe("trace-123");
      expect(cp.id).toBeDefined();
      expect(cp.createdAt).toBeInstanceOf(Date);
    });

    it("should create a checkpoint with custom type", () => {
      const cp = checkpointClient.create("trace-123", {
        name: "milestone-cp",
        type: "milestone",
      });

      expect(cp.type).toBe("milestone");
    });

    it("should create a checkpoint with observation ID", () => {
      const cp = checkpointClient.create("trace-123", {
        name: "obs-checkpoint",
        observationId: "obs-456",
      });

      expect(cp.observationId).toBe("obs-456");
    });

    it("should track files when provided", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({
        size: 1024,
      } as any);
      vi.mocked(readFileSync).mockReturnValue(Buffer.from("test content"));

      const cp = checkpointClient.create("trace-123", {
        name: "file-checkpoint",
        files: ["/path/to/file1.ts", "/path/to/file2.ts"],
      });

      expect(cp.totalFiles).toBe(2);
      expect(cp.filesChanged).toEqual(["/path/to/file1.ts", "/path/to/file2.ts"]);
    });

    it("should calculate total size from files", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({
        size: 500,
      } as any);
      vi.mocked(readFileSync).mockReturnValue(Buffer.from("test"));

      const cp = checkpointClient.create("trace-123", {
        name: "size-checkpoint",
        files: ["/path/file1.ts", "/path/file2.ts"],
      });

      expect(cp.totalSizeBytes).toBe(1000); // 500 + 500
    });

    it("should handle nonexistent files gracefully", () => {
      vi.mocked(existsSync).mockReturnValue(false);

      const cp = checkpointClient.create("trace-123", {
        name: "no-files-checkpoint",
        files: ["/nonexistent/file.ts"],
      });

      expect(cp.totalFiles).toBe(1);
      expect(cp.totalSizeBytes).toBe(0);
    });

    it("should send event to API when client is enabled", () => {
      const cp = checkpointClient.create("trace-123", {
        name: "api-checkpoint",
        type: "auto",
      });

      expect(mockClient._addEvent).toHaveBeenCalled();
      const eventArg = mockClient._addEvent.mock.calls[0][0];
      expect(eventArg.type).toBe("checkpoint-create");
      expect(eventArg.body.traceId).toBe("trace-123");
      expect(eventArg.body.name).toBe("api-checkpoint");
      expect(eventArg.body.type).toBe("auto");
    });

    it("should not send event when client is disabled", () => {
      mockClient.enabled = false;

      const cp = checkpointClient.create("trace-123", {
        name: "disabled-checkpoint",
      });

      expect(mockClient._addEvent).not.toHaveBeenCalled();
      expect(cp.name).toBe("disabled-checkpoint"); // Still returns info
    });
  });

  describe("getGitInfo", () => {
    it("should get git commit sha", () => {
      vi.mocked(execSync).mockImplementation((cmd: string) => {
        if (cmd.includes("rev-parse HEAD")) {
          return "abc123def456\n";
        }
        return "";
      });

      const gitInfo = checkpointClient.getGitInfo();

      expect(gitInfo.commitSha).toBe("abc123def456");
    });

    it("should get git branch", () => {
      vi.mocked(execSync).mockImplementation((cmd: string) => {
        if (cmd.includes("--abbrev-ref HEAD")) {
          return "feature-branch\n";
        }
        return "";
      });

      const gitInfo = checkpointClient.getGitInfo();

      expect(gitInfo.branch).toBe("feature-branch");
    });

    it("should get repo URL", () => {
      vi.mocked(execSync).mockImplementation((cmd: string) => {
        if (cmd.includes("remote.origin.url")) {
          return "https://github.com/test/repo.git\n";
        }
        return "";
      });

      const gitInfo = checkpointClient.getGitInfo();

      expect(gitInfo.repoUrl).toBe("https://github.com/test/repo.git");
    });

    it("should handle git command failures gracefully", () => {
      vi.mocked(execSync).mockImplementation(() => {
        throw new Error("git not found");
      });

      const gitInfo = checkpointClient.getGitInfo();

      expect(gitInfo.commitSha).toBeUndefined();
      expect(gitInfo.branch).toBeUndefined();
      expect(gitInfo.repoUrl).toBeUndefined();
    });
  });

  describe("checkpoint with git info", () => {
    it("should include git info by default", () => {
      vi.mocked(execSync).mockImplementation((cmd: string) => {
        if (cmd.includes("rev-parse HEAD")) return "commit123\n";
        if (cmd.includes("--abbrev-ref")) return "main\n";
        return "";
      });

      const cp = checkpointClient.create("trace-123", {
        name: "git-checkpoint",
        includeGitInfo: true,
      });

      expect(cp.gitCommitSha).toBe("commit123");
      expect(cp.gitBranch).toBe("main");
    });

    it("should exclude git info when requested", () => {
      vi.mocked(execSync).mockReturnValue("should-not-appear\n");

      const cp = checkpointClient.create("trace-123", {
        name: "no-git-checkpoint",
        includeGitInfo: false,
      });

      expect(cp.gitCommitSha).toBeUndefined();
      expect(cp.gitBranch).toBeUndefined();
    });
  });
});

describe("CheckpointType", () => {
  it("should support all checkpoint types", () => {
    const types: Array<"manual" | "auto" | "tool_call" | "error" | "milestone" | "restore"> = [
      "manual",
      "auto",
      "tool_call",
      "error",
      "milestone",
      "restore",
    ];

    const mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };
    const client = new CheckpointClient(mockClient);

    for (const type of types) {
      const cp = client.create("trace-123", {
        name: `${type}-checkpoint`,
        type,
      });
      expect(cp.type).toBe(type);
    }
  });
});

describe("withCheckpoint", () => {
  it("should create checkpoint and execute function", async () => {
    const mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };

    const result = await withCheckpoint(
      mockClient as any,
      "trace-123",
      { name: "scope-checkpoint" },
      async (cp) => {
        expect(cp.name).toBe("scope-checkpoint");
        return "result-value";
      }
    );

    expect(result).toBe("result-value");
  });

  it("should propagate errors from function", async () => {
    const mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };

    await expect(
      withCheckpoint(
        mockClient as any,
        "trace-123",
        { name: "error-checkpoint" },
        async () => {
          throw new Error("Test error");
        }
      )
    ).rejects.toThrow("Test error");
  });
});

describe("CheckpointInfo", () => {
  it("should have all expected fields", () => {
    const mockClient = {
      enabled: false,
      _addEvent: vi.fn(),
    };
    const client = new CheckpointClient(mockClient);

    const cp = client.create("trace-123", {
      name: "full-checkpoint",
      type: "milestone",
      observationId: "obs-456",
      description: "Test description",
      files: [],
    });

    expect(cp).toHaveProperty("id");
    expect(cp).toHaveProperty("name");
    expect(cp).toHaveProperty("type");
    expect(cp).toHaveProperty("traceId");
    expect(cp).toHaveProperty("observationId");
    expect(cp).toHaveProperty("filesChanged");
    expect(cp).toHaveProperty("totalFiles");
    expect(cp).toHaveProperty("totalSizeBytes");
    expect(cp).toHaveProperty("createdAt");
  });
});
