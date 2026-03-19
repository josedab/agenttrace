"use client";

import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface FederatedDashboard {
  instances: number;
  totalTraces: number;
  aggregatedMetrics: Record<string, number>;
  lastUpdated: string;
}

export interface FederatedQueryResult {
  results: Record<string, unknown>[];
  instancesQueried: number;
  executionTimeMs: number;
}

export function useFederatedDashboard() {
  return useQuery({
    queryKey: ["federated-analytics-dashboard"],
    queryFn: () => api.federatedAnalytics.getDashboard() as Promise<FederatedDashboard>,
  });
}

export function useFederatedQuery() {
  return useMutation({
    mutationFn: (data: { query: string; filters?: Record<string, unknown> }) =>
      api.federatedAnalytics.query(data) as Promise<FederatedQueryResult>,
  });
}
