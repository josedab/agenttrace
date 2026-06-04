import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ApiError, api, createApiClient, createGraphQLClient, setApiProjectId } from '../api';

const API_URL = 'http://localhost:8080';

describe('ApiError', () => {
  it('creates an error with status and message', () => {
    const error = new ApiError(404, 'Not found');
    expect(error.status).toBe(404);
    expect(error.message).toBe('Not found');
    expect(error.name).toBe('ApiError');
    expect(error).toBeInstanceOf(Error);
  });
});

describe('createApiClient', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    mockFetch.mockReset();
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    setApiProjectId(undefined);
    vi.restoreAllMocks();
    vi.useRealTimers();
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
      statusText: 'Error',
      json: () => Promise.resolve({ message }),
    });
  }

  function getRequestHeaders(callIndex = 0) {
    const options = mockFetch.mock.calls[callIndex]?.[1] as RequestInit | undefined;
    return new Headers(options?.headers);
  }

  it('makes GET requests with auth header', async () => {
    const client = createApiClient('test-token');
    mockSuccessResponse({ id: 1 });

    const result = await client.get('/api/traces');

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces`,
      expect.objectContaining({
        method: 'GET',
      })
    );
    const headers = getRequestHeaders();
    expect(headers.get('Authorization')).toBe(['Bearer', 'test-token'].join(' '));
    expect(headers.get('Content-Type')).toBe('application/json');
    expect(result).toEqual({ id: 1 });
  });

  it('makes POST requests with JSON body', async () => {
    const client = createApiClient('test-token');
    mockSuccessResponse({ created: true });

    const result = await client.post('/api/traces', { name: 'test' });

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces`,
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'test' }),
      })
    );
    expect(result).toEqual({ created: true });
  });

  it('makes PUT requests', async () => {
    const client = createApiClient('test-token');
    mockSuccessResponse({ updated: true });

    await client.put('/api/traces/1', { name: 'updated' });

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces/1`,
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'updated' }),
      })
    );
  });

  it('makes PATCH requests', async () => {
    const client = createApiClient('test-token');
    mockSuccessResponse({ patched: true });

    await client.patch('/api/traces/1', { name: 'patched' });

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces/1`,
      expect.objectContaining({
        method: 'PATCH',
        body: JSON.stringify({ name: 'patched' }),
      })
    );
  });

  it('makes DELETE requests', async () => {
    const client = createApiClient('test-token');
    mockSuccessResponse({ deleted: true });

    await client.delete('/api/traces/1');

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces/1`,
      expect.objectContaining({ method: 'DELETE' })
    );
  });

  it('works without a token', async () => {
    const client = createApiClient();
    mockSuccessResponse({ data: 'public' });

    await client.get('/api/public');

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/public`,
      expect.objectContaining({ method: 'GET' })
    );
    const headers = getRequestHeaders();
    expect(headers.get('Content-Type')).toBe('application/json');
    expect(headers.has('Authorization')).toBe(false);
  });

  it('throws ApiError on non-ok response with message', async () => {
    const client = createApiClient('test-token');
    mockErrorResponse(404, 'Trace not found');

    await expect(client.get('/api/traces/unknown')).rejects.toThrow(ApiError);
    await expect(client.get('/api/traces/unknown')).rejects.toMatchObject({
      status: 404,
      message: 'Trace not found',
    });
  });

  it('throws ApiError with statusText when JSON parse fails', async () => {
    const client = createApiClient('test-token');
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      json: () => Promise.reject(new Error('Invalid JSON')),
    });

    await expect(client.get('/api/fail')).rejects.toThrow(ApiError);
  });

  it('sends POST with undefined body when no data provided', async () => {
    const client = createApiClient('test-token');
    mockSuccessResponse({});

    await client.post('/api/traces');

    expect(mockFetch).toHaveBeenCalledWith(
      `${API_URL}/api/traces`,
      expect.objectContaining({
        method: 'POST',
        body: undefined,
      })
    );
  });
});

describe('createGraphQLClient', () => {
  it('creates a GraphQL client with the correct endpoint', () => {
    const client = createGraphQLClient('test-token');
    expect(client).toBeDefined();
  });

  it('creates a GraphQL client without token', () => {
    const client = createGraphQLClient();
    expect(client).toBeDefined();
  });
});

