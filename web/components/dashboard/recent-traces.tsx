'use client';

import * as React from 'react';
import Link from 'next/link';
import { formatDistanceToNow } from 'date-fns';
import { ArrowRight, Clock, Layers } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface RecentTrace {
  id: string;
  name: string | null;
  input: unknown;
  output: unknown;
  startTime: string;
  latency: number | null;
  totalCost: number | null;
  level: 'DEBUG' | 'DEFAULT' | 'WARNING' | 'ERROR';
}

interface RecentTracesProps {
  traces: RecentTrace[];
}

export function RecentTraces({ traces }: RecentTracesProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Recent Traces</CardTitle>
          <CardDescription>Latest agent executions</CardDescription>
        </div>
        <Button variant="ghost" size="sm" asChild>
          <Link href="/traces" className="flex items-center gap-1">
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {traces.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <Layers className="mb-4 h-12 w-12 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">No traces yet</p>
              <p className="mt-1 text-xs text-muted-foreground">
                Start instrumenting your agents to see traces here
              </p>
            </div>
          ) : (
            traces.map((trace) => (
              <Link key={trace.id} href={`/traces/${trace.id}`} className="block">
                <div className="flex items-center justify-between rounded-lg border p-3 transition-colors hover:bg-muted/50">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="truncate text-sm font-medium">
                        {trace.name || trace.id.slice(0, 8)}
                      </p>
                      <LevelBadge level={trace.level} />
                    </div>
                    <p className="mt-1 truncate text-xs text-muted-foreground">
                      {trace.input ? truncate(trace.input, 60) : 'No input'}
                    </p>
                  </div>
                  <div className="ml-4 flex items-center gap-4">
                    {trace.latency && (
                      <div className="flex items-center gap-1 text-xs text-muted-foreground">
                        <Clock className="h-3 w-3" />
                        {formatLatency(trace.latency)}
                      </div>
                    )}
                    {trace.totalCost !== null && (
                      <div className="text-xs text-muted-foreground">
                        ${trace.totalCost.toFixed(4)}
                      </div>
                    )}
                    <div className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(trace.startTime), {
                        addSuffix: true,
                      })}
                    </div>
                  </div>
                </div>
              </Link>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function LevelBadge({ level }: { level: RecentTrace['level'] }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        'text-xs',
        level === 'ERROR' && 'border-destructive text-destructive',
        level === 'WARNING' && 'border-yellow-500 text-yellow-500',
        level === 'DEBUG' && 'border-blue-500 text-blue-500'
      )}
    >
      {level}
    </Badge>
  );
}

function truncate(value: unknown, length: number): string {
  const str = typeof value === 'string' ? value : (JSON.stringify(value) ?? String(value));
  if (str.length <= length) return str;
  return str.slice(0, length) + '...';
}

function formatLatency(ms: number): string {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + 's';
  }
  return ms.toFixed(0) + 'ms';
}
