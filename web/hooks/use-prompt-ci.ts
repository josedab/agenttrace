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
