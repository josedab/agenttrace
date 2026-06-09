import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';

vi.mock('@/lib/api', () => ({
  api: {
    traces: {
      list: vi.fn(),
      count: vi.fn(),
      sessions: vi.fn(),
      stats: vi.fn(),
    },
  },
}));

import { useTraces, useTraceCount, useTraceSessions, useTraceStats } from '../use-traces';
import { api, type Trace } from '@/lib/api';

const mockedApi = vi.mocked(api, { deep: true, partial: true });
const timestamp = '2025-01-01T00:00:00.000Z';

function createTrace(overrides: Partial<Trace> = {}): Trace {
  return {
    id: 'trace-1',
    name: 'Test Trace',
    startTime: timestamp,
    endTime: null,
    input: null,
    output: null,
    metadata: null,
    tags: [],
    userId: null,
    sessionId: null,
    release: null,
    version: null,
    level: 'DEFAULT',
    statusMessage: null,
    public: false,
    latency: null,
    totalCost: null,
    usage: null,
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
    },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe('useTraces', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches traces with default filters', async () => {
    mockedApi.traces.list.mockResolvedValue({
      traces: [createTrace()],
      totalCount: 1,
      hasMore: false,
    });

    const { result } = renderHook(() => useTraces(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.traces.list).toHaveBeenCalledWith(expect.objectContaining({ limit: 50 }));
    expect(result.current.data?.pages[0].traces).toHaveLength(1);
  });

  it('passes filters to the API', async () => {
    mockedApi.traces.list.mockResolvedValue({
      traces: [],
      totalCount: 0,
      hasMore: false,
    });

    const filters = { search: 'test', level: 'ERROR' };
    const { result } = renderHook(() => useTraces(filters), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.traces.list).toHaveBeenCalledWith(
      expect.objectContaining({ search: 'test', level: 'ERROR', limit: 50 })
    );
  });

  it('indicates hasNextPage when more traces are available', async () => {
    mockedApi.traces.list.mockResolvedValue({
      traces: [createTrace()],
      totalCount: 2,
      hasMore: true,
    });

    const { result } = renderHook(() => useTraces(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.hasNextPage).toBe(true);
  });

  it('indicates no next page when all traces are loaded', async () => {
    mockedApi.traces.list.mockResolvedValue({
      traces: [createTrace()],
      totalCount: 1,
      hasMore: false,
    });

    const { result } = renderHook(() => useTraces(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.hasNextPage).toBe(false);
  });

  it('handles API errors', async () => {
    mockedApi.traces.list.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useTraces(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toBe('Network error');
  });
});

describe('useTraceCount', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches trace count', async () => {
    mockedApi.traces.count.mockResolvedValue({ count: 42 });

    const { result } = renderHook(() => useTraceCount(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ count: 42 });
  });

  it('passes filters to count endpoint', async () => {
    mockedApi.traces.count.mockResolvedValue({ count: 5 });

    renderHook(() => useTraceCount({ level: 'ERROR' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() =>
      expect(mockedApi.traces.count).toHaveBeenCalledWith(
        expect.objectContaining({ level: 'ERROR' })
      )
    );
  });
});

describe('useTraceSessions', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches sessions with default limit', async () => {
    mockedApi.traces.sessions.mockResolvedValue([]);

    const { result } = renderHook(() => useTraceSessions(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.traces.sessions).toHaveBeenCalledWith({ limit: 10 });
  });

  it('respects custom limit', async () => {
    mockedApi.traces.sessions.mockResolvedValue([]);

    renderHook(() => useTraceSessions(25), { wrapper: createWrapper() });

    await waitFor(() => expect(mockedApi.traces.sessions).toHaveBeenCalledWith({ limit: 25 }));
  });
});

describe('useTraceStats', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches stats with default date range', async () => {
    mockedApi.traces.stats.mockResolvedValue({ totalTraces: 100 });

    const { result } = renderHook(() => useTraceStats(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.traces.stats).toHaveBeenCalledWith({ dateRange: '7d' });
  });

  it('accepts custom date range', async () => {
    mockedApi.traces.stats.mockResolvedValue({ totalTraces: 50 });

    renderHook(() => useTraceStats('30d'), { wrapper: createWrapper() });

    await waitFor(() => expect(mockedApi.traces.stats).toHaveBeenCalledWith({ dateRange: '30d' }));
  });
});
