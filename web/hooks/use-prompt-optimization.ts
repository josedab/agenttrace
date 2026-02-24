"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface PromptOptimization {
  id: string;
  promptId: string;
  status: "pending" | "running" | "completed" | "failed";
  variants: PromptVariant[];
  failurePatterns: FailurePattern[];
  improvementScore: number;
  startedAt: string;
  completedAt?: string;
}

export interface PromptVariant {
  id: string;
  optimizationId: string;
  content: string;
  score: number;
  metrics: Record<string, number>;
  status: "pending" | "approved" | "rejected";
  createdAt: string;
}

export interface FailurePattern {
  id: string;
  pattern: string;
  frequency: number;
  severity: "low" | "medium" | "high";
  examples: string[];
  suggestedFix: string;
}

export interface OptimizationConfig {
  maxVariants: number;
  evaluationMetrics: string[];
  autoApproveThreshold: number;
  targetModel: string;
  enabled: boolean;
}

export function useStartOptimization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { promptId: string; config?: Partial<OptimizationConfig> }) =>
      api.promptOptimization.start(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-optimizations"] }),
  });
}

export function useOptimization(id: string | null) {
  return useQuery({
    queryKey: ["prompt-optimization", id],
    queryFn: () =>
      api.promptOptimization.get(id!) as Promise<PromptOptimization>,
    enabled: !!id,
  });
}

export function useOptimizations() {
  return useQuery({
    queryKey: ["prompt-optimizations"],
    queryFn: () =>
      api.promptOptimization.list() as Promise<PromptOptimization[]>,
  });
}

export function useOptConfig() {
  return useQuery({
    queryKey: ["prompt-optimization-config"],
    queryFn: () =>
      api.promptOptimization.getConfig() as Promise<OptimizationConfig>,
  });
}

export function useUpdateOptConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<OptimizationConfig>) =>
      api.promptOptimization.updateConfig(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-optimization-config"] }),
  });
}

export function useApproveVariant() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.promptOptimization.approveVariant(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-optimizations"] }),
  });
}

export function useRejectVariant() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.promptOptimization.rejectVariant(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["prompt-optimizations"] }),
  });
}
