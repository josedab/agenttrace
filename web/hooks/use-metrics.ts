"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useDashboardMetrics(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["dashboard-metrics", dateRange],
    queryFn: () => api.analytics.getDashboardMetrics({ dateRange }),
    refetchInterval: 60000, // Refetch every minute
  });
}

export function useAnalyticsOverview() {
  return useQuery({
    queryKey: ["analytics-overview"],
    queryFn: () => api.analytics.getOverview(),
    refetchInterval: 60000,
  });
}

export function useCostAnalytics(options: {
  dateRange: string;
  groupBy: string;
}) {
  return useQuery({
    queryKey: ["cost-analytics", options.dateRange, options.groupBy],
    queryFn: () => api.analytics.getCostAnalytics(options),
  });
}

export function useLatencyAnalytics(options: {
  dateRange: string;
  groupBy: string;
}) {
  return useQuery({
    queryKey: ["latency-analytics", options.dateRange, options.groupBy],
    queryFn: () => api.analytics.getLatencyAnalytics(options),
  });
}

export function useUsageAnalytics(options: { dateRange: string }) {
  return useQuery({
    queryKey: ["usage-analytics", options.dateRange],
    queryFn: () => api.analytics.getUsageAnalytics(options),
  });
}

export function useTraceVolumeOverTime(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["trace-volume", dateRange],
    queryFn: () => api.analytics.getTraceVolume({ dateRange }),
  });
}

export function useCostOverTime(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["cost-over-time", dateRange],
    queryFn: () => api.analytics.getCostOverTime({ dateRange }),
  });
}

export function useLatencyPercentiles(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["latency-percentiles", dateRange],
    queryFn: () => api.analytics.getLatencyPercentiles({ dateRange }),
  });
}

export function useModelUsage(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["model-usage", dateRange],
    queryFn: () => api.analytics.getModelUsage({ dateRange }),
  });
}

export function useTopTracesByTokens(limit: number = 10) {
  return useQuery({
    queryKey: ["top-traces-tokens", limit],
    queryFn: () => api.analytics.getTopTracesByTokens({ limit }),
  });
}

export function useTopTracesByCost(limit: number = 10) {
  return useQuery({
    queryKey: ["top-traces-cost", limit],
    queryFn: () => api.analytics.getTopTracesByCost({ limit }),
  });
}

export function useRecentErrors(limit: number = 10) {
  return useQuery({
    queryKey: ["recent-errors", limit],
    queryFn: () => api.analytics.getRecentErrors({ limit }),
  });
}

// Project-level metrics
export function useProjectMetrics(projectId: string) {
  return useQuery({
    queryKey: ["project-metrics", projectId],
    queryFn: () => api.analytics.getProjectMetrics(projectId),
    enabled: !!projectId,
  });
}
