import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { DigestPreview } from '@/components/outcomes/digest-preview';

const digest = {
  title: 'Agent outcome digest',
  summary: 'Eight successful runs.',
  highlights: ['CI passed on main'],
  attention: ['Two reverted commits'],
};

afterEach(() => {
  vi.clearAllMocks();
});

describe('DigestPreview', () => {
  it('announces a successful copy without visual-only feedback', async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    render(<DigestPreview digest={digest} />);

    await user.click(screen.getByRole('button', { name: 'Copy digest to clipboard' }));

    expect(writeText).toHaveBeenCalledWith(
      'Agent outcome digest\nEight successful runs.\n• CI passed on main\n! Two reverted commits'
    );
    expect(await screen.findByText('Digest copied to clipboard')).toBeInTheDocument();
  });

  it('explains a blocked clipboard instead of failing silently', async () => {
    const user = userEvent.setup();
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
    });
    render(<DigestPreview digest={digest} />);

    await user.click(screen.getByRole('button', { name: 'Copy digest to clipboard' }));

    expect(await screen.findByText(/Copying is blocked in this browser/i)).toBeInTheDocument();
    // The digest text remains available for manual selection.
    expect(screen.getByText('Agent outcome digest')).toBeInTheDocument();
  });

  it('handles browsers without clipboard support', async () => {
    const user = userEvent.setup();
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: undefined });
    render(<DigestPreview digest={digest} />);

    await user.click(screen.getByRole('button', { name: 'Copy digest to clipboard' }));

    expect(await screen.findByText(/Copying is blocked in this browser/i)).toBeInTheDocument();
  });
});
