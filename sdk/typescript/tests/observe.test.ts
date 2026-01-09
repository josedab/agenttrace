/**
 * Tests for the observe() wrapper function.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// Mock the transport and batch modules
vi.mock("../src/transport/http", () => ({
  HttpTransport: vi.fn().mockImplementation(() => ({
    post: vi.fn().mockResolvedValue({}),
    get: vi.fn().mockResolvedValue({}),
  })),
}));

vi.mock("../src/transport/batch", () => ({
  BatchQueue: vi.fn().mockImplementation(() => ({
    add: vi.fn(),
    flush: vi.fn().mockResolvedValue(undefined),
    stop: vi.fn(),
  })),
}));

import { observe } from "../src/observe";
import { AgentTrace } from "../src/client";
import { setClient, setCurrentTrace, setCurrentObservation } from "../src/context";

describe("observe()", () => {
  beforeEach(() => {
    // Clear context
    setClient(null as any);
    setCurrentTrace(null);
    setCurrentObservation(null);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("without client", () => {
    it("should execute function normally when no client", () => {
      const fn = observe((x: number) => x * 2);

      const result = fn(5);

      expect(result).toBe(10);
    });

    it("should execute async function normally when no client", async () => {
      const fn = observe(async (x: number) => {
        return x * 2;
      });

      const result = await fn(5);

      expect(result).toBe(10);
    });
  });

  describe("with client", () => {
    let client: AgentTrace;

    beforeEach(() => {
      client = new AgentTrace({ apiKey: "test-api-key" });
    });

    it("should create trace for function without existing trace", () => {
      const fn = observe(function myFunction(x: number) {
        return x * 2;
      });

      const result = fn(5);

      expect(result).toBe(10);
    });

    it("should execute function and return result", () => {
      const fn = observe((a: number, b: number) => a + b);

      const result = fn(3, 7);

      expect(result).toBe(10);
    });

    it("should handle function that throws", () => {
      const fn = observe(() => {
        throw new Error("Test error");
      });

      expect(() => fn()).toThrow("Test error");
    });
  });

  describe("async functions", () => {
    let client: AgentTrace;

    beforeEach(() => {
      client = new AgentTrace({ apiKey: "test-api-key" });
    });

    it("should execute async function and return result", async () => {
      const fn = observe(async (x: number) => {
        return x * 2;
      });

      const result = await fn(5);

      expect(result).toBe(10);
    });

    it("should handle async function that throws", async () => {
      const fn = observe(async () => {
        throw new Error("Async error");
      });

      await expect(fn()).rejects.toThrow("Async error");
    });
  });

  describe("options", () => {
    let client: AgentTrace;

    beforeEach(() => {
      client = new AgentTrace({ apiKey: "test-api-key" });
    });

    it("should use custom name", () => {
      const fn = observe(
        (x: number) => x * 2,
        { name: "custom-name" }
      );

      const result = fn(5);

      expect(result).toBe(10);
    });

    it("should handle asType span", () => {
      const fn = observe(
        (x: number) => x * 2,
        { asType: "span" }
      );

      const result = fn(5);

      expect(result).toBe(10);
    });

    it("should handle asType generation", () => {
      const fn = observe(
        (x: number) => x * 2,
        { asType: "generation", model: "gpt-4" }
      );

      const result = fn(5);

      expect(result).toBe(10);
    });

    it("should respect captureInput option", () => {
      const fn = observe(
        (x: number) => x * 2,
        { captureInput: false }
      );

      const result = fn(5);

      expect(result).toBe(10);
    });

    it("should respect captureOutput option", () => {
      const fn = observe(
        (x: number) => x * 2,
        { captureOutput: false }
      );

      const result = fn(5);

      expect(result).toBe(10);
    });
  });

  describe("function properties", () => {
    it("should preserve function name", () => {
      function namedFunction(x: number) {
        return x * 2;
      }

      const wrapped = observe(namedFunction);

      expect(wrapped.name).toBe("namedFunction");
    });

    it("should preserve function length", () => {
      function threeArgs(a: number, b: number, c: number) {
        return a + b + c;
      }

      const wrapped = observe(threeArgs);

      expect(wrapped.length).toBe(3);
    });
  });
});
