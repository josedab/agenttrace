"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useOrchestrationSessions() {
  return useQuery({
    queryKey: ["orchestration-sessions"],
    queryFn: () => api.orchestration.listSessions(),
  });
}

export function useOrchestrationSession(id: string) {
  return useQuery({
    queryKey: ["orchestration-sessions", id],
    queryFn: () => api.orchestration.getSession(id),
    enabled: !!id,
  });
}

export function useCreateSession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; agentIds: string[]; config?: Record<string, unknown> }) =>
      api.orchestration.createSession(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["orchestration-sessions"] }),
  });
}

export function useExecuteCommand(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (cmd: { type: string; payload?: Record<string, unknown> }) =>
      api.orchestration.executeCommand(sessionId, cmd),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["orchestration-sessions", sessionId] }),
  });
}

export function useAddBreakpoint(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (bp: { agentId: string; condition?: string }) =>
      api.orchestration.addBreakpoint(sessionId, bp),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["orchestration-sessions", sessionId] }),
  });
}
