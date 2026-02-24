"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CostForecast {
  projectId: string;
  currentDailyCost: number;
  projectedDailyCost: number;
  projectedMonthlyCost: number;
  projectedYearlyCost: number;
  confidenceInterval: [number, number];
  dailyProjections: { date: string; projected: number; low: number; high: number }[];
  optimizationPotential: number;
  budgetStatus: string;
}

export interface CostRecommendation {
  id: string;
  currentModel: string;
  recommendedModel: string;
  traceCount: number;
  estimatedSavingsPerMonth: number;
  qualityImpactEstimate: number;
  confidence: number;
  status: string;
}

export interface CostReport {
  totalCost: number;
  totalTokens: number;
  traceCount: number;
  costByModel: { model: string; totalCost: number; traceCount: number }[];
  costByDay: { date: string; cost: number; tokens: number }[];
  forecast: CostForecast;
  roi: { totalSavings: number; savingsPercent: number };
}

export function useCostForecast() {
  return useQuery({
    queryKey: ["cost-forecast"],
    queryFn: () =>
      api.costOptimizer.getForecast() as Promise<CostForecast>,
    refetchInterval: 60000,
  });
}

export function useCostRecommendations() {
  return useQuery({
    queryKey: ["cost-recommendations"],
    queryFn: () =>
      api.costOptimizer.getRecommendations() as Promise<{ recommendations: CostRecommendation[] }>,
  });
}

export function useCostReport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (period?: { startDate: string; endDate: string }) =>
      api.costOptimizer.generateReport(period) as Promise<CostReport>,
  });
}

export function useConfigureAutopilot() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (config: {
      enabled?: boolean;
      maxBudgetDaily?: number;
      maxBudgetMonthly?: number;
      optimizationLevel?: string;
    }) => api.costOptimizer.configureAutopilot(config),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["cost-forecast"] }),
  });
}
