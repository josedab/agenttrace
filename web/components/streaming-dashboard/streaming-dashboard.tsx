'use client';

import * as React from 'react';
import { Activity, DollarSign, Hash, AlertTriangle, Zap, Clock, Cpu, Radio } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useStreamingDashboard } from '@/hooks/use-streaming-dashboard';

const statusConfig = {
  streaming: 'bg-green-500/10 text-green-700 border-green-500/20',
  processing: 'bg-blue-500/10 text-blue-700 border-blue-500/20',
  completing: 'bg-yellow-500/10 text-yellow-700 border-yellow-500/20',
};

export function StreamingDashboard() {
  const { data: metrics, isLoading } = useStreamingDashboard();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid gap-4 md:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-28 animate-pulse rounded-lg bg-muted" />
          ))}
        </div>
        <div className="h-64 animate-pulse rounded-lg bg-muted" />
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="py-12 text-center text-muted-foreground">
        <Radio className="mx-auto mb-4 h-12 w-12 opacity-50" />
        <p className="text-lg font-medium">Live Dashboard</p>
        <p className="mt-2 text-sm">Metrics will appear here when streaming sessions are active</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Sessions
            </CardTitle>
            <Activity className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.activeSessions}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Cost</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${metrics.totalCost.toFixed(4)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Tokens
            </CardTitle>
            <Hash className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.totalTokens.toLocaleString()}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Error Count</CardTitle>
            <AlertTriangle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{metrics.errorCount}</div>
          </CardContent>
        </Card>
      </div>

      {metrics.activeStreams.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Zap className="h-5 w-5 text-yellow-500" />
              Active Streams
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {metrics.activeStreams.map((stream) => (
              <div
                key={stream.id}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <div className="relative h-2 w-2">
                    <div className="absolute h-2 w-2 animate-ping rounded-full bg-green-500" />
                    <div className="absolute h-2 w-2 rounded-full bg-green-500" />
                  </div>
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{stream.id}</p>
                    <div className="mt-1 flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {stream.model}
                      </Badge>
                      <Badge
                        variant="outline"
                        className={cn('text-xs', statusConfig[stream.status])}
                      >
                        {stream.status}
                      </Badge>
                    </div>
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-4 text-sm text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    <span>{stream.elapsedSeconds}s</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Cpu className="h-3 w-3" />
                    <span>{stream.tokensPerSecond.toFixed(1)} tok/s</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <DollarSign className="h-3 w-3" />
                    <span>${stream.costPerMinute.toFixed(4)}/min</span>
                  </div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {metrics.topModels.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Top Models Usage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-hidden rounded-lg border">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="p-3 text-left font-medium">Model</th>
                    <th className="p-3 text-right font-medium">Sessions</th>
                    <th className="p-3 text-right font-medium">Tokens</th>
                    <th className="p-3 text-right font-medium">Cost</th>
                  </tr>
                </thead>
                <tbody>
                  {metrics.topModels.map((model) => (
                    <tr key={model.model} className="border-b last:border-0">
                      <td className="p-3 font-mono">{model.model}</td>
                      <td className="p-3 text-right">{model.sessions}</td>
                      <td className="p-3 text-right">{model.totalTokens.toLocaleString()}</td>
                      <td className="p-3 text-right">${model.totalCost.toFixed(4)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
