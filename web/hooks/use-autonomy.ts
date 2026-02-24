"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface AutonomyConfig {
  agentId: string;
  level: string;
  permissions: string[];
  constraints: Record<string, unknown>;
  updatedAt: string;
}

export interface AutonomyDashboard {
  configs: AutonomyConfig[];
  stats: { totalAgents: number; byLevel: Record<string, number> };
}

export interface TrustEvent {
  agentId: string;
  trustScore: number;
  reason: string;
  timestamp: string;
}

export function useAutonomyDashboard() {
  return useQuery({
    queryKey: ["autonomy-dashboard"],
    queryFn: () => api.autonomy.getDashboard() as Promise<AutonomyDashboard>,
    refetchInterval: 30000,
  });
}

export function useSetAutonomy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { agentId: string; level: string; permissions?: string[] }) =>
      api.autonomy.set(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["autonomy-dashboard"] }),
  });
}

export function useTrustEvolution(agentId: string | null) {
  return useQuery({
    queryKey: ["trust-evolution", agentId],
    queryFn: () => api.autonomy.getTrust(agentId!) as Promise<TrustEvent[]>,
    enabled: !!agentId,
  });
}
