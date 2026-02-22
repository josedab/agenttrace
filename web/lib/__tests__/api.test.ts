import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ApiError, createApiClient, createGraphQLClient } from "../api";

const API_URL = "http://localhost:8080";

describe("ApiError", () => {
  it("creates an error with status and message", () => {
    const error = new ApiError(404, "Not found");
    expect(error.status).toBe(404);
    expect(error.message).toBe("Not found");
    expect(error.name).toBe("ApiError");
    expect(error).toBeInstanceOf(Error);
  });
});

describe("createApiClient", () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", mockFetch);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mockSuccessResponse(data: unknown) {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(data),
    });
  }

  function mockErrorResponse(status: number, message: string) {
    mockFetch.mockResolvedValue({
      ok: false,
      status,
      statusText: "Error",
      json: () => Promise.resolve({ message }),
    });
  }

  it("makes GET requests with auth header", async () => {
    const client = createApiClient("test-token");
    mockSuccessResponse({ id: 1 });

    const result = await client.get("/api/traces");

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces`,
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
        }),
      })
    );
    expect(result).toEqual({ id: 1 });
  });

  it("makes POST requests with JSON body", async () => {
    const client = createApiClient("test-token");
    mockSuccessResponse({ created: true });

    const result = await client.post("/api/traces", { name: "test" });

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces`,
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ name: "test" }),
      })
    );
    expect(result).toEqual({ created: true });
  });

  it("makes PUT requests", async () => {
    const client = createApiClient("test-token");
    mockSuccessResponse({ updated: true });

    await client.put("/api/traces/1", { name: "updated" });

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces/1`,
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({ name: "updated" }),
      })
    );
  });

  it("makes PATCH requests", async () => {
    const client = createApiClient("test-token");
    mockSuccessResponse({ patched: true });

    await client.patch("/api/traces/1", { name: "patched" });

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces/1`,
      expect.objectContaining({
        method: "PATCH",
        body: JSON.stringify({ name: "patched" }),
      })
    );
  });

  it("makes DELETE requests", async () => {
    const client = createApiClient("test-token");
    mockSuccessResponse({ deleted: true });

    await client.delete("/api/traces/1");

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces/1`,
      expect.objectContaining({ method: "DELETE" })
    );
  });

  it("works without a token", async () => {
    const client = createApiClient();
    mockSuccessResponse({ data: "public" });

    await client.get("/api/public");

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/public`,
      expect.objectContaining({
        headers: expect.objectContaining({
          "Content-Type": "application/json",
        }),
      })
    );
    // No Authorization header
    const headers = mockFetch.mock.calls[0][1].headers;
    expect(headers).not.toHaveProperty("Authorization");
  });

  it("throws ApiError on non-ok response with message", async () => {
    const client = createApiClient("test-token");
    mockErrorResponse(404, "Trace not found");

    await expect(client.get("/api/traces/unknown")).rejects.toThrow(ApiError);
    await expect(client.get("/api/traces/unknown")).rejects.toMatchObject({
      status: 404,
      message: "Trace not found",
    });
  });

  it("throws ApiError with statusText when JSON parse fails", async () => {
    const client = createApiClient("test-token");
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: () => Promise.reject(new Error("Invalid JSON")),
    });

    await expect(client.get("/api/fail")).rejects.toThrow(ApiError);
  });

  it("sends POST with undefined body when no data provided", async () => {
    const client = createApiClient("test-token");
    mockSuccessResponse({});

    await client.post("/api/traces");

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces`,
      expect.objectContaining({
        method: "POST",
        body: undefined,
      })
    );
  });
});

describe("createGraphQLClient", () => {
  it("creates a GraphQL client with the correct endpoint", () => {
    const client = createGraphQLClient("test-token");
    expect(client).toBeDefined();
  });

  it("creates a GraphQL client without token", () => {
    const client = createGraphQLClient();
    expect(client).toBeDefined();
  });
});
