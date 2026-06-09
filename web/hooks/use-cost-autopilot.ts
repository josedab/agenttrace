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

interface CostAutopilotDashboardWithRecommendations extends CostAutopilotDashboard {
  recommendations?: CostRoutingRecommendation[];
}

export interface ModelPricing {
  model: string;
  provider: string;
  inputPricePerMToken: number;
  outputPricePerMToken: number;
  contextWindow: number;
  category: string;
}

export interface CostAttribution {
  id: string;
  entityType: "agent" | "team" | "project" | "model";
  entityName: string;
  totalCost: number;
  traceCount: number;
  tokenCount: number;
  avgCostPerTrace: number;
  period: string;
}

export interface CostRoutingRecommendation {
  id: string;
  currentModel: string;
  recommendedModel: string;
  estimatedSavingsPerMonth: number;
  qualityImpact: "none" | "minimal" | "moderate";
  confidence: number;
  taskType: string;
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

export function useCostAttribution(period: string = "current_month", groupBy: string = "model") {
  return useQuery({
    queryKey: ["cost-attribution", period, groupBy],
    queryFn: () => api.costAttribution.getBreakdown({ period, groupBy }) as Promise<CostAttribution[]>,
    enabled: !!period,
  });
}

export function useCostRoutingRecommendations() {
  return useQuery({
    queryKey: ["cost-routing-recommendations"],
    queryFn: () =>
      (
        api.costAutopilot.getDashboard() as Promise<CostAutopilotDashboardWithRecommendations>
      ).then((d) => d.recommendations ?? []) as Promise<CostRoutingRecommendation[]>,
  });
}

export function useApplyRoutingRecommendation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (recommendationId: string) =>
      api.costAutopilot.createRule({ name: `Applied recommendation ${recommendationId}`, condition: "auto", action: "model_switch" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cost-routing-recommendations"] });
      queryClient.invalidateQueries({ queryKey: ["cost-autopilot-rules"] });
    },
  });
}

export function useToggleAutopilotRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.costAutopilot.createRule({ name: id, condition: "toggle", action: enabled ? "enable" : "disable" }),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["cost-autopilot-rules"] }),
  });
}
