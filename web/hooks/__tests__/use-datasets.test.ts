import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';

vi.mock('@/lib/api', () => ({
  api: {
    datasets: {
      list: vi.fn(),
      get: vi.fn(),
      listItems: vi.fn(),
      listRuns: vi.fn(),
      getRun: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      addItem: vi.fn(),
      updateItem: vi.fn(),
      deleteItem: vi.fn(),
      createRun: vi.fn(),
    },
  },
}));

import {
  useDatasets,
  useDataset,
  useDatasetItems,
  useCreateDataset,
  useDeleteDataset,
  useDatasetRuns,
} from '../use-datasets';
import { api, type Dataset, type DatasetItem, type DatasetRun } from '@/lib/api';

const mockedApi = vi.mocked(api, { deep: true, partial: true });
const timestamp = '2025-01-01T00:00:00.000Z';

function createDataset(overrides: Partial<Dataset> = {}): Dataset {
  return {
    id: 'ds-1',
    name: 'Dataset',
    description: null,
    metadata: null,
    itemCount: 0,
    runCount: 0,
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides,
  };
}

function createDatasetItem(overrides: Partial<DatasetItem> = {}): DatasetItem {
  return {
    id: 'item-1',
    datasetId: 'ds-1',
    input: {},
    expectedOutput: null,
    metadata: null,
    sourceTraceId: null,
    sourceObservationId: null,
    status: 'ACTIVE',
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides,
  };
}

function createDatasetRun(overrides: Partial<DatasetRun> = {}): DatasetRun {
  return {
    id: 'run-1',
    datasetId: 'ds-1',
    name: 'Run',
    description: null,
    metadata: null,
    status: 'PENDING',
    itemCount: 0,
    totalCount: 0,
    completedCount: 0,
    failedCount: 0,
    avgScore: null,
    totalCost: null,
    createdAt: timestamp,
    updatedAt: timestamp,
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

describe('useDatasets', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches all datasets', async () => {
    const mockDatasets = [
      createDataset({ id: 'ds-1', name: 'Dataset 1' }),
      createDataset({ id: 'ds-2', name: 'Dataset 2' }),
    ];
    mockedApi.datasets.list.mockResolvedValue(mockDatasets);

    const { result } = renderHook(() => useDatasets(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDatasets);
    expect(result.current.data).toHaveLength(2);
  });

  it('handles empty dataset list', async () => {
    mockedApi.datasets.list.mockResolvedValue([]);

    const { result } = renderHook(() => useDatasets(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles fetch errors', async () => {
    mockedApi.datasets.list.mockRejectedValue(new Error('Server error'));

    const { result } = renderHook(() => useDatasets(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toBe('Server error');
  });
});

describe('useDataset', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single dataset by ID', async () => {
    const mockDataset = createDataset({
      id: 'ds-1',
      name: 'Dataset 1',
      itemCount: 10,
    });
    mockedApi.datasets.get.mockResolvedValue(mockDataset);

    const { result } = renderHook(() => useDataset('ds-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.datasets.get).toHaveBeenCalledWith('ds-1');
    expect(result.current.data).toEqual(mockDataset);
  });

  it('does not fetch when datasetId is empty', async () => {
    const { result } = renderHook(() => useDataset(''), {
      wrapper: createWrapper(),
    });

    // Should not trigger a fetch
    expect(result.current.fetchStatus).toBe('idle');
    expect(mockedApi.datasets.get).not.toHaveBeenCalled();
  });
});

describe('useDatasetItems', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches paginated dataset items', async () => {
    mockedApi.datasets.listItems.mockResolvedValue({
      data: [createDatasetItem({ input: { text: 'hello' } })],
      totalCount: 2,
    });

    const { result } = renderHook(() => useDatasetItems('ds-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.datasets.listItems).toHaveBeenCalledWith('ds-1', {
      offset: 0,
      limit: 50,
    });
    expect(result.current.hasNextPage).toBe(true);
  });

  it('does not fetch when datasetId is empty', async () => {
    const { result } = renderHook(() => useDatasetItems(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useDatasetRuns', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches runs for a dataset', async () => {
    const mockRuns = [createDatasetRun({ status: 'COMPLETED' })];
    mockedApi.datasets.listRuns.mockResolvedValue(mockRuns);

    const { result } = renderHook(() => useDatasetRuns('ds-1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockRuns);
  });
});

describe('useCreateDataset', () => {
  beforeEach(() => vi.clearAllMocks());

  it('creates a dataset and invalidates cache', async () => {
    const newDataset = createDataset({
      id: 'ds-new',
      name: 'New Dataset',
    });
    mockedApi.datasets.create.mockResolvedValue(newDataset);

    const { result } = renderHook(() => useCreateDataset(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({ name: 'New Dataset' });

    expect(mockedApi.datasets.create).toHaveBeenCalledWith({ name: 'New Dataset' });
  });
});

describe('useDeleteDataset', () => {
  beforeEach(() => vi.clearAllMocks());

  it('deletes a dataset', async () => {
    mockedApi.datasets.delete.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteDataset(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync('ds-1');

    expect(mockedApi.datasets.delete).toHaveBeenCalledWith('ds-1');
  });
});
