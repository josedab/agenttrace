'use client';

import * as React from 'react';
import {
  AlertTriangle,
  CheckCircle2,
  CircleDollarSign,
  GitCommitHorizontal,
  GitPullRequest,
  RefreshCw,
  ShieldCheck,
  Workflow,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { DigestPreview } from '@/components/outcomes/digest-preview';
import { useOutcomeDigest, useOutcomeOverview } from '@/hooks/use-outcomes';
import type {
  LinkedOutcome,
  OutcomeBreakdown,
  OutcomeOptionalMetric,
  OutcomeWindow,
} from '@/lib/outcomes';
import { cn } from '@/lib/utils';

const outcomeWindows: readonly OutcomeWindow[] = ['24h', '7d', '30d', '90d'];

export function OutcomeDashboard() {
  const [period, setPeriod] = React.useState<OutcomeWindow>('30d');
  const [showDigest, setShowDigest] = React.useState(false);
  const overviewQuery = useOutcomeOverview(period);
  const digestQuery = useOutcomeDigest(period, showDigest);

  if (overviewQuery.isLoading) {
    return <OutcomeDashboardSkeleton />;
  }

  if (overviewQuery.isError || !overviewQuery.data) {
    return (
      <Card className="border-destructive/30">
        <CardContent className="flex min-h-56 flex-col items-center justify-center gap-3 text-center">
          <AlertTriangle className="h-8 w-8 text-destructive" />
          <div>
            <p className="font-medium">Outcome analytics could not be loaded</p>
            <p className="mt-1 text-sm text-muted-foreground">
              The underlying trace, git, or CI store may be temporarily unavailable.
            </p>
          </div>
          <Button variant="outline" onClick={() => overviewQuery.refetch()}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  const overview = overviewQuery.data;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3 border-y py-3">
        <p className="text-sm text-muted-foreground">
          Real project data from traces, linked commits, and CI runs.
        </p>
        <div className="flex items-center gap-1" aria-label="Analytics period">
          {outcomeWindows.map((value) => (
            <Button
              key={value}
              type="button"
              size="sm"
              variant={period === value ? 'default' : 'ghost'}
              aria-pressed={period === value}
              onClick={() => {
                setPeriod(value);
                setShowDigest(false);
              }}
            >
              {value}
            </Button>
          ))}
        </div>
      </div>

      {overview.availability.unavailable.length > 0 ? (
        <div className="flex items-start gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-3">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-600" />
          <div>
            <p className="text-sm font-medium">Partial outcome coverage</p>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Unavailable for this period: {overview.availability.unavailable.join(', ')}.
            </p>
          </div>
        </div>
      ) : null}

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="Agent success"
          value={formatPercent(overview.runs.successRate)}
          detail={`${overview.runs.successful} of ${overview.runs.total} runs`}
          icon={CheckCircle2}
          available={overview.runs.successRate.available}
        />
        <MetricCard
          label="CI pass rate"
          value={formatPercent(overview.ci.passRate)}
          detail={`${overview.ci.passed} passed · ${overview.ci.failed} failed`}
          icon={ShieldCheck}
          available={overview.ci.passRate.available}
        />
        <MetricCard
          label="Cost / success"
          value={formatCurrencyMetric(overview.cost.costPerSuccessfulOutcome)}
          detail={`${formatCurrency(overview.cost.totalCost)} total spend`}
          icon={CircleDollarSign}
          available={overview.cost.costPerSuccessfulOutcome.available}
        />
        <MetricCard
          label="Regression signals"
          value={
            overview.availability.ciData
              ? overview.sourceControl.regressionSignals.toLocaleString()
              : 'Unavailable'
          }
          detail={`${overview.sourceControl.revertSignals} revert signal(s)`}
          icon={AlertTriangle}
          available={overview.availability.ciData}
          warning={overview.sourceControl.regressionSignals > 0}
        />
      </div>

      <OutcomeFlow
        runs={overview.runs.total}
        commits={overview.sourceControl.linkedCommits}
        ciRuns={overview.ci.linkedRuns}
        pullRequests={overview.sourceControl.linkedPullRequests}
      />

      <div className="grid gap-6 xl:grid-cols-2">
        <BreakdownCard
          title="By agent"
          items={overview.byAgent}
          unavailableLabel="No traces in this period included an agent_name attribution."
        />
        <BreakdownCard
          title="By model"
          items={overview.byModel}
          unavailableLabel="No generation observations in this period included a model."
        />
      </div>

      <RecentOutcomes
        outcomes={overview.recentOutcomes}
        gitAvailable={overview.availability.gitData}
      />

      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <div>
            <CardTitle className="text-base">Team digest</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">
              Channel-neutral report content for GitHub, Slack, Discord, or webhooks.
            </p>
          </div>
          <Button variant="outline" onClick={() => setShowDigest(true)} disabled={showDigest}>
            Generate report
          </Button>
        </CardHeader>
        {showDigest ? (
          <CardContent>
            {digestQuery.isLoading ? (
              <Skeleton className="h-36 w-full" />
            ) : digestQuery.isError || !digestQuery.data ? (
              <p className="text-sm text-destructive">The digest could not be generated.</p>
            ) : (
              <DigestPreview digest={digestQuery.data} />
            )}
          </CardContent>
        ) : null}
      </Card>
    </div>
  );
}

function MetricCard({
  label,
  value,
  detail,
  icon: Icon,
  available,
  warning = false,
}: {
  label: string;
  value: string;
  detail: string;
  icon: React.ComponentType<{ className?: string }>;
  available: boolean;
  warning?: boolean;
}) {
  return (
    <Card className={cn(warning && 'border-amber-500/40')}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">
          {label}
        </CardTitle>
        <Icon className={cn('h-4 w-4', warning ? 'text-amber-600' : 'text-muted-foreground')} />
      </CardHeader>
      <CardContent>
        <p
          className={cn(
            'text-2xl font-semibold tabular-nums',
            !available && 'text-muted-foreground'
          )}
        >
          {value}
        </p>
        <p className="mt-1 text-xs text-muted-foreground">{detail}</p>
      </CardContent>
    </Card>
  );
}

function OutcomeFlow({
  runs,
  commits,
  ciRuns,
  pullRequests,
}: {
  runs: number;
  commits: number;
  ciRuns: number;
  pullRequests: number;
}) {
  const stages = [
    { label: 'Agent runs', value: runs, icon: Workflow },
    { label: 'Linked commits', value: commits, icon: GitCommitHorizontal },
    { label: 'Linked CI runs', value: ciRuns, icon: ShieldCheck },
    { label: 'Pull requests', value: pullRequests, icon: GitPullRequest },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Outcome coverage</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid overflow-hidden rounded-lg border md:grid-cols-4">
          {stages.map((stage, index) => (
            <div
              key={stage.label}
              className={cn(
                'relative flex items-center gap-3 px-4 py-4',
                index > 0 && 'border-t md:border-l md:border-t-0'
              )}
            >
              <stage.icon className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="text-xl font-semibold tabular-nums">{stage.value.toLocaleString()}</p>
                <p className="text-xs text-muted-foreground">{stage.label}</p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function BreakdownCard({
  title,
  items,
  unavailableLabel,
}: {
  title: string;
  items: OutcomeBreakdown[];
  unavailableLabel: string;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {items.length === 0 ? (
          <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
            {unavailableLabel}
          </p>
        ) : (
          <div className="divide-y">
            {items.slice(0, 8).map((item) => (
              <div key={item.name} className="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-4 py-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{item.name}</p>
                  <p className="text-xs text-muted-foreground">{item.runs} runs</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium tabular-nums">
                    {formatPercent(item.successRate)}
                  </p>
                  <p className="text-xs text-muted-foreground">success</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium tabular-nums">
                    {formatCurrencyMetric(item.costPerSuccessfulOutcome)}
                  </p>
                  <p className="text-xs text-muted-foreground">per success</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function RecentOutcomes({
  outcomes,
  gitAvailable,
}: {
  outcomes: LinkedOutcome[];
  gitAvailable: boolean;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Recent linked outcomes</CardTitle>
      </CardHeader>
      <CardContent>
        {outcomes.length === 0 ? (
          <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
            {gitAvailable
              ? 'No linked commit reached a CI outcome in this period.'
              : 'Link traces to commits to unlock commit and pull request outcomes.'}
          </p>
        ) : (
          <div className="divide-y">
            {outcomes.map((outcome) => (
              <div
                key={outcome.commitSha}
                className="grid gap-3 py-3 md:grid-cols-[minmax(0,1fr)_auto_auto]"
              >
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">
                    {outcome.commitMessage || 'Commit without a recorded message'}
                  </p>
                  <p className="mt-1 font-mono text-xs text-muted-foreground">
                    {outcome.commitSha.slice(0, 10)}
                    {outcome.branch ? ` · ${outcome.branch}` : ''}
                  </p>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Workflow className="h-3.5 w-3.5" />
                  {outcome.traceCount} trace{outcome.traceCount === 1 ? '' : 's'}
                </div>
                <OutcomeStatus outcome={outcome} />
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function OutcomeStatus({ outcome }: { outcome: LinkedOutcome }) {
  if (!outcome.ciStatus) {
    return <Badge variant="outline">CI unavailable</Badge>;
  }

  const failed = outcome.ciStatus === 'failure';
  const label = outcome.prNumber
    ? `PR #${outcome.prNumber} · ${outcome.ciStatus}`
    : outcome.ciStatus;

  return (
    <Badge
      variant={failed ? 'destructive' : 'outline'}
      className="justify-self-start md:justify-self-end"
    >
      {label}
    </Badge>
  );
}

function OutcomeDashboardSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-12 w-full" />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className="h-32 w-full" />
        ))}
      </div>
      <Skeleton className="h-44 w-full" />
      <div className="grid gap-6 xl:grid-cols-2">
        <Skeleton className="h-72 w-full" />
        <Skeleton className="h-72 w-full" />
      </div>
    </div>
  );
}

function formatPercent(metric: OutcomeOptionalMetric) {
  return metric.available && metric.value !== null
    ? `${(metric.value * 100).toFixed(1)}%`
    : 'Unavailable';
}

function formatCurrencyMetric(metric: OutcomeOptionalMetric) {
  return metric.available && metric.value !== null ? formatCurrency(metric.value) : 'Unavailable';
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: value < 1 ? 4 : 2,
    maximumFractionDigits: value < 1 ? 4 : 2,
  }).format(value);
}
