/**
 * Tests for the AgentTrace TypeScript SDK client.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// Mock the transport and batch modules before importing client
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

import { AgentTrace, Trace, Span, Generation } from "../src/client";
import { setClient, getClient } from "../src/context";

describe("AgentTrace Client", () => {
  beforeEach(() => {
    // Clear any global client
    setClient(null as any);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("initialization", () => {
    it("should initialize with required config", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
      });

      expect(client.apiKey).toBe("test-api-key");
      expect(client.host).toBe("https://api.agenttrace.io");
      expect(client.enabled).toBe(true);
    });

    it("should accept custom host", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
        host: "https://custom.example.com",
      });

      expect(client.host).toBe("https://custom.example.com");
    });

    it("should strip trailing slash from host", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
        host: "https://example.com/",
      });

      expect(client.host).toBe("https://example.com");
    });

    it("should allow disabling the client", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
        enabled: false,
      });

      expect(client.enabled).toBe(false);
    });

    it("should set itself as the global client", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
      });

      expect(getClient()).toBe(client);
    });
  });

  describe("trace creation", () => {
    it("should create a trace", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
      });

      const trace = client.trace({ name: "test-trace" });

      expect(trace).toBeInstanceOf(Trace);
      expect(trace.name).toBe("test-trace");
      expect(trace.id).toBeDefined();
    });

    it("should create trace with custom id", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
      });

      const trace = client.trace({
        name: "test-trace",
        id: "custom-id",
      });

      expect(trace.id).toBe("custom-id");
    });

    it("should create trace with metadata", () => {
      const client = new AgentTrace({
        apiKey: "test-api-key",
      });

      const trace = client.trace({
        name: "test-trace",
        userId: "user-123",
        sessionId: "session-456",
        metadata: { key: "value" },
        tags: ["tag1", "tag2"],
      });

      expect(trace.userId).toBe("user-123");
      expect(trace.sessionId).toBe("session-456");
      expect(trace.metadata).toEqual({ key: "value" });
      expect(trace.tags).toEqual(["tag1", "tag2"]);
    });
  });
});

describe("Trace", () => {
  let client: AgentTrace;
  let trace: Trace;

  beforeEach(() => {
    client = new AgentTrace({
      apiKey: "test-api-key",
    });
    trace = client.trace({ name: "test-trace" });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("span creation", () => {
    it("should create a span", () => {
      const span = trace.span({ name: "test-span" });

      expect(span).toBeInstanceOf(Span);
      expect(span.name).toBe("test-span");
      expect(span.traceId).toBe(trace.id);
    });

    it("should create span with parent observation", () => {
      const span1 = trace.span({ name: "parent-span" });
      const span2 = trace.span({
        name: "child-span",
        parentObservationId: span1.id,
      });

      expect(span2.parentObservationId).toBe(span1.id);
    });
  });

  describe("generation creation", () => {
    it("should create a generation", () => {
      const generation = trace.generation({
        name: "test-generation",
        model: "gpt-4",
      });

      expect(generation).toBeInstanceOf(Generation);
      expect(generation.name).toBe("test-generation");
      expect(generation.model).toBe("gpt-4");
      expect(generation.traceId).toBe(trace.id);
    });

    it("should create generation with model parameters", () => {
      const generation = trace.generation({
        name: "test-generation",
        model: "gpt-4",
        modelParameters: { temperature: 0.7, maxTokens: 1000 },
      });

      expect(generation.modelParameters).toEqual({
        temperature: 0.7,
        maxTokens: 1000,
      });
    });
  });

  describe("update", () => {
    it("should update trace properties", () => {
      trace.update({
        userId: "new-user",
        output: { result: "success" },
      });

      expect(trace.userId).toBe("new-user");
      expect(trace.output).toEqual({ result: "success" });
    });

    it("should merge metadata", () => {
      trace.update({ metadata: { key1: "value1" } });
      trace.update({ metadata: { key2: "value2" } });

      expect(trace.metadata).toEqual({ key1: "value1", key2: "value2" });
    });
  });

  describe("end", () => {
    it("should end the trace", () => {
      trace.end();

      expect(trace.endTime).toBeDefined();
    });

    it("should end with output", () => {
      trace.end({ output: "final result" });

      expect(trace.output).toBe("final result");
      expect(trace.endTime).toBeDefined();
    });

    it("should not end twice", () => {
      trace.end({ output: "first" });
      const firstEndTime = trace.endTime;

      trace.end({ output: "second" });

      expect(trace.output).toBe("first");
      expect(trace.endTime).toBe(firstEndTime);
    });
  });

  describe("score", () => {
    it("should add score to trace", () => {
      const scoreSpy = vi.spyOn(client, "score");

      trace.score("accuracy", 0.95);

      expect(scoreSpy).toHaveBeenCalledWith({
        traceId: trace.id,
        name: "accuracy",
        value: 0.95,
        dataType: undefined,
        comment: undefined,
      });
    });

    it("should add score with options", () => {
      const scoreSpy = vi.spyOn(client, "score");

      trace.score("success", true, {
        dataType: "BOOLEAN",
        comment: "Task completed",
      });

      expect(scoreSpy).toHaveBeenCalledWith({
        traceId: trace.id,
        name: "success",
        value: true,
        dataType: "BOOLEAN",
        comment: "Task completed",
      });
    });
  });
});

describe("Span", () => {
  let client: AgentTrace;
  let trace: Trace;
  let span: Span;

  beforeEach(() => {
    client = new AgentTrace({ apiKey: "test-api-key" });
    trace = client.trace({ name: "test-trace" });
    span = trace.span({ name: "test-span" });
  });

  it("should have correct initial state", () => {
    expect(span.startTime).toBeDefined();
    expect(span.endTime).toBeUndefined();
    expect(span.level).toBe("DEFAULT");
  });

  it("should end with output", () => {
    span.end({ output: "span result" });

    expect(span.output).toBe("span result");
    expect(span.endTime).toBeDefined();
  });
});

describe("Generation", () => {
  let client: AgentTrace;
  let trace: Trace;
  let generation: Generation;

  beforeEach(() => {
    client = new AgentTrace({ apiKey: "test-api-key" });
    trace = client.trace({ name: "test-trace" });
    generation = trace.generation({
      name: "test-generation",
      model: "gpt-4",
    });
  });

  it("should have correct initial state", () => {
    expect(generation.startTime).toBeDefined();
    expect(generation.endTime).toBeUndefined();
    expect(generation.model).toBe("gpt-4");
  });

  it("should update generation", () => {
    generation.update({
      output: "generated text",
      usage: { inputTokens: 100, outputTokens: 50, totalTokens: 150 },
    });

    expect(generation.output).toBe("generated text");
    expect(generation.usage).toEqual({
      inputTokens: 100,
      outputTokens: 50,
      totalTokens: 150,
    });
  });

  it("should end with all options", () => {
    generation.end({
      output: "final output",
      usage: { inputTokens: 100, outputTokens: 50 },
      model: "gpt-4-turbo",
    });

    expect(generation.output).toBe("final output");
    expect(generation.usage).toEqual({ inputTokens: 100, outputTokens: 50 });
    expect(generation.model).toBe("gpt-4-turbo");
    expect(generation.endTime).toBeDefined();
  });
});
