"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useHandoffChain(traceId: string) {
  return useQuery({
    queryKey: ["handoffs", "chain", traceId],
    queryFn: () => api.handoffs.getChain(traceId),
    enabled: !!traceId,
  });
}

export function useHandoffStats() {
  return useQuery({
    queryKey: ["handoffs", "stats"],
    queryFn: () => api.handoffs.getStats(),
  });
}

export function useInitiateHandoff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => api.handoffs.initiate(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["handoffs"] }),
  });
}
