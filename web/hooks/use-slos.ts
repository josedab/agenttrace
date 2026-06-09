"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CreateSLOInput {
  name: string;
  metric: string;
  target: number;
  window: string;
  description?: string;
}

export function useSLOs() {
  return useQuery({
    queryKey: ["slos"],
    queryFn: () => api.slos.list(),
  });
}

export function useCreateSLO() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSLOInput) => api.slos.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["slos"] }),
  });
}

export function useSLOReport() {
  return useQuery({
    queryKey: ["slos", "report"],
    queryFn: () => api.slos.getReport(),
  });
}

export function useSLOHistory(id: string) {
  return useQuery({
    queryKey: ["slos", id, "history"],
    queryFn: () => api.slos.getHistory(id),
    enabled: !!id,
  });
}
