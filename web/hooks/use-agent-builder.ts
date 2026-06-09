"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useBlueprints() {
  return useQuery({
    queryKey: ["agent-builder-blueprints"],
    queryFn: () => api.agentBuilder.list(),
  });
}

export function useBlueprint(id: string) {
  return useQuery({
    queryKey: ["agent-builder-blueprints", id],
    queryFn: () => api.agentBuilder.get(id),
    enabled: !!id,
  });
}

export function useGenerateBlueprint() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { taskDescription: string; constraints?: Record<string, unknown> }) =>
      api.agentBuilder.generate(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["agent-builder-blueprints"] }),
  });
}

export function useDeployBlueprint() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.agentBuilder.deploy(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["agent-builder-blueprints"] }),
  });
}
