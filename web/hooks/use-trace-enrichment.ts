"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useEnrichmentRules() {
  return useQuery({
    queryKey: ["enrichment-rules"],
    queryFn: () => api.get("/api/public/trace-enrichment/rules"),
  });
}

export function useCreateEnrichmentRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; condition: string; action: string; config?: Record<string, unknown> }) =>
      api.post("/api/public/trace-enrichment/rules", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enrichment-rules"] });
    },
  });
}

export function useUpdateEnrichmentRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ ruleId, data }: { ruleId: string; data: { name?: string; condition?: string; action?: string; config?: Record<string, unknown> } }) =>
      api.put(`/api/public/trace-enrichment/rules/${ruleId}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enrichment-rules"] });
    },
  });
}

export function useDeleteEnrichmentRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (ruleId: string) =>
      api.delete(`/api/public/trace-enrichment/rules/${ruleId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enrichment-rules"] });
    },
  });
}

export function useEnrichmentSources() {
  return useQuery({
    queryKey: ["enrichment-sources"],
    queryFn: () => api.get("/api/public/trace-enrichment/sources"),
  });
}

export function useTestEnrichmentRule() {
  return useMutation({
    mutationFn: (data: { condition: string; action: string; sampleTraceId?: string }) =>
      api.post("/api/public/trace-enrichment/rules/test", data),
  });
}
