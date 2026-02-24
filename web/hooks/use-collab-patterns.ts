"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CollabPattern {
  id: string;
  name: string;
  type: string;
  description: string;
  config: Record<string, unknown>;
  createdAt: string;
}

export interface PatternDeployment {
  id: string;
  patternId: string;
  status: string;
  startedAt: string;
  completedAt?: string;
}

export interface PatternAnalytics {
  patternId: string;
  deployments: number;
  successRate: number;
  avgDuration: number;
  costSavings: number;
}

export function useCollabPatterns() {
  return useQuery({
    queryKey: ["collab-patterns"],
    queryFn: () => api.collabPatterns.list() as Promise<CollabPattern[]>,
    refetchInterval: 30000,
  });
}

export function useDeployPattern() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { config?: Record<string, unknown> } }) =>
      api.collabPatterns.deploy(id, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["collab-patterns"] }),
  });
}

export function usePatternAnalytics(patternId: string | null) {
  return useQuery({
    queryKey: ["pattern-analytics", patternId],
    queryFn: () => api.collabPatterns.getAnalytics(patternId!) as Promise<PatternAnalytics>,
    enabled: !!patternId,
  });
}
