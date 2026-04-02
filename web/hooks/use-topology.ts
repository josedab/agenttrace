"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TopologyAgent {
  id: string;
  name: string;
  role: string;
  status: "active" | "idle" | "error" | "completed";
  messagesSent: number;
  messagesReceived: number;
  avgLatencyMs: number;
  errorCount: number;
}

export interface TopologyEdge {
  source: string;
  target: string;
  messageCount: number;
  avgLatencyMs: number;
  errorRate: number;
  weight: number;
}

export interface TopologyBottleneck {
  agentId: string;
  agentName: string;
  bottleneckType: string;
  severity: "critical" | "high" | "medium" | "low";
  avgLatencyMs: number;
  suggestion: string;
}

export interface TopologyAnalytics {
  sessionId: string;
  agentCount: number;
  agents: TopologyAgent[];
  edges: TopologyEdge[];
  clusters: { id: string; agents: string[] }[];
  metrics: Record<string, number>;
  topologyType: string;
  totalMessages: number;
  totalHandoffs: number;
  avgResponseTimeMs: number;
  criticalPath: string[];
  bottlenecks: TopologyBottleneck[];
  healthScore: number;
}

export interface DelegationStep {
  agentId: string;
  action: string;
  timestamp: string;
  duration: number;
  status: string;
  fromAgent?: string;
  toAgent?: string;
}

export interface DelegationChain {
  id: string;
  sessionId: string;
  initiatorId: string;
  steps: DelegationStep[];
  totalDuration: number;
  status: "completed" | "failed" | "in_progress";
  createdAt: string;
}

export interface MessageFlow {
  sourceAgent: string;
  targetAgent: string;
  messageType: string;
  timestamp: string;
  latencyMs: number;
  payload?: Record<string, unknown>;
}

export function useTopologyAnalytics(sessionId: string) {
  return useQuery({
    queryKey: ["topology-analytics", sessionId],
    queryFn: () =>
      api.multiAgent.getTopologyAnalytics(sessionId) as Promise<TopologyAnalytics>,
    enabled: !!sessionId,
    refetchInterval: 5000,
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

export function useTopologyGraph(sessionId: string) {
  return useQuery({
    queryKey: ["topology-graph", sessionId],
    queryFn: () =>
      api.multiAgent.getTopologyGraph(sessionId) as Promise<{
        agents: TopologyAgent[];
        edges: TopologyEdge[];
        topologyType: string;
      }>,
    enabled: !!sessionId,
    refetchInterval: 5000,
  });
}

export function useMultiAgentSessions() {
  return useQuery({
    queryKey: ["multi-agent-sessions"],
    queryFn: () => api.multiAgent.listSessions() as Promise<{
      sessions: { id: string; name: string; agentCount: number; status: string; createdAt: string }[];
    }>,
  });
}

export function useAnalyzeMultiAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { sessionId: string }) =>
      api.multiAgent.analyze(data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["topology-analytics", variables.sessionId] });
    },
  });
}
