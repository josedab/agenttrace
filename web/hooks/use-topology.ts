"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TopologyAnalytics {
  sessionId: string;
  agentCount: number;
  edges: { source: string; target: string; weight: number }[];
  clusters: { id: string; agents: string[] }[];
  metrics: Record<string, number>;
}

export interface DelegationChain {
  id: string;
  steps: { agentId: string; action: string; timestamp: string; duration: number }[];
  totalDuration: number;
  status: "completed" | "failed" | "in_progress";
}

export function useTopologyAnalytics(sessionId: string) {
  return useQuery({
    queryKey: ["topology-analytics", sessionId],
    queryFn: () =>
      api.multiAgent.getTopologyAnalytics(sessionId) as Promise<TopologyAnalytics>,
    enabled: !!sessionId,
  });
}

export function useDelegationChains(sessionId: string) {
  return useQuery({
    queryKey: ["delegation-chains", sessionId],
    queryFn: () =>
      api.multiAgent.getDelegationChains(sessionId) as Promise<DelegationChain[]>,
    enabled: !!sessionId,
  });
}
