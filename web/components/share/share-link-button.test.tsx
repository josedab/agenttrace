import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ShareLinkButton } from '@/components/share/share-link-button';
import { shareLinksApi } from '@/lib/share-links';

vi.mock('@/lib/share-links', async () => {
  const actual = await vi.importActual<typeof import('@/lib/share-links')>('@/lib/share-links');
  return {
    ...actual,
    shareLinksApi: {
      ...actual.shareLinksApi,
      createTrace: vi.fn(),
    },
  };
});

const toastMock = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}));

vi.mock('sonner', () => ({ toast: toastMock }));

function renderButton() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <ShareLinkButton resourceType="trace" resourceId="trace-1" />
    </QueryClientProvider>
  );
}

afterEach(() => {
  vi.clearAllMocks();
});

describe('ShareLinkButton', () => {
  it('shows the created link for manual copying when the clipboard is blocked', async () => {
    vi.mocked(shareLinksApi.createTrace).mockResolvedValue({
      id: 'link-1',
      url: '/share/token-1',
      token: 'token-1',
      expiresAt: '2026-08-01T00:00:00Z',
      resourceType: 'trace',
      resourceId: 'trace-1',
      redactionVersion: 1,
      createdBy: 'user-1',
      createdAt: '2026-07-25T00:00:00Z',
    });
    const user = userEvent.setup();
    // userEvent installs its own clipboard stub, so the denial is applied after
    // setup to model a browser that blocks clipboard writes.
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
    });
    renderButton();

    await user.click(screen.getByRole('button', { name: /share redacted view/i }));

    const field = await screen.findByLabelText('Share link');
    expect(field).toHaveValue(`${window.location.origin}/share/token-1`);
    // The link was created, so the user must not be told that creation failed.
    expect(toastMock.error).not.toHaveBeenCalled();
    expect(toastMock.warning).toHaveBeenCalled();
  });

  it('reports a creation failure without a manual copy field', async () => {
    vi.mocked(shareLinksApi.createTrace).mockRejectedValue(
      new Error('share links are disabled')
    );

    const user = userEvent.setup();
    renderButton();

    await user.click(screen.getByRole('button', { name: /share redacted view/i }));

    await waitFor(() => expect(toastMock.error).toHaveBeenCalledWith('share links are disabled'));
    expect(screen.queryByLabelText('Share link')).not.toBeInTheDocument();
  });
});
