import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { SharedResourcePage } from '@/components/share/shared-resource-page';
import { API_URL } from '@/lib/api';
import type { SharedResourceView } from '@/lib/share-links';

// Each test gets its own QueryClient so that the `['shared-resource', token]`
// cache from one test can never leak into, or mask a regression in, another.
function renderShared(token: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <SharedResourcePage token={token} />
    </QueryClientProvider>
  );
}

function jsonResponse(body: unknown, ok = true, status = 200) {
  return {
    ok,
    status,
    json: () => Promise.resolve(body),
  } as Response;
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe('SharedResourcePage', () => {
  it('shows a loading placeholder before the shared resource resolves', () => {
    vi.spyOn(global, 'fetch').mockReturnValue(new Promise(() => {}));

    renderShared('token-1');

    expect(screen.queryByText('Shared view unavailable')).not.toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'AgentTrace share' })).toBeInTheDocument();
  });

  it('shows one generic message for an invalid, expired, or revoked link, hiding server detail', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      jsonResponse({ error: 'token row not found in database' }, false, 404)
    );

    renderShared('token-1');

    expect(await screen.findByText('Shared view unavailable')).toBeInTheDocument();
    expect(screen.getByText('This link is invalid, expired, or revoked.')).toBeInTheDocument();
    expect(screen.queryByText(/token row not found/i)).not.toBeInTheDocument();
  });

  it('renders a redacted trace timeline for a trace share', async () => {
    const view: SharedResourceView = {
      resourceType: 'trace',
      expiresAt: '2026-08-01T00:00:00Z',
      trace: {
        traceId: 'trace-1',
        name: 'Checkout agent run',
        startTime: '2026-07-25T00:00:00Z',
        durationMs: 4200,
        summary: {
          totalEvents: 12,
          llmCalls: 3,
          toolCalls: 5,
          fileOperations: 0,
          terminalCommands: 0,
          checkpoints: 1,
          errors: 1,
          totalTokens: 900,
          totalCost: 0.12,
        },
        events: [
          {
            type: 'llm_call',
            timestamp: '2026-07-25T00:00:01Z',
            title: 'Draft response',
            status: 'success',
            model: 'gpt-5',
          },
          {
            type: 'tool_call',
            timestamp: '2026-07-25T00:00:02Z',
            title: 'Run tests',
            status: 'error',
          },
        ],
      },
    };
    vi.spyOn(global, 'fetch').mockResolvedValue(jsonResponse(view));

    renderShared('token-1');

    expect(await screen.findByText('Checkout agent run')).toBeInTheDocument();
    expect(screen.getByText('Draft response')).toBeInTheDocument();
    expect(screen.getByText('llm call · gpt-5')).toBeInTheDocument();
    expect(screen.getByText('Run tests')).toBeInTheDocument();
    expect(screen.getByText('error')).toBeInTheDocument();
    expect(screen.getByText('Redacted timeline')).toBeInTheDocument();
    expect(
      screen.getByText(/Prompts, outputs, commands, file paths, diffs, credentials/i)
    ).toBeInTheDocument();
  });

  it('renders an original-vs-replay comparison for a replay plan share', async () => {
    const view: SharedResourceView = {
      resourceType: 'replay_plan',
      expiresAt: '2026-08-01T00:00:00Z',
      replayPlan: {
        planId: 'plan-1',
        traceId: 'trace-1',
        status: 'completed',
        capabilities: {
          canInspectTimeline: true,
          canReplayRecordedGeneration: true,
          canExecuteInSandbox: false,
          generationCount: 4,
          unsupportedReasons: [],
          safetyNotice: 'Replays never call external tools.',
        },
        comparison: {
          originalGenerationCount: 4,
          replayGenerationCount: 4,
          originalTokens: 1200,
          replayTokens: 1180,
          originalCost: 0.2,
          replayProviderCost: 0.19,
          equivalent: true,
          verdict: 'equivalent_output',
          notes: ['Outputs matched on the final assistant turn.'],
        },
      },
    };
    vi.spyOn(global, 'fetch').mockResolvedValue(jsonResponse(view));

    renderShared('token-1');

    expect(await screen.findByText('Replay plan · completed')).toBeInTheDocument();
    expect(screen.getByText('Original vs replay')).toBeInTheDocument();
    expect(screen.getByText('4 → 4')).toBeInTheDocument();
    expect(screen.getByText('1200 → 1180')).toBeInTheDocument();
    expect(screen.getByText('equivalent output')).toBeInTheDocument();
    expect(screen.getByText(/Outputs matched on the final assistant turn\./)).toBeInTheDocument();
    expect(screen.getByText('unavailable')).toBeInTheDocument(); // sandbox execution
  });

  it('URL-encodes the token when requesting the shared resource', async () => {
    const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue(
      jsonResponse({ resourceType: 'trace', expiresAt: '2026-08-01T00:00:00Z' })
    );
    const rawToken = 'abc def/123+xyz';

    renderShared(rawToken);

    await screen.findByRole('heading', { name: 'AgentTrace share' });
    expect(fetchMock).toHaveBeenCalledWith(
      `${API_URL}/api/share/${encodeURIComponent(rawToken)}`,
      expect.objectContaining({ cache: 'no-store' })
    );
  });

  it('always fetches with cache: "no-store" so a revoked link cannot be served stale', async () => {
    const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue(
      jsonResponse({ resourceType: 'trace', expiresAt: '2026-08-01T00:00:00Z' })
    );

    renderShared('token-1');

    await screen.findByRole('heading', { name: 'AgentTrace share' });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, options] = fetchMock.mock.calls[0];
    expect(options).toMatchObject({ cache: 'no-store' });
  });
});
