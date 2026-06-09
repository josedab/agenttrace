import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useEvalHubPackages } from '@/hooks/use-eval-hub';
import { evalHubApi } from '@/lib/eval-hub';

vi.mock('@/lib/eval-hub', async () => {
  const actual = await vi.importActual<typeof import('@/lib/eval-hub')>('@/lib/eval-hub');
  return {
    ...actual,
    evalHubApi: {
      ...actual.evalHubApi,
      listPackages: vi.fn(),
    },
  };
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('useEvalHubPackages', () => {
  it('keeps search and kind filters in the query contract', async () => {
    vi.mocked(evalHubApi.listPackages).mockResolvedValue({
      packages: [],
      totalCount: 0,
      hasMore: false,
    });
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(() => useEvalHubPackages({ query: 'quality', kind: 'dataset' }), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(evalHubApi.listPackages).toHaveBeenCalledWith({
      query: 'quality',
      kind: 'dataset',
    });
  });
});
