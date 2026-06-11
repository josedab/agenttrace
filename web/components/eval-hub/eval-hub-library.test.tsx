import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { EvalHubLibrary } from '@/components/eval-hub/eval-hub-library';
import { evalHubApi } from '@/lib/eval-hub';

vi.mock('@/lib/eval-hub', async () => {
  const actual = await vi.importActual<typeof import('@/lib/eval-hub')>('@/lib/eval-hub');
  return {
    ...actual,
    evalHubApi: {
      ...actual.evalHubApi,
      listPackages: vi.fn(),
      run: vi.fn(),
      fork: vi.fn(),
    },
  };
});

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api');
  return {
    ...actual,
    getApiProjectId: () => 'project-1',
  };
});

const ownedPackage = {
  id: 'package-1',
  ownerProjectId: 'project-1',
  organizationId: 'org-1',
  kind: 'dataset' as const,
  name: 'Quality dataset',
  description: 'Regression cases',
  visibility: 'private' as const,
  latestVersion: 3,
  publishedBy: 'user-1',
  createdAt: '2026-07-25T00:00:00Z',
  updatedAt: '2026-07-25T00:00:00Z',
};

function renderLibrary() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <EvalHubLibrary />
    </QueryClientProvider>
  );
}

afterEach(() => {
  vi.clearAllMocks();
});

describe('EvalHubLibrary', () => {
  it('labels the search and filter controls for assistive technology', async () => {
    vi.mocked(evalHubApi.listPackages).mockResolvedValue({
      packages: [ownedPackage],
      totalCount: 1,
      hasMore: false,
    });

    renderLibrary();

    expect(await screen.findByLabelText('Search packages')).toBeInTheDocument();
    expect(screen.getByLabelText('Asset type')).toBeInTheDocument();
    expect(
      await screen.findByRole('button', { name: 'Run Quality dataset in this project' })
    ).toBeInTheDocument();
  });

  it('reuses one idempotency key for a repeated run of the same version', async () => {
    vi.mocked(evalHubApi.listPackages).mockResolvedValue({
      packages: [ownedPackage],
      totalCount: 1,
      hasMore: false,
    });
    vi.mocked(evalHubApi.run).mockRejectedValue(new Error('network unavailable'));

    const user = userEvent.setup();
    renderLibrary();

    const runButton = await screen.findByRole('button', {
      name: 'Run Quality dataset in this project',
    });

    await user.click(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(runButton).toBeEnabled());

    await user.click(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalledTimes(2));

    const firstKey = vi.mocked(evalHubApi.run).mock.calls[0][1]?.idempotencyKey;
    const secondKey = vi.mocked(evalHubApi.run).mock.calls[1][1]?.idempotencyKey;
    expect(firstKey).toBeTruthy();
    // A retry after a failed attempt must not start a second server-side run.
    expect(secondKey).toBe(firstKey);
  });

  it('keeps the idempotency key while the durable run is still ready', async () => {
    vi.mocked(evalHubApi.listPackages).mockResolvedValue({
      packages: [ownedPackage],
      totalCount: 1,
      hasMore: false,
    });
    vi.mocked(evalHubApi.run).mockResolvedValue({
      id: 'run-1',
      projectId: 'project-1',
      packageId: 'package-1',
      packageVersion: 3,
      status: 'ready',
      createdBy: 'user-1',
      startedAt: '2026-07-25T00:00:00Z',
    });

    const user = userEvent.setup();
    renderLibrary();

    const runButton = await screen.findByRole('button', {
      name: 'Run Quality dataset in this project',
    });

    await user.click(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(runButton).toBeEnabled());

    await user.click(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalledTimes(2));

    const firstKey = vi.mocked(evalHubApi.run).mock.calls[0][1]?.idempotencyKey;
    const secondKey = vi.mocked(evalHubApi.run).mock.calls[1][1]?.idempotencyKey;
    expect(firstKey).toBeTruthy();
    expect(secondKey).toBe(firstKey);
  });

  it('uses one idempotency key across a rapid double-click', async () => {
    vi.mocked(evalHubApi.listPackages).mockResolvedValue({
      packages: [ownedPackage],
      totalCount: 1,
      hasMore: false,
    });
    vi.mocked(evalHubApi.run).mockResolvedValue({
      id: 'run-1',
      projectId: 'project-1',
      packageId: 'package-1',
      packageVersion: 3,
      status: 'running',
      createdBy: 'user-1',
      startedAt: '2026-07-25T00:00:00Z',
    });

    const user = userEvent.setup();
    renderLibrary();

    const runButton = await screen.findByRole('button', {
      name: 'Run Quality dataset in this project',
    });
    await user.dblClick(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalled());

    const keys = vi
      .mocked(evalHubApi.run)
      .mock.calls.map((call) => call[1]?.idempotencyKey)
      .filter(Boolean);
    expect(keys).not.toHaveLength(0);
    expect(new Set(keys).size).toBe(1);
  });

  it('issues a new idempotency key after a run completes', async () => {
    vi.mocked(evalHubApi.listPackages).mockResolvedValue({
      packages: [ownedPackage],
      totalCount: 1,
      hasMore: false,
    });
    vi.mocked(evalHubApi.run).mockResolvedValue({
      id: 'run-1',
      projectId: 'project-1',
      packageId: 'package-1',
      packageVersion: 3,
      status: 'completed',
      createdBy: 'user-1',
      startedAt: '2026-07-25T00:00:00Z',
      completedAt: '2026-07-25T00:01:00Z',
    });

    const user = userEvent.setup();
    renderLibrary();

    const runButton = await screen.findByRole('button', {
      name: 'Run Quality dataset in this project',
    });

    await user.click(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(runButton).toBeEnabled());

    await user.click(runButton);
    await waitFor(() => expect(evalHubApi.run).toHaveBeenCalledTimes(2));

    const firstKey = vi.mocked(evalHubApi.run).mock.calls[0][1]?.idempotencyKey;
    const secondKey = vi.mocked(evalHubApi.run).mock.calls[1][1]?.idempotencyKey;
    expect(firstKey).toBeTruthy();
    expect(secondKey).not.toBe(firstKey);
  });
});
