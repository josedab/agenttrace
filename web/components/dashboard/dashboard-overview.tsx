"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { MetricsCard } from "@/components/dashboard/metrics-card";
import { TraceVolumeChart } from "@/components/dashboard/trace-volume-chart";
import { CostBreakdownChart } from "@/components/dashboard/cost-breakdown-chart";
import { LatencyPercentileChart } from "@/components/dashboard/latency-percentile-chart";
import { RecentTraces } from "@/components/dashboard/recent-traces";
import { Activity, DollarSign, Clock, Layers, TrendingUp, TrendingDown } from "lucide-react";

export function DashboardOverview() {
  const { data: metrics, isLoading } = useQuery({
    queryKey: ["dashboard-metrics"],
    queryFn: () => api.metrics.getDashboard(),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  if (isLoading || !metrics) {
    return <DashboardOverviewSkeleton />;
  }

  return (
    <div className="space-y-6">
      {/* Metrics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricsCard
          title="Total Traces"
          value={formatNumber(metrics.totalTraces)}
          change={metrics.tracesChange}
          icon={Layers}
          trend={metrics.tracesChange >= 0 ? "up" : "down"}
        />
        <MetricsCard
          title="Total Cost"
          value={formatCurrency(metrics.totalCost)}
          change={metrics.costChange}
          icon={DollarSign}
          trend={metrics.costChange <= 0 ? "up" : "down"} // Lower cost is better
        />
        <MetricsCard
          title="Avg Latency"
          value={formatLatency(metrics.avgLatency)}
          change={metrics.latencyChange}
          icon={Clock}
          trend={metrics.latencyChange <= 0 ? "up" : "down"} // Lower latency is better
        />
        <MetricsCard
          title="Active Sessions"
          value={formatNumber(metrics.activeSessions)}
          change={metrics.sessionsChange}
          icon={Activity}
          trend={metrics.sessionsChange >= 0 ? "up" : "down"}
        />
      </div>

      {/* Charts */}
      <div className="grid gap-6 lg:grid-cols-2">
        <TraceVolumeChart data={metrics.traceVolume || []} />
        <CostBreakdownChart data={metrics.costBreakdown || []} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <LatencyPercentileChart data={metrics.latencyPercentiles || []} />
        <RecentTraces traces={metrics.recentTraces || []} />
      </div>
    </div>
  );
}

function DashboardOverviewSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <div
            key={i}
            className="h-32 rounded-lg border bg-card animate-pulse"
          />
        ))}
      </div>
      <div className="grid gap-6 lg:grid-cols-2">
        {[...Array(2)].map((_, i) => (
          <div
            key={i}
            className="h-80 rounded-lg border bg-card animate-pulse"
          />
        ))}
      </div>
    </div>
  );
}

function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + "M";
  }
  if (value >= 1000) {
    return (value / 1000).toFixed(1) + "K";
  }
  return value.toString();
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
  }).format(value);
}

function formatLatency(ms: number): string {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + "s";
  }
  return ms.toFixed(0) + "ms";
}
