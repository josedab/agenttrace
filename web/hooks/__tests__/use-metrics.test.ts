import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';

vi.mock('@/lib/api', () => ({
  api: {
    analytics: {
      getDashboardMetrics: vi.fn(),
      getOverview: vi.fn(),
      getCostAnalytics: vi.fn(),
      getLatencyAnalytics: vi.fn(),
      getUsageAnalytics: vi.fn(),
      getTraceVolume: vi.fn(),
      getCostOverTime: vi.fn(),
      getLatencyPercentiles: vi.fn(),
      getModelUsage: vi.fn(),
      getTopTracesByTokens: vi.fn(),
      getTopTracesByCost: vi.fn(),
      getRecentErrors: vi.fn(),
      getProjectMetrics: vi.fn(),
    },
  },
}));

import {
  useDashboardMetrics,
  useCostAnalytics,
  useLatencyAnalytics,
  useUsageAnalytics,
  useProjectMetrics,
  useModelUsage,
  useRecentErrors,
} from '../use-metrics';
import {
  api,
  type CostAnalytics,
  type DashboardMetrics,
  type LatencyAnalytics,
  type UsageAnalytics,
} from '@/lib/api';

const mockedApi = vi.mocked(api, { deep: true, partial: true });

function createDashboardMetrics(overrides: Partial<DashboardMetrics> = {}): DashboardMetrics {
  return {
    totalTraces: 0,
    totalCost: 0,
    avgLatency: 0,
    activeSessions: 0,
    tracesChange: 0,
    costChange: 0,
    latencyChange: 0,
    sessionsChange: 0,
    traceVolume: [],
    costBreakdown: [],
    latencyPercentiles: [],
    recentTraces: [],
    ...overrides,
  };
}

function createCostAnalytics(overrides: Partial<CostAnalytics> = {}): CostAnalytics {
  return {
    totalCost: 0,
    costChange: 0,
    inputCost: 0,
    outputCost: 0,
    avgCostPerTrace: 0,
    costOverTime: [],
    costByGroup: [],
    breakdown: [],
    ...overrides,
  };
}

function createLatencyAnalytics(overrides: Partial<LatencyAnalytics> = {}): LatencyAnalytics {
  return {
    p50: 0,
    p95: 0,
    p99: 0,
    avg: 0,
    avgChange: 0,
    totalRequests: 0,
    latencyOverTime: [],
    distribution: [],
    breakdown: [],
    ...overrides,
  };
}

function createUsageAnalytics(overrides: Partial<UsageAnalytics> = {}): UsageAnalytics {
  return {
    totalTraces: 0,
    tracesChange: 0,
    totalGenerations: 0,
    inputTokens: 0,
    outputTokens: 0,
    volumeOverTime: [],
    tokenUsageOverTime: [],
    modelDistribution: [],
    topTraces: [],
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

describe('useDashboardMetrics', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches dashboard metrics with default range', async () => {
    const mockMetrics = createDashboardMetrics({
      totalTraces: 1000,
      totalCost: 42.5,
      avgLatency: 250,
      activeSessions: 15,
    });
    mockedApi.analytics.getDashboardMetrics.mockResolvedValue(mockMetrics);

    const { result } = renderHook(() => useDashboardMetrics(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getDashboardMetrics).toHaveBeenCalledWith({
      dateRange: '7d',
    });
    expect(result.current.data).toEqual(mockMetrics);
  });

  it('accepts custom date range', async () => {
    mockedApi.analytics.getDashboardMetrics.mockResolvedValue(createDashboardMetrics());

    renderHook(() => useDashboardMetrics('30d'), {
      wrapper: createWrapper(),
    });

    await waitFor(() =>
      expect(mockedApi.analytics.getDashboardMetrics).toHaveBeenCalledWith({
        dateRange: '30d',
      })
    );
  });
});

describe('useCostAnalytics', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches cost analytics with options', async () => {
    mockedApi.analytics.getCostAnalytics.mockResolvedValue(createCostAnalytics());

    const { result } = renderHook(() => useCostAnalytics({ dateRange: '7d', groupBy: 'model' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getCostAnalytics).toHaveBeenCalledWith({
      dateRange: '7d',
      groupBy: 'model',
    });
  });
});

describe('useLatencyAnalytics', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches latency analytics', async () => {
    mockedApi.analytics.getLatencyAnalytics.mockResolvedValue(createLatencyAnalytics());

    const { result } = renderHook(() => useLatencyAnalytics({ dateRange: '7d', groupBy: 'name' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getLatencyAnalytics).toHaveBeenCalledWith({
      dateRange: '7d',
      groupBy: 'name',
    });
  });
});

describe('useUsageAnalytics', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches usage analytics', async () => {
    mockedApi.analytics.getUsageAnalytics.mockResolvedValue(createUsageAnalytics());

    const { result } = renderHook(() => useUsageAnalytics({ dateRange: '7d' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getUsageAnalytics).toHaveBeenCalledWith({
      dateRange: '7d',
    });
  });
});

describe('useProjectMetrics', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches project metrics when projectId is provided', async () => {
    mockedApi.analytics.getProjectMetrics.mockResolvedValue({
      ...createDashboardMetrics(),
      totalTraces: 50,
    });

    const { result } = renderHook(() => useProjectMetrics('proj-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getProjectMetrics).toHaveBeenCalledWith('proj-1');
  });

  it('does not fetch when projectId is empty', async () => {
    const { result } = renderHook(() => useProjectMetrics(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockedApi.analytics.getProjectMetrics).not.toHaveBeenCalled();
  });
});

describe('useModelUsage', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches model usage with default limit', async () => {
    mockedApi.analytics.getModelUsage.mockResolvedValue([]);

    const { result } = renderHook(() => useModelUsage(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getModelUsage).toHaveBeenCalledWith({
      dateRange: '7d',
    });
  });
});

describe('useRecentErrors', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches recent errors with default limit', async () => {
    mockedApi.analytics.getRecentErrors.mockResolvedValue([]);

    const { result } = renderHook(() => useRecentErrors(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.analytics.getRecentErrors).toHaveBeenCalledWith({ limit: 10 });
  });

  it('respects custom limit', async () => {
    mockedApi.analytics.getRecentErrors.mockResolvedValue([]);

    renderHook(() => useRecentErrors(25), { wrapper: createWrapper() });

    await waitFor(() =>
      expect(mockedApi.analytics.getRecentErrors).toHaveBeenCalledWith({ limit: 25 })
    );
  });
});
