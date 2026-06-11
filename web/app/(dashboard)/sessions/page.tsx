'use client';

import { formatDistanceToNow } from 'date-fns';
import {
  Activity,
  AlertCircle,
  Bookmark,
  Clock3,
  Coins,
  type LucideIcon,
  Workflow,
} from 'lucide-react';

import { PageHeader } from '@/components/layout/page-header';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useTraceSessions } from '@/hooks/use-traces';

export default function SessionsPage() {
  const { data: sessions = [], isLoading, error } = useTraceSessions(100);
  const traceCount = sessions.reduce((total, session) => total + session.traceCount, 0);
  const totalCost = sessions.reduce((total, session) => total + session.totalCost, 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Sessions"
        description="Review the 100 most recent grouped agent execution contexts"
        icon={Workflow}
      />

      <div className="grid gap-4 sm:grid-cols-3">
        <SummaryCard
          label="Recent sessions"
          value={sessions.length.toLocaleString()}
          icon={Workflow}
        />
        <SummaryCard label="Recent traces" value={traceCount.toLocaleString()} icon={Activity} />
        <SummaryCard label="Recent cost" value={formatCost(totalCost)} icon={Coins} />
      </div>

      {isLoading ? (
        <SessionsLoading />
      ) : error ? (
        <Card className="border-destructive/40">
          <CardContent className="flex items-center gap-3 py-8 text-destructive">
            <AlertCircle className="h-5 w-5" />
            <p>Sessions could not be loaded.</p>
          </CardContent>
        </Card>
      ) : sessions.length === 0 ? (
        <Card className="overflow-hidden">
          <CardContent className="relative flex min-h-64 flex-col items-center justify-center text-center">
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,hsl(var(--muted))_0,transparent_65%)] opacity-70" />
            <div className="relative flex max-w-md flex-col items-center">
              <div className="mb-4 rounded-2xl border bg-background p-4 shadow-sm">
                <Workflow className="h-8 w-8 text-muted-foreground" />
              </div>
              <h2 className="text-lg font-semibold">No sessions yet</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Add a session ID to related traces to see multi-step agent work grouped here.
              </p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 xl:grid-cols-2">
          {sessions.map((session) => (
            <Card
              key={session.id}
              className="group overflow-hidden transition-colors hover:border-primary/40"
            >
              <CardHeader className="border-b bg-muted/25 pb-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="min-w-0">
                    <CardTitle className="truncate font-mono text-base">{session.id}</CardTitle>
                    <CardDescription className="mt-2 flex items-center gap-1.5">
                      <Clock3 className="h-3.5 w-3.5" />
                      Last active {relativeTime(session.lastTraceTime)}
                    </CardDescription>
                  </div>
                  <div className="flex items-center gap-2">
                    {session.bookmarked && (
                      <Bookmark className="h-4 w-4 fill-current text-primary" />
                    )}
                    <Badge variant={session.public ? 'secondary' : 'outline'}>
                      {session.public ? 'Public' : 'Private'}
                    </Badge>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="grid grid-cols-3 gap-4 pt-5">
                <SessionMetric label="Traces" value={session.traceCount.toLocaleString()} />
                <SessionMetric label="Tokens" value={session.totalTokens.toLocaleString()} />
                <SessionMetric label="Cost" value={formatCost(session.totalCost)} />
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

function SummaryCard({
  label,
  value,
  icon: Icon,
}: {
  label: string;
  value: string;
  icon: LucideIcon;
}) {
  return (
    <Card>
      <CardContent className="flex items-center justify-between p-5">
        <div>
          <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">
            {label}
          </p>
          <p className="mt-2 text-2xl font-semibold tabular-nums">{value}</p>
        </div>
        <div className="rounded-xl bg-muted p-3">
          <Icon className="h-5 w-5 text-muted-foreground" />
        </div>
      </CardContent>
    </Card>
  );
}

function SessionMetric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 font-medium tabular-nums">{value}</p>
    </div>
  );
}

function SessionsLoading() {
  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {Array.from({ length: 4 }, (_, index) => (
        <Card key={index}>
          <CardHeader>
            <Skeleton className="h-5 w-56" />
            <Skeleton className="h-4 w-36" />
          </CardHeader>
          <CardContent className="grid grid-cols-3 gap-4">
            <Skeleton className="h-10" />
            <Skeleton className="h-10" />
            <Skeleton className="h-10" />
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function relativeTime(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'unknown' : formatDistanceToNow(date, { addSuffix: true });
}

function formatCost(value: number) {
  return value < 0.01 && value > 0 ? `$${value.toFixed(4)}` : `$${value.toFixed(2)}`;
}
