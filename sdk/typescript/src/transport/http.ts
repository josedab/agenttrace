/**
 * HTTP transport for sending data to AgentTrace.
 */

export interface HttpTransportConfig {
  host: string;
  apiKey: string;
  timeout?: number;
  maxRetries?: number;
  publicKey?: string;
}

export interface BatchEvent {
  type: string;
  body: Record<string, unknown>;
}

/**
 * HTTP transport for communicating with the AgentTrace API.
 */
export class HttpTransport {
  private host: string;
  private apiKey: string;
  private publicKey?: string;
  private timeout: number;
  private maxRetries: number;

  constructor(config: HttpTransportConfig) {
    this.host = config.host.replace(/\/$/, "");
    this.apiKey = config.apiKey;
    this.publicKey = config.publicKey;
    this.timeout = config.timeout ?? 10000;
    this.maxRetries = config.maxRetries ?? 3;
  }

  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.apiKey}`,
      "User-Agent": "agenttrace-typescript/0.1.0",
    };

    if (this.publicKey) {
      headers["X-Langfuse-Public-Key"] = this.publicKey;
    }

    return headers;
  }

  /**
   * Send a batch of events to the AgentTrace API.
   */
  async sendBatch(events: BatchEvent[]): Promise<boolean> {
    if (events.length === 0) return true;

    const url = `${this.host}/api/public/ingestion`;
    const payload = { batch: events };

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(url, {
          method: "POST",
          headers: this.getHeaders(),
          body: JSON.stringify(payload),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (response.status === 200 || response.status === 201 || response.status === 207) {
          // Log partial failures
          if (response.status === 207) {
            try {
              const result = await response.json();
              if (result.errors) {
                console.warn("Partial batch failures:", result.errors);
              }
            } catch {
              // Ignore parse errors
            }
          }
          return true;
        }

        if (response.status === 429) {
          // Rate limited
          const retryAfter = parseInt(response.headers.get("Retry-After") || "5", 10);
          console.warn(`Rate limited, waiting ${retryAfter}s`);
          await this.sleep(retryAfter * 1000);
          continue;
        }

        if (response.status >= 500) {
          // Server error - retry with backoff
          const waitTime = Math.pow(2, attempt) * 500;
          console.warn(
            `Server error ${response.status}, retrying in ${waitTime}ms (attempt ${attempt + 1}/${this.maxRetries})`
          );
          await this.sleep(waitTime);
          continue;
        }

        // Client error - don't retry
        console.error(`Client error ${response.status}:`, await response.text());
        return false;
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          const waitTime = Math.pow(2, attempt) * 500;
          console.warn(
            `Request timeout, retrying in ${waitTime}ms (attempt ${attempt + 1}/${this.maxRetries})`
          );
          await this.sleep(waitTime);
          continue;
        }

        console.error("Request error:", error);
        const waitTime = Math.pow(2, attempt) * 500;
        await this.sleep(waitTime);
      }
    }

    console.error(`Failed to send batch after ${this.maxRetries} attempts`);
    return false;
  }

  /**
   * Make a GET request to the API.
   */
  async get<T>(path: string, params?: Record<string, string>): Promise<T | null> {
    let url = `${this.host}${path}`;
    if (params) {
      const searchParams = new URLSearchParams(params);
      url += `?${searchParams.toString()}`;
    }

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(url, {
          method: "GET",
          headers: this.getHeaders(),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (response.status === 200) {
          return (await response.json()) as T;
        }

        if (response.status === 404) {
          return null;
        }

        if (response.status === 429) {
          const retryAfter = parseInt(response.headers.get("Retry-After") || "5", 10);
          await this.sleep(retryAfter * 1000);
          continue;
        }

        if (response.status >= 500) {
          const waitTime = Math.pow(2, attempt) * 500;
          await this.sleep(waitTime);
          continue;
        }

        console.error(`GET ${path} failed:`, response.status);
        return null;
      } catch (error) {
        const waitTime = Math.pow(2, attempt) * 500;
        await this.sleep(waitTime);
      }
    }

    return null;
  }

  /**
   * Make a POST request to the API.
   */
  async post<T>(path: string, data: Record<string, unknown>): Promise<T | null> {
    const url = `${this.host}${path}`;

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(url, {
          method: "POST",
          headers: this.getHeaders(),
          body: JSON.stringify(data),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (response.status === 200 || response.status === 201) {
          return (await response.json()) as T;
        }

        if (response.status === 429) {
          const retryAfter = parseInt(response.headers.get("Retry-After") || "5", 10);
          await this.sleep(retryAfter * 1000);
          continue;
        }

        if (response.status >= 500) {
          const waitTime = Math.pow(2, attempt) * 500;
          await this.sleep(waitTime);
          continue;
        }

        console.error(`POST ${path} failed:`, response.status);
        return null;
      } catch (error) {
        const waitTime = Math.pow(2, attempt) * 500;
        await this.sleep(waitTime);
      }
    }

    return null;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}
