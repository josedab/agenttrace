import { QueryClient, QueryClientProvider, type UseQueryResult } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

import { EvalHubDatasets, EvalHubEvaluators } from '@/components/eval-hub/eval-hub-assets';
import type { Dataset, Evaluator } from '@/lib/api';

// jsdom does not implement PointerEvent capture, which Radix Select relies on.
beforeAll(() => {
  Element.prototype.hasPointerCapture = Element.prototype.hasPointerCapture || (() => false);
  Element.prototype.setPointerCapture = Element.prototype.setPointerCapture || (() => {});
  Element.prototype.releasePointerCapture = Element.prototype.releasePointerCapture || (() => {});
  Element.prototype.scrollIntoView = Element.prototype.scrollIntoView || (() => {});
});

const toastMock = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }));
vi.mock('sonner', () => ({ toast: toastMock }));

const publishState = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
vi.mock('@/hooks/use-eval-hub', () => ({
  usePublishEvalHubPackage: () => publishState,
}));

vi.mock('@/components/evals/create-evaluator-dialog', () => ({
  CreateEvaluatorDialog: () => <button type="button">Create evaluator (stub)</button>,
}));
vi.mock('@/components/evals/evaluator-list', () => ({
  EvaluatorList: ({ evaluators }: { evaluators: Evaluator[] }) => (
    <ul aria-label="Evaluator results">
      {evaluators.map((evaluator) => (
        <li key={evaluator.id}>{evaluator.name}</li>
      ))}
    </ul>
  ),
}));
vi.mock('@/components/evals/evaluator-list-skeleton', () => ({
  EvaluatorListSkeleton: () => <p role="status">Loading evaluators</p>,
}));
vi.mock('@/components/datasets/create-dataset-dialog', () => ({
  CreateDatasetDialog: () => <button type="button">Create dataset (stub)</button>,
}));

function renderWith(node: React.ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={queryClient}>{node}</QueryClientProvider>);
}

function queryResult<T>(overrides: Partial<UseQueryResult<T, Error>>): UseQueryResult<T, Error> {
  return {
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
    ...overrides,
  } as unknown as UseQueryResult<T, Error>;
}

afterEach(() => {
  vi.clearAllMocks();
  publishState.isPending = false;
});

