"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CostAlert {
  id: string;
  ruleId: string;
  ruleName: string;
  severity: "info" | "warning" | "critical";
  message: string;
  currentValue: number;
  threshold: number;
  acknowledged: boolean;
  triggeredAt: string;
  acknowledgedAt?: string;
}

export interface CostAlertRule {
  id: string;
  name: string;
  condition: CostAlertCondition;
  severity: "info" | "warning" | "critical";
  enabled: boolean;
  channels: string[];
  createdAt: string;
}

export interface CircuitBreakerConfig {
  enabled: boolean;
  maxCostPerMinute: number;
  maxCostPerHour: number;
  maxCostPerDay: number;
  action: "block" | "throttle" | "alert";
  cooldownMinutes: number;
  lastTriggeredAt?: string;
}

export interface CostAlertCondition {
  metric: "total_cost" | "cost_per_trace" | "cost_rate" | "model_cost";
  operator: "gt" | "gte" | "lt" | "lte";
  threshold: number;
  window: string;
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<CostAlertRule, "id" | "createdAt">) =>
      api.costAlerting.createRule(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["cost-alert-rules"] }),
  });
}

export function useAlertRules() {
  return useQuery({
    queryKey: ["cost-alert-rules"],
    queryFn: () =>
      api.costAlerting.listRules() as Promise<CostAlertRule[]>,
  });
}

export function useCostAlerts() {
  return useQuery({
    queryKey: ["cost-alerts"],
    queryFn: () =>
      api.costAlerting.listAlerts() as Promise<CostAlert[]>,
    refetchInterval: 30000,
  });
}

export function useAcknowledgeCostAlert() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.costAlerting.acknowledgeAlert(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["cost-alerts"] }),
  });
}

export function useCircuitBreaker() {
  return useQuery({
    queryKey: ["circuit-breaker"],
    queryFn: () =>
      api.costAlerting.getCircuitBreaker() as Promise<CircuitBreakerConfig>,
  });
}

export function useUpdateCircuitBreaker() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<CircuitBreakerConfig>) =>
      api.costAlerting.updateCircuitBreaker(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["circuit-breaker"] }),
  });
}

export function useCheckCost() {
  return useMutation({
    mutationFn: (data: { amount: number; model?: string }) =>
      api.costAlerting.checkCost(data),
  });
}
