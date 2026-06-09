import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useReplayCapabilities } from '@/hooks/use-replay-plans';
import { replayPlansApi } from '@/lib/replay-plans';

vi.mock('@/lib/replay-plans', async () => {
  const actual = await vi.importActual<typeof import('@/lib/replay-plans')>('@/lib/replay-plans');
  return {
    ...actual,
    replayPlansApi: {
      ...actual.replayPlansApi,
      getCapabilities: vi.fn(),
    },
  };
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('useReplayCapabilities', () => {
  it('keeps checkpoint and mode in the capability query', async () => {
    vi.mocked(replayPlansApi.getCapabilities).mockResolvedValue({
      canInspectTimeline: true,
      canReplayRecordedGeneration: true,
      canExecuteInSandbox: false,
      hasCheckpoint: true,
      hasFileOperations: true,
      hasTerminalCommands: true,
      generationCount: 2,
      unsupportedReasons: [],
      safetyNotice: 'No host execution.',
    });
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(
      () => useReplayCapabilities('trace-1', 'checkpoint-1', 'recorded_generation'),
      { wrapper }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(replayPlansApi.getCapabilities).toHaveBeenCalledWith('trace-1', {
      checkpointId: 'checkpoint-1',
      mode: 'recorded_generation',
    });
  });
});
