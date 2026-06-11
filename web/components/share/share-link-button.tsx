'use client';

import * as React from 'react';
import { Check, Link2 } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useCreateShareLink } from '@/hooks/use-share-links';
import type { ShareResourceType } from '@/lib/share-links';

export function ShareLinkButton({
  resourceType,
  resourceId,
  variant = 'outline',
}: {
  resourceType: ShareResourceType;
  resourceId: string;
  variant?: 'default' | 'outline' | 'ghost';
}) {
  const [copied, setCopied] = React.useState(false);
  const [manualURL, setManualURL] = React.useState<string | null>(null);
  const createLink = useCreateShareLink(resourceType, resourceId);
  const resetTimer = React.useRef<ReturnType<typeof setTimeout> | null>(null);
  const fallbackFieldId = React.useId();

  React.useEffect(
    () => () => {
      if (resetTimer.current) {
        clearTimeout(resetTimer.current);
      }
    },
    []
  );

  const create = React.useCallback(async () => {
    let absoluteURL: string;
    try {
      const result = await createLink.mutateAsync(undefined);
      absoluteURL = result.url.startsWith('http')
        ? result.url
        : `${window.location.origin}${result.url}`;
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to create share link');
      return;
    }

    try {
      if (!navigator.clipboard?.writeText) {
        throw new Error('Clipboard access is unavailable');
      }
      await navigator.clipboard.writeText(absoluteURL);
      setManualURL(null);
      setCopied(true);
      toast.success('Redacted share link copied');
      resetTimer.current = setTimeout(() => setCopied(false), 2000);
    } catch {
      // The link exists even when the browser blocks clipboard access, so it is
      // shown for manual copying instead of being reported as a failure.
      setManualURL(absoluteURL);
      toast.warning('Share link created; copy it manually below.');
    }
  }, [createLink]);

  return (
    <div className="space-y-2">
      <Button
        type="button"
        size="sm"
        variant={variant}
        onClick={create}
        disabled={createLink.isPending}
        aria-busy={createLink.isPending}
      >
        {copied ? (
          <Check aria-hidden="true" className="mr-2 h-4 w-4" />
        ) : (
          <Link2 aria-hidden="true" className="mr-2 h-4 w-4" />
        )}
        {createLink.isPending ? 'Creating…' : copied ? 'Copied' : 'Share redacted view'}
      </Button>
      {manualURL ? (
        <div className="space-y-1" aria-live="polite">
          <Label htmlFor={fallbackFieldId}>Share link</Label>
          <Input
            id={fallbackFieldId}
            readOnly
            value={manualURL}
            onFocus={(e) => e.target.select()}
          />
        </div>
      ) : null}
    </div>
  );
}
