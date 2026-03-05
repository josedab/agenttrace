"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useCostForecast(period: string = "monthly", days: number = 30) {
  return useQuery({
    queryKey: ["cost-forecast", period, days],
    queryFn: () =>
      api.get(`/api/public/cost-forecast?period=${period}&days=${days}`),
  });
}

export function useCostSimulation() {
  return useMutation({
    mutationFn: (data: { modelChanges?: Record<string, string>; volumeMultiplier?: number; period?: string }) =>
      api.post("/api/public/cost-forecast/simulate", data),
  });
}

export function useCostHistory(period: string = "daily") {
  return useQuery({
    queryKey: ["cost-history", period],
    queryFn: () => api.get(`/api/public/cost-forecast/history?period=${period}`),
  });
}

export function useCreateBudgetPlan() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; budgetLimit: number; period: string; alertThreshold?: number }) =>
      api.post("/api/public/cost-forecast/budget-plans", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cost-forecast"] });
    },
  });
}
