'use client';

import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  Clock,
  GitBranch,
  Play,
  ShieldAlert,
  TerminalSquare,
} from 'lucide-react';

import { ReplayTimeline } from '@/components/replay/replay-timeline';
import { ShareLinkButton } from '@/components/share/share-link-button';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import {
  useCreateReplayPlan,
  useExecuteReplayPlan,
  useReplayCapabilities,
} from '@/hooks/use-replay-plans';
import { api, type Checkpoint, type Trace } from '@/lib/api';
import type { ReplayExecutionMode, ReplayPlan, ReplayPlanComparison } from '@/lib/replay-plans';

export function ReplayDashboard() {
  const searchParams = useSearchParams();
  const traceId = searchParams.get('traceId');

  if (!traceId) {
    return <TraceReplayPicker />;
  }

  return <TraceReplayDebugger traceId={traceId} />;
}

function TraceReplayPicker() {
  const router = useRouter();
  const tracesQuery = useQuery({
    queryKey: ['replay-trace-picker'],
    queryFn: () => api.traces.list({ limit: 20, offset: 0 }),
  });

  if (tracesQuery.isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className="h-24 w-full" />
        ))}
      </div>
    );
  }

  if (tracesQuery.isError) {
    return (
      <Card className="border-destructive/30">
        <CardContent className="flex min-h-48 flex-col items-center justify-center gap-3 text-center">
          <AlertTriangle className="h-7 w-7 text-destructive" />
          <p className="font-medium">Recent traces could not be loaded.</p>
          <Button variant="outline" onClick={() => tracesQuery.refetch()}>
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  const traces = tracesQuery.data?.traces ?? [];
  if (traces.length === 0) {
    return (
      <Card>
        <CardContent className="flex min-h-48 flex-col items-center justify-center text-center">
          <Play className="h-8 w-8 text-muted-foreground" />
          <p className="mt-3 font-medium">No traces are available to replay</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Ingest an agent trace first; replay never substitutes demo events.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      <p className="text-sm text-muted-foreground">
        Choose a real trace to inspect its timeline and replay prerequisites.
      </p>
      {traces.map((trace: Trace) => (
        <button
          key={trace.id}
          type="button"
          onClick={() => router.push(`/replay?traceId=${encodeURIComponent(trace.id)}`)}
          className="flex w-full items-center justify-between gap-4 rounded-xl border bg-card px-4 py-4 text-left transition-colors hover:border-foreground/30 hover:bg-muted/30"
        >
          <div className="min-w-0">
            <p className="truncate font-medium">{trace.name || 'Unnamed trace'}</p>
            <p className="mt-1 font-mono text-xs text-muted-foreground">{trace.id}</p>
          </div>
          <div className="flex shrink-0 items-center gap-3 text-xs text-muted-foreground">
            <Clock className="h-3.5 w-3.5" />
            {new Date(trace.startTime).toLocaleString()}
            <Play className="h-4 w-4 text-foreground" />
          </div>
        </button>
      ))}
    </div>
  );
}

function TraceReplayDebugger({ traceId }: { traceId: string }) {
  const router = useRouter();

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3 border-y py-3">
        <Button variant="ghost" size="sm" onClick={() => router.push('/replay')}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Choose another trace
        </Button>
        <div className="flex max-w-full items-center gap-3">
          <span className="truncate font-mono text-xs text-muted-foreground">{traceId}</span>
          <ShareLinkButton resourceType="trace" resourceId={traceId} />
        </div>
      </div>

      <ReplayPlanPanel traceId={traceId} />
      <ReplayTimeline traceId={traceId} />
    </div>
  );
}

function ReplayPlanPanel({ traceId }: { traceId: string }) {
  const [checkpointId, setCheckpointId] = React.useState<string>('origin');
  const [mode, setMode] = React.useState<ReplayExecutionMode>('recorded_generation');
  const [plan, setPlan] = React.useState<ReplayPlan | null>(null);
  const selectedCheckpoint = checkpointId === 'origin' ? undefined : checkpointId;

  const checkpointsQuery = useQuery({
    queryKey: ['replay-checkpoints', traceId],
    queryFn: () =>
      api
        .get<{
          data: Checkpoint[];
        }>(`/api/public/traces/${encodeURIComponent(traceId)}/checkpoints`)
        .then((response) => response.data),
  });
  const capabilityQuery = useReplayCapabilities(traceId, selectedCheckpoint, mode);
  const createPlan = useCreateReplayPlan(traceId);
  const executePlan = useExecuteReplayPlan();

  const create = async () => {
    const created = await createPlan.mutateAsync({
      mode,
      checkpointId: selectedCheckpoint,
    });
    setPlan(created);
  };

  const execute = async () => {
    if (!plan) return;
    const completed = await executePlan.mutateAsync(plan.id);
    setPlan(completed);
  };

  const capabilities = plan?.capabilities ?? capabilityQuery.data;
  const comparison = plan?.result?.comparison;

  return (
    <Card className="overflow-hidden">
      <CardHeader className="border-b bg-muted/20">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <CardTitle className="text-base">Safe replay plan</CardTitle>
            <p className="mt-1 max-w-2xl text-sm text-muted-foreground">
              Select a checkpoint and inspect exactly what AgentTrace can replay. The API host never
              runs captured commands, tools, or file writes.
            </p>
          </div>
          {plan ? <ReplayPlanStatus status={plan.status} /> : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-5 pt-5">
        <div className="grid gap-4 md:grid-cols-2">
          <label className="space-y-2 text-sm font-medium">
            Checkpoint
            <Select value={checkpointId} onValueChange={setCheckpointId}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="origin">Trace origin</SelectItem>
                {(checkpointsQuery.data ?? []).map((checkpoint) => (
                  <SelectItem key={checkpoint.id} value={checkpoint.id}>
                    {checkpoint.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </label>

          <label className="space-y-2 text-sm font-medium">
            Execution mode
            <Select value={mode} onValueChange={(value) => setMode(value as ReplayExecutionMode)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="recorded_generation">Recorded generation replay</SelectItem>
                <SelectItem value="sandbox">Sandbox re-execution</SelectItem>
              </SelectContent>
            </Select>
          </label>
        </div>

        {capabilityQuery.isLoading ? (
          <Skeleton className="h-24 w-full" />
        ) : capabilityQuery.isError || !capabilities ? (
          <p className="rounded-lg border border-destructive/30 p-4 text-sm text-destructive">
            Replay prerequisites could not be assessed.
          </p>
        ) : (
          <CapabilitySummary capabilities={capabilities} />
        )}

        <div className="flex flex-wrap items-center gap-3">
          <Button
            onClick={create}
            disabled={
              createPlan.isPending ||
              capabilityQuery.isLoading ||
              !capabilities ||
              plan?.status === 'running'
            }
          >
            {createPlan.isPending ? 'Creating…' : 'Create replay plan'}
          </Button>
          {plan?.status === 'ready' ? (
            <Button variant="outline" onClick={execute} disabled={executePlan.isPending}>
              <Play className="mr-2 h-4 w-4" />
              {executePlan.isPending ? 'Replaying…' : 'Replay recorded generations'}
            </Button>
          ) : null}
          {plan?.status === 'completed' ? (
            <ShareLinkButton resourceType="replay_plan" resourceId={plan.id} />
          ) : null}
          {createPlan.isError || executePlan.isError ? (
            <p className="text-sm text-destructive">
              {createPlan.error?.message || executePlan.error?.message}
            </p>
          ) : null}
        </div>

        {plan?.status === 'unsupported' ? (
          <p className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-4 text-sm">
            This request was saved as unsupported. Configure the listed prerequisites before
            attempting re-execution.
          </p>
        ) : null}
        {comparison ? <ReplayComparison comparison={comparison} /> : null}
      </CardContent>
    </Card>
  );
}

function CapabilitySummary({
  capabilities,
}: {
  capabilities: {
    canReplayRecordedGeneration: boolean;
    canExecuteInSandbox: boolean;
    hasCheckpoint: boolean;
    hasFileOperations: boolean;
    hasTerminalCommands: boolean;
    generationCount: number;
    unsupportedReasons: string[];
    safetyNotice: string;
  };
}) {
  const items = [
    {
      label: `${capabilities.generationCount} recorded generation(s)`,
      available: capabilities.canReplayRecordedGeneration,
      icon: Play,
    },
    {
      label: capabilities.hasCheckpoint ? 'Checkpoint selected' : 'Replay from trace origin',
      available: true,
      icon: GitBranch,
    },
    {
      label: capabilities.canExecuteInSandbox ? 'Sandbox configured' : 'Sandbox unavailable',
      available: capabilities.canExecuteInSandbox,
      icon: ShieldAlert,
    },
    {
      label:
        capabilities.hasTerminalCommands || capabilities.hasFileOperations
          ? 'Captured side effects remain inspect-only'
          : 'No captured side effects',
      available: true,
      icon: TerminalSquare,
    },
  ];

  return (
    <div className="space-y-3 rounded-lg border p-4">
      <div className="grid gap-3 md:grid-cols-2">
        {items.map((item) => (
          <div key={item.label} className="flex items-center gap-2 text-sm">
            <item.icon
              className={`h-4 w-4 ${item.available ? 'text-emerald-600' : 'text-amber-600'}`}
            />
            {item.label}
          </div>
        ))}
      </div>
      <p className="text-xs text-muted-foreground">{capabilities.safetyNotice}</p>
      {capabilities.unsupportedReasons.length > 0 ? (
        <ul className="space-y-1 text-sm text-amber-700 dark:text-amber-300">
          {capabilities.unsupportedReasons.map((reason) => (
            <li key={reason}>• {reason}</li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

function ReplayPlanStatus({ status }: { status: ReplayPlan['status'] }) {
  const completed = status === 'completed';
  const unsupported = status === 'unsupported' || status === 'failed';
  return (
    <Badge variant={unsupported ? 'destructive' : 'outline'}>
      {completed ? <CheckCircle2 className="mr-1 h-3 w-3" /> : null}
      {status}
    </Badge>
  );
}

function ReplayComparison({ comparison }: { comparison: ReplayPlanComparison }) {
  return (
    <div className="space-y-4 rounded-lg border border-emerald-500/30 bg-emerald-500/5 p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="font-medium">Original vs replay branch</p>
          <p className="text-sm text-muted-foreground">
            Deterministic output hashes matched; side-effecting events were not executed.
          </p>
        </div>
        <Badge variant="outline">{comparison.verdict.replaceAll('_', ' ')}</Badge>
      </div>
      <div className="grid gap-3 text-sm sm:grid-cols-3">
        <ComparisonValue
          label="Generations"
          original={comparison.originalGenerationCount}
          replay={comparison.replayGenerationCount}
        />
        <ComparisonValue
          label="Tokens"
          original={comparison.originalTokens}
          replay={comparison.replayTokens}
        />
        <ComparisonValue
          label="Provider cost"
          original={`$${comparison.originalCost.toFixed(4)}`}
          replay={`$${comparison.replayProviderCost.toFixed(4)}`}
        />
      </div>
      <ul className="space-y-1 text-xs text-muted-foreground">
        {comparison.notes.map((note) => (
          <li key={note}>• {note}</li>
        ))}
      </ul>
    </div>
  );
}

function ComparisonValue({
  label,
  original,
  replay,
}: {
  label: string;
  original: string | number;
  replay: string | number;
}) {
  return (
    <div className="rounded-md bg-background/80 p-3">
      <p className="text-xs uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className="mt-1 tabular-nums">
        {original} <span className="text-muted-foreground">→</span> {replay}
      </p>
    </div>
  );
}
