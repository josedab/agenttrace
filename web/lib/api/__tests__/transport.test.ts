import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  ApiError,
  createSearchParams,
  getApiAccessToken,
  getApiProjectId,
  isApiErrorPayload,
  normalizeApiErrorMessage,
  parseJsonRecord,
  parseJsonValue,
  setApiAccessToken,
  setApiProjectId,
  unsupportedApiFeature,
} from '../transport';
import { api } from '../../api';

const API_URL = 'http://localhost:8080';

describe('createSearchParams', () => {
  it('serializes primitives, dates and arrays while skipping empty values', () => {
    const params = createSearchParams({
      name: 'greeting',
      limit: 25,
      active: true,
      since: new Date('2025-01-01T00:00:00.000Z'),
      tags: ['a', 'b', 'c'],
      cursor: undefined,
      empty: '',
    });

    expect(params.get('name')).toBe('greeting');
    expect(params.get('limit')).toBe('25');
    expect(params.get('active')).toBe('true');
    expect(params.get('since')).toBe('2025-01-01T00:00:00.000Z');
    expect(params.get('tags')).toBe('a,b,c');
    expect(params.has('cursor')).toBe(false);
    expect(params.has('empty')).toBe(false);
  });

  it('returns an empty query string when given no params', () => {
    expect(createSearchParams().toString()).toBe('');
  });
});

describe('parseJsonValue / parseJsonRecord', () => {
  it('parses JSON strings and leaves non-strings untouched', () => {
    expect(parseJsonValue('{"a":1}')).toEqual({ a: 1 });
    expect(parseJsonValue('not json')).toBe('not json');
    expect(parseJsonValue(42)).toBe(42);
  });

  it('returns records only for object payloads', () => {
    expect(parseJsonRecord('{"a":1}')).toEqual({ a: 1 });
    expect(parseJsonRecord('[1,2,3]')).toBeNull();
    expect(parseJsonRecord('"scalar"')).toBeNull();
    expect(parseJsonRecord(null)).toBeNull();
  });
});

describe('unsupportedApiFeature', () => {
  it('rejects with a 501 ApiError describing the feature', async () => {
    await expect(unsupportedApiFeature('Widgets')).rejects.toBeInstanceOf(ApiError);
    await expect(unsupportedApiFeature('Widgets')).rejects.toMatchObject({
      status: 501,
      message: 'Widgets is not available in the current API',
    });
  });
});

describe('ApiError type guard', () => {
  it('narrows unknown errors via instanceof', () => {
    const err: unknown = new ApiError(418, 'teapot');
    expect(err instanceof ApiError).toBe(true);
    if (err instanceof ApiError) {
      expect(err.status).toBe(418);
    }
  });
});

describe('API error payload normalization', () => {
  it('recognizes only non-empty string messages', () => {
    expect(isApiErrorPayload({ message: 'Not found' })).toBe(true);
    expect(isApiErrorPayload({ message: 404 })).toBe(false);
    expect(isApiErrorPayload({ message: '' })).toBe(false);
    expect(isApiErrorPayload(null)).toBe(false);
  });

  it('uses the response fallback for malformed payloads', () => {
    expect(normalizeApiErrorMessage({ message: 'Denied' }, 'Forbidden')).toBe('Denied');
    expect(normalizeApiErrorMessage(null, 'Forbidden')).toBe('Forbidden');
    expect(normalizeApiErrorMessage('failure', '')).toBe('Unknown error');
  });
});

describe('fetchWithAuth (via composed api)', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    mockFetch.mockReset();
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    setApiAccessToken(undefined);
    setApiProjectId(undefined);
    vi.restoreAllMocks();
  });

  function headersFor(callIndex = 0) {
    const options = mockFetch.mock.calls[callIndex]?.[1] as RequestInit | undefined;
    return new Headers(options?.headers);
  }

  it('injects the session Authorization and X-Project-ID headers', async () => {
    setApiAccessToken('session-token');
    setApiProjectId('project-42');
    expect(getApiAccessToken()).toBe('session-token');
    expect(getApiProjectId()).toBe('project-42');

    mockFetch.mockResolvedValue({ ok: true, status: 200, json: () => Promise.resolve({ ok: 1 }) });

    await api.get('/api/thing');

    const headers = headersFor();
    expect(headers.get('Authorization')).toBe('Bearer session-token');
    expect(headers.get('X-Project-ID')).toBe('project-42');
    expect(headers.get('Content-Type')).toBe('application/json');
  });

  it('omits auth headers when no session token or project is set', async () => {
    mockFetch.mockResolvedValue({ ok: true, status: 200, json: () => Promise.resolve({}) });

    await api.get('/api/public/thing');

    const headers = headersFor();
    expect(headers.has('Authorization')).toBe(false);
    expect(headers.has('X-Project-ID')).toBe(false);
  });

  it('returns undefined for 204 No Content responses', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 204,
      json: () => Promise.reject(new Error('no body')),
    });

    await expect(api.delete('/api/thing/1')).resolves.toBeUndefined();
  });

  it('normalizes error payloads into an ApiError with the server message', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 404,
      statusText: 'Not Found',
      json: () => Promise.resolve({ message: 'Trace not found' }),
    });

    await expect(api.get('/api/thing/missing')).rejects.toMatchObject({
      status: 404,
      message: 'Trace not found',
    });
  });

  it('falls back to statusText when the error payload has no message', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 400,
      statusText: 'Bad Request',
      json: () => Promise.resolve({}),
    });

    await expect(api.get('/api/thing')).rejects.toMatchObject({
      status: 400,
      message: 'Bad Request',
    });
  });

  it('falls back safely when the error payload is null', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 502,
      statusText: 'Bad Gateway',
      json: () => Promise.resolve(null),
    });

    await expect(api.get('/api/thing')).rejects.toMatchObject({
      status: 502,
      message: 'Bad Gateway',
    });
  });

  it('falls back to a generic message when the error body cannot be parsed', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      statusText: '',
      json: () => Promise.reject(new Error('Invalid JSON')),
    });

    await expect(api.get('/api/thing')).rejects.toMatchObject({
      status: 500,
      message: 'Unknown error',
    });
  });

  it('routes flexible query params through the shared helper (no unsafe casts)', async () => {
    mockFetch.mockResolvedValue({ ok: true, status: 200, json: () => Promise.resolve({}) });

    await api.diffAnalysis.list({ limit: 10, offset: 20 });

    const url = mockFetch.mock.calls[0]?.[0] as string;
    expect(url).toBe(`${API_URL}/api/public/diff-analysis?limit=10&offset=20`);
  });
});
