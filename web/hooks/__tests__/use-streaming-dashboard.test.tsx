import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
  },
}));

import { api } from '@/lib/api';
import { useStreamingDashboard } from '../use-streaming-dashboard';

const mockedGet = vi.mocked(api.get);

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe('useStreamingDashboard', () => {
  beforeEach(() => vi.clearAllMocks());

  it('normalizes backend live metrics', async () => {
    mockedGet.mockResolvedValue({
      activeSessions: 1,
      totalCost: 0.5,
      totalTokens: 100,
      errorCount: 0,
      activeStreams: [
        {
          traceId: 'trace-1',
          activeSpans: 1,
          completedSpans: 3,
          totalTokens: 100,
          totalCost: 0.5,
          errorCount: 0,
          elapsedMs: 2500,
          tokensPerSecond: 40,
          costPerMinute: 0.1,
        },
      ],
      topModels: [
        {
          model: 'gpt-test',
          requestCount: 2,
          totalTokens: 100,
          totalCost: 0.5,
        },
      ],
    });

    const { result } = renderHook(() => useStreamingDashboard(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.activeStreams[0]).toMatchObject({
      id: 'trace-1',
      model: 'Trace stream',
      elapsedSeconds: 2,
      progress: 75,
      status: 'streaming',
    });
    expect(result.current.data?.topModels[0]?.sessions).toBe(2);
  });
});
