"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function usePredictions() {
  return useQuery({
    queryKey: ["predictions"],
    queryFn: () => api.predictions.list(),
  });
}

export function usePredictCost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { taskDescription: string; model?: string; constraints?: Record<string, unknown> }) =>
      api.predictions.predict(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["predictions"] }),
  });
}

export function useRequestApproval() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.predictions.requestApproval(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["predictions"] }),
  });
}

export function useDecideApproval() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: { approved: boolean; reason?: string } }) =>
      api.predictions.decideApproval(id, decision),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["predictions"] }),
  });
}