describe('EvalHubEvaluators', () => {
  it('shows a loading skeleton while evaluators are fetched', () => {
    renderWith(<EvalHubEvaluators query={queryResult<Evaluator[]>({ isLoading: true })} />);

    expect(screen.getByRole('status')).toHaveTextContent('Loading evaluators');
  });

  it('offers a retry when evaluators fail to load', async () => {
    const refetch = vi.fn();
    const user = userEvent.setup();
    renderWith(
      <EvalHubEvaluators
        query={queryResult<Evaluator[]>({ isError: true, error: new Error('boom'), refetch })}
      />
    );

    expect(screen.getByText('Evaluators could not be loaded.')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Retry' }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('renders an explicit empty state with a create action', () => {
    renderWith(<EvalHubEvaluators query={queryResult<Evaluator[]>({ data: [] })} />);

    expect(screen.getByText('No evaluators yet')).toBeInTheDocument();
    // One "create" affordance lives in the header and a second in the empty-state action.
    expect(screen.getAllByRole('button', { name: 'Create evaluator (stub)' })).toHaveLength(2);
  });

  it('renders the evaluator list once data resolves', () => {
    const evaluators = [
      { id: 'eval-1', name: 'Correctness judge' } as Evaluator,
      { id: 'eval-2', name: 'Latency guard' } as Evaluator,
    ];
    renderWith(<EvalHubEvaluators query={queryResult<Evaluator[]>({ data: evaluators })} />);

    expect(screen.getByRole('list', { name: 'Evaluator results' })).toBeInTheDocument();
    expect(screen.getByText('Correctness judge')).toBeInTheDocument();
    expect(screen.getByText('Latency guard')).toBeInTheDocument();
  });
});

const datasetFixture: Dataset = {
  id: 'dataset-1',
  name: 'Support regressions',
  description: 'Curated tickets',
  itemCount: 12,
  runCount: 3,
} as Dataset;

describe('EvalHubDatasets', () => {
  it('shows a skeleton grid while datasets are fetched, not the empty or list states', () => {
    renderWith(<EvalHubDatasets query={queryResult<Dataset[]>({ isLoading: true })} />);

    expect(screen.queryByText('No datasets yet')).not.toBeInTheDocument();
    expect(screen.queryByText('Support regressions')).not.toBeInTheDocument();
    expect(screen.queryByText('Datasets could not be loaded.')).not.toBeInTheDocument();
  });

  it('offers a retry when datasets fail to load', async () => {
    const refetch = vi.fn();
    const user = userEvent.setup();
    renderWith(
      <EvalHubDatasets
        query={queryResult<Dataset[]>({ isError: true, error: new Error('boom'), refetch })}
      />
    );

    expect(screen.getByText('Datasets could not be loaded.')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Retry' }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('renders an empty state and disables publishing with no datasets', () => {
    renderWith(<EvalHubDatasets query={queryResult<Dataset[]>({ data: [] })} />);

    expect(screen.getByText('No datasets yet')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /publish/i })).toBeDisabled();
  });

  it('links to each dataset and enables publishing when data resolves', () => {
    renderWith(<EvalHubDatasets query={queryResult<Dataset[]>({ data: [datasetFixture] })} />);

    expect(screen.getByText('Support regressions')).toBeInTheDocument();
    expect(screen.getByText('12 items · 3 runs')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /open/i })).toHaveAttribute(
      'href',
      '/datasets/dataset-1'
    );
    expect(screen.getByRole('button', { name: /publish/i })).toBeEnabled();
  });

  it('publishes the selected dataset with the chosen visibility', async () => {
    const user = userEvent.setup();
    publishState.mutate.mockImplementation((_input, options) => {
      options?.onSuccess?.();
    });
    renderWith(<EvalHubDatasets query={queryResult<Dataset[]>({ data: [datasetFixture] })} />);

    await user.click(screen.getByRole('button', { name: /publish/i }));
    await user.click(screen.getByRole('combobox', { name: /dataset/i }));
    await user.click(await screen.findByRole('option', { name: 'Support regressions' }));
    await user.click(screen.getByRole('combobox', { name: /visibility/i }));
    await user.click(await screen.findByRole('option', { name: 'Organization' }));
    await user.click(screen.getByRole('button', { name: 'Publish version' }));

    expect(publishState.mutate).toHaveBeenCalledWith(
      expect.objectContaining({
        kind: 'dataset',
        sourceResourceId: 'dataset-1',
        visibility: 'organization',
      }),
      expect.anything()
    );
    await waitFor(() => expect(toastMock.success).toHaveBeenCalledWith('Dataset package published'));
    // The dialog closes after a successful publish.
    expect(screen.queryByRole('button', { name: 'Publish version' })).not.toBeInTheDocument();
  });

  it('reports a publish failure without closing the dialog', async () => {
    const user = userEvent.setup();
    publishState.mutate.mockImplementation((_input, options) => {
      options?.onError?.(new Error('quota exceeded'));
    });
    renderWith(<EvalHubDatasets query={queryResult<Dataset[]>({ data: [datasetFixture] })} />);

    await user.click(screen.getByRole('button', { name: /publish/i }));
    await user.click(screen.getByRole('combobox', { name: /dataset/i }));
    await user.click(await screen.findByRole('option', { name: 'Support regressions' }));
    await user.click(screen.getByRole('button', { name: 'Publish version' }));

    await waitFor(() => expect(toastMock.error).toHaveBeenCalledWith('quota exceeded'));
    expect(screen.getByRole('button', { name: 'Publish version' })).toBeInTheDocument();
  });
});
