"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useRCAReports() {
  return useQuery({
    queryKey: ["rca-reports"],
    queryFn: () => api.rca.listReports(),
  });
}

export function useRCAReport(id: string) {
  return useQuery({
    queryKey: ["rca-reports", id],
    queryFn: () => api.rca.getReport(id),
    enabled: !!id,
  });
}

export function useAnalyzeTrace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; depth?: string }) =>
      api.rca.analyze(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["rca-reports"] }),
  });
}
