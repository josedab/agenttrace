import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@/lib/api', () => ({
  API_URL: 'https://api.example.test',
  getApiAccessToken: vi.fn(),
  getApiProjectId: vi.fn(),
}));

import { getApiAccessToken, getApiProjectId } from '@/lib/api';
import { useTraceStream } from '../use-streaming';

const mockedGetApiAccessToken = vi.mocked(getApiAccessToken);
const mockedGetApiProjectId = vi.mocked(getApiProjectId);

describe('useTraceStream', () => {
  beforeEach(() => {
    mockedGetApiAccessToken.mockReturnValue('header.jwt.token');
    mockedGetApiProjectId.mockReturnValue('project-1');
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.unstubAllGlobals();
  });

  it('authenticates with a header and parses CRLF-delimited events', async () => {
    let streamController: ReadableStreamDefaultController<Uint8Array> | undefined;
    const stream = new ReadableStream<Uint8Array>({
      start(controller) {
        streamController = controller;
        controller.enqueue(
          new TextEncoder().encode(
            [
              'event: metrics',
              'data: {"traceId":"trace-1","activeSpans":1}',
              '',
              'event: activity',
              'data: {"id":"activity-1","traceId":"trace-1","type":"tool"}',
              '',
              '',
            ].join('\r\n')
          )
        );
      },
    });
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      body: stream,
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result, unmount } = renderHook(() => useTraceStream('trace-1'));

    await waitFor(() => {
      expect(result.current.metrics?.traceId).toBe('trace-1');
      expect(result.current.activities).toHaveLength(1);
    });
    expect(result.current.connected).toBe(true);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('https://api.example.test/api/public/traces/trace-1/stream?follow=true');
    expect(url).not.toContain('header.jwt.token');
    expect(new Headers(init.headers).get('Authorization')).toBe('Bearer header.jwt.token');

    expect(new Headers(init.headers).get('X-Project-ID')).toBe('project-1');

    const signal = init.signal as AbortSignal;
    act(() => unmount());
    expect(signal.aborted).toBe(true);
    streamController?.close();
  });
});
