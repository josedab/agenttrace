"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useCodeImpact(traceId: string | null) {
  return useQuery({
    queryKey: ["code-impact", traceId],
    queryFn: () => api.get(`/api/public/traces/${traceId}/code-impact`),
    enabled: !!traceId,
  });
}

export function useCodeImpactSummary() {
  return useQuery({
    queryKey: ["code-impact-summary"],
    queryFn: () => api.get("/api/public/code-impact/summary"),
  });
}

export function useCodeImpactFileTree(traceId: string | null) {
  return useQuery({
    queryKey: ["code-impact-file-tree", traceId],
    queryFn: () =>
      api.get(`/api/public/code-impact/file-tree?traceId=${traceId}`),
    enabled: !!traceId,
  });
}
