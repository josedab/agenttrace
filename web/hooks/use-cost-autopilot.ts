"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CostHotspot {
  id: string;
  resource: string;
  cost: number;
  trend: "rising" | "stable" | "declining";
  percentage: number;
}

export interface AutopilotRule {
  id: string;
  name: string;
  condition: string;
  action: string;
  enabled: boolean;
  createdAt: string;
}

export interface CostPrediction {
  date: string;
  predictedCost: number;
  confidence: number;
  budgetRemaining: number;
}

export interface CostAutopilotDashboard {
  totalSavings: number;
  activeRules: number;
  hotspotCount: number;
  predictions: CostPrediction[];
}

export function useCostHotspots(days: number = 30) {
  return useQuery({
    queryKey: ["cost-hotspots", days],
    queryFn: () => api.costAutopilot.getHotspots(days) as Promise<CostHotspot[]>,
  });
}

export function useCostAutopilotRules() {
  return useQuery({
    queryKey: ["cost-autopilot-rules"],
    queryFn: () => api.costAutopilot.getRules() as Promise<AutopilotRule[]>,
  });
}

export function useCreateAutopilotRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; condition: string; action: string }) =>
      api.costAutopilot.createRule(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["cost-autopilot-rules"] }),
  });
}

export function useCostPredictions(days: number = 30, budget: number = 0) {
  return useQuery({
    queryKey: ["cost-predictions", days, budget],
    queryFn: () => api.costAutopilot.getPredictions(days, budget) as Promise<CostPrediction[]>,
  });
}

export function useCostAutopilotDashboard() {
  return useQuery({
    queryKey: ["cost-autopilot-dashboard"],
    queryFn: () => api.costAutopilot.getDashboard() as Promise<CostAutopilotDashboard>,
  });
}
