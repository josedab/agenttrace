/**
 * Tests for the AgentTrace TypeScript SDK file operations module.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { existsSync, statSync } from "fs";

// Mock fs and crypto
vi.mock("fs", () => ({
  existsSync: vi.fn(),
  statSync: vi.fn(),
}));

vi.mock("crypto", () => ({
  createHash: vi.fn(() => ({
    update: vi.fn().mockReturnThis(),
    digest: vi.fn().mockReturnValue("abc123hash"),
  })),
}));

import { FileOperationClient, FileOperationInfo, withFileOp } from "../src/fileops";
import type { FileOperationType } from "../src/fileops";

describe("FileOperationClient", () => {
  let mockClient: any;
  let fileOpClient: FileOperationClient;

  beforeEach(() => {
    mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };
    fileOpClient = new FileOperationClient(mockClient);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("track", () => {
    it("should track a read operation", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({
        size: 1024,
        mode: 0o644,
      } as any);

      const op = fileOpClient.track("trace-123", {
        operation: "read",
        filePath: "/path/to/file.ts",
      });

      expect(op.operation).toBe("read");
      expect(op.filePath).toBe("/path/to/file.ts");
      expect(op.traceId).toBe("trace-123");
      expect(op.id).toBeDefined();
      expect(op.success).toBe(true);
    });

    it("should track a write operation with line changes", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({
        size: 2048,
        mode: 0o644,
      } as any);

      const op = fileOpClient.track("trace-123", {
        operation: "update",
        filePath: "/path/to/file.ts",
        linesAdded: 10,
        linesRemoved: 5,
      });

      expect(op.operation).toBe("update");
      expect(op.linesAdded).toBe(10);
      expect(op.linesRemoved).toBe(5);
    });

    it("should track a create operation", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({
        size: 512,
        mode: 0o644,
      } as any);

      const op = fileOpClient.track("trace-123", {
        operation: "create",
        filePath: "/path/to/newfile.ts",
        contentAfter: "const x = 1;",
      });

      expect(op.operation).toBe("create");
      expect(op.filePath).toBe("/path/to/newfile.ts");
    });

    it("should track a delete operation", () => {
      vi.mocked(existsSync).mockReturnValue(false);

      const op = fileOpClient.track("trace-123", {
        operation: "delete",
        filePath: "/path/to/deleted.ts",
      });

      expect(op.operation).toBe("delete");
      expect(op.fileSize).toBe(0);
    });

    it("should track a rename operation with new path", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({
        size: 1024,
        mode: 0o644,
      } as any);

      const op = fileOpClient.track("trace-123", {
        operation: "rename",
        filePath: "/old/path/file.ts",
        newPath: "/new/path/file.ts",
      });

      expect(op.operation).toBe("rename");
      expect(op.filePath).toBe("/old/path/file.ts");
      expect(op.newPath).toBe("/new/path/file.ts");
    });

    it("should track with observation ID", () => {
      const op = fileOpClient.track("trace-123", {
        operation: "read",
        filePath: "/path/file.ts",
        observationId: "obs-456",
      });

      expect(op.observationId).toBe("obs-456");
    });

    it("should auto-calculate lines changed from content", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({ size: 100, mode: 0o644 } as any);

      const op = fileOpClient.track("trace-123", {
        operation: "update",
        filePath: "/path/file.ts",
        contentBefore: "line1\nline2\nline3",
        contentAfter: "line1\nline2\nline4\nline5",
      });

      // line4 and line5 are new, line3 was removed
      expect(op.linesAdded).toBe(2);
      expect(op.linesRemoved).toBe(1);
    });

    it("should calculate duration", () => {
      const startedAt = new Date("2024-01-01T10:00:00Z");
      const completedAt = new Date("2024-01-01T10:00:01Z");

      const op = fileOpClient.track("trace-123", {
        operation: "read",
        filePath: "/path/file.ts",
        startedAt,
        completedAt,
      });

      expect(op.durationMs).toBe(1000);
    });

    it("should track failed operation", () => {
      const op = fileOpClient.track("trace-123", {
        operation: "update",
        filePath: "/path/file.ts",
        success: false,
        errorMessage: "Permission denied",
      });

      expect(op.success).toBe(false);
    });

    it("should send event to API when enabled", () => {
      vi.mocked(existsSync).mockReturnValue(true);
      vi.mocked(statSync).mockReturnValue({ size: 100, mode: 0o644 } as any);

      fileOpClient.track("trace-123", {
        operation: "update",
        filePath: "/path/file.ts",
      });

      expect(mockClient._addEvent).toHaveBeenCalled();
      const eventArg = mockClient._addEvent.mock.calls[0][0];
      expect(eventArg.type).toBe("file-operation-create");
      expect(eventArg.body.traceId).toBe("trace-123");
      expect(eventArg.body.operation).toBe("update");
    });

    it("should not send event when client is disabled", () => {
      mockClient.enabled = false;

      const op = fileOpClient.track("trace-123", {
        operation: "read",
        filePath: "/path/file.ts",
      });

      expect(mockClient._addEvent).not.toHaveBeenCalled();
      expect(op.operation).toBe("read"); // Still returns info
    });
  });
});

describe("FileOperationType", () => {
  it("should support all operation types", () => {
    const types: FileOperationType[] = [
      "create",
      "read",
      "update",
      "delete",
      "rename",
      "copy",
      "move",
      "chmod",
    ];

    const mockClient = {
      enabled: false,
      _addEvent: vi.fn(),
    };
    const client = new FileOperationClient(mockClient);

    for (const type of types) {
      const op = client.track("trace-123", {
        operation: type,
        filePath: `/path/file-${type}.ts`,
      });
      expect(op.operation).toBe(type);
    }
  });
});

describe("withFileOp", () => {
  it("should track file operation and execute function", async () => {
    const mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };

    const result = await withFileOp(
      mockClient as any,
      "trace-123",
      "read",
      "/path/file.ts",
      async () => {
        return "file content";
      }
    );

    expect(result).toBe("file content");
    expect(mockClient._addEvent).toHaveBeenCalled();
  });

  it("should track failed operation on error", async () => {
    const mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };

    await expect(
      withFileOp(mockClient as any, "trace-123", "read", "/path/file.ts", async () => {
        throw new Error("Read failed");
      })
    ).rejects.toThrow("Read failed");

    expect(mockClient._addEvent).toHaveBeenCalled();
    const eventArg = mockClient._addEvent.mock.calls[0][0];
    expect(eventArg.body.success).toBe(false);
    expect(eventArg.body.errorMessage).toBe("Read failed");
  });

  it("should track with custom options", async () => {
    const mockClient = {
      enabled: true,
      _addEvent: vi.fn(),
    };

    await withFileOp(
      mockClient as any,
      "trace-123",
      "update",
      "/path/file.ts",
      async (ctx) => {
        ctx.contentBefore = "old content";
        ctx.contentAfter = "new content";
        return "done";
      },
      { toolName: "editor", reason: "bug fix" }
    );

    expect(mockClient._addEvent).toHaveBeenCalled();
    const eventArg = mockClient._addEvent.mock.calls[0][0];
    expect(eventArg.body.toolName).toBe("editor");
    expect(eventArg.body.reason).toBe("bug fix");
  });
});

describe("FileOperationInfo", () => {
  it("should have all expected fields", () => {
    const mockClient = {
      enabled: false,
      _addEvent: vi.fn(),
    };
    const client = new FileOperationClient(mockClient);

    vi.mocked(existsSync).mockReturnValue(true);
    vi.mocked(statSync).mockReturnValue({ size: 500, mode: 0o644 } as any);

    const op = client.track("trace-123", {
      operation: "update",
      filePath: "/path/file.ts",
      observationId: "obs-456",
      newPath: "/new/path.ts",
      linesAdded: 5,
      linesRemoved: 2,
    });

    expect(op).toHaveProperty("id");
    expect(op).toHaveProperty("traceId");
    expect(op).toHaveProperty("observationId");
    expect(op).toHaveProperty("operation");
    expect(op).toHaveProperty("filePath");
    expect(op).toHaveProperty("newPath");
    expect(op).toHaveProperty("fileSize");
    expect(op).toHaveProperty("linesAdded");
    expect(op).toHaveProperty("linesRemoved");
    expect(op).toHaveProperty("success");
    expect(op).toHaveProperty("durationMs");
    expect(op).toHaveProperty("startedAt");
    expect(op).toHaveProperty("completedAt");
  });
});
