/**
 * Prompt management for fetching and compiling prompts.
 */

import { getClient } from "./context";

export interface PromptVersionData {
  id: string;
  version: number;
  prompt: string;
  config: Record<string, unknown>;
  labels: string[];
  createdAt: string;
}

export interface ChatMessage {
  role: "system" | "user" | "assistant" | "function";
  content: string;
}

/**
 * Represents a specific version of a prompt.
 */
export class PromptVersion {
  public readonly id: string;
  public readonly version: number;
  public readonly prompt: string;
  public readonly config: Record<string, unknown>;
  public readonly labels: string[];
  public readonly createdAt: string;

  constructor(data: PromptVersionData) {
    this.id = data.id;
    this.version = data.version;
    this.prompt = data.prompt;
    this.config = data.config;
    this.labels = data.labels;
    this.createdAt = data.createdAt;
  }

  /**
   * Compile the prompt with variables.
   *
   * @example
   * ```typescript
   * const compiled = prompt.compile({ name: "Alice", topic: "Python" });
   * ```
   */
  compile(variables: Record<string, unknown>): string {
    let result = this.prompt;

    for (const [key, value] of Object.entries(variables)) {
      // Support both {{var}} and {var} syntax
      result = result.replace(new RegExp(`\\{\\{${key}\\}\\}`, "g"), String(value));
      result = result.replace(new RegExp(`\\{${key}\\}`, "g"), String(value));
    }

    return result;
  }

  /**
   * Compile as chat messages.
   *
   * Expects prompt to be in the format:
   * ```
   * system: You are a helpful assistant.
   * user: Hello, {{name}}!
   * assistant: Hi {{name}}, how can I help?
   * ```
   */
  compileChat(variables: Record<string, unknown>): ChatMessage[] {
    const compiled = this.compile(variables);
    const messages: ChatMessage[] = [];

    const lines = compiled.trim().split("\n");
    let currentRole: ChatMessage["role"] | null = null;
    let currentContent: string[] = [];

    for (const line of lines) {
      // Check for role prefix
      const roleMatch = line.match(/^(system|user|assistant|function):\s*(.*)$/i);

      if (roleMatch) {
        // Save previous message
        if (currentRole && currentContent.length > 0) {
          messages.push({
            role: currentRole,
            content: currentContent.join("\n").trim(),
          });
        }

        currentRole = roleMatch[1].toLowerCase() as ChatMessage["role"];
        currentContent = roleMatch[2] ? [roleMatch[2]] : [];
      } else {
        currentContent.push(line);
      }
    }

    // Don't forget the last message
    if (currentRole && currentContent.length > 0) {
      messages.push({
        role: currentRole,
        content: currentContent.join("\n").trim(),
      });
    }

    return messages;
  }

  /**
   * Extract variable names from the prompt.
   */
  getVariables(): string[] {
    const doubleBrace = this.prompt.match(/\{\{(\w+)\}\}/g) || [];
    const singleBrace = this.prompt.match(/\{(\w+)\}/g) || [];

    const variables = new Set<string>();

    for (const match of doubleBrace) {
      variables.add(match.replace(/\{\{|\}\}/g, ""));
    }
    for (const match of singleBrace) {
      variables.add(match.replace(/\{|\}/g, ""));
    }

    return Array.from(variables);
  }
}

// Cache for prompts
const promptCache = new Map<string, { prompt: PromptVersion; cachedAt: number }>();
let defaultCacheTtl = 60000; // 1 minute

export interface GetPromptOptions {
  name: string;
  version?: number;
  label?: string;
  fallback?: string;
  cacheTtl?: number;
}

/**
 * Prompt management class.
 */
export class Prompt {
  /**
   * Fetch a prompt from the server.
   *
   * @example
   * ```typescript
   * // Get latest version
   * const prompt = await Prompt.get({ name: "my-prompt" });
   *
   * // Get specific version
   * const prompt = await Prompt.get({ name: "my-prompt", version: 2 });
   *
   * // Get by label
   * const prompt = await Prompt.get({ name: "my-prompt", label: "production" });
   *
   * // With fallback
   * const prompt = await Prompt.get({
   *   name: "my-prompt",
   *   fallback: "Default prompt text"
   * });
   * ```
   */
  static async get(options: GetPromptOptions): Promise<PromptVersion> {
    const cacheKey = getCacheKey(options);
    const cacheTtl = options.cacheTtl ?? defaultCacheTtl;

    // Check cache
    const cached = promptCache.get(cacheKey);
    if (cached && Date.now() - cached.cachedAt < cacheTtl) {
      return cached.prompt;
    }

    const client = getClient();
    if (!client) {
      return getFallback(options);
    }

    try {
      // Build query params
      const params: Record<string, string> = { name: options.name };
      if (options.version !== undefined) {
        params.version = String(options.version);
      }
      if (options.label !== undefined) {
        params.label = options.label;
      }

      // Make API request through transport
      const response = await (client as any)._transport.get<PromptVersionData>(
        "/api/public/prompts",
        params
      );

      if (response && response.id) {
        const promptVersion = new PromptVersion(response);

        // Update cache
        promptCache.set(cacheKey, { prompt: promptVersion, cachedAt: Date.now() });

        return promptVersion;
      }

      return getFallback(options);
    } catch (error) {
      console.warn("Failed to fetch prompt:", error);
      return getFallback(options);
    }
  }

  /**
   * Set the default cache TTL for all prompts.
   */
  static setCacheTtl(ttl: number): void {
    defaultCacheTtl = ttl;
  }

  /**
   * Clear the prompt cache.
   */
  static clearCache(): void {
    promptCache.clear();
  }

  /**
   * Invalidate cache for a specific prompt.
   */
  static invalidate(name: string): void {
    for (const key of promptCache.keys()) {
      if (key.startsWith(name)) {
        promptCache.delete(key);
      }
    }
  }
}

/**
 * Convenience function to fetch a prompt.
 */
export async function getPrompt(options: GetPromptOptions): Promise<PromptVersion> {
  return Prompt.get(options);
}

function getCacheKey(options: GetPromptOptions): string {
  const parts = [options.name];
  if (options.version !== undefined) {
    parts.push(`v${options.version}`);
  }
  if (options.label !== undefined) {
    parts.push(`l:${options.label}`);
  }
  return parts.join(":");
}

function getFallback(options: GetPromptOptions): PromptVersion {
  if (options.fallback !== undefined) {
    return new PromptVersion({
      id: "fallback",
      version: 0,
      prompt: options.fallback,
      config: {},
      labels: ["fallback"],
      createdAt: new Date().toISOString(),
    });
  }
  throw new Error(`Prompt '${options.name}' not found and no fallback provided`);
}
