"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface PromptBaseline {
  id: string;
  name: string;
  promptId: string;
  version: string;
  scores: Record<string, number>;
  sampleCount: number;
  createdAt: string;
}

export interface PromptCIRun {
  id: string;
  baselineId: string;
  baselineName: string;
  status: "pending" | "running" | "passed" | "failed" | "error";
  comparisons: ScoreComparison[];
  triggeredBy: string;
  ciProvider?: string;
  startedAt: string;
  completedAt?: string;
}

export interface ScoreComparison {
  metric: string;
  baselineValue: number;
  currentValue: number;
  delta: number;
  deltaPercent: number;
  passed: boolean;
  threshold: number;
}

export function usePromptBaselines() {
  return useQuery({
    queryKey: ["prompt-baselines"],
    queryFn: () =>
      api.promptCI.listBaselines() as Promise<PromptBaseline[]>,
  });
}

export function useCreateBaseline() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; promptId: string; version: string }) =>
      api.promptCI.createBaseline(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-baselines"] }),
  });
}

export function useRunComparison() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { baselineId: string; promptVersion: string }) =>
      api.promptCI.runComparison(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-runs"] }),
  });
}

export function usePromptCIRuns() {
  return useQuery({
    queryKey: ["prompt-ci-runs"],
    queryFn: () =>
      api.promptCI.listRuns() as Promise<PromptCIRun[]>,
  });
}

export interface GateConfig {
  id: string;
  name: string;
  metrics: string[];
  thresholds: Record<string, number>;
  enforceMode: "block" | "warn" | "log";
  createdAt: string;
}

export interface RegressionEntry {
  id: string;
  runId: string;
  metric: string;
  previousValue: number;
  currentValue: number;
  detectedAt: string;
}

export interface PromptCIDashboardStats {
  totalBaselines: number;
  totalRuns: number;
  passRate: number;
  recentRegressions: number;
}

export function usePromptCIGateConfigs() {
  return useQuery({
    queryKey: ["prompt-ci-gate-configs"],
    queryFn: () => api.promptCI.listGateConfigs() as Promise<GateConfig[]>,
  });
}

export function useCreateGateConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; metrics: string[]; thresholds: Record<string, number>; enforceMode: string }) =>
      api.promptCI.createGateConfig(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-gate-configs"] }),
  });
}

export function useUpdateGateConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<GateConfig> }) =>
      api.promptCI.updateGateConfig(id, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-gate-configs"] }),
  });
}

export function useRegressionHistory() {
  return useQuery({
    queryKey: ["prompt-ci-regressions"],
    queryFn: () => api.promptCI.getRegressionHistory() as Promise<RegressionEntry[]>,
  });
}

export function usePromptCIDashboardStats() {
  return useQuery({
    queryKey: ["prompt-ci-dashboard"],
    queryFn: () => api.promptCI.getDashboardStats() as Promise<PromptCIDashboardStats>,
  });
}

export interface CIWebhookPayload {
  provider: "github" | "gitlab" | "custom";
  branch: string;
  commitSha: string;
  prNumber?: number;
  gateConfigId?: string;
  scores?: Record<string, number>;
}

export interface CIWebhookResult {
  passed: boolean;
  exitCode: number;
  overallSeverity: string;
  summary: string;
  blockReason?: string;
  metricResults: {
    metricName: string;
    baselineValue: number;
    currentValue: number;
    thresholdPercent: number;
    actualChangePercent: number;
    passed: boolean;
    severity: string;
  }[];
}

export function useTriggerCIWebhook() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CIWebhookPayload) =>
      api.promptCI.triggerCIWebhook(data) as Promise<CIWebhookResult>,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-runs"] });
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-regressions"] });
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-dashboard"] });
    },
  });
}

export function useGenerateCIConfig(provider: string) {
  return useQuery({
    queryKey: ["prompt-ci-config", provider],
    queryFn: () => api.promptCI.generateCIConfig(provider),
    enabled: !!provider,
  });
}
