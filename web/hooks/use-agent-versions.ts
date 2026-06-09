"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useAgentVersions(agentName?: string) {
  return useQuery({
    queryKey: ["agent-versions", agentName],
    queryFn: () => api.agentVersions.list(agentName),
  });
}

export function useAgentVersion(id: string) {
  return useQuery({
    queryKey: ["agent-versions", "detail", id],
    queryFn: () => api.agentVersions.get(id),
    enabled: !!id,
  });
}

export function useActiveAgentVersions(agentName?: string) {
  return useQuery({
    queryKey: ["agent-versions", "active", agentName],
    queryFn: () => api.agentVersions.getActive(agentName),
  });
}

export function useCreateVersion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { agentName: string; config: Record<string, unknown>; changeLog?: string }) =>
      api.agentVersions.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["agent-versions"] }),
  });
}

export function useRollback() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.agentVersions.rollback(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["agent-versions"] }),
  });
}

export function useDiffVersions() {
  return useMutation({
    mutationFn: (data: { versionIdA: string; versionIdB: string }) =>
      api.agentVersions.diff(data),
  });
}
