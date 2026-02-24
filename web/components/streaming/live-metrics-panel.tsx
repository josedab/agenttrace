"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { LiveMetrics } from "@/hooks/use-streaming";
import {
  Activity,
  Clock,
  DollarSign,
  Zap,
  FileText,
  Terminal,
  AlertTriangle,
} from "lucide-react";

interface LiveMetricsPanelProps {
  metrics: LiveMetrics | null;
  connected: boolean;
}

export function LiveMetricsPanel({
  metrics,
  connected,
}: LiveMetricsPanelProps) {
  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  };

  const formatCost = (cost: number) => `$${cost.toFixed(4)}`;

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Badge variant={connected ? "default" : "destructive"}>
          <span
            className={`mr-1.5 h-2 w-2 rounded-full inline-block ${connected ? "bg-green-400 animate-pulse" : "bg-red-400"}`}
          />
          {connected ? "Live" : "Disconnected"}
        </Badge>
        {metrics && (
          <span className="text-sm text-muted-foreground">
            Updated {new Date(metrics.lastUpdated).toLocaleTimeString()}
          </span>
        )}
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <Card>
          <CardHeader className="pb-2 pt-4 px-4">
            <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <Clock className="h-3 w-3" /> Elapsed
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            <div className="text-2xl font-bold">
              {metrics ? formatDuration(metrics.elapsedMs) : "--"}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2 pt-4 px-4">
            <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <DollarSign className="h-3 w-3" /> Total Cost
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            <div className="text-2xl font-bold">
              {metrics ? formatCost(metrics.totalCost) : "--"}
            </div>
            {metrics && metrics.costPerMinute > 0 && (
              <p className="text-xs text-muted-foreground">
                {formatCost(metrics.costPerMinute)}/min
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2 pt-4 px-4">
            <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <Zap className="h-3 w-3" /> Tokens
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            <div className="text-2xl font-bold">
              {metrics ? metrics.totalTokens.toLocaleString() : "--"}
            </div>
            {metrics && metrics.tokensPerSecond > 0 && (
              <p className="text-xs text-muted-foreground">
                {metrics.tokensPerSecond.toFixed(0)} tok/s
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2 pt-4 px-4">
            <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <Activity className="h-3 w-3" /> Spans
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            <div className="text-2xl font-bold">
              {metrics
                ? `${metrics.activeSpans}/${metrics.completedSpans}`
                : "--"}
            </div>
            <p className="text-xs text-muted-foreground">active/completed</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <Card>
          <CardContent className="flex items-center gap-2 py-3 px-4">
            <FileText className="h-4 w-4 text-blue-500" />
            <div>
              <div className="text-lg font-semibold">
                {metrics?.filesModified ?? 0}
              </div>
              <div className="text-xs text-muted-foreground">
                Files modified
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-2 py-3 px-4">
            <Terminal className="h-4 w-4 text-green-500" />
            <div>
              <div className="text-lg font-semibold">
                {metrics?.terminalCommands ?? 0}
              </div>
              <div className="text-xs text-muted-foreground">Commands</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="flex items-center gap-2 py-3 px-4">
            <AlertTriangle className="h-4 w-4 text-red-500" />
            <div>
              <div className="text-lg font-semibold">
                {metrics?.errorCount ?? 0}
              </div>
              <div className="text-xs text-muted-foreground">Errors</div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
