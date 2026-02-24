"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface MultiAgentSession {
  id: string;
  name: string;
  agents: AgentNode[];
  messages: AgentMessage[];
  bottlenecks: Bottleneck[];
  status: "active" | "completed" | "failed";
  startedAt: string;
  completedAt?: string;
}

export interface AgentNode {
  id: string;
  name: string;
  role: string;
  status: "idle" | "active" | "waiting" | "completed" | "error";
  messagesSent: number;
  messagesReceived: number;
  avgResponseTime: number;
}

export interface AgentMessage {
  id: string;
  fromAgent: string;
  toAgent: string;
  type: string;
  content: string;
  timestamp: string;
  latency: number;
}

export interface Bottleneck {
  agentId: string;
  agentName: string;
  type: "slow_response" | "high_queue" | "error_rate" | "dependency_wait";
  severity: "low" | "medium" | "high" | "critical";
  description: string;
  detectedAt: string;
}

export function useMultiAgentSessions() {
  return useQuery({
    queryKey: ["multi-agent-sessions"],
    queryFn: () =>
      api.multiAgent.listSessions() as Promise<MultiAgentSession[]>,
  });
}

export function useAnalyzeMultiAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { sessionId: string; analysisType?: string }) =>
      api.multiAgent.analyze(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["multi-agent-sessions"] }),
  });
}

export function useMultiAgentSession(sessionId: string | null) {
  return useQuery({
    queryKey: ["multi-agent-session", sessionId],
    queryFn: () =>
      api.multiAgent.getSession(sessionId!) as Promise<MultiAgentSession>,
    enabled: !!sessionId,
    refetchInterval: 5000,
  });
}
