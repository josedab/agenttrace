import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';

vi.mock('@/lib/api', () => ({
  api: {
    scores: {
      list: vi.fn(),
      get: vi.fn(),
      listByTrace: vi.fn(),
      listByObservation: vi.fn(),
      stats: vi.fn(),
      distribution: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      names: vi.fn(),
    },
  },
}));

import {
  useScores,
  useScore,
  useScoresByTrace,
  useScoresByObservation,
  useScoreStats,
  useScoreDistribution,
  useCreateScore,
  useDeleteScore,
  useScoreNames,
} from '../use-scores';
import { api, type Score } from '@/lib/api';

const mockedApi = vi.mocked(api, { deep: true, partial: true });
const timestamp = '2025-01-01T00:00:00.000Z';

function createScore(overrides: Partial<Score> = {}): Score {
  return {
    id: 'score-1',
    traceId: 'trace-1',
    observationId: null,
    projectId: 'project-1',
    name: 'accuracy',
    value: 0.95,
    stringValue: null,
    dataType: 'NUMERIC',
    source: 'API',
    comment: null,
    createdAt: timestamp,
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

describe('useScores', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches scores with default filters', async () => {
    mockedApi.scores.list.mockResolvedValue({
      scores: [createScore()],
      totalCount: 1,
      hasMore: false,
    });

    const { result } = renderHook(() => useScores(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.list).toHaveBeenCalledWith(expect.objectContaining({ limit: 50 }));
    expect(result.current.data?.pages[0].scores).toHaveLength(1);
  });

  it('passes filters to the API', async () => {
    mockedApi.scores.list.mockResolvedValue({
      scores: [],
      totalCount: 0,
      hasMore: false,
    });

    const filters = { scoreName: 'accuracy', minScore: 0.5 };
    const { result } = renderHook(() => useScores(filters), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.list).toHaveBeenCalledWith(
      expect.objectContaining({ scoreName: 'accuracy', minScore: 0.5, limit: 50 })
    );
  });

  it('indicates hasNextPage when more scores are available', async () => {
    mockedApi.scores.list.mockResolvedValue({
      scores: [createScore()],
      totalCount: 2,
      hasMore: true,
    });

    const { result } = renderHook(() => useScores(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.hasNextPage).toBe(true);
  });

  it('indicates no next page when all scores are loaded', async () => {
    mockedApi.scores.list.mockResolvedValue({
      scores: [createScore()],
      totalCount: 1,
      hasMore: false,
    });

    const { result } = renderHook(() => useScores(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.hasNextPage).toBe(false);
  });
});

describe('useScore', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single score by ID', async () => {
    const mockScore = createScore();
    mockedApi.scores.get.mockResolvedValue(mockScore);

    const { result } = renderHook(() => useScore('score-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.get).toHaveBeenCalledWith('score-1');
    expect(result.current.data).toEqual(mockScore);
  });

  it('does not fetch when scoreId is empty', async () => {
    const { result } = renderHook(() => useScore(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockedApi.scores.get).not.toHaveBeenCalled();
  });
});

describe('useScoresByTrace', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches scores for a trace', async () => {
    const mockScores = [
      createScore(),
      createScore({ id: 'score-2', name: 'relevance', value: 0.8 }),
    ];
    mockedApi.scores.listByTrace.mockResolvedValue(mockScores);

    const { result } = renderHook(() => useScoresByTrace('trace-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.listByTrace).toHaveBeenCalledWith('trace-1');
    expect(result.current.data).toHaveLength(2);
  });

  it('does not fetch when traceId is empty', async () => {
    const { result } = renderHook(() => useScoresByTrace(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockedApi.scores.listByTrace).not.toHaveBeenCalled();
  });
});

describe('useScoresByObservation', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches scores for an observation', async () => {
    mockedApi.scores.listByObservation.mockResolvedValue([
      createScore({
        observationId: 'obs-1',
        name: 'quality',
        value: 4,
      }),
    ]);

    const { result } = renderHook(() => useScoresByObservation('obs-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.listByObservation).toHaveBeenCalledWith('obs-1');
  });

  it('does not fetch when observationId is empty', async () => {
    const { result } = renderHook(() => useScoresByObservation(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useScoreStats', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches stats for a score name', async () => {
    mockedApi.scores.stats.mockResolvedValue({ name: 'accuracy', count: 100 });

    const { result } = renderHook(() => useScoreStats('accuracy'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.stats).toHaveBeenCalledWith({ name: 'accuracy' });
  });

  it('does not fetch stats without a score name', async () => {
    const { result } = renderHook(() => useScoreStats(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockedApi.scores.stats).not.toHaveBeenCalled();
  });
});

describe('useScoreDistribution', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches distribution for a score name', async () => {
    mockedApi.scores.distribution.mockResolvedValue({ buckets: [] });

    const { result } = renderHook(() => useScoreDistribution('accuracy'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.scores.distribution).toHaveBeenCalledWith({
      scoreName: 'accuracy',
      dateRange: '7d',
    });
  });

  it('does not fetch when scoreName is empty', async () => {
    const { result } = renderHook(() => useScoreDistribution(''), { wrapper: createWrapper() });

    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreateScore', () => {
  beforeEach(() => vi.clearAllMocks());

  it('creates a score and calls API', async () => {
    const newScore = createScore({ id: 'score-new', value: 0.9 });
    mockedApi.scores.create.mockResolvedValue(newScore);

    const { result } = renderHook(() => useCreateScore(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({
      traceId: 'trace-1',
      name: 'accuracy',
      value: 0.9,
    });

    expect(mockedApi.scores.create).toHaveBeenCalledWith({
      traceId: 'trace-1',
      name: 'accuracy',
      value: 0.9,
    });
  });
});

describe('useDeleteScore', () => {
  beforeEach(() => vi.clearAllMocks());

  it('deletes a score', async () => {
    mockedApi.scores.delete.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteScore(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync('score-1');

    expect(mockedApi.scores.delete).toHaveBeenCalledWith('score-1');
  });
});

describe('useScoreNames', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches available score names', async () => {
    mockedApi.scores.names.mockResolvedValue(['accuracy', 'relevance', 'quality']);

    const { result } = renderHook(() => useScoreNames(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(['accuracy', 'relevance', 'quality']);
  });
});
