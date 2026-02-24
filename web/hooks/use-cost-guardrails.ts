"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CostGuardrailDashboard {
  totalBudget: number;
  currentSpend: number;
  activePolicies: number;
  recentViolations: CostGuardrailViolation[];
  forecast: CostForecast;
  policies: CostGuardrailPolicy[];
}

export interface CostGuardrailPolicy {
  id: string;
  name: string;
  type: "hard_limit" | "soft_limit" | "rate_limit" | "forecast_alert";
  threshold: number;
  window: string;
  action: "block" | "alert" | "throttle";
  enabled: boolean;
  createdAt: string;
}

export interface CostGuardrailViolation {
  id: string;
  policyId: string;
  policyName: string;
  type: string;
  amount: number;
  threshold: number;
  occurredAt: string;
  resolved: boolean;
}

export interface CostForecast {
  projectedSpend: number;
  confidence: number;
  period: string;
  trend: "increasing" | "decreasing" | "stable";
  breakdown: { category: string; amount: number }[];
}

export function useCostGuardrailDashboard() {
  return useQuery({
    queryKey: ["cost-guardrail-dashboard"],
    queryFn: () =>
      api.costGuardrails.getDashboard() as Promise<CostGuardrailDashboard>,
    refetchInterval: 30000,
  });
}

export function useCreateGuardrailPolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<CostGuardrailPolicy, "id" | "createdAt">) =>
      api.costGuardrails.createPolicy(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["cost-guardrail-dashboard"] }),
  });
}

export function useCheckBudget() {
  return useMutation({
    mutationFn: (data: { amount: number; category?: string }) =>
      api.costGuardrails.checkBudget(data),
  });
}

export function useCostForecast() {
  return useQuery({
    queryKey: ["cost-forecast"],
    queryFn: () =>
      api.costGuardrails.getForecast() as Promise<CostForecast>,
  });
}

export function useGuardrailViolations() {
  return useQuery({
    queryKey: ["guardrail-violations"],
    queryFn: () =>
      api.costGuardrails.listViolations() as Promise<CostGuardrailViolation[]>,
  });
}
