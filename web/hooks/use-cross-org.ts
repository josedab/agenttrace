"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CrossOrgReport {
  benchmarks: CrossOrgBenchmark[];
  generatedAt: string;
}

export interface CrossOrgBenchmark {
  metric: string;
  yourValue: number;
  industryMedian: number;
  industryP90: number;
  percentile: number;
}

export interface IndustryStats {
  category: string;
  metrics: Record<string, { median: number; p90: number; count: number }>;
}

export function useCrossOrgReport() {
  return useQuery({
    queryKey: ["cross-org-report"],
    queryFn: () => api.crossOrg.getReport() as Promise<CrossOrgReport>,
    refetchInterval: 60000,
  });
}

export function useSubmitCrossOrg() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { metrics: Record<string, number>; category?: string }) =>
      api.crossOrg.submit(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["cross-org-report"] }),
  });
}

export function useIndustryStats(category: string | null) {
  return useQuery({
    queryKey: ["industry-stats", category],
    queryFn: () => api.crossOrg.getIndustry(category!) as Promise<IndustryStats>,
    enabled: !!category,
  });
}
