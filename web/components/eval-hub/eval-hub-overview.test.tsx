import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { EvalHubOverview } from '@/components/eval-hub/eval-hub-overview';

describe('EvalHubOverview', () => {
  it('shows placeholders instead of counts while loading', () => {
    render(
      <EvalHubOverview
        evaluatorCount={undefined}
        datasetCount={undefined}
        packageCount={undefined}
        runCount={undefined}
        loading
        onNavigate={vi.fn()}
      />
    );

    // Real counts must not render as "0" while the underlying queries are still loading.
    expect(screen.queryByText('0')).not.toBeInTheDocument();
    expect(screen.getByText('Evaluators')).toBeInTheDocument();
    expect(screen.getByText('Datasets')).toBeInTheDocument();
  });

  it('renders resolved counts for each surface', () => {
    render(
      <EvalHubOverview
        evaluatorCount={4}
        datasetCount={12}
        packageCount={7}
        runCount={0}
        loading={false}
        onNavigate={vi.fn()}
      />
    );

    expect(screen.getByText('4')).toBeInTheDocument();
    expect(screen.getByText('12')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    // A resolved zero must still render as "0", distinct from the loading placeholder.
    expect(screen.getByText('0')).toBeInTheDocument();
  });

  it('invokes the navigation callback with the target view for each card', async () => {
    const onNavigate = vi.fn();
    const user = userEvent.setup();
    render(
      <EvalHubOverview
        evaluatorCount={1}
        datasetCount={2}
        packageCount={3}
        runCount={4}
        loading={false}
        onNavigate={onNavigate}
      />
    );

    await user.click(screen.getByRole('button', { name: /evaluators/i }));
    expect(onNavigate).toHaveBeenNthCalledWith(1, 'evaluators');

    await user.click(screen.getByRole('button', { name: /datasets/i }));
    expect(onNavigate).toHaveBeenNthCalledWith(2, 'datasets');

    await user.click(screen.getByRole('button', { name: /packages/i }));
    expect(onNavigate).toHaveBeenNthCalledWith(3, 'library');

    await user.click(screen.getByRole('button', { name: /runs/i }));
    expect(onNavigate).toHaveBeenNthCalledWith(4, 'runs');

    expect(onNavigate).toHaveBeenCalledTimes(4);
  });
});
