'use client';

import type { UseQueryResult } from '@tanstack/react-query';
import { Play } from 'lucide-react';

import {
  EvalHubEmptyState,
  EvalHubErrorState,
  EvalHubGridSkeleton,
} from '@/components/eval-hub/eval-hub-states';
import { Badge } from '@/components/ui/badge';
import type { EvalHubRunList } from '@/lib/eval-hub';

export function EvalHubRuns({ query }: { query: UseQueryResult<EvalHubRunList, Error> }) {
  if (query.isLoading) {
    return <EvalHubGridSkeleton />;
  }
  if (query.isError) {
    return (
      <EvalHubErrorState
        label="Eval Hub runs could not be loaded."
        onRetry={() => query.refetch()}
      />
    );
  }
  const runs = query.data?.runs ?? [];
  if (runs.length === 0) {
    return (
      <EvalHubEmptyState
        icon={Play}
        title="No Eval Hub runs yet"
        description="Run a project-owned package from the Community & private tab."
      />
    );
  }

  return (
    <div className="divide-y rounded-xl border bg-card">
      {runs.map((run) => (
        <div key={run.id} className="grid gap-3 px-4 py-4 md:grid-cols-[minmax(0,1fr)_auto_auto]">
          <div className="min-w-0">
            <p className="truncate font-mono text-sm">{run.id}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Package {run.packageId.slice(0, 8)} · v{run.packageVersion}
            </p>
          </div>
          <Badge variant={run.status === 'failed' ? 'destructive' : 'outline'}>{run.status}</Badge>
          <p className="text-xs text-muted-foreground md:text-right">
            {new Date(run.startedAt).toLocaleString()}
          </p>
          {run.capabilityMessage ? (
            <p className="text-sm text-muted-foreground md:col-span-3">{run.capabilityMessage}</p>
          ) : null}
        </div>
      ))}
    </div>
  );
}
