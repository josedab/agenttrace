'use client';

import * as React from 'react';
import { Copy } from 'lucide-react';

import { Button } from '@/components/ui/button';

export interface OutcomeDigestContent {
  title: string;
  summary: string;
  highlights: string[];
  attention: string[];
}

/**
 * DigestPreview renders the canonical digest text and offers a clipboard copy.
 * Clipboard access is denied in several browsers and in insecure contexts, so a
 * failure is announced and the text stays selectable for manual copying.
 */
export function DigestPreview({ digest }: { digest: OutcomeDigestContent }) {
  const rendered = React.useMemo(
    () =>
      [
        digest.title,
        digest.summary,
        ...digest.highlights.map((item) => `• ${item}`),
        ...digest.attention.map((item) => `! ${item}`),
      ].join('\n'),
    [digest]
  );
  const [copyState, setCopyState] = React.useState<'idle' | 'copied' | 'failed'>('idle');

  const copyDigest = React.useCallback(async () => {
    try {
      if (!navigator.clipboard?.writeText) {
        throw new Error('Clipboard access is unavailable');
      }
      await navigator.clipboard.writeText(rendered);
      setCopyState('copied');
    } catch {
      // Clipboard access is denied in several browsers and in insecure
      // contexts; the digest text stays selectable so copying is still possible.
      setCopyState('failed');
    }
  }, [rendered]);

  return (
    <div className="relative rounded-lg border bg-muted/20 p-4">
      <Button
        size="sm"
        variant="ghost"
        className="absolute right-2 top-2"
        onClick={copyDigest}
        aria-label="Copy digest to clipboard"
      >
        <Copy aria-hidden="true" className="mr-2 h-4 w-4" />
        Copy
      </Button>
      <p aria-live="polite" className="sr-only">
        {copyState === 'copied'
          ? 'Digest copied to clipboard'
          : copyState === 'failed'
            ? 'Copying failed; select the digest text below to copy it manually'
            : ''}
      </p>
      {copyState === 'failed' ? (
        <p className="mb-2 pr-24 text-sm text-amber-700 dark:text-amber-300">
          Copying is blocked in this browser. Select the digest text below to copy it manually.
        </p>
      ) : null}
      <p className="pr-24 font-medium">{digest.title}</p>
      <p className="mt-2 text-sm text-muted-foreground">{digest.summary}</p>
      {digest.highlights.length > 0 ? (
        <ul className="mt-4 space-y-1 text-sm">
          {digest.highlights.map((item) => (
            <li key={item}>• {item}</li>
          ))}
        </ul>
      ) : null}
      {digest.attention.length > 0 ? (
        <ul className="mt-4 space-y-1 text-sm text-amber-700 dark:text-amber-300">
          {digest.attention.map((item) => (
            <li key={item}>! {item}</li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
