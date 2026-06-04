import { GraphQLClient } from 'graphql-request';

const configuredApiUrl = process.env.NEXT_PUBLIC_API_URL?.trim();
export const API_URL =
  configuredApiUrl || (process.env.NODE_ENV === 'production' ? '' : 'http://localhost:8080');

let sessionAccessToken: string | undefined;
let activeProjectId: string | undefined;

export function setApiAccessToken(token?: string) {
  sessionAccessToken = token;
}

export function getApiAccessToken() {
  return sessionAccessToken;
}

export function setApiProjectId(projectId?: string) {
  activeProjectId = projectId;
}

export function getApiProjectId() {
  return activeProjectId;
}

export function requireApiProjectId() {
  if (!activeProjectId) {
    throw new ApiError(400, 'No active project is available');
  }
  return activeProjectId;
}

/**
 * Fetch wrapper with error handling
 */
export async function fetchWithAuth<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_URL}${endpoint}`;

  const headers = new Headers(options.headers);
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  if (sessionAccessToken && !headers.has('Authorization')) {
    headers.set('Authorization', ['Bearer', sessionAccessToken].join(' '));
  }
  if (activeProjectId && !headers.has('X-Project-ID')) {
    headers.set('X-Project-ID', activeProjectId);
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const errorPayload: unknown = await response.json().catch(() => undefined);
    throw new ApiError(
      response.status,
      normalizeApiErrorMessage(errorPayload, response.statusText)
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

type QueryParam = string | number | boolean | Date | string[] | undefined;

/** A flexible, JSON-friendly bag of query-string parameters. */
export type QueryParams = Record<string, QueryParam>;

export function createSearchParams(params: object = {}): URLSearchParams {
  const searchParams = new URLSearchParams();
  const entries: Array<[string, unknown]> = Object.entries(params);

  for (const [key, value] of entries) {
    if (value === undefined || value === '') continue;

    if (value instanceof Date) {
      searchParams.set(key, value.toISOString());
    } else if (Array.isArray(value)) {
      searchParams.set(key, value.join(','));
    } else if (
      typeof value === 'string' ||
      typeof value === 'number' ||
      typeof value === 'boolean'
    ) {
      searchParams.set(key, String(value satisfies QueryParam));
    }
  }

  return searchParams;
}

/**
 * Custom API error class
 */
export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export function isApiErrorPayload(value: unknown): value is { message: string } {
  if (value === null || typeof value !== 'object') {
    return false;
  }

  const message = Reflect.get(value, 'message');
  return typeof message === 'string' && message.length > 0;
}

export function normalizeApiErrorMessage(payload: unknown, fallback: string): string {
  if (isApiErrorPayload(payload)) {
    return payload.message;
  }

  return fallback || 'Unknown error';
}

export function unsupportedApiFeature<T>(feature: string): Promise<T> {
  return Promise.reject(new ApiError(501, `${feature} is not available in the current API`));
}

export function parseJsonValue(value: unknown): unknown {
  if (typeof value !== 'string' || value.length === 0) {
    return value;
  }

  try {
    return JSON.parse(value);
  } catch {
    return value;
  }
}

export function parseJsonRecord(value: unknown): Record<string, unknown> | null {
  const parsed = parseJsonValue(value);
  if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
    return parsed as Record<string, unknown>;
  }
  return null;
}

export function createApiClient(token?: string) {
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = ['Bearer', token].join(' ');
  }

  return {
    get: <T>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: 'GET', headers }),

    post: <T>(endpoint: string, data?: unknown) =>
      fetchWithAuth<T>(endpoint, {
        method: 'POST',
        headers,
        body: data ? JSON.stringify(data) : undefined,
      }),

    put: <T>(endpoint: string, data?: unknown) =>
      fetchWithAuth<T>(endpoint, {
        method: 'PUT',
        headers,
        body: data ? JSON.stringify(data) : undefined,
      }),

    patch: <T>(endpoint: string, data?: unknown) =>
      fetchWithAuth<T>(endpoint, {
        method: 'PATCH',
        headers,
        body: data ? JSON.stringify(data) : undefined,
      }),

    delete: <T>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: 'DELETE', headers }),
  };
}

/**
 * Create GraphQL client
 */
export function createGraphQLClient(token?: string) {
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = ['Bearer', token].join(' ');
  }

  return new GraphQLClient(`${API_URL}/graphql`, { headers });
}
