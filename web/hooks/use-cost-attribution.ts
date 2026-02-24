"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CostAttribution {
  id: string;
  traceId: string;
  feature: string;
  team: string;
  cost: number;
  businessOutcome: string;
  roi?: number;
  createdAt: string;
}

export interface CostAttributionReport {
  totalCost: number;
  totalROI: number;
  byFeature: Record<string, { cost: number; roi: number }>;
  byTeam: Record<string, { cost: number; roi: number }>;
  generatedAt: string;
}

export function useAttributeCost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; feature: string; team: string; cost: number; businessOutcome: string }) =>
      api.costAttribution.attribute(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["cost-attribution"] }),
  });
}

export function useCostAttributionReport() {
  return useQuery({
    queryKey: ["cost-attribution-report"],
    queryFn: () => api.costAttribution.getReport() as Promise<CostAttributionReport>,
    refetchInterval: 60000,
  });
}

export function useCostAttributions() {
  return useQuery({
    queryKey: ["cost-attribution"],
    queryFn: () => api.costAttribution.list() as Promise<CostAttribution[]>,
  });
}
