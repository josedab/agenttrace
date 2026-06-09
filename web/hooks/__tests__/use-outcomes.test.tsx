import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useOutcomeOverview } from '@/hooks/use-outcomes';
import { outcomesApi } from '@/lib/outcomes';

vi.mock('@/lib/outcomes', async () => {
  const actual = await vi.importActual<typeof import('@/lib/outcomes')>('@/lib/outcomes');
  return {
    ...actual,
    outcomesApi: {
      ...actual.outcomesApi,
      getOverview: vi.fn(),
    },
  };
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('useOutcomeOverview', () => {
  it('loads the selected project outcome window', async () => {
    vi.mocked(outcomesApi.getOverview).mockResolvedValue({
      projectId: 'project-1',
      period: { from: '2026-07-01T00:00:00Z', to: '2026-07-25T00:00:00Z' },
      runs: {
        total: 2,
        successful: 1,
        failed: 1,
        inProgress: 0,
        successRate: { value: 0.5, available: true },
      },
      ci: {
        total: 0,
        passed: 0,
        failed: 0,
        cancelled: 0,
        inProgress: 0,
        linkedRuns: 0,
        passRate: { value: null, available: false },
      },
      sourceControl: {
        linkedCommits: 0,
        linkedTraces: 0,
        linkedPullRequests: 0,
        regressionSignals: 0,
        revertSignals: 0,
      },
      cost: {
        totalCost: 1,
        costPerSuccessfulOutcome: { value: 1, available: true },
      },
      byAgent: [],
      byModel: [],
      recentOutcomes: [],
      availability: {
        traceData: true,
        gitData: false,
        ciData: false,
        pullRequestData: false,
        agentAttribution: false,
        modelAttribution: false,
        unavailable: ['CI outcomes'],
      },
      generatedAt: '2026-07-25T00:00:00Z',
    });

    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(() => useOutcomeOverview('7d'), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(outcomesApi.getOverview).toHaveBeenCalledWith('7d');
    expect(result.current.data?.runs.successRate.value).toBe(0.5);
  });
});
