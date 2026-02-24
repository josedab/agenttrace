"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface FederatedDashboard {
  totalInstances: number;
  activeInstances: number;
  totalMetrics: number;
  insights: FederatedInsight[];
  benchmarks: FederatedBenchmark[];
}

export interface FederatedInstance {
  id: string;
  name: string;
  url: string;
  status: "active" | "inactive" | "syncing" | "error";
  lastSync: string;
  metricsCount: number;
  version: string;
}

export interface FederatedBenchmark {
  id: string;
  metricType: string;
  instanceCount: number;
  aggregatedValue: number;
  percentile: number;
  rank: number;
  lastUpdated: string;
}

export interface FederatedInsight {
  id: string;
  type: string;
  title: string;
  description: string;
  severity: "info" | "warning" | "critical";
  affectedInstances: string[];
  detectedAt: string;
}

export function useFederatedDashboard() {
  return useQuery({
    queryKey: ["federated-dashboard"],
    queryFn: () =>
      api.federatedAggregation.getDashboard() as Promise<FederatedDashboard>,
    refetchInterval: 30000,
  });
}

export function useRegisterInstance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; url: string }) =>
      api.federatedAggregation.registerInstance(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["federated-dashboard"] }),
  });
}

export function useFederatedBenchmarks() {
  return useQuery({
    queryKey: ["federated-benchmarks"],
    queryFn: () =>
      api.federatedAggregation.getBenchmarks() as Promise<FederatedBenchmark[]>,
  });
}

export function useFederatedInsights() {
  return useQuery({
    queryKey: ["federated-insights"],
    queryFn: () =>
      api.federatedAggregation.getInsights() as Promise<FederatedInsight[]>,
  });
}
