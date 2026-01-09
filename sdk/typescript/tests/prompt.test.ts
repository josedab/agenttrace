/**
 * Tests for prompt management functionality.
 */

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { PromptVersion, Prompt } from "../src/prompt";

describe("PromptVersion", () => {
  describe("compile", () => {
    it("should compile prompt with double brace variables", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Hello, {{name}}! Welcome to {{place}}.",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const result = prompt.compile({ name: "Alice", place: "AgentTrace" });

      expect(result).toBe("Hello, Alice! Welcome to AgentTrace.");
    });

    it("should compile prompt with single brace variables", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Hello, {name}!",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const result = prompt.compile({ name: "Bob" });

      expect(result).toBe("Hello, Bob!");
    });

    it("should compile prompt with mixed brace styles", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Hello, {{name}}! Your score is {score}.",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const result = prompt.compile({ name: "Charlie", score: 95 });

      expect(result).toBe("Hello, Charlie! Your score is 95.");
    });

    it("should handle multiple occurrences of same variable", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "{{name}} said hello. {{name}} is happy.",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const result = prompt.compile({ name: "Eve" });

      expect(result).toBe("Eve said hello. Eve is happy.");
    });

    it("should leave unmatched variables as-is", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Hello, {{name}}! Your {{missing}} is here.",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const result = prompt.compile({ name: "Frank" });

      expect(result).toBe("Hello, Frank! Your {{missing}} is here.");
    });
  });

  describe("compileChat", () => {
    it("should parse simple chat format", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: `system: You are a helpful assistant.
user: Hello, {{name}}!`,
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const messages = prompt.compileChat({ name: "Grace" });

      expect(messages).toHaveLength(2);
      expect(messages[0]).toEqual({
        role: "system",
        content: "You are a helpful assistant.",
      });
      expect(messages[1]).toEqual({
        role: "user",
        content: "Hello, Grace!",
      });
    });

    it("should handle multi-line content", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: `system: You are a helpful assistant.
Be concise.
user: Tell me about {{topic}}.`,
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const messages = prompt.compileChat({ topic: "TypeScript" });

      expect(messages).toHaveLength(2);
      expect(messages[0].content).toBe("You are a helpful assistant.\nBe concise.");
      expect(messages[1].content).toBe("Tell me about TypeScript.");
    });

    it("should support all role types", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: `system: System message
user: User message
assistant: Assistant message
function: Function result`,
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const messages = prompt.compileChat({});

      expect(messages).toHaveLength(4);
      expect(messages.map((m) => m.role)).toEqual([
        "system",
        "user",
        "assistant",
        "function",
      ]);
    });

    it("should be case-insensitive for roles", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: `System: System message
USER: User message
Assistant: Assistant message`,
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const messages = prompt.compileChat({});

      expect(messages.map((m) => m.role)).toEqual([
        "system",
        "user",
        "assistant",
      ]);
    });
  });

  describe("getVariables", () => {
    it("should extract double brace variables", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Hello, {{name}}! Welcome to {{place}}.",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const variables = prompt.getVariables();

      expect(variables).toContain("name");
      expect(variables).toContain("place");
      expect(variables).toHaveLength(2);
    });

    it("should extract single brace variables", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Score: {score}, Level: {level}",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const variables = prompt.getVariables();

      expect(variables).toContain("score");
      expect(variables).toContain("level");
    });

    it("should deduplicate variables", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "{{name}} said hello. {{name}} is {{name}}.",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const variables = prompt.getVariables();

      expect(variables).toEqual(["name"]);
    });

    it("should handle mixed brace styles", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "{{name}}, {score}, {{name}}, {level}",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const variables = prompt.getVariables();

      expect(variables).toContain("name");
      expect(variables).toContain("score");
      expect(variables).toContain("level");
      expect(variables).toHaveLength(3);
    });

    it("should return empty array for no variables", () => {
      const prompt = new PromptVersion({
        id: "test-id",
        version: 1,
        prompt: "Hello, world!",
        config: {},
        labels: [],
        createdAt: "2024-01-01T00:00:00Z",
      });

      const variables = prompt.getVariables();

      expect(variables).toEqual([]);
    });
  });
});

describe("Prompt", () => {
  beforeEach(() => {
    Prompt.clearCache();
  });

  describe("cache management", () => {
    it("should clear cache", () => {
      Prompt.clearCache();
      // No error means success
    });

    it("should set cache TTL", () => {
      Prompt.setCacheTtl(30000);
      // No error means success
    });

    it("should invalidate cache by name", () => {
      Prompt.invalidate("test-prompt");
      // No error means success
    });
  });
});
