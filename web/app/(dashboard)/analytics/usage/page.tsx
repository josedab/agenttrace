"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { Activity, Zap, FileText, TrendingUp, TrendingDown, Calendar } from "lucide-react";

import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TraceVolumeChart } from "@/components/dashboard/trace-volume-chart";
import { TokenUsageChart } from "@/components/analytics/token-usage-chart";
import { ModelDistributionChart } from "@/components/analytics/model-distribution-chart";

export default function UsageAnalyticsPage() {
  const [dateRange, setDateRange] = React.useState("7d");

  const { data: usageData, isLoading } = useQuery({
    queryKey: ["usage-analytics", dateRange],
    queryFn: () => api.analytics.getUsageAnalytics({ dateRange }),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Usage Analytics"
        description="Track your LLM usage and token consumption."
      />

      {/* Filters */}
      <div className="flex items-center gap-4">
        <Select value={dateRange} onValueChange={setDateRange}>
          <SelectTrigger className="w-[150px]">
            <Calendar className="h-4 w-4 mr-2" />
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="24h">Last 24 hours</SelectItem>
            <SelectItem value="7d">Last 7 days</SelectItem>
            <SelectItem value="30d">Last 30 days</SelectItem>
            <SelectItem value="90d">Last 90 days</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Traces</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold">
                  {usageData?.totalTraces?.toLocaleString() ?? 0}
                </span>
                {usageData?.tracesChange !== undefined && (
                  <span
                    className={cn(
                      "text-xs flex items-center",
                      usageData.tracesChange >= 0
                        ? "text-green-500"
                        : "text-red-500"
                    )}
                  >
                    {usageData.tracesChange >= 0 ? (
                      <TrendingUp className="h-3 w-3 mr-0.5" />
                    ) : (
                      <TrendingDown className="h-3 w-3 mr-0.5" />
                    )}
                    {Math.abs(usageData.tracesChange).toFixed(1)}%
                  </span>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Generations</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                {usageData?.totalGenerations?.toLocaleString() ?? 0}
              </span>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Input Tokens</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                {formatTokenCount(usageData?.inputTokens ?? 0)}
              </span>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Output Tokens</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                {formatTokenCount(usageData?.outputTokens ?? 0)}
              </span>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Trace Volume</CardTitle>
            <CardDescription>Traces and generations over time</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <TraceVolumeChart data={usageData?.volumeOverTime || []} />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Token Usage</CardTitle>
            <CardDescription>Input vs output tokens</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <TokenUsageChart data={usageData?.tokenUsageOverTime || []} />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Model distribution */}
      <Card>
        <CardHeader>
          <CardTitle>Model Distribution</CardTitle>
          <CardDescription>Usage breakdown by model</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <Skeleton className="h-64 w-full" />
              <div className="space-y-3">
                {[...Array(5)].map((_, i) => (
                  <Skeleton key={i} className="h-10 w-full" />
                ))}
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ModelDistributionChart data={usageData?.modelDistribution || []} />
              <div className="space-y-3">
                {usageData?.modelDistribution?.map((model: any) => (
                  <div key={model.name} className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium">{model.name}</span>
                      <span className="text-muted-foreground">
                        {model.count?.toLocaleString()} ({model.percentage?.toFixed(1)}%)
                      </span>
                    </div>
                    <Progress value={model.percentage} className="h-2" />
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Top traces by token usage */}
      <Card>
        <CardHeader>
          <CardTitle>Top Traces by Token Usage</CardTitle>
          <CardDescription>Traces with highest token consumption</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <div className="border rounded-lg">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Trace ID</TableHead>
                    <TableHead>Model</TableHead>
                    <TableHead className="text-right">Input Tokens</TableHead>
                    <TableHead className="text-right">Output Tokens</TableHead>
                    <TableHead className="text-right">Total</TableHead>
                    <TableHead className="text-right">Cost</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {usageData?.topTraces?.map((trace: any) => (
                    <TableRow key={trace.id}>
                      <TableCell className="font-mono text-sm">
                        {trace.id.slice(0, 12)}...
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{trace.model}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        {trace.inputTokens?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {trace.outputTokens?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {(trace.inputTokens + trace.outputTokens)?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        ${trace.cost?.toFixed(4)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function formatTokenCount(count: number): string {
  if (count >= 1_000_000) {
    return `${(count / 1_000_000).toFixed(1)}M`;
  }
  if (count >= 1_000) {
    return `${(count / 1_000).toFixed(1)}K`;
  }
  return count.toString();
}
