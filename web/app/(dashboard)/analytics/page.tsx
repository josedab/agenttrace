"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { DollarSign, Clock, Activity, TrendingUp, TrendingDown } from "lucide-react";

import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TraceVolumeChart } from "@/components/dashboard/trace-volume-chart";
import { CostBreakdownChart } from "@/components/dashboard/cost-breakdown-chart";
import { LatencyPercentileChart } from "@/components/dashboard/latency-percentile-chart";

export default function AnalyticsPage() {
  const { data: metrics, isLoading } = useQuery({
    queryKey: ["analytics-overview"],
    queryFn: () => api.analytics.getOverview(),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Analytics"
        description="Monitor your LLM usage, costs, and performance."
      />

      {/* Overview metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <MetricCard
          title="Total Traces"
          value={metrics?.totalTraces}
          change={metrics?.tracesChange}
          icon={Activity}
          isLoading={isLoading}
        />
        <MetricCard
          title="Total Cost"
          value={metrics?.totalCost}
          change={metrics?.costChange}
          icon={DollarSign}
          prefix="$"
          isLoading={isLoading}
        />
        <MetricCard
          title="Avg Latency"
          value={metrics?.avgLatency}
          change={metrics?.latencyChange}
          icon={Clock}
          suffix="ms"
          isLoading={isLoading}
          invertChange
        />
        <MetricCard
          title="Tokens Used"
          value={metrics?.totalTokens}
          change={metrics?.tokensChange}
          icon={TrendingUp}
          isLoading={isLoading}
        />
      </div>

      {/* Quick links */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Link href="/analytics/cost">
          <Card className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <DollarSign className="h-5 w-5" />
                Cost Analytics
              </CardTitle>
              <CardDescription>
                Detailed cost breakdown by model, project, and time period.
              </CardDescription>
            </CardHeader>
          </Card>
        </Link>
        <Link href="/analytics/latency">
          <Card className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5" />
                Latency Analytics
              </CardTitle>
              <CardDescription>
                P50, P95, P99 latency percentiles and trends.
              </CardDescription>
            </CardHeader>
          </Card>
        </Link>
        <Link href="/analytics/usage">
          <Card className="hover:border-primary/50 transition-colors cursor-pointer">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Usage Analytics
              </CardTitle>
              <CardDescription>
                Token usage, trace volume, and model distribution.
              </CardDescription>
            </CardHeader>
          </Card>
        </Link>
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
              <TraceVolumeChart data={metrics?.traceVolume || []} />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Cost by Model</CardTitle>
            <CardDescription>Cost distribution across models</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <CostBreakdownChart data={metrics?.costByModel || []} />
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Latency Percentiles</CardTitle>
          <CardDescription>P50, P95, P99 latency over time</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-64 w-full" />
          ) : (
            <LatencyPercentileChart data={metrics?.latencyPercentiles || []} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

interface MetricCardProps {
  title: string;
  value?: number;
  change?: number;
  icon: React.ComponentType<{ className?: string }>;
  prefix?: string;
  suffix?: string;
  isLoading?: boolean;
  invertChange?: boolean;
}

function MetricCard({
  title,
  value,
  change,
  icon: Icon,
  prefix,
  suffix,
  isLoading,
  invertChange,
}: MetricCardProps) {
  const isPositive = invertChange ? (change ?? 0) < 0 : (change ?? 0) > 0;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <>
            <div className="text-2xl font-bold">
              {prefix}
              {value?.toLocaleString() ?? 0}
              {suffix}
            </div>
            {change !== undefined && (
              <p
                className={cn(
                  "text-xs flex items-center gap-1",
                  isPositive ? "text-green-500" : "text-red-500"
                )}
              >
                {isPositive ? (
                  <TrendingUp className="h-3 w-3" />
                ) : (
                  <TrendingDown className="h-3 w-3" />
                )}
                {Math.abs(change).toFixed(1)}% from last period
              </p>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
