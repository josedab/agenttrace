'use client';

import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, Brain, Clock, LockKeyhole, ShieldCheck } from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { API_URL } from '@/lib/api';
import type { SharedResourceView } from '@/lib/share-links';

export function SharedResourcePage({ token }: { token: string }) {
  const query = useQuery({
    queryKey: ['shared-resource', token],
    queryFn: async () => {
      const response = await fetch(`${API_URL}/api/share/${encodeURIComponent(token)}`, {
        cache: 'no-store',
      });
      if (!response.ok) {
        throw new Error('This link is invalid, expired, or revoked.');
      }
      return (await response.json()) as SharedResourceView;
    },
    retry: false,
  });

  return (
    <main className="min-h-screen bg-[#f3f1eb] px-4 py-10 text-[#171916] dark:bg-[#151714] dark:text-[#eceee8]">
      <div className="mx-auto max-w-5xl space-y-6">
        <header className="border-current/15 flex flex-wrap items-start justify-between gap-4 border-b pb-5">
          <div>
            <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              <LockKeyhole className="h-3.5 w-3.5" />
              Read-only redacted view
            </div>
            <h1 className="mt-3 text-3xl font-semibold tracking-tight">AgentTrace share</h1>
          </div>
          {query.data ? (
            <Badge variant="outline">
              Expires {new Date(query.data.expiresAt).toLocaleString()}
            </Badge>
          ) : null}
        </header>

        {query.isLoading ? (
          <div className="space-y-4">
            <Skeleton className="h-36 w-full" />
            <Skeleton className="h-80 w-full" />
          </div>
        ) : query.isError || !query.data ? (
          <Card className="border-destructive/30">
            <CardContent className="flex min-h-64 flex-col items-center justify-center text-center">
              <AlertTriangle className="h-9 w-9 text-destructive" />
              <p className="mt-4 font-medium">Shared view unavailable</p>
              <p className="mt-1 text-sm text-muted-foreground">
                {query.error?.message || 'This link is invalid, expired, or revoked.'}
              </p>
            </CardContent>
          </Card>
        ) : query.data.trace ? (
          <SharedTrace trace={query.data.trace} />
        ) : query.data.replayPlan ? (
          <SharedReplayPlan plan={query.data.replayPlan} />
        ) : null}

        <p className="text-center text-xs text-muted-foreground">
          Prompts, outputs, commands, file paths, diffs, credentials, and project identifiers are
          omitted server-side.
        </p>
      </div>
    </main>
  );
}

function SharedTrace({ trace }: { trace: NonNullable<SharedResourceView['trace']> }) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{trace.name}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-3">
          <SharedMetric label="Events" value={trace.summary.totalEvents} />
          <SharedMetric label="Model calls" value={trace.summary.llmCalls} />
          <SharedMetric label="Errors" value={trace.summary.errors} />
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Redacted timeline</CardTitle>
        </CardHeader>
        <CardContent className="divide-y">
          {trace.events.map((event, index) => (
            <div
              key={`${event.timestamp}-${index}`}
              className="grid gap-2 py-3 sm:grid-cols-[auto_minmax(0,1fr)_auto]"
            >
              <Clock className="mt-0.5 h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium">{event.title}</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  {event.type.replaceAll('_', ' ')}
                  {event.model ? ` · ${event.model}` : ''}
                </p>
              </div>
              <Badge variant={event.status === 'error' ? 'destructive' : 'outline'}>
                {event.status}
              </Badge>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

function SharedReplayPlan({ plan }: { plan: NonNullable<SharedResourceView['replayPlan']> }) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5" />
            Replay plan · {plan.status}
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <SharedMetric label="Recorded generations" value={plan.capabilities.generationCount} />
          <SharedMetric
            label="Sandbox execution"
            value={plan.capabilities.canExecuteInSandbox ? 'available' : 'unavailable'}
          />
        </CardContent>
      </Card>
      {plan.comparison ? (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Brain className="h-4 w-4" />
              Original vs replay
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-3">
              <SharedMetric
                label="Generations"
                value={`${plan.comparison.originalGenerationCount} → ${plan.comparison.replayGenerationCount}`}
              />
              <SharedMetric
                label="Tokens"
                value={`${plan.comparison.originalTokens} → ${plan.comparison.replayTokens}`}
              />
              <SharedMetric label="Verdict" value={plan.comparison.verdict.replaceAll('_', ' ')} />
            </div>
            <ul className="space-y-1 text-sm text-muted-foreground">
              {plan.comparison.notes.map((note) => (
                <li key={note}>• {note}</li>
              ))}
            </ul>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

function SharedMetric({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-[0.15em] text-muted-foreground">
        {label}
      </p>
      <p className="mt-1 text-xl font-semibold tabular-nums">{value}</p>
    </div>
  );
}
