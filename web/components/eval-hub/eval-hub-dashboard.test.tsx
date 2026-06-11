import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { EvalHubDashboard } from '@/components/eval-hub/eval-hub-dashboard';
import { api } from '@/lib/api';
import { useEvalHubPackages, useEvalHubRuns } from '@/hooks/use-eval-hub';

const routerState = vi.hoisted(() => ({
  replace: vi.fn(),
  searchParams: new URLSearchParams(),
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({ replace: routerState.replace }),
  useSearchParams: () => routerState.searchParams,
}));

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api')>('@/lib/api');
  return {
    ...actual,
    api: {
      ...actual.api,
      evaluators: { ...actual.api.evaluators, list: vi.fn() },
      datasets: { ...actual.api.datasets, list: vi.fn() },
    },
  };
});

vi.mock('@/hooks/use-eval-hub', async () => {
  const actual = await vi.importActual<typeof import('@/hooks/use-eval-hub')>(
    '@/hooks/use-eval-hub'
  );
  return {
    ...actual,
    useEvalHubPackages: vi.fn(),
    useEvalHubRuns: vi.fn(),
  };
});

vi.mock('@/components/eval-hub/eval-hub-overview', () => ({
  EvalHubOverview: (props: {
    evaluatorCount?: number;
    datasetCount?: number;
    packageCount?: number;
    runCount?: number;
    loading: boolean;
    onNavigate: (view: string) => void;
  }) => (
    <div>
      <h2>Overview surface</h2>
      <p>loading:{String(props.loading)}</p>
      <p>evaluators:{props.evaluatorCount ?? 'none'}</p>
      <p>datasets:{props.datasetCount ?? 'none'}</p>
      <p>packages:{props.packageCount ?? 'none'}</p>
      <p>runs:{props.runCount ?? 'none'}</p>
      <button type="button" onClick={() => props.onNavigate('runs')}>
        Go to runs
      </button>
    </div>
  ),
}));
vi.mock('@/components/eval-hub/eval-hub-assets', () => ({
  EvalHubEvaluators: () => <h2>Evaluators surface</h2>,
  EvalHubDatasets: () => <h2>Datasets surface</h2>,
}));
vi.mock('@/components/eval-hub/eval-hub-library', () => ({
  EvalHubLibrary: () => <h2>Library surface</h2>,
}));
vi.mock('@/components/eval-hub/eval-hub-runs', () => ({
  EvalHubRuns: () => <h2>Runs surface</h2>,
}));

function renderDashboard() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <EvalHubDashboard />
    </QueryClientProvider>
  );
}

function setView(view: string | null) {
  routerState.searchParams = view ? new URLSearchParams({ view }) : new URLSearchParams();
}

afterEach(() => {
  vi.clearAllMocks();
  setView(null);
});

describe('EvalHubDashboard', () => {
  it('defaults to the overview surface with no query param', async () => {
    vi.mocked(api.evaluators.list).mockResolvedValue([]);
    vi.mocked(api.datasets.list).mockResolvedValue([]);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: { packages: [], totalCount: 0, hasMore: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: { runs: [], totalCount: 0, hasMore: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);

    renderDashboard();

    expect(await screen.findByText('Overview surface')).toBeInTheDocument();
    expect(screen.queryByText('Evaluators surface')).not.toBeInTheDocument();
    expect(screen.queryByText('Datasets surface')).not.toBeInTheDocument();
    expect(screen.queryByText('Library surface')).not.toBeInTheDocument();
    expect(screen.queryByText('Runs surface')).not.toBeInTheDocument();
  });

  it('falls back to overview for an invalid view query param', async () => {
    setView('not-a-real-view');
    vi.mocked(api.evaluators.list).mockResolvedValue([]);
    vi.mocked(api.datasets.list).mockResolvedValue([]);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);

    renderDashboard();

    expect(await screen.findByText('Overview surface')).toBeInTheDocument();
  });

  it.each([
    ['evaluators', 'Evaluators surface'],
    ['datasets', 'Datasets surface'],
    ['library', 'Library surface'],
    ['runs', 'Runs surface'],
  ])('renders only the %s surface for its view query param', async (view, expectedHeading) => {
    setView(view);
    vi.mocked(api.evaluators.list).mockResolvedValue([]);
    vi.mocked(api.datasets.list).mockResolvedValue([]);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);

    renderDashboard();

    expect(await screen.findByText(expectedHeading)).toBeInTheDocument();
    for (const heading of [
      'Overview surface',
      'Evaluators surface',
      'Datasets surface',
      'Library surface',
      'Runs surface',
    ]) {
      if (heading !== expectedHeading) {
        expect(screen.queryByText(heading)).not.toBeInTheDocument();
      }
    }
  });

  it('passes loading state and resolved counts through to the overview surface', async () => {
    vi.mocked(api.evaluators.list).mockResolvedValue([{ id: 'e-1' }, { id: 'e-2' }] as never);
    vi.mocked(api.datasets.list).mockResolvedValue([{ id: 'd-1' }] as never);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: { packages: [], totalCount: 9, hasMore: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: { runs: [], totalCount: 5, hasMore: false },
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);

    renderDashboard();

    expect(await screen.findByText('loading:false')).toBeInTheDocument();
    expect(screen.getByText('evaluators:2')).toBeInTheDocument();
    expect(screen.getByText('datasets:1')).toBeInTheDocument();
    expect(screen.getByText('packages:9')).toBeInTheDocument();
    expect(screen.getByText('runs:5')).toBeInTheDocument();
  });

  it('reports loading while any underlying eval hub query is still pending', async () => {
    vi.mocked(api.evaluators.list).mockResolvedValue([]);
    vi.mocked(api.datasets.list).mockResolvedValue([]);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);

    renderDashboard();

    expect(await screen.findByText('loading:true')).toBeInTheDocument();
  });

  it('navigates via router.replace to a query-param URL when a non-overview tab is selected', async () => {
    vi.mocked(api.evaluators.list).mockResolvedValue([]);
    vi.mocked(api.datasets.list).mockResolvedValue([]);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);
    const user = userEvent.setup();

    renderDashboard();

    await user.click(screen.getByRole('tab', { name: 'Evaluators' }));
    expect(routerState.replace).toHaveBeenCalledWith('/evals?view=evaluators');

    await user.click(screen.getByRole('tab', { name: 'Community & private' }));
    expect(routerState.replace).toHaveBeenCalledWith('/evals?view=library');
  });

  it('navigates via router.replace to the bare path when the overview tab is selected', async () => {
    setView('runs');
    vi.mocked(api.evaluators.list).mockResolvedValue([]);
    vi.mocked(api.datasets.list).mockResolvedValue([]);
    vi.mocked(useEvalHubPackages).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubPackages>);
    vi.mocked(useEvalHubRuns).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as unknown as ReturnType<typeof useEvalHubRuns>);
    const user = userEvent.setup();

    renderDashboard();
    expect(await screen.findByText('Runs surface')).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: 'Overview' }));
    expect(routerState.replace).toHaveBeenCalledWith('/evals');
  });
});
