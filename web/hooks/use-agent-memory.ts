"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useMemoryAnalysis(traceId: string) {
  return useQuery({
    queryKey: ["memory", "analysis", traceId],
    queryFn: () => api.memory.analyze({ traceId }),
    enabled: !!traceId,
  });
}

export function useMemoryOptimizations() {
  return useQuery({
    queryKey: ["memory", "optimizations"],
    queryFn: () => api.memory.getOptimizations(),
  });
}