describe('registered API contracts', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    mockFetch.mockReset();
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    setApiProjectId(undefined);
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('reads score lists from the scores field and normalizes categorical values', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          scores: [
            {
              id: 'score-1',
              projectId: 'project-1',
              traceId: 'trace-1',
              name: 'quality',
              source: 'API',
              dataType: 'CATEGORICAL',
              stringValue: 'good',
              createdAt: '2025-01-01T00:00:00.000Z',
              updatedAt: '2025-01-01T00:00:00.000Z',
            },
          ],
          totalCount: 1,
          hasMore: false,
        }),
    });

    const result = await api.scores.list();

    expect(result.scores[0]?.value).toBe('good');
    expect(result.totalCount).toBe(1);
  });

  it('normalizes an empty session list to an array', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          data: null,
          totalCount: 0,
          hasMore: false,
        }),
    });

    await expect(api.traces.sessions({ limit: 100 })).resolves.toEqual([]);
  });

  it('normalizes observation trees for trace consumers', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          observation: {
            id: 'root',
            traceId: 'trace-1',
            projectId: 'project-1',
            type: 'SPAN',
            name: 'root',
            startTime: '2025-01-01T00:00:00.000Z',
            level: 'DEFAULT',
          },
          children: [
            {
              observation: {
                id: 'child',
                traceId: 'trace-1',
                projectId: 'project-1',
                type: 'SPAN',
                name: 'child',
                startTime: '2025-01-01T00:00:01.000Z',
                level: 'DEFAULT',
              },
            },
          ],
        }),
    });

    const observations = await api.observations.listByTrace('trace-1');

    expect(observations).toHaveLength(2);
    expect(observations[0]?.id).toBe('root');
    expect(observations[1]?.id).toBe('child');
  });

  it('unwraps checkpoint data and derives file metadata', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          data: [
            {
              id: 'checkpoint-1',
              projectId: 'project-1',
              traceId: 'trace-1',
              name: 'Before edit',
              type: 'manual',
              filesSnapshot: JSON.stringify({
                'src/index.ts': { size: 42, hash: 'abc123' },
              }),
              totalFiles: 1,
              totalSizeBytes: 42,
              createdAt: '2025-01-01T00:00:00.000Z',
            },
          ],
          totalCount: 1,
          hasMore: false,
        }),
    });

    const result = await api.checkpoints.list('project-1');

    expect(result.checkpoints[0]).toMatchObject({
      name: 'Before edit',
      fileCount: 1,
      totalSize: 42,
      files: [{ path: 'src/index.ts', size: 42, hash: 'abc123' }],
    });
  });

  it('converts API key expiration presets to expiresAt', async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2025-01-01T00:00:00.000Z'));
    setApiProjectId('project-1');
    mockFetch.mockResolvedValue({
      ok: true,
      status: 201,
      json: () =>
        Promise.resolve({
          id: 'key-1',
          name: 'CI',
          publicKey: 'pk-at-public',
          secretKey: 'sk-at-secret',
          secretKeyPreview: 'cret',
          scopes: ['traces:write'],
          expiresAt: '2025-04-01T00:00:00.000Z',
          createdAt: '2025-01-01T00:00:00.000Z',
        }),
    });

    await api.apiKeys.create({
      name: 'CI',
      scopes: ['traces:write'],
      expiresIn: '90d',
    });

    const options = mockFetch.mock.calls[0]?.[1] as RequestInit;
    expect(JSON.parse(options.body as string)).toEqual({
      name: 'CI',
      scopes: ['traces:write'],
      expiresAt: '2025-04-01T00:00:00.000Z',
    });
  });

  it('resolves a prompt name before deleting through the ID route', async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () =>
          Promise.resolve({
            id: 'prompt-id',
            projectId: 'project-1',
            name: 'greeting',
            type: 'text',
            tags: [],
            createdAt: '2025-01-01T00:00:00.000Z',
            updatedAt: '2025-01-01T00:00:00.000Z',
          }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.resolve(undefined),
      });

    await api.prompts.delete('greeting');

    expect(mockFetch.mock.calls[1]?.[0]).toBe(`${API_URL}/api/public/prompts/prompt-id`);
    expect(mockFetch.mock.calls[1]?.[1]).toEqual(expect.objectContaining({ method: 'DELETE' }));
  });
});
