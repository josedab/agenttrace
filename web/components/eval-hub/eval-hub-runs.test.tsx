import { QueryClient, QueryClientProvider, type UseQueryResult } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { EvalHubRuns } from '@/components/eval-hub/eval-hub-runs';
import type { EvalHubRunList } from '@/lib/eval-hub';

const queryClient = new QueryClient();

function renderRuns(query: UseQueryResult<EvalHubRunList, Error>) {
  return render(
    <QueryClientProvider client={queryClient}>
      <EvalHubRuns query={query} />
    </QueryClientProvider>
  );
}

describe('EvalHubRuns', () => {
  it('renders an explicit empty state', () => {
    renderRuns({
      data: { runs: [], totalCount: 0, hasMore: false },
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<EvalHubRunList, Error>);

    expect(screen.getByText('No Eval Hub runs yet')).toBeInTheDocument();
  });

  it('renders unsupported capability messages', () => {
    renderRuns({
      data: {
        totalCount: 1,
        hasMore: false,
        runs: [
          {
            id: 'run-1',
            projectId: 'project-1',
            packageId: 'package-1',
            packageVersion: 1,
            status: 'unsupported',
            capabilityMessage: 'Fork this package before running it.',
            createdBy: 'user-1',
            startedAt: '2026-07-25T00:00:00Z',
          },
        ],
      },
      isLoading: false,
      isError: false,
    } as unknown as UseQueryResult<EvalHubRunList, Error>);

    expect(screen.getByText('unsupported')).toBeInTheDocument();
    expect(screen.getByText('Fork this package before running it.')).toBeInTheDocument();
  });
});
