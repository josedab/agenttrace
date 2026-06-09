"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useChaosExperiments() {
  return useQuery({
    queryKey: ["chaos", "experiments"],
    queryFn: () => api.chaos.list(),
  });
}

export function useCreateExperiment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Record<string, unknown>) => api.chaos.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["chaos"] }),
  });
}

export function useRunExperiment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.chaos.run(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["chaos"] }),
  });
}

export function useResilienceScorecard(agentName: string) {
  return useQuery({
    queryKey: ["chaos", "scorecard", agentName],
    queryFn: () => api.chaos.getScorecard(agentName),
    enabled: !!agentName,
  });
}
