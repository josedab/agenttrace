import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { OutcomeDashboard } from '@/components/outcomes/outcome-dashboard';
import { useOutcomeDigest, useOutcomeOverview } from '@/hooks/use-outcomes';
import type { OutcomeDigest, OutcomeOverview } from '@/lib/outcomes';

vi.mock('@/hooks/use-outcomes', () => ({
  useOutcomeOverview: vi.fn(),
  useOutcomeDigest: vi.fn(),
}));

function overview(overrides: Partial<OutcomeOverview> = {}): OutcomeOverview {
  return {
    projectId: 'project-1',
    period: { from: '2026-06-25T00:00:00Z', to: '2026-07-25T00:00:00Z' },
    runs: {
      total: 20,
      successful: 18,
      failed: 2,
      inProgress: 0,
      successRate: { value: 0.9, available: true },
    },
    ci: {
      total: 10,
      passed: 9,
      failed: 1,
      cancelled: 0,
      inProgress: 0,
      linkedRuns: 10,
      passRate: { value: 0.9, available: true },
    },
    sourceControl: {
      linkedCommits: 8,
      linkedTraces: 15,
      linkedPullRequests: 4,
      regressionSignals: 0,
      revertSignals: 0,
    },
    cost: {
      totalCost: 12.5,
      costPerSuccessfulOutcome: { value: 0.69, available: true },
    },
    byAgent: [],
    byModel: [],
    recentOutcomes: [],
    availability: {
      traceData: true,
      gitData: true,
      ciData: true,
      pullRequestData: true,
      agentAttribution: true,
      modelAttribution: true,
      unavailable: [],
    },
    generatedAt: '2026-07-25T00:00:00Z',
    ...overrides,
  };
}

function mockOverviewQuery(overrides: Record<string, unknown> = {}) {
  const value = {
    data: undefined,
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
    ...overrides,
  };
  vi.mocked(useOutcomeOverview).mockReturnValue(value as unknown as ReturnType<
    typeof useOutcomeOverview
  >);
  return value;
}

function mockDigestQuery(overrides: Record<string, unknown> = {}) {
  const value = {
    data: undefined,
    isLoading: false,
    isError: false,
    ...overrides,
  };
  vi.mocked(useOutcomeDigest).mockReturnValue(value as unknown as ReturnType<
    typeof useOutcomeDigest
  >);
  return value;
}

afterEach(() => {
  vi.clearAllMocks();
});

describe('OutcomeDashboard', () => {
  it('shows a loading placeholder without metrics or the period switcher', () => {
    mockOverviewQuery({ isLoading: true });
    mockDigestQuery();

    render(<OutcomeDashboard />);

    expect(screen.queryByLabelText('Analytics period')).not.toBeInTheDocument();
    expect(screen.queryByText('Agent success')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Retry' })).not.toBeInTheDocument();
  });

  it('shows a retry action when the overview request fails, which calls refetch', async () => {
    const refetch = vi.fn();
    mockOverviewQuery({ isError: true, refetch });
    mockDigestQuery();
    const user = userEvent.setup();

    render(<OutcomeDashboard />);

    expect(screen.getByText('Outcome analytics could not be loaded')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /retry/i }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('renders real metrics and flags partial availability', () => {
    mockOverviewQuery({
      data: overview({
        ci: {
          total: 0,
          passed: 0,
          failed: 0,
          cancelled: 0,
          inProgress: 0,
          linkedRuns: 0,
          passRate: { value: null, available: false },
        },
        availability: {
          traceData: true,
          gitData: true,
          ciData: false,
          pullRequestData: false,
          agentAttribution: true,
          modelAttribution: true,
          unavailable: ['CI outcomes', 'Pull request outcomes'],
        },
      }),
    });
    mockDigestQuery();

    render(<OutcomeDashboard />);

    expect(screen.getByText('90.0%')).toBeInTheDocument(); // run success rate
    expect(screen.getByText('18 of 20 runs')).toBeInTheDocument();
    expect(screen.getByText('Partial outcome coverage')).toBeInTheDocument();
    expect(
      screen.getByText('Unavailable for this period: CI outcomes, Pull request outcomes.')
    ).toBeInTheDocument();
    // The CI pass rate card must show "Unavailable" rather than a stale/zero number.
    const ciCard = screen.getByText('CI pass rate').closest('div');
    expect(ciCard).not.toBeNull();
    expect(within(ciCard!.parentElement!).getByText('Unavailable')).toBeInTheDocument();
  });

  it('does not show the partial-coverage banner when every source is available', () => {
    mockOverviewQuery({ data: overview() });
    mockDigestQuery();

    render(<OutcomeDashboard />);

    expect(screen.queryByText('Partial outcome coverage')).not.toBeInTheDocument();
  });

  it('requests a new window from the overview hook when a period button is clicked', async () => {
    mockOverviewQuery({ data: overview() });
    mockDigestQuery();
    const user = userEvent.setup();

    render(<OutcomeDashboard />);

    expect(vi.mocked(useOutcomeOverview).mock.calls.at(-1)?.[0]).toBe('30d');

    await user.click(screen.getByRole('button', { name: '7d', pressed: false }));

    expect(vi.mocked(useOutcomeOverview).mock.calls.at(-1)?.[0]).toBe('7d');
    expect(screen.getByRole('button', { name: '7d' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: '30d' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('requests and renders a digest report only after "Generate report" is clicked', async () => {
    mockOverviewQuery({ data: overview() });
    mockDigestQuery();
    const user = userEvent.setup();

    render(<OutcomeDashboard />);

    expect(screen.queryByText('Weekly digest summary')).not.toBeInTheDocument();
    expect(vi.mocked(useOutcomeDigest).mock.calls.at(-1)?.[1]).toBe(false);

    mockDigestQuery({
      data: {
        projectId: 'project-1',
        period: { from: '2026-06-25T00:00:00Z', to: '2026-07-25T00:00:00Z' },
        title: 'Weekly digest summary',
        summary: 'Outcomes trended positively this period.',
        highlights: ['CI stayed green'],
        attention: [],
        generatedAt: '2026-07-25T00:00:00Z',
      } satisfies OutcomeDigest,
    });

    await user.click(screen.getByRole('button', { name: 'Generate report' }));

    expect(vi.mocked(useOutcomeDigest).mock.calls.at(-1)?.[1]).toBe(true);
    expect(await screen.findByText('Weekly digest summary')).toBeInTheDocument();
    expect(screen.getByText('Outcomes trended positively this period.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Generate report' })).toBeDisabled();
  });

  it('shows a digest-specific error without failing the whole dashboard', async () => {
    mockOverviewQuery({ data: overview() });
    mockDigestQuery({ isError: true });
    const user = userEvent.setup();

    render(<OutcomeDashboard />);
    await user.click(screen.getByRole('button', { name: 'Generate report' }));

    expect(screen.getByText('The digest could not be generated.')).toBeInTheDocument();
  });
});
